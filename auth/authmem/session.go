package authmem

import (
	"fmt"
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/config"
	"sync"
	"time"
)

type SessionManager struct {
	mutex   sync.RWMutex
	byKey   map[string]*Session
	users   *UserManager
	cookie  config.CookieConfig
	keyFunc auth.KeyFunc
	nowFunc func() time.Time
}

// NewSession creates a new session using a key generated with the given user
// ID, auth.KeyFunc, and time generator
func (m *SessionManager) NewSession(id int64) (*Session, error) {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session := &Session{
		userID:  id,
		manager: m,
	}

	// Set the expires from the cookie config
	session.expires = m.nowFunc().Add(m.cookie.Age)

	// Generate a new session key
	var err error
	for {
		session.key, err = m.keyFunc()
		if err != nil {
			return session, err
		}
		if _, exists := m.byKey[session.key]; !exists {
			break
		}
	}
	m.byKey[session.key] = session
	return session, nil
}

// GetSession returns the session with the given key
func (m *SessionManager) GetSession(key string) (*Session, error) {
	// Lock the mutex for reading
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// Get the session at the given key
	session, exists := m.byKey[key]
	if !exists {
		return session, fmt.Errorf("authmem: no session with key %s exists", key)
	}
	return session, nil
}

// DeleteSession deletes the session with the given key.
func (m *SessionManager) DeleteSession(key string) error {
	// Lock the mutex for writing
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Delete the session at the given key
	delete(m.byKey, key)
	return nil
}

// NewSessionManager creates a new SessionManager with default implementations
// of keyFunc and nowFunc.
func NewSessionManager(c config.CookieConfig, u *UserManager) *SessionManager {
	// Use default functions
	return &SessionManager{
		byKey:   make(map[string]*Session),
		users:   u,
		cookie:  c,
		keyFunc: auth.RandomKey,
		nowFunc: time.Now,
	}
}

type Session struct {
	key     string
	userID  int64
	expires time.Time
	manager *SessionManager
}

// Key returns the session's key.
func (s *Session) Key() string {
	return s.key
}

// Expires returns the session's expiration.
func (s *Session) Expires() time.Time {
	return s.expires
}

// Users returns the User with the session's user id.
func (s *Session) User() (auth.User, error) {
	return s.manager.users.GetUser(s.userID)
}

func (s *Session) Delete() error {
	return s.manager.DeleteSession(s.key)
}
