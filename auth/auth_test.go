package auth

import (
	"testing"

	sql "github.com/aodin/aspect"
	"github.com/aodin/volta/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	// Create a mock auth using the testing database
	assert := assert.New(t)

	// Create the test schema
	conn, dbtx := initSchemas(t, Users, Sessions, Tokens)
	defer dbtx.Rollback()
	defer conn.Close()

	// Some of these operations are transactional
	tx := sql.FakeTx(dbtx)

	// Create a mock Auth and test its methods
	auth := Mock(config.Default, tx)

	var valid, invalid User
	var err error

	// Create a user, session, and a token
	user, err := auth.CreateUser("a@example.com", "admin", "guy", "secret")
	require.Nil(t, err, "Error during user creation")
	require.True(t, user.Exists(), "Failed to create user")

	session := auth.sessions.Create(user)
	require.True(t, session.Exists(), "Failed to create session")
	assert.Equal(user.ID, session.UserID)

	token := auth.tokens.ForeverToken(user)
	assert.Equal(user.ID, token.UserID)

	// Duplicate users cannot be created
	invalid, err = auth.CreateUser("a@example.com", "admin", "guy", "secret")
	require.NotNil(t, err, "Failed to error when creating duplicate user")
	require.False(t, invalid.Exists(), "Invalid user was created")

	// Update the user's existing token to perform auth by user token
	auth.ResetUserToken(&user)
	assert.NotEqual("", user.Token, "No user token was set")
	assert.False(user.TokenSetAt.IsZero(), "No user token timestamp was set")

	// Attempt auth by password
	valid, err = auth.ByPassword("a@example.com", "secret")
	require.Nil(t, err, "Could not auth by password")
	assert.Equal(user.ID, valid.ID)

	// Incorrect password
	_, err = auth.ByPassword("a@example.com", "1234")
	assert.NotNil(err, "Incorrect password should have errored during auth")

	// User that does not exist
	_, err = auth.ByPassword("b@example.com", "secret")
	assert.NotNil(err, "Missing email should have errored auth by password")

	// Attempt auth by session
	valid = auth.BySession(session.Key)
	assert.True(valid.Exists(), "An invalid user was returned by session key")

	// Session that does not exist
	invalid = auth.BySession("")
	assert.False(
		invalid.Exists(),
		"A valid user returned from a session that does not exist",
	)

	// Attempt auth by user token (the token field on the user schema)
	valid, err = auth.ByUserToken(user.ID, user.Token)
	assert.Nil(err, "An invalid user was returned by user token")
	assert.True(valid.Exists())

	invalid, err = auth.ByUserToken(0, user.Token)
	assert.NotNil(err, "An valid user was returned from a zero id")
	assert.False(invalid.Exists())

	invalid, err = auth.ByUserToken(user.ID, "")
	assert.NotNil(err, "An valid user was returned from an empty token")
	assert.False(invalid.Exists())

	// Attempt auth by token (used in APIs)
	valid = auth.ByToken(user.ID, token.Key)
	assert.True(valid.Exists(), "An invalid user was returned by token")

	invalid = auth.ByToken(user.ID, "")
	assert.False(invalid.Exists(), "A valid user returned from an empty token")
}
