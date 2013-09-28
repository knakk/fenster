package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/knakk/fenster/sparql"
)

const (
	version = "0.1"
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

var (
	templates = template.Must(template.ParseFiles("data/html/index.html", "data/html/error.html"))
	conf      Config
)

type mainHandler struct{}

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.Redirect(w, r, conf.UI.RootRedirectTo, http.StatusFound)
		return
	}

	uri := conf.BaseURI + r.URL.Path
	q := fmt.Sprintf(query, uri, uri)
	res, err := sparql.Query(conf.QuadStore.Endpoint, q,
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond, time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond)
	if err != nil {
		errorHandler(w, r, err.Error()+". Refresh to try again.\n\nYou can increase the timeout values in Fensters configuration file.", http.StatusInternalServerError)
		return
	}

	if len(res.Results.Bindings) == 0 {
		errorHandler(w, r, "This URI has no information", http.StatusNotFound)
		return
	}

	solutions := res.Solutions()
	subj := rejectWhereEmpty("o", &solutions)
	obj := rejectWhereEmpty("s", &solutions)
	data := struct {
		Title              interface{}
		Name, Version, URI string
		AsSubject          *[]map[string]interface{}
		AsObject           *[]map[string]interface{}
		AsSubjectSize      int
		AsObjectSize       int
		Images             []string
	}{
		findTitle(&conf.UI.TitlePredicates, &solutions),
		"Fenster",
		string(version),
		uri,
		subj,
		obj,
		len(*subj) - 1,
		len(*obj) - 1,
		findImages(&solutions),
	}

	err = templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// errorHandler serves 404 & 500 error pages
func errorHandler(w http.ResponseWriter, r *http.Request, msg string, status int) {
	w.WriteHeader(status)
	data := struct {
		ErrorCode int
		ErrorMsg  string
	}{
		status,
		msg,
	}

	err := templates.ExecuteTemplate(w, "error.html", data)
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

// init: load config.ini
func init() {
	if _, err := toml.DecodeFile("config.ini", &conf); err != nil {
		log.Fatal("Couldn't parse config file: ", err)
	}
}

func main() {
	mux := http.NewServeMux()
	var handler mainHandler
	mux.HandleFunc("/robots.txt", serveFile("data/robots.txt"))
	mux.HandleFunc("/css/styles.css", serveFile("data/css/styles.css"))
	mux.HandleFunc("/favicon.ico", serveFile("data/favicon.ico"))
	mux.Handle("/", handler)

	fmt.Printf("Listening on port %d ...\n", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), mux)
}
