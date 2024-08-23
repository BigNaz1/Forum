package RebootForums

import (
	"database/sql"
	"log"
	"net/http"
	"time"
)

func SessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("session_token")
		if err != nil {
			newToken, err := generateSessionToken()
			if err != nil {
				log.Printf("Error generating session token: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			expiry := time.Now().Add(24 * time.Hour)
			err = UpsertSession(nil, newToken, expiry, true)
			if err != nil {
				log.Printf("Error creating guest session: %v", err)
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "session_token",
				Value:   newToken,
				Expires: expiry,
			})
		} else {
			var exists bool
			err = DB.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE token = ?)", token.Value).Scan(&exists)
			if err != nil || !exists {
				http.SetCookie(w, &http.Cookie{
					Name:    "session_token",
					Value:   "",
					Expires: time.Now().Add(-1 * time.Hour),
				})
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			err = UpdateSessionActivity(token.Value)
			if err != nil {
				log.Printf("Error updating session activity: %v", err)
			}
		}
		next.ServeHTTP(w, r)
	}
}
func UpsertSession(userID *int, token string, expiry time.Time, isGuest bool) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if userID != nil && !isGuest {
		_, err = tx.Exec("DELETE FROM sessions WHERE user_id = ?", *userID)
		if err != nil {
			return err
		}
	}
	query := `
    INSERT INTO sessions (user_id, token, expiry, is_guest, last_activity, created_at)
    VALUES (?, ?, ?, ?, ?, ?)
    ON CONFLICT(token) DO UPDATE SET
    user_id = ?, expiry = ?, is_guest = ?, last_activity = ?
    `
	now := time.Now()
	_, err = tx.Exec(query, userID, token, expiry, isGuest, now, now,
		userID, expiry, isGuest, now)
	if err != nil {
		return err
	}
	return tx.Commit()
}
func UpdateSessionActivity(token string) error {
	_, err := DB.Exec("UPDATE sessions SET last_activity = ? WHERE token = ?", time.Now(), token)
	return err
}
func GetActiveSessions() (int, int, error) {
	var registeredCount, guestCount int
	err := DB.QueryRow(`
		SELECT 
			COUNT(CASE WHEN is_guest = 0 THEN 1 END) as registered_count,
			COUNT(CASE WHEN is_guest = 1 THEN 1 END) as guest_count
		FROM sessions
		WHERE last_activity > ?
	`, time.Now().Add(-5*time.Minute)).Scan(&registeredCount, &guestCount)
	return registeredCount, guestCount, err
}
func CleanupSessions() {
	_, err := DB.Exec("DELETE FROM sessions WHERE expiry < ?", time.Now())
	if err != nil {
		log.Printf("Error cleaning up sessions: %v", err)
	}
}
func init() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			CleanupSessions()
		}
	}()
}
func GetSessionDuration(token string) (time.Duration, error) {
	var createdAt time.Time
	var lastActivity time.Time
	err := DB.QueryRow("SELECT created_at, last_activity FROM sessions WHERE token = ?", token).Scan(&createdAt, &lastActivity)
	if err != nil {
		return 0, err
	}
	return lastActivity.Sub(createdAt), nil
}
func DeleteSession(token string) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
		return err
	}
	return nil
}
func GetUserFromSession(r *http.Request) (*User, error) {
	c, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			// No cookie found, which is fine for guest users
			return nil, nil
		}
		// For any other error, log it and return
		log.Printf("Error getting session cookie: %v", err)
		return nil, err
	}

	sessionToken := c.Value
	var userID int
	var isGuest bool
	err = DB.QueryRow("SELECT user_id, is_guest FROM sessions WHERE token = ?", sessionToken).Scan(&userID, &isGuest)
	if err != nil {
		if err == sql.ErrNoRows {
			// Session not found in database, treat as guest
			log.Printf("Session not found for token: %s", sessionToken)
			return nil, nil
		}
		// For any other database error, log it and return
		log.Printf("Database error when fetching session: %v", err)
		return nil, err
	}

	if isGuest {
		// Guest session, return nil user
		return nil, nil
	}

	user, err := GetUserByID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, which shouldn't happen for a valid session
			log.Printf("User not found for session user_id: %d", userID)
			return nil, nil
		}
		// For any other error getting user, log it and return
		log.Printf("Error getting user by ID: %v", err)
		return nil, err
	}

	return user, nil
}

func GetUserByID(id int) (*User, error) {
    var user User
    err := DB.QueryRow("SELECT id, username, email FROM users WHERE id = ?", id).Scan(&user.ID, &user.Username, &user.Email)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
