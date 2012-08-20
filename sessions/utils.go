package sessions

import (
	"crypto/rand"
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