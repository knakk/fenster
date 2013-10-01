package main

import "testing"

func TestPrefixify(t *testing.T) {
	prefixes := [][]string{
		{"dc", "http://purl.org/dc/terms/"},
		{"foaf", "http://xmlns.com/foaf/0.1/"}}

	var (
		u1 = "<http://purl.org/dc/terms/title>"
		u2 = "http://xmlns.com/foaf/0.1/name"
		u3 = "<http://purl.org/dc/terms/deeper/path>"
	)

	var resTests = []struct {
		in  interface{}
		out interface{}
	}{
		{prefixify(&prefixes, u1), "dc:title"},
		{prefixify(&prefixes, u2), "foaf:name"},
		{prefixify(&prefixes, u3), u3},
	}

	for i, tt := range resTests {
		if tt.in != tt.out {
			t.Errorf("%d) expected %v, got %v", i, tt.out, tt.in)
		}
	}
}
