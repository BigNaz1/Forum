package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	RebootForums "RebootForums/Handlers"

	_ "github.com/mattn/go-sqlite3"
)

func makeHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r)
	}
}

func main() {
	// Initialize database
	err := RebootForums.InitDB("./forum.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer RebootForums.DB.Close()

	// Create tables
	err = RebootForums.CreateTables()
	if err != nil {
		log.Fatal("Failed to create tables:", err)
	}

	// Add this line to ensure the updated_at column exists
	err = RebootForums.AddUpdatedAtColumn()
	if err != nil {
		log.Fatal("Failed to add updated_at column:", err)
	}

	// Get the absolute path to the templates directory
	templatesDir, err := filepath.Abs("./templates")
	if err != nil {
		log.Fatal("Failed to get absolute path for templates directory:", err)
	}
	log.Printf("Templates directory: %s", templatesDir)

	// Set the templates directory in the RebootForums package
	RebootForums.SetTemplatesDir(templatesDir)

	// Update the routes
	mux := http.NewServeMux()

	// Set up routes
	mux.HandleFunc("/", RebootForums.HomeHandler)
	mux.HandleFunc("POST /register", makeHandler(RebootForums.RegisterHandler))
	mux.HandleFunc("POST /login", makeHandler(RebootForums.LoginHandler))
	mux.HandleFunc("/logout", makeHandler(RebootForums.LogoutHandler))
	// Post-related routes
	mux.HandleFunc("/create-post", makeHandler(RebootForums.CreatePostFormHandler))
	mux.HandleFunc("/post/", makeHandler(RebootForums.ViewPostHandler))
	mux.HandleFunc("DELETE /delete-post/", makeHandler(RebootForums.DeletePostHandler))
	mux.HandleFunc("/like-post", makeHandler(RebootForums.LikePostHandler))
	mux.HandleFunc("/like-comment", makeHandler(RebootForums.LikeCommentHandler))
	mux.HandleFunc("/add-comment", makeHandler(RebootForums.AddCommentHandler))
	// Explicit error routes
	mux.HandleFunc("/400", RebootForums.Error400Handler)
	mux.HandleFunc("/404", RebootForums.Error404Handler)
	mux.HandleFunc("/500", RebootForums.Error500Handler)

	// Serve static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
