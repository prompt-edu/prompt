package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"

	promptSDK "github.com/prompt-edu/prompt-sdk"
)

var ErrEmptyKey = errors.New("AI_ENCRYPTION_KEY environment variable is not set")

var ErrShortCiphertext = errors.New("ciphertext too short")

var ErrPlaceholderKey = errors.New("AI_ENCRYPTION_KEY is still the placeholder from .env.template, which is public; generate one with: openssl rand -base64 32")

// Shipped in the public .env templates, so only accepted with DEBUG=true.
const placeholderKey = "bG9jYWwtYWkta2V5LW5vdC1hLXJlYWwtc2VjcmV0ISE="

func ValidateKey() error {
	_, err := getKey()
	return err
}

func getKey() ([]byte, error) {
	raw := promptSDK.GetEnv("AI_ENCRYPTION_KEY", "")
	if raw == "" {
		return nil, ErrEmptyKey
	}
	if raw == placeholderKey && promptSDK.GetEnv("DEBUG", "false") != "true" {
		return nil, ErrPlaceholderKey
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, errors.New("AI_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}
	return key, nil
}

func newGCM() (cipher.AEAD, error) {
	key, err := getKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func Encrypt(plaintext []byte) ([]byte, error) {
	gcm, err := newGCM()
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func Decrypt(data []byte) ([]byte, error) {
	gcm, err := newGCM()
	if err != nil {
		return nil, err
	}
	if len(data) < gcm.NonceSize() {
		return nil, ErrShortCiphertext
	}
	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
