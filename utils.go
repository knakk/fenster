package main

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/knakk/fenster/rdf"
)

// prefixify returns the prefixed form of an URI if the prefix and namespaces
// is found in the prefixes array, which must have the following form:
// ["dc", "http://purl.org/dc/terms/"], ["foaf", "http://xmlns....etc"]
//
//It returns the string unmodified if no match is found.
func prefixify(prefixes *[][]string, uri string) string {
	uriOriginal := "" + uri

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
	included := make([]map[string]interface{}, 1)
	for _, m := range *rdfMap {
		if m[key] != nil {
			tm := make(map[string]interface{})
			for k, v := range m {
				if k != "g" && k != "p" && strings.HasPrefix(v.String(), "<"+conf.BaseURI) {
					link := fmt.Sprintf("<a href='/%v'>%v</a>", v.String()[25:len(v.String())-1], template.HTMLEscapeString(v.String()))
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
				return m["o"].Value()
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
