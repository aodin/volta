package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"log"
)

// RandomKey generates a new session key. It does so by producing 24
// random bytes that are encoded in URL safe base64, for output of 32 chars.
func RandomKey() string {
	return RandomKeyN(24)
}

// RandomKeyN generates a new Base 64 encoded random string. N is the length
// of the random bytes, not the final encoded string.
func RandomKeyN(n int) string {
	return EncodeBase64String(RandomBytes(n))
}

// RandomBytes returns random bytes from the crypto/rand Reader or it panics.
func RandomBytes(n int) []byte {
	key := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		log.Panicf("auth: could not generate random bytes: %s", err)
	}
	return key
}

// KeyFunc is the function type that will be used to generate new session keys.
type KeyFunc func() string

// RandomKey should meet the KeyFunc signature
var _ KeyFunc = RandomKey

type mockKeyFunc bool

// Return the hardcoded key the first time the func is run, then vary
func (ran *mockKeyFunc) Key() string {
	if !*ran {
		*ran = true
		return "mock"
	}
	return RandomKey()
}

// TODO standardize the usage of []byte arrays versus string

// ConstantTimeStringCompare is wrapper around subtle.ConstantTimeCompare that
// takes two strings as parameters and returns a boolean instead of an int.
func ConstantTimeStringCompare(x, y string) bool {
	return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1
}

// EncodeBase64String is a wrapper around the standard base64 encoding call.
func EncodeBase64String(input []byte) string {
	return base64.URLEncoding.EncodeToString(input)
}
