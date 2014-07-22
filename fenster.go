package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	//"github.com/gorilla/handlers"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
	"github.com/rcrowley/go-metrics"
)

const (
	version = "0.2"
	qSelect = `
		SELECT *
		WHERE { GRAPH ?g { { <%s> ?p ?o } UNION { ?s ?p <%s> } } }
		LIMIT %d`
	qCount = `
		SELECT COUNT(?s) AS ?maxO, COUNT(?o) as ?maxS
		WHERE { GRAPH ?g { { <%s> ?p ?o } UNION { ?s ?p <%s> } } }`
	qConstruct = `
		CONSTRUCT { GRAPH ?g { <%s> ?p ?o . ?s ?p <%s> } }
		WHERE { GRAPH ?g { { <%s> ?p ?o } UNION { ?s ?p <%s> } } }`
)

var (
	templates = template.Must(template.ParseFiles("data/html/index.html", "data/html/error.html"))
	conf      Config
	repo      *remoteRepo
	status    *appMetrics
)

type mainHandler struct{}

// rdfHandler serves the quads in TriG syntax
// http://wifo5-03.informatik.uni-mannheim.de/bizer/trig/
func rdfHandler(w http.ResponseWriter, r *http.Request) {
	uri := conf.BaseURI + strings.TrimSuffix(r.URL.Path, ".rdf")
	format := "rdf"
	q := fmt.Sprintf(qConstruct, uri, uri, uri, uri)

	resp, err := repo.Query(conf.QuadStore.Endpoint, q, format)
	if err != nil {
		errorHandler(w, r, err.Error()+". Refresh to try again.\n\nYou can increase the timeout values in Fensters configuration file.", http.StatusInternalServerError)
		return
	}
	defer resp.Close()

	w.Header().Set("Content-Type", "application/x-trig")
	io.Copy(w, resp)
}

// jsonHandler serves the raw "application/sparql-results+json" results from
// the SPARQL endpoint
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	uri := conf.BaseURI + strings.TrimSuffix(r.URL.Path, ".json")
	q := fmt.Sprintf(qSelect, uri, uri, conf.QuadStore.ResultsLimit)
	format := "json"

	resp, err := repo.Query(conf.QuadStore.Endpoint, q, format)
	if err != nil {
		errorHandler(w, r,
			err.Error()+`. Refresh to try again.\n\n
			You can increase the timeout values in Fensters configuration file.`,
			http.StatusInternalServerError)
		return
	}
	defer resp.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp)
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

	q := fmt.Sprintf(qSelect, uri, uri, conf.QuadStore.ResultsLimit)
	resp, err := repo.Query(conf.QuadStore.Endpoint, q, "json")
	if err != nil {
		//println(err.Error())
		errorHandler(w, r,
			err.Error()+". Refresh to try again.\n\nYou can increase the timeout"+
				" values in Fensters configuration file.", http.StatusInternalServerError)
		return
	}
	defer resp.Close()

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

	var maxS, maxO int
	if len(res.Results.Bindings) >= conf.QuadStore.ResultsLimit {
		// Fetch solution counts, if we hit the results limit
		q := fmt.Sprintf(qCount, uri, uri)
		// TODO use shorter timeouts? This is not vital information
		resp, err := repo.Query(conf.QuadStore.Endpoint, q, "json")
		if err == nil {
			res, err := sparql.ParseJSON(resp)
			if err == nil {
				b := res.Bindings()
				maxS = b["maxS"][0].(*rdf.Literal).Value.(int)
				maxO = b["maxO"][0].(*rdf.Literal).Value.(int)
			}
			resp.Close()
		}

	}

	solutions := res.Solutions()
	subj := rejectWhereEmpty("o", &solutions)
	obj := rejectWhereEmpty("s", &solutions)
	data := struct {
		Title               interface{}
		License, LicenseURL string
		Endpoint            string
		Name, Version, URI  string
		AsSubject           *[]map[string]interface{}
		AsObject            *[]map[string]interface{}
		AsSubjectSize       int
		AsObjectSize        int
		MaxSubject          int
		MaxObject           int
		Images              []string
	}{
		findTitle(&conf.UI.TitlePredicates, &solutions),
		conf.License,
		conf.LicenseURL,
		conf.QuadStore.Endpoint,
		"Fenster",
		string(version),
		uri,
		subj,
		obj,
		len(*subj) - 1,
		len(*obj) - 1,
		maxS,
		maxO,
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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(status.Export())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	// Load config file
	if _, err := toml.DecodeFile("config.ini", &conf); err != nil {
		log.Fatal("Couldn't parse config file: ", err)
	}

	// Setup remote repository
	repo = newRepo(
		conf.QuadStore.Endpoint,
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond,
		time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond,
	)

	// Register metrics
	status = registerMetrics()

	// HTTP routing
	mux := http.NewServeMux()
	var handler mainHandler
	mux.HandleFunc("/robots.txt", serveFile("data/robots.txt"))
	mux.HandleFunc("/css/styles.css", serveFile("data/css/styles.css"))
	mux.HandleFunc("/favicon.ico", serveFile("data/favicon.ico"))
	mux.HandleFunc("/.status", statusHandler)
	mux.Handle("/", Timed(CountedByStatusXX(handler, "status", metrics.DefaultRegistry),
		"responseTime",
		metrics.DefaultRegistry))

	fmt.Printf("Listening on port %d ...\n", conf.ServePort)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.ServePort), mux) //handlers.CompressHandler(mux))
}
