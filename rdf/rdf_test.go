package rdf

import (
	"fmt"
	"testing"
)

func TestTermUri(t *testing.T) {
	var u1 Uri = "http://data.deichman.no/resource/tnr_823184"
	var u2 Uri = "http://data.deichman.no/person/x2041802300"
	var u3 = NewUri("http://data.deichman.no/resource/tnr_823184")

	if u1.Type() != URI {
		t.Errorf("Excepcted %v to be of type URI, got type %v", u1, u1.Type())
	}

	if u1.Eq(u2) {
		t.Errorf("Excepted %v to be unequal %v", u1, u2)
	}

	if !u1.Eq(u3) {
		t.Errorf("Excepted %v to be equal %v", u1, u2)
	}

	if fmt.Sprint(u1) != "<http://data.deichman.no/resource/tnr_823184>" {
		t.Errorf("Excepcted %v to be formatted as \"<http://data.deichman.no/resource/tnr_823184>\"", u1)
	}
}

func TestTermLiteral(t *testing.T) {
	l1, _ := NewLiteral(42)
	l2, _ := NewLiteral(42.00001)
	l3, _ := NewLiteral(true)
	l4, _ := NewLiteral(false)
	l5, _ := NewLiteral("fisk@nno")
	l6, _ := NewLiteral("fisk@no")
	l7, _ := NewLiteral("fisk")

	_, err := NewLiteral([]int{1, 2, 3})
	if err == nil {
		t.Errorf("Expected an error creating Literal, got nil")
	}

	var eqTests = []struct {
		a, b Literal
		res  bool
	}{
		{l1, l2, false},
		{l1, l3, false},
		{l3, l4, false},
		{l5, l6, false},
		{l6, l7, false},
	}

	for _, tt := range eqTests {
		if tt.a.Eq(tt.b) != tt.res {
			t.Errorf("Expected %v.Eq(%v) to be %v, got %v", tt.a, tt.b, tt.res, tt.a.Eq(tt.b))
		}
	}

	var formatTests = []struct {
		l Literal
		s string
	}{
		{l1, "42^^<http://www.w3.org/2001/XMLSchema#int>"},
		{l2, "42.00001^^<http://www.w3.org/2001/XMLSchema#float>"},
		{l3, "true^^<http://www.w3.org/2001/XMLSchema#boolean>"},
		{l4, "false^^<http://www.w3.org/2001/XMLSchema#boolean>"},
		{l5, "fisk@nno"},
		{l7, "fisk^^<http://www.w3.org/2001/XMLSchema#string>"},
	}

	for _, tt := range formatTests {
		if tt.l.String() != tt.s {
			t.Errorf("Expected formatting %v, got %v", tt.s, tt.l.String())
		}
	}
}

func TestTermBlankNode(t *testing.T) {
	b1 := Blank("p1")
	b2 := Blank("p2")
	b3 := Blank("p1")

	if b1.String() != "_:p1" {
		t.Errorf("Expected formatting \"_:p1\", got %v", b1)
	}

	if b1 == b2 {
		t.Errorf("Expected %v to be unequal %v", b1, b2)
	}

	if b1 != b3 {
		t.Errorf("Expected %v to be equal %v", b1, b3)
	}
}
