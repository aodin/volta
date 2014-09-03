package authdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUsers(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Users.Name, "users")
}
