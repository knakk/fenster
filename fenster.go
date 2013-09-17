package main

import (
	"fmt"
	"html/template"
	"net/http"
)

const version = "0.0"

var templates = template.Must(template.ParseFiles("data/html/index.html"))

type mainHandler struct{}

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := struct {
		Name, Version string
	}{
		"Fenster",
		string(version),
	}

	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

func main() {
	mux := http.NewServeMux()
	var handler mainHandler
	mux.Handle("/", handler)
	mux.HandleFunc("/css/styles.css", serveFile("data/css/styles.css"))
	mux.HandleFunc("/robots.txt", serveFile("data/robots.txt"))

	fmt.Println("Listening on localhost:4000 ...")
	http.ListenAndServe("localhost:4000", mux)
}
