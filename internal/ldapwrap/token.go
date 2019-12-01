package ldapwrap

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"io"
	"time"
)

func GenerateToken(nickname string) (tokenString string, err error) {
	random := bytes.NewBuffer(make([]byte, 0, 32))
	_, err = io.CopyN(random, rand.Reader, 32)
	if err != nil {
		return "", fmt.Errorf("unable to generate random token: %s", err)
	}
	token := Token{
		ValidUntil: time.Now().Add(24 * time.Hour),
		Nickname:   nickname,
		Random:     random.Bytes(),
	}
	w := bytes.NewBuffer([]byte{})
	err = gob.NewEncoder(w).Encode(token)
	if err != nil {
		return "", fmt.Errorf("unable to gob encode token: %s", err)
	}
	return base64.URLEncoding.EncodeToString(w.Bytes()), nil
}

func ValidateToken(tokenString string, nickname string) (ok bool, err error) {
	tokenBytes, err := base64.URLEncoding.DecodeString(tokenString)
	if err != nil {
		return false, fmt.Errorf("unable to decode base64: %s\n%s", err, tokenString)
	}
	token := Token{}
	r := bytes.NewBuffer(tokenBytes)
	err = gob.NewDecoder(r).Decode(&token)
	if err != nil {
		return false, fmt.Errorf("unable to decode token: %s", err)
	}
	if token.Nickname != nickname {
		return false, fmt.Errorf("invalid nickname in token: %s", token.Nickname)
	}
	return true, nil
}
