package auth

import (
	"crypto/sha1"
	"github.com/aodin/volta/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionManager(t *testing.T) {
	assert := assert.New(t)

	// Create a simple test hasher
	hasher := NewPBKDF2Hasher("test", 1, sha1.New)

	// Create a new in-memory user manager
	users := UsersInMemory(hasher)

	// Create a new user
	admin, err := users.Create("admin", "admin")
	assert.Nil(err)
	assert.Equal(admin.Name(), "admin")

	// Create a new sessions manager
	sessions := SessionsInMemory(config.DefaultCookie)

	session, err := sessions.Create(admin)
	assert.Nil(err)

	// Get the user back out of the session
	id := session.User()
	authAdmin, err := users.Get(Fields{"id": id})
	assert.Nil(err)

	assert.Equal(admin.Name(), authAdmin.Name())
	assert.Equal(admin.Password(), authAdmin.Password())

	// TODO test the handling of duplicate session keys
}
