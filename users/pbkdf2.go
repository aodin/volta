package users

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
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

*/

func key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	prf := hmac.New(h, password)
	hashLen := prf.Size()
	numBlocks := (keyLen + hashLen - 1) / hashLen

	var buf [4]byte
	dk := make([]byte, 0, numBlocks*hashLen)
	U := make([]byte, hashLen)
	for block := 1; block <= numBlocks; block++ {
		// N.B.: || means concatenation, ^ means XOR
		// for each block T_i = U_1 ^ U_2 ^ ... ^ U_iter
		// U_1 = PRF(password, salt || uint(i))
		prf.Reset()
		prf.Write(salt)
		buf[0] = byte(block >> 24)
		buf[1] = byte(block >> 16)
		buf[2] = byte(block >> 8)
		buf[3] = byte(block)
		prf.Write(buf[:4])
		dk = prf.Sum(dk)
		T := dk[len(dk)-hashLen:]
		copy(U, T)

		// U_n = PRF(password, U_(n-1))
		for n := 2; n <= iter; n++ {
			prf.Reset()
			prf.Write(U)
			U = U[:0]
			U = prf.Sum(U)
			for x := range U {
				T[x] ^= U[x]
			}
		}
	}
	return dk[:keyLen]
}

func Pbkdf2(cleartext, salt []byte, rounds int, h func() hash.Hash) []byte {
	// Use the hash Size as the keyLen
	return key(cleartext, salt, rounds, h().Size(), h)
}

// TODO declare private?
type PBKDF2_Base struct {
	BaseHasher
	Digest func() hash.Hash // TODO move to base hasher?
}

func (pbkH *PBKDF2_Base) Encode(cleartext, salt string) string {
	// TODO these []byte conversions are a bit silly
	rounds := 10000
	hashed := EncodeBase64String(Pbkdf2([]byte(cleartext), []byte(salt), rounds, pbkH.Digest))
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
	hashed := Pbkdf2([]byte(cleartext), []byte(salt), rounds, pbkH.Digest)
	return ConstantTimeStringCompare(EncodeBase64String(hashed), splitHash[3])
}

func init() {
	pbkdf2_sha256 := &PBKDF2_Base{BaseHasher{Algorithm: "pbkdf2_sha256", Rounds: 10000}, sha256.New}
	Register(pbkdf2_sha256.Algorithm, pbkdf2_sha256)

	pbkdf2_sha1 := &PBKDF2_Base{BaseHasher{Algorithm: "pbkdf2_sha1", Rounds: 10000}, sha1.New}
	Register(pbkdf2_sha1.Algorithm, pbkdf2_sha1)
}
