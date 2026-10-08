package session

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateKey generates a secure random key of the specified length.
// For AES-256 encryption, use 32 bytes.
// For HMAC signing, use at least 32 bytes.
func GenerateKey(length int) ([]byte, error) {
	key := make([]byte, length)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateKeyHex generates a secure random key and returns it as a hex string.
func GenerateKeyHex(length int) (string, error) {
	key, err := GenerateKey(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// MustGenerateKey generates a key and panics if it fails.
func MustGenerateKey(length int) []byte {
	key, err := GenerateKey(length)
	if err != nil {
		panic("session: failed to generate key: " + err.Error())
	}
	return key
}

// MustGenerateKeyHex generates a key as hex string and panics if it fails.
func MustGenerateKeyHex(length int) string {
	key, err := GenerateKeyHex(length)
	if err != nil {
		panic("session: failed to generate key: " + err.Error())
	}
	return key
}