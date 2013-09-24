package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/knakk/fenster/rdf"
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

var templates = template.Must(template.ParseFiles("data/html/index.html", "data/html/error.html"))
var conf Config

type mainHandler struct{}

func rejectWhereEmpty(key string, rdfMap *[]map[string]rdf.Term) *[]map[string]interface{} {
	included := make([]map[string]interface{}, 1)
	for _, m := range *rdfMap {
		if m[key] != nil {
			tm := make(map[string]interface{})
			for k, v := range m {
				if k != "g" && k != "p" && strings.HasPrefix(v.String(), "<"+conf.BaseURI) {
					link := fmt.Sprintf("<a href='/%v'>%v</a>", v.String()[25:len(v.String())-1], template.HTMLEscapeString(prefixify(v.String())))
					tm[k] = template.HTML(link)
				} else {
					tm[k] = prefixify(v.String())
				}
			}
			included = append(included, tm)
		}
	}
	return &included
}

func trimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func prefixify(uri string) string {
	if !conf.Vocab.Enabled {
		return uri
	}
	for _, prefixPair := range conf.Vocab.Dict {
		if strings.HasPrefix(uri, "<"+prefixPair[1]) {
			return trimSuffix(strings.Replace(uri, prefixPair[1], prefixPair[0]+":", 1)[1:], ">")
		}
	}
	return uri
}

func findTitle(rdfMap *[]map[string]rdf.Term) interface{} {
	if len(conf.UI.TitlePredicates) == 0 {
		return false
	}

	for _, m := range *rdfMap {
		for _, p := range conf.UI.TitlePredicates {
			if m["p"].String() == "<"+p+">" {
				return m["o"].Value()
			}
		}
	}
	return false
}

func findImages(rdfMap *[]map[string]rdf.Term) []string {
	images := make([]string, 0)
	if !conf.UI.ShowImages {
		return images
	}
	for _, m := range *rdfMap {
		for _, p := range conf.UI.ImagePredicates {
			if m["p"].String() == "<"+p+">" {
				images = append(images, trimSuffix(m["o"].String()[1:], ">"))
				if len(images) == conf.UI.NumImages {
					return images
				}
			}
		}
	}
	return images
}

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

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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
		ShortURI           string
	}{
		findTitle(&solutions),
		"Fenster",
		string(version),
		uri,
		subj,
		obj,
		len(*subj) - 1,
		len(*obj) - 1,
		findImages(&solutions),
		prefixify("<" + uri + ">"),
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
	mux.Handle("/", handler)

	fmt.Printf("Listening on port %d ...\n", conf.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), mux)
}
