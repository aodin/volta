package auth

import (
	"testing"

	sql "github.com/aodin/aspect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aodin/volta/config"
)

func initSchemas(t *testing.T, tables ...*sql.TableElem) (*sql.DB, sql.Transaction) {
	// Connect to the database specified in the test db.json config
	// Default to the Travis CI settings if no file is found
	conf, err := sql.ParseTestConfig("./db.json")
	if err != nil {
		t.Fatalf(
			"auth: failed to parse test configuration, test aborted: %s",
			err,
		)
	}
	conn, err := sql.Connect(conf.Driver, conf.Credentials())
	require.Nil(t, err)

	// Perform all tests in a transaction and always rollback
	tx, err := conn.Begin()
	require.Nil(t, err)

	// Create the given schemas
	for _, table := range tables {
		_, err = tx.Execute(table.Create())
		require.Nil(t, err)
	}
	return conn, tx
}

func TestSessions(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Sessions.Name, "sessions")

	conn, tx := initSchemas(t, Users, Sessions, Tokens)
	defer tx.Rollback()
	defer conn.Close()

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
}
