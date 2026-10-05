package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	assert.Equal(t, "http://localhost:8092/ai/api", Module("development", "localhost:8080").BaseURL)
	assert.Equal(t, "https://prompt.example.org/ai/api", Module("production", "prompt.example.org").BaseURL,
		"it sits behind the same host as core")
	assert.Equal(t, ServiceName, Module("production", "prompt.example.org").Name)
}
