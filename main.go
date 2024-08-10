package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Set up routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/comment", commentHandler)

	// Start the server
	fmt.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the Forum!")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user registration
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement user login
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