package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokens(t *testing.T) {
	assert := assert.New(t)

	// Get a blank DB and create the schemas
	tx, _ := getConn(t).Must().Begin()
	defer tx.Rollback()
	initSchema(tx, Users, Sessions, Tokens)

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

	// There should now be two tokens
	assert.EqualValues(2, tokens.Count())

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

	// There should be one token left in the database
	assert.EqualValues(1, tokens.Count())

	// Delete the user, it should clear the remaining token
	assert.Nil(user.Delete(), "Deleting a user should not return an error")
	assert.EqualValues(0, tokens.Count())
}
