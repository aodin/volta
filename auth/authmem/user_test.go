package authmem

import (
	"github.com/aodin/volta/auth"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUserManager(t *testing.T) {
	assert := assert.New(t)

	// Create a new in-memory manager
	manager := NewUserManager()
	admin, err := manager.NewUser("admin", true)
	assert.Nil(err)

	// This user should implement the volta.auth User interface
	var authUser auth.User = admin

	// Its fields should be accessible through the interface getters
	assert.Equal(authUser.ID(), 1)
	assert.Equal(authUser.Username(), "admin")
	assert.Equal(authUser.IsAdmin(), true)

	// TODO Duplicate a user?
	err = admin.Create()
	assert.Nil(err)
	assert.Equal(admin.id, 2)
}
