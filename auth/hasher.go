package auth

import (
	"fmt"
	"hash"
)

// Hasher is the target interface for included hashers.
type Hasher interface {
	Encode(string, string) string
	Salt() string
	Verify(string, string) bool
	Algorithm() string
}

// MakePassword hashes the given cleartext string using the given Hasher
func MakePassword(h Hasher, cleartext string) string {
	return h.Encode(cleartext, h.Salt())
}

// CheckPassword verifies the given cleartext password against the given
// encoded string using the given hasher.
func CheckPassword(h Hasher, cleartext, encoded string) bool {
	return h.Verify(cleartext, encoded)
}

var hashers = make(map[string]Hasher)

// RegisterHasher adds a new Hasher to the registry with the given name.
func RegisterHasher(name string, hasher Hasher) {
	if hasher == nil {
		panic("auth: attempting to register a nil Hasher")
	}
	if _, duplicate := hashers[name]; duplicate {
		panic("auth: register called twice for Hasher " + name)
	}
	hashers[name] = hasher
}

// GetHasher returns the Hasher in the registry with the given name.
func GetHasher(name string) (Hasher, error) {
	hasher, ok := hashers[name]
	if !ok {
		return nil, fmt.Errorf(
			"auth: unknown hasher %s (did you remember to import it?)", name,
		)
	}
	return hasher, nil
}

// BaseHasher is the parent of all included Hashers
type BaseHasher struct {
	algorithm string
}

// Salt generates nine random bytes encoded to base64 for use as a salt.
func (h *BaseHasher) Salt() string {
	return EncodeBase64String(RandomBytes(9))
}

// Algorithm returns the algorithm of this Hasher
func (h *BaseHasher) Algorithm() string {
	return h.algorithm
}

func NewBaseHasher(algorithm string) BaseHasher {
	return BaseHasher{algorithm: algorithm}
}

type mockHasher struct {
	PBKDF2_Base
}

func (h *mockHasher) Salt() string {
	return h.algorithm
}

func MockHasher(name string, n int, digest func() hash.Hash) *mockHasher {
	return &mockHasher{PBKDF2_Base{NewBaseHasher(name), n, digest}}
}
