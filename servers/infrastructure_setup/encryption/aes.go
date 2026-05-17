// Package encryption provides AES-256-GCM encryption for storing provider credentials.
//
// The stored format is: 12-byte nonce || ciphertext.
// A fresh random nonce is generated for each Encrypt call.
// The encryption key is read from the ENCRYPTION_KEY environment variable,
// which must be a base64-encoded 32-byte key.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	promptSDK "github.com/prompt-edu/prompt-sdk"
)

// ErrEmptyKey is returned when no encryption key is configured.
var ErrEmptyKey = errors.New("ENCRYPTION_KEY environment variable is not set")

// ErrShortCiphertext is returned when the ciphertext is too short to contain a nonce.
var ErrShortCiphertext = errors.New("ciphertext too short")

// getKey reads and decodes the AES key from the environment.
func getKey() ([]byte, error) {
	raw := promptSDK.GetEnv("ENCRYPTION_KEY", "")
	if raw == "" {
		return nil, ErrEmptyKey
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns nonce||ciphertext.
func Encrypt(plaintext []byte) ([]byte, error) {
	key, err := getKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize()) // 12 bytes for GCM
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Prepend nonce to ciphertext so Decrypt knows the nonce.
	result := make([]byte, len(nonce)+len(ciphertext))
	copy(result, nonce)
	copy(result[len(nonce):], ciphertext)
	return result, nil
}

// Decrypt decrypts data produced by Encrypt (expects nonce||ciphertext format).
func Decrypt(data []byte) ([]byte, error) {
	key, err := getKey()
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrShortCiphertext
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
