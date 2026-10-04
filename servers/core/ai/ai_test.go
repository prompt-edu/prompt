package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerURL(t *testing.T) {
	t.Setenv("AI_ENABLED", "")
	assert.Empty(t, ServerURL(), "AI is off unless it is switched on")

	t.Setenv("AI_ENABLED", "true")
	t.Setenv("ENVIRONMENT", "development")
	assert.Equal(t, "http://localhost:8092/ai/api", ServerURL())

	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("CORE_HOST", "prompt.example.org")
	assert.Equal(t, "https://prompt.example.org/ai/api", ServerURL(), "it sits behind the same host as core")
}
