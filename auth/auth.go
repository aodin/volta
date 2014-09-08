package auth

import (
	"github.com/aodin/volta/config"
	"net/http"
	"time"
)

// TODO distinguish between errors and nil responses
// TODO An interfaces the implements all these required parameters
func Login(w http.ResponseWriter, r *http.Request, username, password string, sessions SessionManager, users UserManager, hasher Hasher, c config.CookieConfig, loginURL string) (bool, error) {
	// Get the requested user
	user, err := users.Get(Fields{"name": username})
	if err != nil {
		return IgnoreUserErrors(w, r, err), err
	}

	// TODO hasher could be obtained from the password string
	if !CheckPassword(hasher, password, user.Password()) {
		return IgnoreUserErrors(w, r, err), err
	}

	// Create a new session
	session, err := sessions.Create(user)
	if err != nil {
		return IgnoreUserErrors(w, r, err), err
	}

	SetCookie(w, c, session)
	http.Redirect(w, r, loginURL, 302)
	return true, nil
}

func GetUserIfValidSession(sessions SessionManager, users UserManager, key string) User {
	return getUserIfValidSession(sessions, users, key, time.Now)
}

func getUserIfValidSession(sessions SessionManager, users UserManager, key string, nowFunc func() time.Time) User {
	session, err := sessions.Get(key)
	if err != nil {
		return nil
	}
	if !session.Expires().After(nowFunc()) {
		return nil
	}
	user, err := users.Get(Fields{"id": session.User()})
	if err != nil {
		return nil
	}
	return user
}

// IsValidSession checks if a session key exists in the given manager.
func IsValidSession(m SessionManager, key string) bool {
	return isValidSession(m, key, time.Now)
}

func isValidSession(m SessionManager, key string, nowFunc func() time.Time) bool {
	session, err := m.Get(key)
	if err != nil {
		return false
	}
	return session.Expires().After(nowFunc())
}
