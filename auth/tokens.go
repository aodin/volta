package auth

import (
	"fmt"
	"time"

	"github.com/aodin/sol"
	"github.com/aodin/sol/postgres"
	"github.com/aodin/sol/types"
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
var Tokens = postgres.Table("tokens",
	sol.Column("key", types.Varchar().NotNull()),
	sol.ForeignKey(
		"user_id",
		Users.C("id"),
		types.Integer().NotNull(),
	).OnDelete(sol.Cascade).OnUpdate(sol.Cascade),
	sol.Column("expires", postgres.Timestamp().WithTimezone()),
	sol.Column(
		"created_at",
		postgres.Timestamp().WithTimezone().NotNull().Default(postgres.Now),
	),
	sol.PrimaryKey("key"),
)

// TokenManager is the internal manager of tokens
type TokenManager struct {
	conn    sol.Conn
	keyFunc KeyFunc
	nowFunc func() time.Time
}

// All returns all tokens for the given user ID
func (m *TokenManager) All(id int64) (tokens []Token) {
	stmt := Tokens.Select().Where(Tokens.C("user_id").Equals(id))
	m.conn.Query(stmt, &tokens)
	return
}

func (m *TokenManager) Count() (count int64) {
	m.conn.Query(sol.Select(sol.Count(Tokens.C("key"))), &count)
	return
}

// Delete removes the token with the given key from the database.
// It will return an error if the token does not have a key or the key
// was not deleted from the database. It will panic on any connection error.
func (m *TokenManager) Delete(key string) error {
	stmt := Tokens.Delete().Where(Tokens.C("key").Equals(key))
	return m.conn.Query(stmt)
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
		stmt := sol.Select(
			Tokens.C("key"),
		).Where(Tokens.C("key").Equals(token.Key)).Limit(1)
		var duplicate string
		m.conn.Query(stmt, &duplicate)
		if duplicate == "" {
			break
		}
	}

	// Insert the token into the database
	m.conn.Query(postgres.Insert(Tokens).Values(token).Returning(), &token)
	return
}

// Get returns the token with the given key. Panic on database error.
func (m *TokenManager) Get(key string) (token Token) {
	stmt := Tokens.Select().Where(Tokens.C("key").Equals(key))
	m.conn.Query(stmt, &token)
	return
}

func NewTokens(conn sol.Conn) *TokenManager {
	return &TokenManager{
		conn:    conn,
		keyFunc: RandomKey,
		nowFunc: time.Now,
	}
}
