package calls

import (
	"testing"

	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizeJSON(t *testing.T) {
	summary := Summarize([]byte(`{"model":"m1","system_fingerprint":"fp_1","choices":[{"message":{"content":"Hello"},"finish_reason":"stop"}],
		"usage":{"prompt_tokens":42,"completion_tokens":7}}`), false)

	assert.Equal(t, "m1", summary.Model)
	assert.Equal(t, "fp_1", summary.SystemFingerprint)
	assert.Equal(t, []callDTO.Choice{{Content: "Hello"}}, summary.Choices)
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

	assert.Equal(t, []callDTO.Choice{{Content: "Hello"}}, summary.Choices)
	assert.Equal(t, "stop", summary.FinishReason)
	require.NotNil(t, summary.CompletionTokens)
	assert.EqualValues(t, 7, *summary.CompletionTokens)
}

func TestSummarizeCutOffStream(t *testing.T) {
	summary := Summarize([]byte("data: {\"model\":\"m1\",\"choices\":[{\"delta\":{\"content\":\"Hel\"}}]}\n\ndata: {\"mod"), true)

	assert.Equal(t, []callDTO.Choice{{Content: "Hel"}}, summary.Choices)
	assert.Empty(t, summary.FinishReason)
	assert.Nil(t, summary.PromptTokens)
}

func TestSummarizeKeepsChoicesApart(t *testing.T) {
	stream := "data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Fir\"}}]}\n\n" +
		"data: {\"choices\":[{\"index\":1,\"delta\":{\"content\":\"Sec\"}}]}\n\n" +
		"data: {\"choices\":[{\"index\":0,\"delta\":{\"content\":\"st\"}}]}\n\n" +
		"data: {\"choices\":[{\"index\":1,\"delta\":{\"content\":\"ond\"}}]}\n\n"

	summary := Summarize([]byte(stream), true)

	assert.Equal(t, []callDTO.Choice{{Index: 0, Content: "First"}, {Index: 1, Content: "Second"}}, summary.Choices)
}

func TestSummarizeAssemblesStreamedToolCallsAndRefusals(t *testing.T) {
	stream := "data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call_1\",\"function\":{\"name\":\"grade\",\"arguments\":\"{\\\"sc\"}}]}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"function\":{\"arguments\":\"ore\\\":2}\"}}]}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":99999999,\"function\":{\"name\":\"ignored\"}}]}}]}\n\n" +
		"data: {\"choices\":[{\"delta\":{\"refusal\":\"I can't\"}}]}\n\n"

	summary := Summarize([]byte(stream), true)

	require.Len(t, summary.Choices, 1)
	assert.Equal(t, "I can't", summary.Choices[0].Refusal)
	assert.Equal(t, []callDTO.ToolCall{{ID: "call_1", Name: "grade", Arguments: `{"score":2}`}}, summary.Choices[0].ToolCalls)
}

func TestSummarizeJSONWithSeveralToolCalls(t *testing.T) {
	summary := Summarize([]byte(`{"choices":[{"index":0,"message":{"content":null,"tool_calls":[
		{"id":"a","function":{"name":"one","arguments":"{}"}},{"id":"b","function":{"name":"two","arguments":"{}"}}]}}]}`), false)

	require.Len(t, summary.Choices, 1)
	assert.Equal(t, []callDTO.ToolCall{{ID: "a", Name: "one", Arguments: "{}"}, {ID: "b", Name: "two", Arguments: "{}"}},
		summary.Choices[0].ToolCalls)
}
