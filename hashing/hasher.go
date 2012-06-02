package hashing

import (
	"fmt"
)

type Hasher interface {
	Encode(string, string) string
	Salt() string
	Verify(string, string) bool
	GetAlgorithm() string // TODO easier way to access properties?
}

func MakePassword(h Hasher, cleartext string) string {
	return h.Encode(cleartext, h.Salt())
}

func CheckPassword(h Hasher, cleartext, encoded string) bool {
	return h.Verify(cleartext, encoded)
}

var hashers = make(map[string]Hasher)

func Register(name string, hasher Hasher) {
	if hasher == nil {
		panic("hashing: Attempting to register a nil Hasher")
	}
    if _, duplicate := hashers[name]; duplicate {
    	panic("hashing: Register called twice for Hasher " + name)
    }
    hashers[name] = hasher
}

func Get(name string) (Hasher, error) {
	hasher, ok := hashers[name]
	if !ok {
		return nil, fmt.Errorf("hashing: unknown hasher %q (did you remember to import it?)", name)
	}
	return hasher, nil
}

// The BaseHasher struct is the parent of all included Hashers
type BaseHasher struct {
	Algorithm string
	Rounds int
}

// TODO should these functions return an error as well?
func (bH *BaseHasher) Encode(cleartext string) string {
	// TODO raise an error (NotImplementedError)
	panic("hashing: BaseHasher has no ability to encode.")
}

func (bH *BaseHasher) Salt() string {
	// Create a random string
	return EncodeBase64String(RandomBytes(9))
}

func (bH *BaseHasher) Verify(cleartext, encoded string) bool {
	// TODO raise an error (NotImplementedError)
	panic("hashing: BaseHasher has no ability to verify.")
}

func (bH *BaseHasher) GetAlgorithm() string {
	return bH.Algorithm
}