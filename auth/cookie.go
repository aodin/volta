package auth

import (
    "net/http"
    "github.com/aodin/volta/config"
)

// Include the cookie on the response
// The cookie's name is taken from the cookie configuration and its value
// is the given session key.
func SetCookie(w http.ResponseWriter, config config.CookieConfig, session Session) {
    // Create the cookie
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
