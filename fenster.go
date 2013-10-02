package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/knakk/fenster/sparql"
)

const (
	version = "0.1"
	query   = `
		SELECT *
		WHERE { GRAPH ?g { { <%s> ?p ?o } UNION { ?s ?p <%s> } } }
		LIMIT %d
		`
	query2 = `
		CONSTRUCT { GRAPH ?g { <%s> ?p ?o . ?s ?p <%s> } }
		WHERE { GRAPH ?g { { <%s> ?p ?o } UNION { ?s ?p <%s> } } }
	   `
)

var (
	templates = template.Must(template.ParseFiles("data/html/index.html", "data/html/error.html"))
	conf      Config
)

type mainHandler struct{}

// rdfHandler serves the quads in TriG syntax
// http://wifo5-03.informatik.uni-mannheim.de/bizer/trig/
func rdfHandler(w http.ResponseWriter, r *http.Request) {
	uri := conf.BaseURI + strings.TrimSuffix(r.URL.Path, ".rdf")
	format := "rdf"
	q := fmt.Sprintf(query2, uri, uri, uri, uri)

	resp, err := sparql.Query(conf.QuadStore.Endpoint, q, format,
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond, time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond)
	if err != nil {
		errorHandler(w, r, err.Error()+". Refresh to try again.\n\nYou can increase the timeout values in Fensters configuration file.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-trig")
	io.WriteString(w, string(resp))
}

// jsonHandler serves the raw "application/sparql-results+json" results from
// the SPARQL endpoint
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	uri := conf.BaseURI + strings.TrimSuffix(r.URL.Path, ".json")
	q := fmt.Sprintf(query, uri, uri, conf.QuadStore.ResultsLimit)
	format := "json"

	resp, err := sparql.Query(conf.QuadStore.Endpoint, q, format,
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond, time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond)
	if err != nil {
		errorHandler(w, r,
			err.Error()+`. Refresh to try again.\n\n
			You can increase the timeout values in Fensters configuration file.`,
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, string(resp))
}

// mainHandler serves the resource HTML presentation, or dispatches to the
// rdfHandler or jsonHandler
func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// redirect from root to info page
		http.Redirect(w, r, conf.UI.RootRedirectTo, http.StatusFound)
		return
	}

	var uri string
	resolved := false
	suffix := regexp.MustCompile(`\.[a-z1-9]+$`).FindString(r.URL.Path)

	switch suffix {
	case "":
		break
	case ".html":
		uri = conf.BaseURI + strings.TrimSuffix(r.URL.Path, ".html")
		resolved = true
	case ".json":
		jsonHandler(w, r)
		return
	case ".rdf":
		rdfHandler(w, r)
		return
	default:
		errorHandler(w, r,
			fmt.Sprintf("Unsupported output format: %s.\n\n"+
				"Valid formats are: html, json, rdf", suffix[1:]), http.StatusBadRequest)
		return
	}

	// The URI should be an exlusive identifier of the resource; so we redirect
	// to URI+.html
	if !resolved {
		http.Redirect(w, r, r.URL.Path+".html", http.StatusFound)
		return
	}

	q := fmt.Sprintf(query, uri, uri, conf.QuadStore.ResultsLimit)
	resp, err := sparql.Query(conf.QuadStore.Endpoint, q, "json",
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond,
		time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond)
	if err != nil {
		errorHandler(w, r,
			err.Error()+". Refresh to try again.\n\nYou can increase the timeout"+
				"values in Fensters configuration file.", http.StatusInternalServerError)
		return
	}

	res, err := sparql.ParseJSON(resp)
	if err != nil {
		errorHandler(w, r,
			"Failed to parse JSON response from remote SPARQL endpoint.",
			http.StatusInternalServerError)
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
		Title               interface{}
		License, LicenseURL string
		Name, Version, URI  string
		AsSubject           *[]map[string]interface{}
		AsObject            *[]map[string]interface{}
		AsSubjectSize       int
		AsObjectSize        int
		Images              []string
	}{
		findTitle(&conf.UI.TitlePredicates, &solutions),
		conf.License,
		conf.LicenseURL,
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

// errorHandler serves 40x & 50x error pages
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
