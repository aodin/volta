package router

import (
	"fmt"
	"log"
	"net/http"
	"runtime"

	"github.com/aodin/volta/auth"
)

// The router is built on: https://github.com/julienschmidt/httprouter
// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.
type Method string

const (
	GET     Method = "GET"
	POST    Method = "POST"
	PUT     Method = "PUT"
	DELETE  Method = "DELETE"
	PATCH   Method = "PATCH"
	OPTIONS Method = "OPTIONS"
)

var All = []Method{GET, POST, PUT, DELETE, PATCH, OPTIONS}

// Router is a http.Handler which can be used to dispatch requests to different
// handler functions via configurable routes
type Router struct {
	trees                 map[Method]*node
	RedirectTrailingSlash bool
	RedirectFixedPath     bool
	auth                  *auth.Auth
}

// Route attaches the given handler on the path for every given method.
func (r *Router) Route(path string, h Handler, methods ...Method) {
	if path[0] != '/' {
		panic("path must begin with '/'")
	}
	if r.trees == nil {
		r.trees = make(map[Method]*node)
	}
	for _, method := range methods {
		root := r.trees[method]
		if root == nil {
			root = new(node)
			r.trees[method] = root
		}
		root.addRoute(path, h)
	}
}

// Handler adapts the given http.Handler so it can be served for every
// given method by the router.
func (r *Router) Handler(path string, h http.Handler, ms ...Method) {
	r.Route(
		path,
		func(w http.ResponseWriter, req *Request) error {
			h.ServeHTTP(w, req.Request)
			return nil
		},
		ms...,
	)
}

// HandleFunc adapts the given http.HandleFunc so it can be served for every
// given method by the router.
func (r *Router) HandleFunc(path string, h http.HandlerFunc, ms ...Method) {
	r.Route(
		path,
		func(w http.ResponseWriter, req *Request) error {
			h(w, req.Request)
			return nil
		},
		ms...,
	)
}

// GET attaches the given handler on the path for the GET method only.
func (r *Router) GET(path string, h Handler) {
	r.Route(path, h, GET)
}

// POST attaches the given handler on the path for the POST method only.
func (r *Router) POST(path string, h Handler) {
	r.Route(path, h, POST)
}

// PATCH attaches the given handler on the path for the PATCH method only.
func (r *Router) PATCH(path string, h Handler) {
	r.Route(path, h, PATCH)
}

// PUT attaches the given handler on the path for the PUT method only.
func (r *Router) PUT(path string, h Handler) {
	r.Route(path, h, PUT)
}

// DELETE attaches the given handler on the path for the DELETE method only.
func (r *Router) DELETE(path string, h Handler) {
	r.Route(path, h, DELETE)
}

// ServeFiles serves files from the given file system path.
func (r *Router) ServeFiles(path string, root http.FileSystem) {
	if len(path) < 10 || path[len(path)-10:] != "/*filepath" {
		panic("path must end with /*filepath")
	}

	fileServer := http.FileServer(root)

	var h Handler = func(w http.ResponseWriter, req *Request) error {
		req.URL.Path = req.Params.ByName("filepath")
		fileServer.ServeHTTP(w, req.Request)
		return nil
	}
	r.GET(path, h)
}

// Lookup allows the manual lookup of a method + path combo.
func (r *Router) Lookup(method Method, path string) (Handler, Params, bool) {
	if root := r.trees[method]; root != nil {
		return root.getValue(path)
	}
	return nil, nil, false
}

// ServeHTTP dispatches the request to the handler whose pattern and method
// matches the request. It will also get an auth.User if the session is
// valid, pass context to the matched handler, and generate any
// parameters requested by the route.
func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Recover from panic
	defer func() {
		if panicked := recover(); panicked != nil {
			// Generate a stack trace
			buf := make([]byte, 1<<16)

			// Set false to dump only this goroutine
			l := runtime.Stack(buf, false)
			log.Printf("%s\n\n%s", panicked, buf[:l])

			// TODO Only display an error in DEBUG
			http.Error(w, fmt.Sprintf("%s", panicked), 500)
		}
	}()

	// Build a new request and attach a user - if there is no valid user
	// the request.User will be an auth.AnonUser
	request := NewRequest(req, router.auth)

	// Record if a handler ran, so if false, a 404 page can be served
	var ranHandler bool
	var err error

	// Determine routes
	if root := router.trees[Method(req.Method)]; root != nil {
		path := req.URL.Path

		if h, ps, tsr := root.getValue(path); h != nil {
			// Add the parameters to the request
			request.Params = ps

			err = h(w, request)
			ranHandler = true
		} else if req.Method != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if req.Method != string(GET) {
				// Temporary redirect, request with same method
				// As of Go 1.3, Go does not support status code 308.
				code = 307
			}

			if tsr && router.RedirectTrailingSlash {
				if path[len(path)-1] == '/' {
					req.URL.Path = path[:len(path)-1]
				} else {
					req.URL.Path = path + "/"
				}
				http.Redirect(w, req, req.URL.String(), code)
				return
			}

			// Try to fix the request path
			if router.RedirectFixedPath {
				fixedPath, found := root.findCaseInsensitivePath(
					CleanPath(path),
					router.RedirectTrailingSlash,
				)
				if found {
					req.URL.Path = string(fixedPath)
					http.Redirect(w, req, req.URL.String(), code)
					return
				}
			}
		}
	}

	// Handle any other user errors
	// TODO Other codes?
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if ranHandler {
		return
	}
	http.NotFound(w, req)
}

func newMockRouter() *Router {
	return New(nil)
}

// New creates a new router instance using the given context
func New(auth *auth.Auth) *Router {
	return &Router{
		RedirectTrailingSlash: true,
		RedirectFixedPath:     true,
		auth:                  auth,
	}
}

// Make sure the Router conforms with the http.Handler interface
var _ http.Handler = newMockRouter()
