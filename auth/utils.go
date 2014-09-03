package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"math/big"
	"strings"
)

func GetRandomString(length int, allowed string) string {
	// allowed Chars should be a byte array?
	randReader := rand.Reader
	max := big.NewInt(int64(len(allowed)))
	randString := make([]string, length)
	allowedChars := []byte(allowed)
	for i := 0; i < length; i++ {
		randInt, _ := rand.Int(randReader, max)
		randString[i] = string(allowedChars[randInt.Int64()])
	}
	return strings.Join(randString, "")
}

// TODO standardize the usage of []byte arrays versus string
func ConstantTimeStringCompare(v1, v2 string) bool {
	// Reimplementation of crypto.subtle.ConstantTimeCompare
	b1 := []byte(v1)
	b2 := []byte(v2)
	if len(b1) != len(b2) {
		return false
	}
	var result byte
	for i := 0; i < len(b1); i++ {
		result |= b1[i] ^ b2[i]
	}
	return subtle.ConstantTimeByteEq(result, 0) == 1
}

// TODO errors instead of panic?
func EncodeBase64String(input []byte) string {
	var buf bytes.Buffer
	e := base64.NewEncoder(base64.StdEncoding, &buf)
	e.Write(input)
	e.Close()
	return buf.String()
}

func RandomBytes(length int) []byte {
	salt := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, salt)
	if err != nil {
		panic("auth: could not generate random bytes")
	}
	return salt
}
