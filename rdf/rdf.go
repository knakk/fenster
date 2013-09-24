package rdf

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type TermType int

// RDF Term types
const (
	BLANK TermType = iota
	LITERAL
	URI
)

// Term is the common interface for RDF Terms
type Term interface {
	String() string
	Type() TermType
	Eq(Term) bool
	Value() interface{}
}

// Uri RDF Term
type Uri string

// String returns the string representation of an Uri
func (u Uri) String() string {
	return "<" + string(u) + ">"
}

// Type returns the RDF Term type (URI)
func (u Uri) Type() TermType {
	return URI
}

// Value returns the Uri as value, i.e the Uri string without the enclosing < >
func (u Uri) Value() interface{} {
	return string(u)
}

// Eq tests the equality of an Uri against another RDF Term
func (u Uri) Eq(t Term) bool {
	if u.Type() == t.Type() {
		if u.String() == t.String() {
			return true
		}
	}
	return false
}

// NewUri returns a new Uri.
// Equivalent of var x Uri = "{{uri}}"
func NewUri(s string) Uri {
	return Uri(s)
}

// Literal RDF Type
type Literal struct {
	Lang     string
	Val      interface{}
	DataType Uri
}

// String returns the string representation of a Literal
func (l Literal) String() string {
	if l.Lang != "" {
		return fmt.Sprintf("%v@%s", l.Val, l.Lang)
	}
	if l.DataType != "" {
		return fmt.Sprintf("%v^^%v", l.Val, l.DataType)
	}
	return fmt.Sprintf("%v", l.Val)
}

// Type returns the RDF Term type (LITERAL)
func (l Literal) Type() TermType {
	return LITERAL
}

// Value returns the Literal value, in the corresponding go type, i.e.
// xsd:integer -> int, xsd:float -> float64 and so on.
//
// For language-tagged literals, it returns the string without the trailing
// language tag.
func (l Literal) Value() interface{} {
	return l.Val
}

// Eq tests the equality of a Literal against another RDF Term
func (l Literal) Eq(t Term) bool {
	if l.Type() == t.Type() {
		if l.String() == t.String() {
			return true
		}
	}
	return false
}

// NewLiteral is a Literal constructor.
// It sets the DataType based on it's input, or returns an error if it fails to
// guess the DataType.
// Suffix with @ to create a language tagged literal, ex: NewLiteral("string@en")
// If Lang is provided, DataType is ignored.
// If you need a custom DataType, you must create the literal with the normal
// struct syntax, ex:
// l := Literal{Val: "my-val", DataType: Uri("my-custom-type")}
func NewLiteral(v interface{}) (Literal, error) {
	switch t := v.(type) {
	default:
		return Literal{}, fmt.Errorf("cannot infer xsd:datatype from %v", t)
	case bool:
		return Literal{Val: t, DataType: Uri("http://www.w3.org/2001/XMLSchema#boolean")}, nil
	case int:
		return Literal{Val: t, DataType: Uri("http://www.w3.org/2001/XMLSchema#int")}, nil
	case string:
		r := regexp.MustCompile(`@[a-z]{2,3}$`)
		if r.MatchString(t) {
			// matches a iso639-2 language code:
			// http://www.loc.gov/standards/iso639-2/php/code_list.php
			tag := r.FindString(t)
			return Literal{Val: strings.Split(t, tag)[0], Lang: tag[1:]}, nil
		}
		return Literal{Val: t, DataType: Uri("http://www.w3.org/2001/XMLSchema#string")}, nil
	case float64:
		return Literal{Val: t, DataType: Uri("http://www.w3.org/2001/XMLSchema#float")}, nil
	case time.Time:
		return Literal{}, fmt.Errorf("not implemented")
		/*    ti, err := time.Parse(time.RFC3339, t)
		      if err == nil {
		        // matches a RFC3339 date, assuming xsd:dateTime
		        return Literal{Val: ti, DataType: Uri("http://www.w3.org/2001/XMLSchema#dateTime")}, nil
		      }*/
	}
}

// Blank Node RDF Term.
//
// Note: It is left to the consumer to enforce a globaly unique identifier.
type Blank string

// String returns the string representation of a blank node
func (b Blank) String() string {
	return "_:" + string(b)
}

// Type returns the RDF Term type (BLANK)
func (b Blank) Type() TermType {
	return BLANK
}

// Value returns the value of a Blank Node (which is nil)
func (b Blank) Value() interface{} {
	return nil
	// or empty string?
}

// Eq tests the equality of a Blank node against another RDF Term
func (b Blank) Eq(t Term) bool {
	if b.Type() == t.Type() {
		if b.String() == t.String() {
			return true
		}
	}
	return false
}

// NewBlank returns a new Blank Node.
func NewBlank(s string) Blank {
	if s == "" {
		return Blank(fmt.Sprint(&s))
	}
	return Blank(s)
}

// Triple represents a RDF statement.
type Triple struct {
	Subject, Predicate, Object Term
}

// Quad represents a RDF statement within a specified graph.
type Quad struct {
	Graph                      Uri
	Subject, Predicate, Object Term
}

// TODO serialization?
// At least Turtle & N3

/*
type TripleGraph []Triple

type QuadGraph []Quad

*/
