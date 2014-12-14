package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	assert := assert.New(t)

	var valid string
	var err error

	// Valid
	valid, err = Normalize(" a@example.com ")
	assert.Nil(err)
	assert.Equal("a@example.com", valid)

	valid, err = Normalize(" a@EXAMPLE.com ")
	assert.Nil(err)
	assert.Equal("a@example.com", valid)

	// Invalid
	_, err = Normalize(" a ")
	assert.NotNil(err)

	_, err = Normalize(" a@ ")
	assert.NotNil(err)

	_, err = Normalize(" @a@ ")
	assert.NotNil(err)
}
