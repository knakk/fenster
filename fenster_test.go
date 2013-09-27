package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestAppnameAndVersionInPage(t *testing.T) {
	handler := &mainHandler{}
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/resource/tnr_1140686")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Got non-200 response: %d\n", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !regexp.MustCompile(`Fenster version (\d\.)+`).MatchString(string(body)) {
		t.Error("Application name and version missing.")
	}
}
