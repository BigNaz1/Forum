package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	GreatForums "GreatForums/Handlers"

	_ "github.com/mattn/go-sqlite3"
)

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func main() {
	// Initialize database
	err := GreatForums.InitDB("./forum.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer GreatForums.DB.Close()

	// Create tables
	err = GreatForums.CreateTables()
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Get the absolute path to the templates directory
	templatesDir, err := filepath.Abs("./templates")
	if err != nil {
		log.Fatal("Failed to get absolute path for templates directory:", err)
	}
	log.Printf("Templates directory: %s", templatesDir)

	// Set the templates directory in the GreatForums package
	GreatForums.SetTemplatesDir(templatesDir)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Set up routes
	mux.HandleFunc("/", GreatForums.HomeHandler) // Remove makeHandler wrapper for HomeHandler
	mux.HandleFunc("/register", makeHandler(GreatForums.RegisterHandler))
	mux.HandleFunc("/login", makeHandler(GreatForums.LoginHandler))
	mux.HandleFunc("/logout", makeHandler(logoutHandler))
	mux.HandleFunc("/post", makeHandler(postHandler))
	mux.HandleFunc("/comment", makeHandler(commentHandler))

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user logout
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement post creation and viewing
}

func commentHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement comment creation and viewing
}
