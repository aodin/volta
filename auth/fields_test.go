package auth

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testUser struct {
	ID      int64  `db:"id"`
	Name    string `db:"name"`
	IsAdmin bool   `db:"is_admin"`
}

func TestFields(t *testing.T) {
	assert := assert.New(t)
	// Create fields
	f := Fields{
		"id":       1,
		"name":     "admin",
		"is_admin": true,
	}

	var user testUser
	err := f.Unmarshal(&user)
	assert.Nil(err)
	assert.Equal(1, user.ID)
	assert.Equal("admin", user.Name)
	assert.Equal(true, user.IsAdmin)
}
