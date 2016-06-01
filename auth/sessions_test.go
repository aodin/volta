package auth

import (
	"testing"

	"github.com/aodin/config"
	"github.com/stretchr/testify/assert"
)

func TestSessions(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Sessions.Name(), "sessions")

	// Get a blank DB and create the schemas
	tx, _ := getConn(t).Must().Begin()
	defer tx.Rollback()
	initSchema(tx, Users, Sessions, Tokens)

	// Create a new user
	users := MockUsers(tx)
	admin, err := users.Create("admin@example.com", "admin", "guy", "secret")
	assert.Nil(err, "User Create() returned an error")
	assert.NotEqual(0, admin.ID, "User ID was not set")
	assert.NotEqual(0, admin.Token, "User token was not set")
	assert.False(admin.TokenSetAt.IsZero(), "Token timestamp was not set")
	assert.Equal("admin@example.com", admin.Email)
	assert.Equal("admin guy", admin.Name())
	assert.NotEqual("", admin.Password, "User password was not set")
	assert.Equal(MakePassword(users.Hasher(), "secret"), admin.Password)
	assert.True(CheckPassword(users.Hasher(), "secret", admin.Password))

	// Create a new session
	sessions := NewSessions(config.DefaultCookie, tx)
	session := sessions.Create(admin)
	assert.Equal(admin.ID, session.UserID)

	// Delete a session
	assert.Nil(session.Delete(), "Deleting a session returned an error")
}
