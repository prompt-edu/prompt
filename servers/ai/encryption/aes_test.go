package encryption

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testKey      = "ZTJlLWFpLXRlc3Qta2V5LW5vdC1hLXJlYWwtc2VjcmU="
	rotatedKey   = "cm90YXRlZC1haS10ZXN0LWtleS1ub3QtYS1zZWNyZXQ="
	testKeyEntry = "1:" + testKey
)

func newTestBox(t *testing.T) *Box {
	t.Setenv("AI_ENCRYPTION_KEYS", testKeyEntry)
	box, err := New()
	require.NoError(t, err)
	return box
}

func TestEncryptRoundTrip(t *testing.T) {
	box := newTestBox(t)
	callID := uuid.New()

	sealed, err := box.Encrypt(callID, Request, []byte("logos-key-1234"))
	require.NoError(t, err)
	assert.NotContains(t, string(sealed), "logos-key-1234")
	assert.Equal(t, byte(1), sealed[0])

	opened, err := box.Decrypt(callID, Request, sealed)
	require.NoError(t, err)
	assert.Equal(t, "logos-key-1234", string(opened))
}

func TestDecryptRejectsTamperedCiphertext(t *testing.T) {
	box := newTestBox(t)
	callID := uuid.New()
	sealed, err := box.Encrypt(callID, Request, []byte("content"))
	require.NoError(t, err)
	sealed[len(sealed)-1] ^= 0xff

	_, err = box.Decrypt(callID, Request, sealed)
	assert.Error(t, err)
}

func TestDecryptRejectsContentOfAnotherCall(t *testing.T) {
	box := newTestBox(t)
	sealed, err := box.Encrypt(uuid.New(), Request, []byte("content"))
	require.NoError(t, err)

	_, err = box.Decrypt(uuid.New(), Request, sealed)
	assert.Error(t, err)
}

func TestDecryptRejectsTheOtherPurpose(t *testing.T) {
	box := newTestBox(t)
	callID := uuid.New()
	sealed, err := box.Encrypt(callID, Request, []byte("content"))
	require.NoError(t, err)

	_, err = box.Decrypt(callID, Response, sealed)
	assert.Error(t, err, "a request swapped into the response field must not decrypt")
}

func TestDecryptRejectsUnknownKeyVersion(t *testing.T) {
	box := newTestBox(t)
	callID := uuid.New()
	sealed, err := box.Encrypt(callID, Request, []byte("content"))
	require.NoError(t, err)
	sealed[0] = 2

	_, err = box.Decrypt(callID, Request, sealed)
	assert.ErrorIs(t, err, ErrUnknownKeyVersion)
}

func TestRotatedKeySealsNewContentAndOpensOldContent(t *testing.T) {
	callID := uuid.New()
	old := newTestBox(t)
	sealedBefore, err := old.Encrypt(callID, Request, []byte("before"))
	require.NoError(t, err)

	t.Setenv("AI_ENCRYPTION_KEYS", testKeyEntry+", 2:"+rotatedKey)
	rotated, err := New()
	require.NoError(t, err)
	sealedAfter, err := rotated.Encrypt(callID, Request, []byte("after"))
	require.NoError(t, err)
	assert.Equal(t, byte(2), sealedAfter[0], "new content uses the highest version")

	opened, err := rotated.Decrypt(callID, Request, sealedBefore)
	require.NoError(t, err)
	assert.Equal(t, "before", string(opened))
	_, err = old.Decrypt(callID, Request, sealedAfter)
	assert.ErrorIs(t, err, ErrUnknownKeyVersion)
}

func TestNewValidatesTheKeys(t *testing.T) {
	t.Setenv("DEBUG", "false")

	for keys, want := range map[string]error{
		"":                    ErrEmptyKey,
		"1:" + placeholderKey: ErrPlaceholderKey,
	} {
		t.Setenv("AI_ENCRYPTION_KEYS", keys)
		_, err := New()
		assert.ErrorIs(t, err, want, keys)
	}
	for _, keys := range []string{
		testKey,
		"0:" + testKey,
		"256:" + testKey,
		"1:c2hvcnQ=",
		testKeyEntry + ",1:" + rotatedKey,
	} {
		t.Setenv("AI_ENCRYPTION_KEYS", keys)
		_, err := New()
		assert.Error(t, err, keys)
	}

	t.Setenv("DEBUG", "true")
	t.Setenv("AI_ENCRYPTION_KEYS", "1:"+placeholderKey)
	_, err := New()
	assert.NoError(t, err, "the placeholder is accepted in a debug deployment")
}
