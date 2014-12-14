// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockResponseWriter struct {
	code int
}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func TestRouter(t *testing.T) {
	router := newMockRouter()
	routed := false

	h := func(w http.ResponseWriter, r *Request) error {
		routed = true
		want := Params{Param{"name", "gopher"}}
		// TODO testify
		if !reflect.DeepEqual(r.Params, want) {
			t.Fatalf("wrong wildcard values: want %v, got %v", want, r.Params)
		}
		return nil
	}

	router.Route("/user/:name", h, GET)

	w := new(mockResponseWriter)

	req, _ := http.NewRequest("GET", "/user/gopher", nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}
}

func TestContext(t *testing.T) {
	// Also test the router with a mock context to trigger the auth
	router := newMockRouter()
	routed := false

	h := func(w http.ResponseWriter, r *Request) error {
		routed = true
		return nil
	}

	router.GET("/signin", h)

	w := new(mockResponseWriter)

	req, _ := http.NewRequest("GET", "/signin", nil)
	router.ServeHTTP(w, req)

	if !routed {
		t.Fatal("routing failed")
	}

	// // Cause a 500
	// router.GET("/500", func(w http.ResponseWriter, r *Request) error {
	// 	panic("Halt, and catch fire")
	// 	return nil
	// })

	// // Recover from a 500
	// r, _ := http.NewRequest("GET", "/500", nil)
	// router.ServeHTTP(w, r)
}

type handlerStruct struct {
	handeled *bool
}

func (h handlerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	*h.handeled = true
}

func TestRouterAPI(t *testing.T) {
	var get, post, put, patch, delete, handler, handleFunc bool
	httpHandler := handlerStruct{&handler}

	router := newMockRouter()
	router.GET("/GET", func(w http.ResponseWriter, r *Request) error {
		get = true
		return nil
	})
	router.POST("/POST", func(w http.ResponseWriter, r *Request) error {
		post = true
		return nil
	})
	router.PUT("/PUT", func(w http.ResponseWriter, r *Request) error {
		put = true
		return nil
	})
	router.PATCH("/PATCH", func(w http.ResponseWriter, r *Request) error {
		patch = true
		return nil
	})
	router.DELETE("/DELETE", func(w http.ResponseWriter, r *Request) error {
		delete = true
		return nil
	})

	// Return a 400 error
	router.GET("/400", func(w http.ResponseWriter, r *Request) error {
		return fmt.Errorf("I AM AN ERROR")
	})

	router.Handler("/Handler", httpHandler, GET)

	router.HandleFunc("/HandleFunc", func(w http.ResponseWriter, r *http.Request) {
		handleFunc = true
	}, GET)

	w := new(mockResponseWriter)

	r, _ := http.NewRequest("GET", "/GET", nil)
	router.ServeHTTP(w, r)
	if !get {
		t.Error("routing GET failed")
	}

	r, _ = http.NewRequest("POST", "/POST", nil)
	router.ServeHTTP(w, r)
	if !post {
		t.Error("routing POST failed")
	}

	r, _ = http.NewRequest("PUT", "/PUT", nil)
	router.ServeHTTP(w, r)
	if !put {
		t.Error("routing PUT failed")
	}

	r, _ = http.NewRequest("PATCH", "/PATCH", nil)
	router.ServeHTTP(w, r)
	if !patch {
		t.Error("routing PATCH failed")
	}

	r, _ = http.NewRequest("DELETE", "/DELETE", nil)
	router.ServeHTTP(w, r)
	if !delete {
		t.Error("routing DELETE failed")
	}

	// Test that a 400 status was written
	w = new(mockResponseWriter)
	r, _ = http.NewRequest("GET", "/400", nil)
	router.ServeHTTP(w, r)
	if w.code != 400 {
		t.Error("router: unexpected status code written by error")
	}

	r, _ = http.NewRequest("GET", "/Handler", nil)
	router.ServeHTTP(w, r)
	if !handler {
		t.Error("routing Handler failed")
	}

	r, _ = http.NewRequest("GET", "/HandleFunc", nil)
	router.ServeHTTP(w, r)
	if !handleFunc {
		t.Error("routing HandleFunc failed")
	}
}

func TestRouterRoot(t *testing.T) {
	router := newMockRouter()
	recv := catchPanic(func() {
		router.GET("noSlashRoot", nil)
	})
	if recv == nil {
		t.Fatal("registering path not beginning with '/' did not panic")
	}
}

func TestRouterNotFound(t *testing.T) {
	handlerFunc := func(_ http.ResponseWriter, _ *Request) error {
		return nil
	}

	router := newMockRouter()
	router.GET("/path", handlerFunc)
	router.GET("/dir/", handlerFunc)

	testRoutes := []struct {
		route  string
		code   int
		header string
	}{
		{"/path/", 301, "map[Location:[/path]]"},   // TSR -/
		{"/dir", 301, "map[Location:[/dir/]]"},     // TSR +/
		{"/PATH", 301, "map[Location:[/path]]"},    // Fixed Case
		{"/DIR/", 301, "map[Location:[/dir/]]"},    // Fixed Case
		{"/PATH/", 301, "map[Location:[/path]]"},   // Fixed Case -/
		{"/DIR", 301, "map[Location:[/dir/]]"},     // Fixed Case +/
		{"/../path", 301, "map[Location:[/path]]"}, // CleanPath
		{"/nope", 404, ""},                         // NotFound
	}
	for _, tr := range testRoutes {
		r, _ := http.NewRequest("GET", tr.route, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if !(w.Code == tr.code && (w.Code == 404 || fmt.Sprint(w.Header()) == tr.header)) {
			t.Errorf("NotFound handling route %s failed: Code=%d, Header=%v", tr.route, w.Code, w.Header())
		}
	}

	r, _ := http.NewRequest("GET", "/nope", nil)
	w := httptest.NewRecorder()

	// Test other method than GET (want 307 instead of 301)
	router.PATCH("/path", handlerFunc)
	r, _ = http.NewRequest("PATCH", "/path/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == 307 && fmt.Sprint(w.Header()) == "map[Location:[/path]]") {
		t.Errorf("Custom NotFound handler failed: Code=%d, Header=%v", w.Code, w.Header())
	}

	// Test special case where no node for the prefix "/" exists
	router = newMockRouter()
	router.GET("/a", handlerFunc)
	r, _ = http.NewRequest("GET", "/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if !(w.Code == 404) {
		t.Errorf("NotFound handling route / failed: Code=%d", w.Code)
	}
}

func TestRouterLookup(t *testing.T) {
	routed := false
	wantHandle := func(_ http.ResponseWriter, _ *Request) error {
		routed = true
		return nil
	}
	wantParams := Params{Param{"name", "gopher"}}

	router := newMockRouter()

	// try empty router first
	handle, _, tsr := router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}

	// insert route and try again
	router.GET("/user/:name", wantHandle)

	handle, params, tsr := router.Lookup("GET", "/user/gopher")
	if handle == nil {
		t.Fatal("Got no handle!")
	} else {
		handle(nil, nil)
		if !routed {
			t.Fatal("Routing failed!")
		}
	}

	if !reflect.DeepEqual(params, wantParams) {
		t.Fatalf("Wrong parameter values: want %v, got %v", wantParams, params)
	}

	handle, _, tsr = router.Lookup("GET", "/user/gopher/")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if !tsr {
		t.Error("Got no TSR recommendation!")
	}

	handle, _, tsr = router.Lookup("GET", "/nope")
	if handle != nil {
		t.Fatalf("Got handle for unregistered pattern: %v", handle)
	}
	if tsr {
		t.Error("Got wrong TSR recommendation!")
	}
}

type mockFileSystem struct {
	opened bool
}

func (mfs *mockFileSystem) Open(name string) (http.File, error) {
	mfs.opened = true
	return nil, errors.New("this is just a mock")
}

func TestRouterServeFiles(t *testing.T) {
	router := newMockRouter()
	mfs := &mockFileSystem{}

	recv := catchPanic(func() {
		router.ServeFiles("/noFilepath", mfs)
	})
	if recv == nil {
		t.Fatal("registering path not ending with '*filepath' did not panic")
	}

	router.ServeFiles("/*filepath", mfs)
	w := new(mockResponseWriter)
	r, _ := http.NewRequest("GET", "/favicon.ico", nil)
	router.ServeHTTP(w, r)
	if !mfs.opened {
		t.Error("serving file failed")
	}
}
