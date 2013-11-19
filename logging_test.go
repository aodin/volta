package volta

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogRequest(t *testing.T) {
	// Start a basic test server
	hello := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	}
	ts := httptest.NewServer(http.HandlerFunc(LogRequest(hello)))
	defer ts.Close()

	// Log to a buffer
	buffer := &bytes.Buffer{}
	log.SetOutput(buffer)

	// Send a request
	_, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// TODO Check the actual content of the logged request
	// Check the content of the logged request, minus the timestamp
	actual := buffer.String()[20:len(buffer.String()) - 1]
	expected := `"GET /" 127.0.0.1 "" "Go 1.1 package http"`
	if actual != expected {
		t.Errorf("Unexpected log entry: %s != %s", actual, expected)
	}
}