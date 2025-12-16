package main

import (
	"html/template"
	"log"
	"net/http"
)

var templates = template.Must(template.ParseGlob("templates/*.html"))

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":  "Home",
		"User":   "Iago",
		"Logged": true,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		log.Printf("template exec error: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}

}

func main() {
	http.HandleFunc("/", indexHandler)

	// Arquivos estáticos
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static")),
		),
	)

	http.ListenAndServe(":8080", nil)
}
