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

var templates = template.Must(template.ParseFiles("data/html/index.html"))
var conf Config

type mainHandler struct{}

func rejectWhereEmpty(key string, rdfMap []map[string]rdf.Term) *[]map[string]interface{} {
	included := make([]map[string]interface{}, 1)
	for _, m := range rdfMap {
		if m[key] != nil {
			tm := make(map[string]interface{})
			for k, v := range m {
				if k != "g" && k != "p" && strings.HasPrefix(v.String(), "<http://data.deichman.no/") {
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
	for _, prefixPair := range conf.Vocab.Dict {
		if strings.HasPrefix(uri, "<"+prefixPair[1]) {
			return trimSuffix(strings.Replace(uri, prefixPair[1], prefixPair[0]+":", 1)[1:], ">")
		}
	}
	return uri
}

func findImages(rdfMap []map[string]rdf.Term) []template.HTML {
	images := make([]template.HTML, 0)
	if !conf.UI.ShowImages {
		return images
	}
	for _, m := range rdfMap {
		if m["p"].String() == "<http://xmlns.com/foaf/0.1/depiction>" {
			images = append(images, template.HTML("<img src=\""+m["o"].String()[1:len(m["o"].String())-1]+"\">"))
		}
	}
	return images
}

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := "http://data.deichman.no" + r.URL.Path
	q := fmt.Sprintf(query, uri, uri)
	res, err := sparql.Query(conf.QuadStore.Endpoint, q, 250*time.Millisecond, 500*time.Millisecond)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Name, Version, URI string
		AsSubject          *[]map[string]interface{}
		AsObject           *[]map[string]interface{}
		Images             []template.HTML
	}{
		"Fenster",
		string(version),
		uri,
		rejectWhereEmpty("o", res.Solutions()),
		rejectWhereEmpty("s", res.Solutions()),
		findImages(res.Solutions()),
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

	fmt.Println("Listening on localhost:4000 ...")
	http.ListenAndServe("localhost:4000", mux)
}
