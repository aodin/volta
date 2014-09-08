package auth

import (
	"crypto/sha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserManager(t *testing.T) {
	assert := assert.New(t)

	// Create a simple test hasher
	hasher := NewPBKDF2Hasher("test", 1, sha1.New)

	// Perform all operations through the UserManager interface
	var manager UserManager = UsersInMemory(hasher)

	// Create a new user
	admin, err := manager.Create("admin", "admin")
	assert.Nil(err)
	assert.Equal(admin.ID(), 1)
	assert.Equal(admin.Name(), "admin")

	// Verify the password
	assert.Equal(CheckPassword(hasher, "admin", admin.Password()), true)
}
