package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/knakk/fenster/rdf"
	"github.com/knakk/fenster/sparql"
)

const (
	version = "0.0"
	query   = `
		SELECT * 
		WHERE
		 { 
		   GRAPH ?g
		    {
		      { <%s> ?p ?o } 
		      UNION
		      { ?s ?p <%s> }
		    }
		 }
		 `
)

var templates = template.Must(template.ParseFiles("data/html/index.html"))

type mainHandler struct{}

func rejectWhereEmpty(key string, rdfMap []map[string]rdf.Term) *[]map[string]interface{} {
	included := make([]map[string]interface{}, 1)
	for _, m := range rdfMap {
		if m[key] != nil {
			tm := make(map[string]interface{})
			for k, v := range m {
				if k != "g" && strings.HasPrefix(v.String(), "<http://data.deichman.no/") {
					link := fmt.Sprintf("<a href='/%v'>%v</a>", v.String()[25:len(v.String())-1], template.HTMLEscapeString(v.String()))
					tm[k] = template.HTML(link)
				} else {
					tm[k] = v.String()
				}
			}
			included = append(included, tm)
		}
	}
	return &included
}

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := "http://data.deichman.no" + r.URL.Path
	q := fmt.Sprintf(query, uri, uri)
	res, err := sparql.Query("http://marc2rdf.deichman.no/sparql", q, 250*time.Millisecond, 500*time.Millisecond)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Name, Version, URI string
		AsSubject          *[]map[string]interface{}
		AsObject           *[]map[string]interface{}
	}{
		"Fenster",
		string(version),
		uri,
		rejectWhereEmpty("o", res.Solutions()),
		rejectWhereEmpty("s", res.Solutions()),
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// serveFile serves a single file from disk
func serveFile(filename string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filename)
	}
}

func main() {
	mux := http.NewServeMux()
	var handler mainHandler
	mux.HandleFunc("/css/styles.css", serveFile("data/css/styles.css"))
	mux.HandleFunc("/robots.txt", serveFile("data/robots.txt"))
	mux.Handle("/", handler)

	fmt.Println("Listening on localhost:4000 ...")
	http.ListenAndServe("localhost:4000", mux)
}
