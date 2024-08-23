package RebootForums

import (
    "html/template"
    "log"
    "net/http"
)

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
    tmpl, err := template.ParseFiles("templates/" + tmplName)
    if err != nil {
        log.Printf("Error parsing template: %v", err)
        http.Error(w, "Error rendering page", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        log.Printf("Error executing template: %v", err)
        http.Error(w, "Error rendering page", http.StatusInternalServerError)
    }
}