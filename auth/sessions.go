package auth

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// Session is an interface for sessions.
// TODO Session data as JSON or map[string]interface{}?
type Session interface {
	Key() string
	Expires() time.Time
	User() (User, error)
	Delete() error
}

// RandomKey generates a new 144 bit session key. It does so by producing 18
// random bytes that are encoded in URL safe base64, for output of 24 chars.
func RandomKey() (string, error) {
	b := make([]byte, 18)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// KeyFunc is the function type that will be used to generate new session keys.
type KeyFunc func() (string, error)
