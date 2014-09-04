package authmem

import (
	"github.com/aodin/volta/auth"
	"github.com/aodin/volta/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSessionManager(t *testing.T) {
	assert := assert.New(t)

	// Create a new in-memory user manager
	users := NewUserManager()
	user, err := users.NewUser("admin", true)
	assert.Nil(err)

	// Create a new in-memory session manager
	sessions := NewSessionManager(config.DefaultCookie, users)
	session, err := sessions.NewSession(user.id)
	assert.Nil(err)

	// Get the session
	s, err := sessions.GetSession(session.key)
	assert.Nil(err)

	// It should implement the auth.Session interface
	var authSession auth.Session = s
	assert.Equal(session.key, authSession.Key())

	// Get the user tied to this session
	u, err := authSession.User()
	assert.Nil(err)
	assert.Equal(user.id, u.ID())
	assert.Equal(user.email, u.Username())
	assert.Equal(user.isAdmin, u.IsAdmin())
}
