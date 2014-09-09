package auth

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	// Get the hasher from the registry
	pbkdf2_sha256, err := GetHasher("pbkdf2_sha256")
	if err != nil {
		// TODO best way to print an error?
		t.Errorf("Failed to get hasher with err %s", err)
	}
	cleartext := "badpassword"
	hashed := MakePassword(pbkdf2_sha256, cleartext)
	// TODO some error output would be appreciated
	verify := CheckPassword(pbkdf2_sha256, cleartext, hashed)
	if !verify {
		t.Errorf("Password did not verify!")
	}
}

// TODO Create a hash and plug it in
// TODO test util functions
