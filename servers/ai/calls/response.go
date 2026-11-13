package calls

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"

	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
)

type Summary struct {
	Model string
	// SystemFingerprint identifies the provider's model build, where the provider reports one.
	SystemFingerprint string
	FinishReason      string
	PromptTokens      *int32
	CompletionTokens  *int32
	Choices           []callDTO.Choice
}

// maxToolCalls bounds the tool call index a provider's stream can make us allocate.
const maxToolCalls = 128

type messagePart struct {
	Content   string `json:"content"`
	Refusal   string `json:"refusal"`
	ToolCalls []struct {
		Index    int    `json:"index"`
		ID       string `json:"id"`
		Function struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"function"`
	} `json:"tool_calls"`
}

type completionChunk struct {
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index        int         `json:"index"`
		FinishReason *string     `json:"finish_reason"`
		Message      messagePart `json:"message"`
		Delta        messagePart `json:"delta"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int32 `json:"prompt_tokens"`
		CompletionTokens int32 `json:"completion_tokens"`
	} `json:"usage"`
}

// Summarize reads a JSON response or a possibly cut-off SSE stream, skipping what does not parse.
func Summarize(raw []byte, streamed bool) Summary {
	if !streamed {
		var chunk completionChunk
		if json.Unmarshal(raw, &chunk) != nil {
			return Summary{}
		}
		var summary Summary
		summary.add(chunk)
		return summary
	}

	var summary Summary
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 0, 64*1024), len(raw)+1)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data:")
		if !ok {
			continue
		}
		var chunk completionChunk
		if json.Unmarshal([]byte(strings.TrimSpace(data)), &chunk) == nil {
			summary.add(chunk)
		}
	}
	return summary
}

func (s *Summary) add(chunk completionChunk) {
	if chunk.Model != "" {
		s.Model = chunk.Model
	}
	if chunk.SystemFingerprint != "" {
		s.SystemFingerprint = chunk.SystemFingerprint
	}
	for _, part := range chunk.Choices {
		choice := s.choice(part.Index)
		choice.Content += part.Message.Content + part.Delta.Content
		choice.Refusal += part.Message.Refusal + part.Delta.Refusal
		for _, call := range part.Message.ToolCalls {
			choice.ToolCalls = append(choice.ToolCalls, callDTO.ToolCall{ID: call.ID, Name: call.Function.Name, Arguments: call.Function.Arguments})
		}
		// A streamed tool call arrives in fragments that share its index.
		for _, call := range part.Delta.ToolCalls {
			if call.Index < 0 || call.Index >= maxToolCalls {
				continue
			}
			for len(choice.ToolCalls) <= call.Index {
				choice.ToolCalls = append(choice.ToolCalls, callDTO.ToolCall{})
			}
			assembled := &choice.ToolCalls[call.Index]
			assembled.ID += call.ID
			assembled.Name += call.Function.Name
			assembled.Arguments += call.Function.Arguments
		}
		if part.FinishReason != nil && *part.FinishReason != "" {
			s.FinishReason = *part.FinishReason
		}
	}
	if chunk.Usage != nil {
		s.PromptTokens = &chunk.Usage.PromptTokens
		s.CompletionTokens = &chunk.Usage.CompletionTokens
	}
}

func (s *Summary) choice(index int) *callDTO.Choice {
	for i := range s.Choices {
		if s.Choices[i].Index == index {
			return &s.Choices[i]
		}
	}
	s.Choices = append(s.Choices, callDTO.Choice{Index: index})
	return &s.Choices[len(s.Choices)-1]
}
