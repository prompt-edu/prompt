package storage

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Without the trailing slash, deleting one phase's prefix would also match every phase
// whose ID merely starts with the same characters.
func TestCoursePhasePrefixEndsInASlash(t *testing.T) {
	prefix := CoursePhasePrefix(uuid.New())

	assert.True(t, strings.HasSuffix(prefix, "/"), "prefix %q must end in a slash", prefix)
	assert.NoError(t, ValidateDeletePrefix(prefix), "the prefix phase deletion passes must be accepted")
}

func TestValidateDeletePrefixRejectsBroadOrUnterminatedPrefixes(t *testing.T) {
	coursePhaseID := uuid.NewString()
	tests := []struct {
		name   string
		prefix string
	}{
		{name: "empty prefix matches the whole bucket", prefix: ""},
		{name: "bare slash", prefix: "/"},
		{name: "top-level directory spans every course phase", prefix: "presentations/"},
		{name: "bare course phase ID", prefix: coursePhaseID},
		{name: "course phase directory without trailing slash", prefix: "presentations/" + coursePhaseID},
		{name: "empty segment", prefix: "presentations//"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.ErrorIs(t, ValidateDeletePrefix(test.prefix), ErrInvalidPrefix)
		})
	}
}

// The guard runs before any S3 call, so a zero adapter without a client proves no request
// is sent for a rejected prefix.
func TestS3AdapterDeletePrefixRejectsInvalidPrefixBeforeCallingS3(t *testing.T) {
	err := (&S3Adapter{}).DeletePrefix(context.Background(), "presentations/"+uuid.NewString())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPrefix)
}
