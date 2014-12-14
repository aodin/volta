package router

import (
	"net/http"
	"net/url"

	"github.com/aodin/volta/auth"
)

type Request struct {
	*http.Request
	Params Params
	User   auth.User
	Values url.Values
}

// Get gets a GET parameter and ONLY a get parameter - never POST form data
func (r *Request) Get(key string) string {
	return r.QueryValues().Get(key)
}

// TODO This function is messy and doesn't make much sense - prefer mocks
// over tricks with context for testing
func (r *Request) QueryValues() url.Values {
	if r.Values == nil {
		if r.Request != nil {
			r.Values = r.Request.URL.Query()
		} else {
			r.Values = url.Values{}
		}
	}
	return r.Values
}

// SetPath allows the URL Path to be easily set - useful for testing.
func (r *Request) SetPath(path string) {
	if r.URL == nil {
		r.URL = &url.URL{}
	}
	r.URL.Path = path
}

// NewRequest wraps the http.Request and adds an auth.User if valid
func NewRequest(r *http.Request, auth *auth.Auth) (request *Request) {
	request = &Request{
		Request: r,
	}

	// For testing, if auth is nil, just return here
	if auth == nil {
		return
	}

	// Cookie will return an ErrNoCookie if not found
	cookie, err := r.Cookie(auth.CookieName())
	if err != nil {
		return
	}

	request.User = auth.BySession(cookie.Value)

	// Do not perform authentication by tokens here - tokens are only good
	// for the API
	return
}
