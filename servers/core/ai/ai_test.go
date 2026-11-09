package ai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModule(t *testing.T) {
	assert.Equal(t, "http://localhost:8092/ai/api", Module("http://localhost:8092", "localhost:8080").BaseURL)
	assert.Equal(t, "https://prompt.example.org/ai/api", Module("", "prompt.example.org").BaseURL,
		"without SERVER_AI_HOST it sits behind the same host as core")
	assert.Equal(t, ServiceName, Module("", "prompt.example.org").Name)
}
