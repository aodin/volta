package auth

import (
	"testing"

	sql "github.com/aodin/aspect"
	_ "github.com/aodin/aspect/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func countTokens(conn sql.Connection) (count int64) {
	conn.MustQueryOne(sql.Select(sql.Count(Tokens.C["key"])), &count)
	return
}

func TestTokens(t *testing.T) {
	assert := assert.New(t)

	conn, tx := initSchemas(t, Users, Sessions, Tokens)
	defer tx.Rollback()
	defer conn.Close()

	users := MockUsers(tx)
	tokens := NewTokens(tx)

	// Modify the expires and key gen func
	var mock mockKeyFunc
	tokens.keyFunc = mock.Key

	// Create a user
	user, err := users.Create("a@example.com", "admin", "guy", "secret")
	require.Nil(t, err, "Error while creating user")

	// Create a token for the user that lasts forever
	token := tokens.ForeverToken(user)
	assert.Equal("mock", token.Key)
	assert.Equal(user.ID, token.UserID)

	// Generate a new token with a key collision. It should collide, then
	// generate a random key.
	repeat := tokens.ForeverToken(user)
	assert.Equal(user.ID, repeat.UserID)

	// Get a token that should exist
	byKey := tokens.Get(repeat.Key)
	assert.True(byKey.Exists(), "Token not returned from Get()")
	assert.Equal(byKey.Key, repeat.Key)
	assert.Equal(byKey.UserID, repeat.UserID)

	// Get a token that shouldn't exist
	invalid := tokens.Get("DNE")
	assert.False(invalid.Exists(), "Token should not exist")

	// Delete a token that exists
	assert.Nil(
		repeat.Delete(),
		"Deleting an existing token returned an error",
	)

	// Attempt to delete a keyless token
	assert.NotNil(Token{}.Delete(), "Deleting a keyless token should error")

	// And one that doesn't
	assert.NotNil(
		tokens.Delete("DNE"),
		"Deleting a token that does not exist should return an error",
	)

	// There should be one token left in the database
	assert.Equal(1, countTokens(tx))

	// Delete the user, it should clear the remaining token
	assert.Nil(user.Delete(), "Deleting a user should not return an error")
	assert.Equal(0, countTokens(tx))
}
