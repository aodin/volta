package auth

import (
	"fmt"
	"time"

	sql "github.com/aodin/aspect"
	"github.com/aodin/aspect/postgres"
)

// Token is a database-backed user API token. The Expires field is nil if
// the token never expires.
type Token struct {
	Key       string        `db:"key"`
	UserID    int64         `db:"user_id"`
	Expires   *time.Time    `db:"expires"`
	CreatedAt time.Time     `db:"created_at,omitempty"`
	manager   *TokenManager `db:"-"`
}

// Delete removes the token with the given key from the database.
// It will return an error if the token does not have a key or the key
// was not deleted from the database. It will panic on any connection error.
func (token Token) Delete() error {
	if !token.Exists() {
		return fmt.Errorf("auth: keyless tokens cannot be deleted")
	}
	return token.manager.Delete(token.Key)
}

// Exists returns true if the token exists
func (token Token) Exists() bool {
	return token.Key != ""
}

// Tokens is the postgres schema for user API tokens.
var Tokens = sql.Table("tokens",
	sql.Column("key", sql.String{NotNull: true}),
	sql.ForeignKey(
		"user_id",
		Users.C["id"],
		sql.Integer{NotNull: true},
	).OnDelete(sql.Cascade),
	sql.Column("expires", sql.Timestamp{WithTimezone: true}),
	sql.Column("created_at", sql.Timestamp{Default: postgres.Now}),
	sql.PrimaryKey("key"),
)

// TokenManager is the internal manager of tokens
type TokenManager struct {
	conn    sql.Connection
	keyFunc KeyFunc
	nowFunc func() time.Time
}

// All returns all tokens for the given user ID
func (m *TokenManager) All(id int64) (tokens []Token) {
	stmt := Tokens.Select().Where(Tokens.C["user_id"].Equals(id))
	m.conn.MustQueryAll(stmt, &tokens)
	return
}

// Delete removes the token with the given key from the database.
// It will return an error if the token does not have a key or the key
// was not deleted from the database. It will panic on any connection error.
func (m *TokenManager) Delete(key string) error {
	stmt := Tokens.Delete().Where(Tokens.C["key"].Equals(key))
	rowsAffected, err := m.conn.MustExecute(stmt).RowsAffected()
	if err != nil {
		return fmt.Errorf("auth: error during rows affected: %s", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("auth: token key %s was not deleted", key)
	}
	return nil
}

// Create creates a new token for the user. It will panic on error. The user
// ID must exist.
func (m *TokenManager) ForeverToken(user User) (token Token) {
	token.UserID = user.ID
	token.manager = m

	// Generate a new token
	for {
		token.Key = m.keyFunc()

		// No duplicates - generate a new key if this key already exists
		stmt := sql.Select(
			Tokens.C["key"],
		).Where(Tokens.C["key"].Equals(token.Key)).Limit(1)
		var duplicate string
		if !m.conn.MustQueryOne(stmt, &duplicate) {
			break
		}
	}

	// Insert the token into the database
	st := postgres.Insert(Tokens).Values(token).Returning(Tokens.Columns()...)
	m.conn.MustQueryOne(st, &token)
	return
}

// Get returns the token with the given key. Panic on database error.
func (m *TokenManager) Get(key string) (token Token) {
	stmt := Tokens.Select().Where(Tokens.C["key"].Equals(key))
	m.conn.MustQueryOne(stmt, &token)
	return
}

func NewTokens(conn sql.Connection) *TokenManager {
	return &TokenManager{
		conn:    conn,
		keyFunc: RandomKey,
		nowFunc: time.Now,
	}
}
