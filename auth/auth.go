package auth

import (
	"github.com/aodin/volta/config"
	"net/http"
	"time"
)

// TODO distinguish between errors and nil responses
func Login(w http.ResponseWriter, username, password string, sessions SessionManager, users UserManager, c config.CookieConfig, loginURL string) bool {
	return false
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
