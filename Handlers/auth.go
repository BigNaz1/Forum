package RebootForums

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		RenderTemplate(w, "register.html", nil)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "All fields are required"})
			return
		}

		var exists bool
		err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)", username, email).Scan(&exists)
		if err != nil {
			log.Printf("Database error during registration: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Database error"})
			return
		}
		if exists {
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Username or email already exists"})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Error creating user"})
			return
		}

		_, err = DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, string(hashedPassword))
		if err != nil {
			log.Printf("Error creating user: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Error creating user"})
			return
		}

		// Retrieve the user ID of the newly registered user
		var userID int
		err = DB.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&userID)
		if err != nil {
			log.Printf("Error retrieving user ID: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Error retrieving user"})
			return
		}

		// Generate a session token
		sessionToken, err := generateSessionToken()
		if err != nil {
			log.Printf("Error generating session token: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Error creating session"})
			return
		}

		expiryTime := time.Now().Add(24 * time.Hour)
		err = UpsertSession(&userID, sessionToken, expiryTime, false)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			RenderTemplate(w, "register.html", map[string]interface{}{"Message": "Error creating session"})
			return
		}

		// Set the session cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiryTime,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
		})

		// Redirect to the homepage or a different page as needed
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		message := ""
		if r.URL.Query().Get("registered") == "true" {
			message = "Registration successful. Please log in."
		}
		RenderTemplate(w, "login.html", map[string]interface{}{"Message": message})
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			RenderTemplate(w, "login.html", map[string]interface{}{"Message": "Username and password are required"})
			return
		}

		var user User
		var hashedPassword string
		err := DB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &hashedPassword)
		if err != nil {
			if err == sql.ErrNoRows {
				RenderTemplate(w, "login.html", map[string]interface{}{"Message": "Invalid username or password"})
			} else {
				log.Printf("Database error during login: %v", err)
				RenderTemplate(w, "login.html", map[string]interface{}{"Message": "An error occurred. Please try again later."})
			}
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
			RenderTemplate(w, "login.html", map[string]interface{}{"Message": "Invalid username or password"})
			return
		}

		// Delete any existing sessions for this user
		_, err = DB.Exec("DELETE FROM sessions WHERE user_id = ?", user.ID)
		if err != nil {
			log.Printf("Error deleting existing sessions: %v", err)
			RenderTemplate(w, "login.html", map[string]interface{}{"Message": "An error occurred. Please try again later."})
			return
		}

		sessionToken, err := generateSessionToken()
		if err != nil {
			log.Printf("Error generating session token: %v", err)
			RenderTemplate(w, "login.html", map[string]interface{}{"Message": "An error occurred. Please try again later."})
			return
		}

		expiryTime := time.Now().Add(24 * time.Hour)
		err = UpsertSession(&user.ID, sessionToken, expiryTime, false)
		if err != nil {
			log.Printf("Error creating session: %v", err)
			RenderTemplate(w, "login.html", map[string]interface{}{"Message": "An error occurred. Please try again later."})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiryTime,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil, // Set Secure flag if using HTTPS
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateSessionToken() (string, error) {
	token := uuid.New().String()
	return token, nil
}

func GetUserByUsername(username string) (*User, error) {
	var user User
	err := DB.QueryRow("SELECT id, username, email, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Email, &user.Password)
	if err != nil {
		log.Printf("Error getting user by username: %v", err)
		return nil, err
	}
	return &user, nil
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_token")
	if err != nil {
		// If there's no session cookie, just redirect to home page
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Delete the session from the database
	err = DeleteSession(c.Value)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
		// Continue with logout even if there's an error
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-1 * time.Hour), // Set expiry in the past
		MaxAge:   -1,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
