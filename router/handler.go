package router

import (
	"net/http"
)

type Handler func(http.ResponseWriter, *Request) error
