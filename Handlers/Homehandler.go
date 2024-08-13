package GreatForums

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Post represents a forum post
type Post struct {
	ID        int
	Title     string
	Content   string
	Author    string
	CreatedAt time.Time
}

func (p Post) FormattedCreatedAt() string {
	return p.CreatedAt.Format("January 2, 2006 at 3:04 PM")
}

// Category represents a forum category
type Category struct {
	ID   int
	Name string
}

// GetRecentPosts fetches recent posts from the database
func GetRecentPosts(limit int) ([]Post, error) {
	// This is a placeholder implementation. Replace with actual database query.
	posts := []Post{
		{ID: 1, Title: "First Post", Content: "This is the first post", Author: "User1", CreatedAt: time.Now()},
		{ID: 2, Title: "Second Post", Content: "This is the second post", Author: "User2", CreatedAt: time.Now()},
	}
	return posts, nil
}

// GetAllCategories fetches all categories from the database
func GetAllCategories() ([]Category, error) {
	// This is a placeholder implementation. Replace with actual database query.
	categories := []Category{
		{ID: 1, Name: "General"},
		{ID: 2, Name: "Technology"},
		{ID: 3, Name: "Sports"},
	}
	return categories, nil
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request for path: %s", r.URL.Path)

	if r.URL.Path != "/" {
		log.Printf("Path is not /, returning 404")
		http.NotFound(w, r)
		return
	}

	// Fetch recent posts
	recentPosts, err := GetRecentPosts(10) // Fetch 10 most recent posts
	if err != nil {
		log.Printf("Failed to fetch recent posts: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Fetch categories
	categories, err := GetAllCategories()
	if err != nil {
		log.Printf("Failed to fetch categories: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prepare data for the template
	data := struct {
		RecentPosts []Post
		Categories  []Category
	}{
		RecentPosts: recentPosts,
		Categories:  categories,
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current working directory: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Construct the full path to the template file
	templatePath := filepath.Join(cwd, "templates", "home.html")
	log.Printf("Attempting to load template from: %s", templatePath)

	// Check if the file exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		log.Printf("Template file does not exist: %s", templatePath)
		http.Error(w, "Template file not found", http.StatusInternalServerError)
		return
	}

	// Parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		log.Printf("Failed to parse template: %v", err)
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Execute the template
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Failed to execute template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully rendered home page")
}
