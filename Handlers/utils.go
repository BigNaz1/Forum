package RebootForums

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, tmplName string, data interface{}) error {
	tmpl, err := template.ParseFiles("templates/" + tmplName)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return fmt.Errorf("error parsing template: %v", err)
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return fmt.Errorf("error executing template: %v", err)
	}

	return nil
}
