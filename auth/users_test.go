package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUsers(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Users.Name(), "users")

	// Get a blank DB and create the schemas
	tx, _ := getConn(t).Must().Begin()
	defer tx.Rollback()
	initSchema(tx, Users, Sessions, Tokens)

	// Create a new users manager with the default hasher - pbkdf2_sha256
	users := NewUsers(tx)

	admin, err := users.CreateSuperuser("a@example.com", "A", "B", "secret")
	require.Nil(t, err, "CreateSuperuser returned an error")
	assert.Equal(fmt.Sprintf("%d: a@example.com", admin.ID), admin.String())

	// Attempt to get a user by an ID that does not exist
	dne, err := users.GetByID(0)
	assert.NotNil(err, "Getting non-existing users by ID should error")
	assert.False(dne.Exists(), "A non-existing user should not exist")

	// Only users with IDs can be deleted
	assert.NotNil(User{}.Delete(), "Delete did not error for a zero ID user")
}
