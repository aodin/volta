package auth

import (
	"fmt"
	"time"

	sql "github.com/aodin/aspect"

	"github.com/aodin/volta/config"
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
var Sessions = sql.Table("sessions",
	sql.Column("key", sql.String{NotNull: true}),
	sql.ForeignKey("user_id", Users.C["id"], sql.Integer{NotNull: true}),
	sql.Column("expires", sql.Timestamp{WithTimezone: true, NotNull: true}),
	sql.PrimaryKey("key"),
)

// SessionManager is the internal manager of sessions
type SessionManager struct {
	conn    sql.Connection
	cookie  config.CookieConfig
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
		stmt := sql.Select(
			Sessions.C["key"],
		).Where(Sessions.C["key"].Equals(session.Key)).Limit(1)
		if !m.conn.MustQueryOne(stmt, &duplicate) {
			break
		}
	}

	// Insert the session
	m.conn.MustExecute(Sessions.Insert().Values(session))
	return
}

// Delete removes the session with the given key from the database.
// It will return an error if the session key was not deleted from the
// database. It will panic on any connection error.
func (m *SessionManager) Delete(key string) error {
	stmt := Sessions.Delete().Where(Sessions.C["key"].Equals(key))
	rowsAffected, err := m.conn.MustExecute(stmt).RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: error during rows affected: %s", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf(
			"auth: session key %s was not deleted - it may not exist", key,
		)
	}
	return nil
}

// Get returns the session with the given key.
func (m *SessionManager) Get(key string) (session Session) {
	stmt := Sessions.Select().Where(Sessions.C["key"].Equals(key))
	m.conn.MustQueryOne(stmt, &session)
	return
}

// NewSessions will create a new internal session manager
func NewSessions(c config.CookieConfig, conn sql.Connection) *SessionManager {
	return &SessionManager{
		conn:    conn,
		cookie:  c,
		keyFunc: RandomKey,
		nowFunc: func() time.Time { return time.Now().In(time.UTC) },
	}
}
