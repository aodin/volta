Volta [![GoDoc](http://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/aodin/volta) [![Build Status](https://travis-ci.org/aodin/volta.svg)](https://travis-ci.org/aodin/volta)
=====

A library for building web applications in Go.


### Auth

Postgres-backed auth users, sessions, and tokens using [aspect](https://github.com/aodin/aspect). Fork it and modify the user to your needs!


### Config

Provides default implementations and JSON-writable configurations for server, database, email, and cookie settings.


### Router

Built upon Julien Schmidt's [httprouter](https://github.com/julienschmidt/httprouter). It adds a User and parameters directly to the request type. It also returns an optional error.


### Templates

Provides `Attrs` and `Templates` instances for simplifying the `html/template` package when using complex template definitions stored in nested directories with global and local attributes.
