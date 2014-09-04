package auth

import (
	"github.com/aodin/volta/config"
	"net/http"
)

// SetCookie writes the cookie to the given http.ResponseWriter.
// The cookie's name is taken from the cookie configuration and its value
// is the given session key.
func SetCookie(w http.ResponseWriter, config config.CookieConfig, session Session) {
	cookie := &http.Cookie{
		Name:     config.Name,
		Value:    session.Key(),
		Path:     config.Path,
		Domain:   config.Domain,
		Expires:  session.Expires(),
		HttpOnly: config.HttpOnly,
		Secure:   config.Secure,
	}
	http.SetCookie(w, cookie)
}
