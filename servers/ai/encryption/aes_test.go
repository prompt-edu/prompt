package encryption

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testKey = "ZTJlLWFpLXRlc3Qta2V5LW5vdC1hLXJlYWwtc2VjcmU="

func TestEncryptRoundTrip(t *testing.T) {
	t.Setenv("AI_ENCRYPTION_KEY", testKey)

	sealed, err := Encrypt([]byte("logos-key-1234"))
	require.NoError(t, err)
	assert.NotContains(t, string(sealed), "logos-key-1234")

	opened, err := Decrypt(sealed)
	require.NoError(t, err)
	assert.Equal(t, "logos-key-1234", string(opened))
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	t.Setenv("AI_ENCRYPTION_KEY", testKey)

	sealed, err := Encrypt([]byte("content"))
	require.NoError(t, err)
	sealed[len(sealed)-1] ^= 0xff

	_, err = Decrypt(sealed)
	assert.Error(t, err)
}

func TestValidateKey(t *testing.T) {
	t.Setenv("DEBUG", "false")

	t.Setenv("AI_ENCRYPTION_KEY", "")
	assert.ErrorIs(t, ValidateKey(), ErrEmptyKey)

	t.Setenv("AI_ENCRYPTION_KEY", placeholderKey)
	assert.ErrorIs(t, ValidateKey(), ErrPlaceholderKey)

	t.Setenv("DEBUG", "true")
	assert.NoError(t, ValidateKey(), "the placeholder is accepted in a debug deployment")

	t.Setenv("AI_ENCRYPTION_KEY", "c2hvcnQ=")
	assert.Error(t, ValidateKey(), "a key that is not 32 bytes long is refused")
}
