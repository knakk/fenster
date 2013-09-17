package sparql

import (
	"log"
	"testing"

	"github.com/knakk/fenster/rdf"
)

// Example response in the format "application/sparql-results+json" taken from
// the W3C specification at http://www.w3.org/TR/rdf-sparql-json-res/
const response = `
{
   "head": {
       "link": [
           "http://www.w3.org/TR/rdf-sparql-XMLres/example.rq"
           ],
       "vars": [
           "x",
           "hpage",
           "name",
           "mbox",
           "age",
           "blurb",
           "friend"
           ]
       },
   "results": {
       "bindings": [
               {
                   "x" : {
                     "type": "bnode",
                     "value": "r1"
                   },

                   "hpage" : {
                     "type": "uri",
                     "value": "http://work.example.org/alice/"
                   },

                   "name" : {
                     "type": "literal",
                     "value": "Alice"
                   },
                   
                   "mbox" : {
                     "type": "literal",
                     "value": ""
                   },

                   "blurb" : {
                     "datatype": "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral",
                     "type": "typed-literal",
                     "value": "<p xmlns=\"http://www.w3.org/1999/xhtml\">My name is <b>alice</b></p>"
                   },

                   "friend" : {
                     "type": "bnode",
                     "value": "r2"
                   }
               },{
                   "x" : {
                     "type": "bnode",
                     "value": "r2"
                   },
                   
                   "hpage" : {
                     "type": "uri",
                     "value": "http://work.example.org/bob/"
                   },
                   
                   "name" : {
                     "type": "literal",
                     "value": "Bob",
                     "xml:lang": "en"
                   },

                   "mbox" : {
                     "type": "uri",
                     "value": "mailto:bob@work.example.org"
                   },

                   "friend" : {
                     "type": "bnode",
                     "value": "r1"
                   }
               }
           ]
       }
   }
`

var res *Results

func init() {
	var err error
	res, err = parse([]byte(response))
	if err != nil {
		log.Fatal("json parse error: ", err)
	}
}

func TestJsonUnmarshal(t *testing.T) {

	var resTests = []struct {
		in  interface{}
		out interface{}
	}{
		{res.Head.Link[0], "http://www.w3.org/TR/rdf-sparql-XMLres/example.rq"},
		{len(res.Head.Vars), 7},
		{len(res.Results.Bindings), 2},
		{res.Results.Bindings[0]["x"].Type, "bnode"},
		{res.Results.Bindings[0]["x"].Value, "r1"},
		{res.Results.Bindings[1]["x"].Value, "r2"},
		{res.Results.Bindings[0]["mbox"].Value, ""},
		{res.Results.Bindings[0]["blurb"].DataType, "http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral"},
		{res.Results.Bindings[1]["name"].Lang, "en"},
	}

	for i, tt := range resTests {
		if tt.in != tt.out {
			t.Errorf("%d) expected %v, got %v", i, tt.in, tt.out)
		}
	}

}

func TestResultBindings(t *testing.T) {
	b := res.Bindings()
	if len(b) != 6 { // not 7, becase "age" is not bound, even if present in head.vars
		t.Errorf("missing bound variables")
	}

	if l, ok := b["blurb"][0].(rdf.Literal); ok {
		if l.DataType != rdf.Uri("http://www.w3.org/1999/02/22-rdf-syntax-ns#XMLLiteral") {
			t.Errorf("just checkin'")
		}
	}
}

func TestResultSolutions(t *testing.T) {
	s := res.Solutions()

	if len(s) != 2 {
		t.Errorf("expected 2 solutions, got %d", len(s))
	}

	if s[1]["name"].String() != "Bob@en" {
		t.Errorf("just checkin'")
	}
}
