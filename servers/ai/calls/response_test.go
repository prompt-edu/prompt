package calls

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeJSON(t *testing.T) {
	summary := Summarize([]byte(`{"model":"m1","choices":[{"message":{"content":"Hello"},"finish_reason":"stop"}],
		"usage":{"prompt_tokens":42,"completion_tokens":7}}`), false)

	assert.Equal(t, "m1", summary.Model)
	assert.Equal(t, "Hello", summary.Text)
	assert.Equal(t, "stop", summary.FinishReason)
	require.NotNil(t, summary.PromptTokens)
	assert.EqualValues(t, 42, *summary.PromptTokens)
	assert.EqualValues(t, 7, *summary.CompletionTokens)
}

func TestSummarizeStream(t *testing.T) {
	stream := "data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"role\":\"assistant\",\"content\":\"\"},\"finish_reason\":null}]}\n\n" +
		"data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"Hel\"},\"finish_reason\":null}]}\n\n" +
		"data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"lo\"},\"finish_reason\":\"stop\"}]}\n\n" +
		"data: {\"model\":\"m1\",\"choices\":[],\"usage\":{\"prompt_tokens\":42,\"completion_tokens\":7}}\n\n" +
		"data: [DONE]\n\n"

	summary := Summarize([]byte(stream), true)

	assert.Equal(t, "Hello", summary.Text)
	assert.Equal(t, "stop", summary.FinishReason)
	require.NotNil(t, summary.CompletionTokens)
	assert.EqualValues(t, 7, *summary.CompletionTokens)
}

func TestSummarizeCutOffStream(t *testing.T) {
	summary := Summarize([]byte("data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\ndata: {\"mod"), true)

	assert.Equal(t, "Hel", summary.Text)
	assert.Empty(t, summary.FinishReason)
	assert.Nil(t, summary.PromptTokens)
}
