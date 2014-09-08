package auth

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/aodin/volta/config"
	"sync"
	"time"
)

type SessionManager interface {
	Create(user User) (Session, error)
	Get(key string) (Session, error)
	Delete(key string) error
}

// Session is an interface for sessions.
type Session interface {
	Key() string
	Expires() time.Time
	User() int64
	Delete() error
	// TODO Session data as JSON or map[string]interface{}?
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

// MemorySessions is an in-memory store of sessions.
type MemorySessions struct {
	mutex   sync.RWMutex
	byKey   map[string]*session
	cookie  config.CookieConfig
	keyFunc KeyFunc
	nowFunc func() time.Time
}

// Create creates a new session using a key generated for the given User
func (m *MemorySessions) Create(user User) (Session, error) {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	s := &session{
		userID:  user.ID(),
		manager: m,
	}

	// Set the expires from the cookie config
	s.expires = m.nowFunc().Add(m.cookie.Age)

	// Generate a new session key
	var err error
	for {
		s.key, err = m.keyFunc()
		if err != nil {
			return s, NewServerError("auth: key generation error: %s", err)
		}
		if _, exists := m.byKey[s.key]; !exists {
			break
		}
	}
	m.byKey[s.key] = s
	return s, nil
}

// Get returns the session with the given key
// Errors should only be returned on server error conditions (such as failed
// database connections). If no session is found, a zero-initialized session
// should be returned instead of an error.
func (m *MemorySessions) Get(key string) (Session, error) {
	// Lock the mutex for reading
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get the session at the given key
	session, ok := m.byKey[key]
	if !ok {
		return session, NewUserError("auth: no session key %s exists", key)
	}
	return session, nil
}

// Delete deletes the session with the given key.
func (m *MemorySessions) Delete(key string) error {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Delete the session at the given key
	delete(m.byKey, key)
	return nil
}

// SessionsInMemory creates a new MemorySessions with default implementations
// of keyFunc and nowFunc.
func SessionsInMemory(c config.CookieConfig) *MemorySessions {
	return &MemorySessions{
		byKey:   make(map[string]*session),
		cookie:  c,
		keyFunc: RandomKey,
		nowFunc: time.Now,
	}
}

type session struct {
	key     string
	userID  int64
	expires time.Time
	manager *MemorySessions
}

// Key returns the session's key.
func (s *session) Key() string {
	return s.key
}

// Expires returns the session's expiration.
func (s *session) Expires() time.Time {
	return s.expires
}

// User returns UserID of the session.
func (s *session) User() int64 {
	return s.userID
}

func (s *session) Delete() error {
	return s.manager.Delete(s.key)
}