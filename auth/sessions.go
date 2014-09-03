package auth

import (
	"time"
)

// TODO should also provide the interface for sessions
type Session interface {
	Create()
	Delete(string)
	Exists(string) bool
	Load()
	Save()
}

type SessionBase struct {
	sessionExpiration time.Time
}

func (s *SessionBase) encode() string {
	// TODO accept any valid json?
	// TODO Should return a base64 encoded string
	return ""
}

func (s *SessionBase) decode(sessionInfo string) {
	// TODO return what? the json, the decoded session info in a struct?
}

func (s *SessionBase) getNewSessionKey() string {
	// TODO needs to check existance
	return GetRandomString(24, "abcdefghijklmnopqrstuvwxyz0123456789")
}

// TODO allow nil as a zero value?
func (s *SessionBase) setExpiration(d time.Duration) {
	s.sessionExpiration = time.Now().Add(d)
}

