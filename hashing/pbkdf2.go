package hashing

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"hash"
	"strconv"
	"strings"
)

/*

The PBKDF2 algorithm is a slightly modified version of:
https://bitbucket.org/taruti/pbkdf2/

Copyright (c) 2010-2011 Taru Karttunen <taruti@taruti.net>

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the "Software"), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.

*/

func Pbkdf2(cleartext []byte, salt []byte, rounds int, hash func() hash.Hash, outlen int) []byte {
	// TODO necessity of outlen? For the app we'll be using the hash Size
	hashSize := hash().Size()
	if outlen == 0 {
		outlen = hashSize
	}
	out := make([]byte, outlen)
	ibuf := make([]byte, 4)
	block := 1
	p := out
	for outlen > 0 {
		clen := outlen
		if clen > hashSize {
			clen = hashSize
		}
		ibuf[0] = byte((block >> 24) & 0xff)
		ibuf[1] = byte((block >> 16) & 0xff)
		ibuf[2] = byte((block >> 8) & 0xff)
		ibuf[3] = byte((block) & 0xff)
		hmac := hmac.New(hash, cleartext)
		hmac.Write(salt)
		hmac.Write(ibuf)
		tmp := hmac.Sum(nil)
		for i := 0; i < clen; i++ {
			p[i] = tmp[i]
		}
		for j := 1; j < rounds; j++ {
			hmac.Reset()
			hmac.Write(tmp)
			tmp = hmac.Sum(nil)
			for k := 0; k < clen; k++ {
				p[k] ^= tmp[k]
			}
		}
		outlen -= clen
		block++
		p = p[clen:]
	}
	return out
}

// TODO declare private?
type PBKDF2_Base struct {
	BaseHasher
	Digest func() hash.Hash // TODO move to base hasher?
}

func (pbkH *PBKDF2_Base) Encode(cleartext, salt string) string {
	// TODO these []byte conversions are a bit silly
	rounds := 10000
	hashed := EncodeBase64String(Pbkdf2([]byte(cleartext), []byte(salt), rounds, pbkH.Digest, 0))
	return strings.Join([]string{pbkH.GetAlgorithm(), fmt.Sprintf("%d", rounds), salt, hashed}, "$")
}

func (pbkH *PBKDF2_Base) Verify(cleartext, encoded string) bool {
	// Split the saved hash apart
	splitHash := strings.SplitN(encoded, "$", 4)

	// The algorithm should match this hasher
	algo := splitHash[0]
	if algo != pbkH.Algorithm {
		return false
	}
	rounds64, err := strconv.ParseInt(splitHash[1], 10, 0)
	if err != nil {
		return false
	}
	rounds := int(rounds64)
	salt := splitHash[2]

	// Generate a new hash using the given cleartext
	hashed := Pbkdf2([]byte(cleartext), []byte(salt), rounds, pbkH.Digest, 0)
	return ConstantTimeStringCompare(EncodeBase64String(hashed), splitHash[3])
}

func init() {
	pbkdf2_sha256 := &PBKDF2_Base{BaseHasher{Algorithm: "pbkdf2_sha256", Rounds: 10000}, sha256.New}
	Register(pbkdf2_sha256.Algorithm, pbkdf2_sha256)

	pbkdf2_sha1 := &PBKDF2_Base{BaseHasher{Algorithm: "pbkdf2_sha1", Rounds: 10000}, sha1.New}
	Register(pbkdf2_sha1.Algorithm, pbkdf2_sha1)
}