package volta

import (
	"log"
	"net/http"
	"strings"
)

// Wrap a request handler function with logging
// TODO Get response time, bytes written, status code returned?
func LogRequest(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
		Log(r)
	}
}

// TODO allow custom log formats
func Log(r *http.Request) {
	log.Printf(`"%s %s" %s "%s" "%s"`, r.Method, r.URL, strings.SplitN(r.RemoteAddr, ":", 2)[0], r.Header.Get("Referer"), r.Header.Get("User-Agent"))
}
