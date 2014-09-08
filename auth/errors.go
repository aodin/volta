package auth

import (
	"fmt"
	"log"
	"net/http"
)

// ServerError is an error that should result in a 5xx class HTTP code
type ServerError struct {
	err  error
	Code int
}

func (e ServerError) Error() string {
	return e.err.Error()
}

func NewServerError(s string, args ...interface{}) ServerError {
	return ServerError{
		err: fmt.Errorf(s, args...),
	}
}

// UserError is an error that should result in a 4xx class HTTP code
type UserError struct {
	err  error
	Code int
}

func (e UserError) Error() string {
	return e.err.Error()
}

func NewUserError(s string, args ...interface{}) UserError {
	return UserError{
		err: fmt.Errorf(s, args...),
	}
}

// HandleErrorHTTP writes the 404 or 500 page to the response and returns
// a boolean indicating if the response should be aborted.
func HandleErrorHTTP(w http.ResponseWriter, r *http.Request, err error) bool {
	switch err.(type) {
	case ServerError:
		// Log and return a 500
		// TODO Ignore the code for now
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return true
	case UserError:
		// 404
		http.NotFound(w, r)
		return true
	}
	return false
}
