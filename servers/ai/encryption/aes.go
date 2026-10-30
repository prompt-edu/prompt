package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
)

var ErrEmptyKey = errors.New("AI_ENCRYPTION_KEYS environment variable is not set")

var ErrShortCiphertext = errors.New("ciphertext too short")

var ErrPlaceholderKey = errors.New("AI_ENCRYPTION_KEYS still holds the placeholder from .env.template, which is public; generate a key with: openssl rand -base64 32")

var ErrUnknownKeyVersion = errors.New("ciphertext was sealed with an unknown key version")

// Shipped in the public .env templates, so only accepted with DEBUG=true.
const placeholderKey = "bG9jYWwtYWkta2V5LW5vdC1hLXJlYWwtc2VjcmV0ISE="

// Box seals with the highest key version and opens by the version byte that leads every ciphertext.
type Box struct {
	aeads   map[byte]cipher.AEAD
	current byte
}

// New reads AI_ENCRYPTION_KEYS as "1:<key>,2:<key>", each key base64 of 32 bytes.
func New() (*Box, error) {
	raw := promptSDK.GetEnv("AI_ENCRYPTION_KEYS", "")
	if strings.TrimSpace(raw) == "" {
		return nil, ErrEmptyKey
	}
	box := &Box{aeads: map[byte]cipher.AEAD{}}
	for _, entry := range strings.Split(raw, ",") {
		version, key, err := parseEntry(strings.TrimSpace(entry))
		if err != nil {
			return nil, err
		}
		if _, duplicate := box.aeads[version]; duplicate {
			return nil, fmt.Errorf("AI_ENCRYPTION_KEYS lists key version %d twice", version)
		}
		block, err := aes.NewCipher(key)
		if err != nil {
			return nil, err
		}
		if box.aeads[version], err = cipher.NewGCM(block); err != nil {
			return nil, err
		}
		box.current = max(box.current, version)
	}
	return box, nil
}

func parseEntry(entry string) (byte, []byte, error) {
	rawVersion, rawKey, ok := strings.Cut(entry, ":")
	version, err := strconv.ParseUint(rawVersion, 10, 8)
	if !ok || err != nil || version == 0 {
		return 0, nil, errors.New("AI_ENCRYPTION_KEYS entries must look like <version 1-255>:<base64 key>")
	}
	if rawKey == placeholderKey && promptSDK.GetEnv("DEBUG", "false") != "true" {
		return 0, nil, ErrPlaceholderKey
	}
	key, err := base64.StdEncoding.DecodeString(rawKey)
	if err != nil {
		return 0, nil, err
	}
	if len(key) != 32 {
		return 0, nil, fmt.Errorf("AI_ENCRYPTION_KEYS key version %d must decode to exactly 32 bytes", version)
	}
	return byte(version), key, nil
}

// Purpose tells a call's request and response apart, so the two cannot be swapped.
type Purpose string

const (
	Request  Purpose = "request"
	Response Purpose = "response"
)

func additionalData(callID uuid.UUID, purpose Purpose) []byte {
	return append(callID[:], purpose...)
}

// Encrypt binds the ciphertext to its call and purpose, so content moved to another call or field does not decrypt.
func (b *Box) Encrypt(callID uuid.UUID, purpose Purpose, plaintext []byte) ([]byte, error) {
	aead := b.aeads[b.current]
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := append([]byte{b.current}, nonce...)
	return aead.Seal(sealed, nonce, plaintext, additionalData(callID, purpose)), nil
}

func (b *Box) Decrypt(callID uuid.UUID, purpose Purpose, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, ErrShortCiphertext
	}
	aead, ok := b.aeads[data[0]]
	if !ok {
		return nil, ErrUnknownKeyVersion
	}
	nonceEnd := 1 + aead.NonceSize()
	if len(data) < nonceEnd {
		return nil, ErrShortCiphertext
	}
	return aead.Open(nil, data[1:nonceEnd], data[nonceEnd:], additionalData(callID, purpose))
}
