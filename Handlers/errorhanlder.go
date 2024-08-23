package RebootForums

import (
	"log"
	"net/http"
)

func Error400Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	err := RenderTemplate(w, "error_400.html", nil)
	if err != nil {
		log.Printf("Error rendering 400 template: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
	}
}

func Error404Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	err := RenderTemplate(w, "error_404.html", nil)
	if err != nil {
		log.Printf("Error rendering 404 template: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
	}
}

func Error500Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	err := RenderTemplate(w, "error_500.html", nil)
	if err != nil {
		log.Printf("Error rendering 500 template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// CustomNotFoundHandler is a wrapper to use Error404Handler for undefined routes
func CustomNotFoundHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			Error404Handler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// ErrorHandler is a middleware that recovers from panics and serves an error page
func ErrorHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error here if you have a logging system
				Error500Handler(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	}
}
