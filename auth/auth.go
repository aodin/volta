package auth

import (
	"time"
)

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
	user, err := session.User()
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
