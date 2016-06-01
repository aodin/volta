package auth

import (
	"net/http"

	"github.com/aodin/config"
)

// SetCookie writes the cookie to the given http.ResponseWriter.
// The cookie's name is taken from the cookie configuration and its value
// is the given session key.
func SetCookie(w http.ResponseWriter, c config.Cookie, session Session) {
	c.Set(w, session.Key, session.Expires)
}
