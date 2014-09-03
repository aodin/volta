package authdb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessions(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(Sessions.Name, "sessions")
}
