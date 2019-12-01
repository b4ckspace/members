package ssha

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
)

type (
	HashAlgo string
)

const (
	SSHA    HashAlgo = "SSHA"
	SSHA256 HashAlgo = "SSHA256"
	SSHA512 HashAlgo = "SSHA512"
)

func Hash(password string, algo HashAlgo) (hashed string, err error) {
	var s hash.Hash
	var saltLen int
	switch algo {
	case SSHA:
		s = sha1.New()
		saltLen = 8
	case SSHA256:
		s = sha256.New()
		saltLen = 8
	case SSHA512:
		s = sha512.New()
		saltLen = 16
	default:
		return "", fmt.Errorf("invalid hash algo")
	}

	salt := make([]byte, saltLen)
	_, err = rand.Read(salt)
	fmt.Println(salt)
	if err != nil {
		return "", fmt.Errorf("unable to generate salt: %s", err)
	}
	_, err = s.Write([]byte(password))
	if err != nil {
		return "", err
	}
	_, err = s.Write(salt)
	if err != nil {
		return "", err
	}
	hash := s.Sum(nil)
	hashWithSalt := append(hash, salt...)
	hash64 := base64.StdEncoding.EncodeToString(hashWithSalt)
	hashed = fmt.Sprintf("{%s}%s", algo, hash64)
	return
}
