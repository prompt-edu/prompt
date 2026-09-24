package testutils

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/prompt-edu/prompt/servers/presentation/storage"
)

// The fake has to refuse what the S3 adapter refuses, or a test could pass with a prefix
// that fails in production.
func TestFakeStorageDeletePrefixRejectsInvalidPrefix(t *testing.T) {
	fake := NewFakeStorage()
	key := storage.CoursePhasePrefix(uuid.New()) + "material.pdf"
	fake.Put(key, "application/pdf", 1)

	for _, prefix := range []string{"", "presentations/"} {
		assert.ErrorIs(t, fake.DeletePrefix(context.Background(), prefix), storage.ErrInvalidPrefix)
	}
	assert.True(t, fake.Has(key), "a rejected prefix must not delete anything")
	assert.Empty(t, fake.Deleted)
}
