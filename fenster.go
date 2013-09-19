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

func findTitle(uri string, rdfMap []map[string]rdf.Term) string {
	if len(conf.UI.TitlePredicates) == 0 {
		return uri
	}

	for _, m := range rdfMap {
		for _, p := range conf.UI.TitlePredicates {
			if m["p"].String() == "<"+p+">" {
				return m["o"].String()
			}
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
		for _, p := range conf.UI.ImagePredicates {
			if m["p"].String() == "<"+p+">" {
				images = append(images, template.HTML("<img src=\""+m["o"].String()[1:len(m["o"].String())-1]+"\">"))
				if len(images) == conf.UI.NumImages {
					return images
				}
			}
		}
	}
	return images
}

func (m mainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	uri := conf.BaseURI + r.URL.Path
	q := fmt.Sprintf(query, uri, uri)
	res, err := sparql.Query(conf.QuadStore.Endpoint, q,
		time.Duration(conf.QuadStore.OpenTimeout)*time.Millisecond, time.Duration(conf.QuadStore.ReadTimeout)*time.Millisecond)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Title              string
		Name, Version, URI string
		AsSubject          *[]map[string]interface{}
		AsObject           *[]map[string]interface{}
		Images             []template.HTML
		ShortURI           string
	}{
		findTitle(uri, res.Solutions()),
		"Fenster",
		string(version),
		uri,
		rejectWhereEmpty("o", res.Solutions()),
		rejectWhereEmpty("s", res.Solutions()),
		findImages(res.Solutions()),
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
