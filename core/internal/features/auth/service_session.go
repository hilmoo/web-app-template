package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/matthewhartstonge/argon2"
)

const (
	// sessionDuration is how long a newly created session remains valid.
	sessionDuration = 7 * 24 * time.Hour
	// randomTokenBytes is the number of random bytes used to generate a session token.
	randomTokenBytes = 32
)

func hashSessionToken(secret, token string) string {
	hashToken := hmac.New(sha256.New, []byte(secret))
	hashToken.Write([]byte(token))
	return base64.URLEncoding.EncodeToString(hashToken.Sum(nil))
}

func newSessionToken() (string, error) {
	b := make([]byte, randomTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashPasswordArgon(plain string) (string, error) {
	argon := argon2.DefaultConfig()
	hash, err := argon.HashEncoded([]byte(plain))
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}
