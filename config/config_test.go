package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	assert := assert.New(t)

	// Parse the local settings.json file
	c, err := Parse()
	assert.Nil(err)

	// Test the parent config methods
	assert.Equal("localhost:9001", c.Address())
	assert.Equal("http://localhost:9001", c.FullAddress())

	c.Port = 80
	c.HTTPS = true
	assert.Equal("https://localhost", c.FullAddress())

	// Test the SMTP config methods
	assert.Equal(`"Example User" <no_reply@example.com>`, c.SMTP.FromAddress())
	assert.Equal("example.com:587", c.SMTP.Address())

	// Test the database config methods when values are missing
	assert.Equal(
		"host=localhost port=5432 dbname=db user=pg password=pass",
		c.Database.Credentials(),
	)

	// Test the default cookie settings
	assert.Equal(336*time.Hour, c.Cookie.Age)
	assert.Equal("", c.Cookie.Domain)
	assert.Equal(false, c.Cookie.HttpOnly)
	assert.Equal("sessionid", c.Cookie.Name)
	assert.Equal("/", c.Cookie.Path)
	assert.Equal(false, c.Cookie.Secure)

	// TODO Test custom cookie settings
}
