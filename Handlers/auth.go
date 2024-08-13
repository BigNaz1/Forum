package GreatForums

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

type User struct {
	ID       int
	Username string
	Email    string
	Password string
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderTemplate(w, "register.html", "")
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			renderTemplate(w, "register.html", "All fields are required")
			return
		}

		var exists bool
		err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)", username, email).Scan(&exists)
		if err != nil {
			renderTemplate(w, "register.html", "Database error")
			return
		}
		if exists {
			renderTemplate(w, "register.html", "Username or email already exists")
			return
		}

		hashedPassword := hashPassword(password)

		_, err = DB.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", username, email, hashedPassword)
		if err != nil {
			renderTemplate(w, "register.html", "Error creating user")
			return
		}

		http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		message := ""
		if r.URL.Query().Get("registered") == "true" {
			message = "Registration successful. Please log in."
		}
		renderTemplate(w, "login.html", message)
		return
	}

	if r.Method == "POST" {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			renderTemplate(w, "login.html", "Username and password are required")
			return
		}

		var user User
		err := DB.QueryRow("SELECT id, username, password FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &user.Password)
		if err != nil {
			if err == sql.ErrNoRows {
				renderTemplate(w, "login.html", "Invalid username or password")
			} else {
				renderTemplate(w, "login.html", "Database error")
			}
			return
		}

		if !checkPasswordHash(password, user.Password) {
			renderTemplate(w, "login.html", "Invalid username or password")
			return
		}

		sessionToken, err := generateSessionToken()
		if err != nil {
			renderTemplate(w, "login.html", "Error generating session token")
			return
		}

		_, err = DB.Exec("INSERT INTO sessions (user_id, token, expiry) VALUES (?, ?, ?)",
			user.ID, sessionToken, time.Now().Add(24*time.Hour))
		if err != nil {
			renderTemplate(w, "login.html", "Error creating session")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_token",
			Value:   sessionToken,
			Expires: time.Now().Add(24 * time.Hour),
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func renderTemplate(w http.ResponseWriter, tmplName, message string) {
	tmpl, err := template.ParseFiles("templates/" + tmplName)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, struct{ Message string }{Message: message})
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash)
}

func checkPasswordHash(password, hash string) bool {
	return hashPassword(password) == hash
}

func generateSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
