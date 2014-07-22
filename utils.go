package main

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/knakk/rdf"
	"github.com/mreiferson/go-httpclient"
)

type remoteRepo struct {
	endpoint string
	client   *http.Client
}

func newRepo(endpoint string, openTimeout, readTimeout time.Duration) *remoteRepo {
	transport := &httpclient.Transport{
		ConnectTimeout:        openTimeout,
		RequestTimeout:        openTimeout + readTimeout,
		ResponseHeaderTimeout: readTimeout,
	}
	client := &http.Client{Transport: transport}
	return &remoteRepo{endpoint: endpoint, client: client}
}

func (r *remoteRepo) Close() {
	//r.client.Transport.Close()
}

// Query sends a request to a remote SPARQL endpoint and returns the unparsed
// response body
func (r *remoteRepo) Query(endpoint string, query string, format string) (io.ReadCloser, error) {
	reqDefaults := url.Values{}
	reqDefaults.Set("query", query)

	switch format {
	case "json":
		reqDefaults.Set("format", "application/sparql-results+json")
	case "rdf":
		reqDefaults.Set("format", "application/x-trig")
	default:
		reqDefaults.Set("format", "application/sparql-results+json")
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s?%v", endpoint, reqDefaults.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %v", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		// trim URL from error message
		i := strings.Index(err.Error(), "dial")
		if i != -1 {
			return nil, errors.New(err.Error()[i:])
		}
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SPARQL endpoint responded with HTTP status code: %v", resp.StatusCode)
	}

	return resp.Body, nil
}

// prefixify returns the prefixed form of an URI if the prefix and namespaces
// is found in the prefixes array, which must have the following form:
// ["dc", "http://purl.org/dc/terms/"], ["foaf", "http://xmlns....etc"]
//
//It returns the string unmodified if no match is found.
func prefixify(prefixes *[][]string, uri string) string {
	uriOriginal := uri

	// trim enclosing < > if present
	if strings.HasPrefix(uri, "<") && strings.HasSuffix(uri, ">") {
		uri = strings.TrimSuffix(uri, ">")[1:]
	}

	for _, prefixPair := range *prefixes {
		if strings.HasPrefix(uri, prefixPair[1]) {
			//avoid prefixing URI's witch has forward slash after the namespaced part
			if regexp.MustCompile(`\/`).MatchString(strings.TrimPrefix(uri, prefixPair[1])) {
				return uriOriginal
			}
			return strings.Replace(uri, prefixPair[1], prefixPair[0]+":", 1)
		}
	}
	return uriOriginal
}

func rejectWhereEmpty(key string, rdfMap *[]map[string]rdf.Term) *[]map[string]interface{} {
	// TODO clean up this function; choose another name too..
	included := make([]map[string]interface{}, 1)
	for _, m := range *rdfMap {
		if m[key] != nil {
			tm := make(map[string]interface{})
			for k, v := range m {
				if k != "g" && k != "p" && strings.HasPrefix(v.String(), "<"+conf.BaseURI) {
					var link string
					if conf.UI.FetchLiterals {
						link = fmt.Sprintf("<div class='relative'><a class=\"resource-link\" href='/%v'>%v</a><div class=\"tooltip\"><strong>%s</strong><div class='literals'>...</div></div></div>",
							v.String()[25:len(v.String())-1], template.HTMLEscapeString(v.String()), template.HTMLEscapeString(v.String()))
					} else {
						link = fmt.Sprintf("<a href='/%v'>%v</a>",
							v.String()[25:len(v.String())-1], template.HTMLEscapeString(v.String()))
					}

					// TODO magic number 25 = len(conf.BaseURI)??
					tm[k] = template.HTML(link)
				} else {
					if conf.Vocab.Enabled {
						tm[k] = prefixify(&conf.Vocab.Dict, v.String())
					} else {
						tm[k] = v.String()
					}
				}
			}
			included = append(included, tm)
		}
	}
	return &included
}

// findTitle returns the first literal object of a collection of triples which
// matches a prediate found in titlePredicates, or false if no match is found.
func findTitle(titlePredicates *[]string, rdfMap *[]map[string]rdf.Term) interface{} {
	if len(*titlePredicates) == 0 {
		return false
	}

	for _, m := range *rdfMap {
		for _, p := range *titlePredicates {
			if m["p"].String() == "<"+p+">" {
				return m["o"].String()
			}
		}
	}
	return false
}

// findImages
func findImages(rdfMap *[]map[string]rdf.Term) []string {
	images := make([]string, 0)
	if !conf.UI.ShowImages {
		return images
	}
	for _, m := range *rdfMap {
		for _, p := range conf.UI.ImagePredicates {
			if m["p"].String() == "<"+p+">" {
				images = append(images, strings.TrimSuffix(m["o"].String()[1:], ">"))
				if len(images) == conf.UI.NumImages {
					return images
				}
			}
		}
	}
	return images
}
