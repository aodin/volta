package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsers(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Users.Name, "users")
}
