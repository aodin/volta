package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParams(t *testing.T) {
	assert := assert.New(t)

	// Create a slice of params
	params := Params{{Key: "id", Value: "1"}}
	assert.Equal("1", params.ByName("id"), "Parameter 'id' should exist")
	assert.Equal("", params.ByName("name"), "Parameter 'name' shouldn't exist")
	assert.EqualValues(1, params.AsID("id"), "Parameter 'id' should be an integer")
	assert.EqualValues(0, params.AsID("name"), "Parameter 'name' should return as 0")

	assert.True(
		params.EqualsAny("id", "0", "1", "2"),
		"EqualsAny should have found a match",
	)
	assert.False(
		params.EqualsAny("id", "a", "b", "c"),
		"EqualsAny should not have succeeded",
	)
	assert.False(
		params.EqualsAny("name", "0", "1", "2"),
		"A parameter that does not exist should not match non-empty values",
	)
}
