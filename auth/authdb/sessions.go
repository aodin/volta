package authdb

import (
	"fmt"
	"github.com/aodin/aspect"
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/config"
	"time"
)

type SessionManager struct {
	db      *aspect.DB
	cookie  config.CookieConfig
	keyFunc auth.KeyFunc
	nowFunc func() time.Time
}

// Create creates a new session using a key generated for the given User
func (m *SessionManager) Create(user auth.User) (auth.Session, error) {
	s := &Session{
		UserField: user.ID(),
	}

	// Set the expires from the cookie config
	s.ExpiresField = m.nowFunc().Add(m.cookie.Age)

	// Generate a new session key
	var err error
	for {
		s.KeyField, err = m.keyFunc()
		if err != nil {
			return s, auth.NewServerError("authdb: key generation error: %s", err)
		}
		// TODO distinguish between no response and an improper query
		var sessions []Session
		stmt := Sessions.Select().Where(Sessions.C["key"].Equals(s.KeyField)).Limit(2)
		if err = m.db.QueryAll(stmt, &sessions); err != nil {
			return s, auth.NewServerError("authdb: error checking if key exists: %s", err)
		}
		if len(sessions) == 0 {
			break
		}
	}

	if _, err = m.db.Execute(Sessions.Insert(s)); err != nil {
		return s, err
	}

	return s, nil
}

// Get returns the session with the given key
// Errors should only be returned on server error conditions (such as failed
// database connections).
func (m *SessionManager) Get(key string) (auth.Session, error) {
	// Get the user at the given name
	var session Session
	stmt := Sessions.Select().Where(Sessions.C["key"].Equals(key))
	if err := m.db.QueryOne(stmt, &session); err != nil {
		// TODO distinguish between no response and an improper query
		return &session, auth.NewUserError("no session with key %s exists", key)
	}
	return &session, nil
}

// Delete deletes the session with the given key.
func (m *SessionManager) Delete(key string) error {
	_, err := m.db.Execute(Sessions.Delete().Where(Sessions.C["key"].Equals(key)))
	if err != nil {
		// TODO When is it a server error and when is it a user error?
		err = fmt.Errorf(
			"authdb: error deleting session with key %s: %s",
			key,
			err,
		)
	}
	return err
}

func NewSessionManager(db *aspect.DB, c config.CookieConfig) *SessionManager {
	return &SessionManager{
		db:      db,
		cookie:  c,
		keyFunc: auth.RandomKey,
		nowFunc: time.Now,
	}
}

// Session is a database-backed session that implements the volta auth.Session
// interface.
type Session struct {
	KeyField     string    `db:"key"`
	UserField    int64     `db:"user_id"`
	ExpiresField time.Time `db:"expires"`
}

// Key returns the session's key.
func (s *Session) Key() string {
	return s.KeyField
}

// Expires returns the session's expiration.
func (s *Session) Expires() time.Time {
	return s.ExpiresField
}

// User returns UserID of the session.
func (s *Session) User() int64 {
	return s.UserField
}

func (s *Session) Delete() error {
	return auth.NewServerError("authdb: delete sessions through the session manager")
}

var Sessions = aspect.Table("sessions",
	aspect.Column("key", aspect.String{}),
	aspect.Column("user_id", aspect.Integer{}),
	aspect.Column("expires", aspect.Timestamp{WithTimezone: true}),
	aspect.PrimaryKey("key"),
)
