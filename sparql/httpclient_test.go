package sparql

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type delayedHandler struct{}

func (h delayedHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	time.Sleep(300 * time.Millisecond)
	io.WriteString(w, "Too late?")
}

func TestHttpClientTimeout(t *testing.T) {
	handler := &delayedHandler{}
	server := httptest.NewServer(handler)
	defer server.Close()

	openTimeout := 250 * time.Millisecond
	readTimeout := 250 * time.Millisecond

	httpClient := newTimeoutClient(openTimeout, readTimeout)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = httpClient.Do(req)
	if err == nil {
		t.Fatalf("Request should have timed out.")
	}

	openTimeout = 350 * time.Millisecond
	readTimeout = 350 * time.Millisecond

	httpClient = newTimeoutClient(openTimeout, readTimeout)

	_, err = httpClient.Do(req)
	if err != nil {
		t.Fatalf("Request should not have timed out.")
	}

}
