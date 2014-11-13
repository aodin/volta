package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTestYAML(t *testing.T) {
	assert := assert.New(t)

	c, err := ParseTestYAML("./test_fixtures/dbconf.yml")
	require.Nil(t, err)
	assert.Equal("postgres", c.Driver)
	assert.Equal("host=localhost port=5432 dbname=db_test user=test password=bad sslmode=disable", c.Credentials())
}
