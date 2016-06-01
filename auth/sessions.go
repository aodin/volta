package auth

import (
	"fmt"
	"time"

	"github.com/aodin/config"
	"github.com/aodin/sol"
	"github.com/aodin/sol/postgres"
	"github.com/aodin/sol/types"
)

// Session is a database-backed user session.
type Session struct {
	Key     string          `db:"key"`
	UserID  int64           `db:"user_id"`
	Expires time.Time       `db:"expires"`
	manager *SessionManager `db:"-"`
}

// Delete removes the session with the given key from the database.
// It will return an error if the session does not have a key or the key
// was not deleted from the database. It will panic on any connection error.
func (session Session) Delete() error {
	if !session.Exists() {
		return fmt.Errorf("auth: keyless sessions cannot be deleted")
	}
	return session.manager.Delete(session.Key)
}

// Exists returns true if the session exists
func (session Session) Exists() bool {
	return session.Key != ""
}

// Sessions is the postgres schema for sessions
var Sessions = postgres.Table("sessions",
	sol.Column("key", types.Varchar().NotNull()),
	sol.ForeignKey(
		"user_id",
		Users.C("id"),
		types.Integer().NotNull(),
	).OnDelete(sol.Cascade).OnUpdate(sol.Cascade),
	sol.Column("expires", postgres.Timestamp().WithTimezone()),
	sol.PrimaryKey("key"),
)

// SessionManager is the internal manager of sessions
type SessionManager struct {
	conn    sol.Conn
	cookie  config.Cookie
	keyFunc KeyFunc
	nowFunc func() time.Time
}

// Create creates a new session using a key generated for the given User
func (m *SessionManager) Create(user User) (session Session) {
	// Set the expires from the cookie config
	session = Session{
		Expires: m.nowFunc().Add(m.cookie.Age),
		UserID:  user.ID,
		manager: m,
	}

	// Generate a new random session key
	for {
		session.Key = m.keyFunc()

		// No duplicates - generate a new key if this key already exists
		var duplicate string
		stmt := sol.Select(
			Sessions.C("key"),
		).Where(Sessions.C("key").Equals(session.Key)).Limit(1)
		m.conn.Query(stmt, &duplicate)
		if duplicate == "" {
			break
		}
	}

	// Insert the session
	m.conn.Query(Sessions.Insert().Values(session))
	return
}

// Delete removes the session with the given key from the database.
func (m *SessionManager) Delete(key string) error {
	stmt := Sessions.Delete().Where(Sessions.C("key").Equals(key))
	return m.conn.Query(stmt)
}

// Get returns the session with the given key.
func (m *SessionManager) Get(key string) (session Session) {
	stmt := Sessions.Select().Where(Sessions.C("key").Equals(key))
	m.conn.Query(stmt, &session)
	return
}

// NewSessions will create a new internal session manager
func NewSessions(c config.Cookie, conn sol.Conn) *SessionManager {
	return &SessionManager{
		conn:    conn,
		cookie:  c,
		keyFunc: RandomKey,
		nowFunc: func() time.Time { return time.Now().In(time.UTC) },
	}
}
