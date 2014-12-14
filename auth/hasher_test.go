package auth

import (
	"crypto/sha1"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistry(t *testing.T) {
	assert := assert.New(t)

	// Get the hasher from the registry
	pbkdf2_sha256, err := GetHasher("pbkdf2_sha256")
	assert.Nil(err, "GetHasher should not return an error")

	// Hash a cleartext and verify
	cleartext := "badpassword"
	hashed := MakePassword(pbkdf2_sha256, cleartext)
	assert.True(CheckPassword(pbkdf2_sha256, cleartext, hashed))

	// Get a hasher that doesn't exist
	_, err = GetHasher("dne")
	assert.NotNil(err)

	// Attempt to register a nil hasher
	func() {
		var panicked interface{}
		defer func() {
			panicked = recover()
		}()
		RegisterHasher("nil", nil)
		if panicked == nil {
			t.Fatalf("auth: registry failed to panic when given a nil hasher")
		}
	}()

	// Create a hasher and register it
	pbkdf2_crap := NewPBKDF2Hasher("pbkdf2_crap", 1, sha1.New)
	RegisterHasher(pbkdf2_crap.algorithm, pbkdf2_crap)

	// Twice
	func() {
		var panicked interface{}
		defer func() {
			panicked = recover()
		}()
		RegisterHasher(pbkdf2_crap.algorithm, pbkdf2_crap)
		if panicked == nil {
			t.Fatalf(
				"auth: registry failed to panic when given a duplicate hasher",
			)
		}
	}()
}
