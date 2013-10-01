// Package sparql includes functions for performing SPARQL requests and parsing
// the results into RDF statements.
package sparql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/knakk/fenster/rdf"
)

// Results holds the unmarshaled sparql/json response.
type Results struct {
	Head    jsonHeader
	Results jsonRes
}

type jsonHeader struct {
	Link []string
	Vars []string
}

type jsonRes struct {
	Distinct bool
	Ordered  bool
	Bindings []map[string]jsonBinding
}

type jsonBinding struct {
	Type     string // Can be "uri", "literal", "typed-literal" or "bnode"
	Value    string
	Lang     string `json:"xml:lang"`
	DataType string
}

// Query sends a request to a remote SPARQL endpoint and returns the unparsed
// response body
func Query(endpoint string, query string, format string, open time.Duration, read time.Duration) ([]byte, error) {
	reqDefaults := url.Values{}
	reqDefaults.Set("query", query)

	switch format {
	case "json", "html":
		reqDefaults.Set("format", "application/sparql-results+json")
	case "xml":
		reqDefaults.Set("format", "application/sparql-results+xml")
	case "n3":
		reqDefaults.Set("format", "text/n3")
	default:
		reqDefaults.Set("format", "application/sparql-results+json")
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%v", endpoint, reqDefaults.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("error preparing http request: %v", err)
	}

	client := newTimeoutClient(open, read)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to remote SPARQL endpoint timed out")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request failed with status code: %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %v", err)
	}

	return body, nil
}

// ParseJSON unmarshals a 'application/sparql-results+json' response into
// the Results struct
func ParseJSON(raw []byte) (*Results, error) {
	var res *Results
	err := json.Unmarshal(raw, &res)
	if err != nil {
		return nil, fmt.Errorf("json parse error: %v", err)
	}

	return res, nil
}

// Bindings holds a map of the bound variables from the SPARQL response, where
// each variable points to one or more RDF Terms.
func (r *Results) Bindings() map[string][]rdf.Term {
	rb := make(map[string][]rdf.Term)
	for _, v := range r.Head.Vars {
		for _, b := range r.Results.Bindings {
			t, err := termFromJSON(b[v])
			if err == nil {
				rb[v] = append(rb[v], t)
			}
		}
	}
	return rb
}

// Solutions returns an array of the solutions, each containing a map of every
// binding in the solution.
func (r *Results) Solutions() []map[string]rdf.Term {
	rs := []map[string]rdf.Term{}
	m := make(map[string]rdf.Term)
	for _, b := range r.Results.Bindings {
		for k, v := range b {
			t, err := termFromJSON(v)
			if err == nil {
				m[k] = t
			}
		}
		rs = append(rs, m)
		m = make(map[string]rdf.Term)
	}
	return rs
}

func termFromJSON(j jsonBinding) (rdf.Term, error) {
	switch j.Type {
	case "bnode":
		return rdf.Blank(j.Value), nil
	case "uri":
		return rdf.Uri(j.Value), nil
	case "literal":
		if j.Lang != "" {
			return rdf.Literal{Val: j.Value, Lang: j.Lang}, nil
		}
		return rdf.Literal{Val: j.Value}, nil
	case "typed-literal":
		return rdf.Literal{Val: j.Value, DataType: rdf.Uri(j.DataType)}, nil
	default:
		return rdf.Literal{}, fmt.Errorf("unknown term type")
	}
}

// Update sends a SPARQL UPDATE request to remote endpoint. It returns whatever
// answer the remote service gave.
func Update(endpoint string, query string) (string, error) {
	return "Not implemented", nil
}
