package auditLog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseIngestKeys_SkipsEmptyAndSupportsRotation(t *testing.T) {
	keys := parseIngestKeys("interview:old, interview:new, assessment:sekret, broken, empty:")

	// A service listed twice keeps both keys so a key can be rotated without
	// downtime.
	require.Equal(t, []string{"old", "new"}, keys["interview"])
	require.Equal(t, []string{"sekret"}, keys["assessment"])
	// An empty key is dropped: subtle.ConstantTimeCompare("", "") returns 1, so a
	// blank configured key would otherwise authenticate every request.
	assert.NotContains(t, keys, "empty")
	// A pair without a ':' is ignored.
	assert.NotContains(t, keys, "broken")
}

func TestMatchesAnyKey(t *testing.T) {
	accepted := []string{"old", "new"}
	assert.True(t, matchesAnyKey("old", accepted))
	assert.True(t, matchesAnyKey("new", accepted))
	assert.False(t, matchesAnyKey("wrong", accepted))
	// No configured key (or an empty accepted list) can never authenticate, in
	// particular not with an empty token.
	assert.False(t, matchesAnyKey("", nil))
	assert.False(t, matchesAnyKey("", []string{}))
}
