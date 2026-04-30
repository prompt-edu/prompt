package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

// generateTestKey sets a random 32-byte base64-encoded ENCRYPTION_KEY in the environment.
func generateTestKey(t *testing.T) {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(key))
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	generateTestKey(t)

	original := []byte(`{"api_key":"secret-api-key","base_url":"https://example.com"}`)
	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(original) {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, original)
	}
}

func TestEncryptProducesUniqueNonces(t *testing.T) {
	generateTestKey(t)

	plaintext := []byte("same plaintext")
	seen := make(map[string]struct{})

	for i := 0; i < 100; i++ {
		enc, err := Encrypt(plaintext)
		if err != nil {
			t.Fatalf("Encrypt failed at iteration %d: %v", i, err)
		}
		key := string(enc[:12]) // first 12 bytes are the nonce
		if _, exists := seen[key]; exists {
			t.Fatal("nonce collision detected - entropy source is broken")
		}
		seen[key] = struct{}{}
	}
}

func TestDecryptWithWrongKeyFails(t *testing.T) {
	generateTestKey(t)

	original := []byte("sensitive credentials")
	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Switch to a different key.
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		t.Fatalf("failed to generate wrong key: %v", err)
	}
	t.Setenv("ENCRYPTION_KEY", base64.StdEncoding.EncodeToString(newKey))

	_, err = Decrypt(encrypted)
	if err == nil {
		t.Error("expected Decrypt to fail with wrong key, but it succeeded")
	}
}

func TestDecryptCorruptedCiphertextFails(t *testing.T) {
	generateTestKey(t)

	original := []byte("sensitive credentials")
	encrypted, err := Encrypt(original)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Corrupt a byte in the ciphertext portion.
	encrypted[len(encrypted)-1] ^= 0xFF

	_, err = Decrypt(encrypted)
	if err == nil {
		t.Error("expected Decrypt to fail with corrupted ciphertext, but it succeeded")
	}
}

func TestDecryptShortDataFails(t *testing.T) {
	generateTestKey(t)

	_, err := Decrypt([]byte("short"))
	if err == nil {
		t.Error("expected Decrypt to fail on short data, but it succeeded")
	}
}

func TestEncryptWithoutKeyFails(t *testing.T) {
	os.Unsetenv("ENCRYPTION_KEY")

	_, err := Encrypt([]byte("data"))
	if err == nil {
		t.Error("expected Encrypt to fail without key, but it succeeded")
	}
}
