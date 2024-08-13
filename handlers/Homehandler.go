package GreatForums

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get recent posts count
	var postCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	if err != nil {
		http.Error(w, "Failed to fetch post count", http.StatusInternalServerError)
		return
	}

	// Get category count
	var categoryCount int
	err = DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	if err != nil {
		http.Error(w, "Failed to fetch category count", http.StatusInternalServerError)
		return
	}

	// Render welcome message with some basic stats
	welcomeMessage := fmt.Sprintf("Welcome to the Great Forums! We have %d posts in %d categories.", postCount, categoryCount)
	fmt.Fprintf(w, welcomeMessage)
}

// InitDB initializes the database connection
func InitDB(dataSourceName string) error {
	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return err
	}
	return DB.Ping()
}
