package calls

import (
	"bufio"
	"bytes"
	"encoding/json"
	"strings"
)

type Summary struct {
	Model            string
	FinishReason     string
	PromptTokens     *int32
	CompletionTokens *int32
	Text             string
}

type completionChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		FinishReason *string `json:"finish_reason"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
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
	for _, choice := range chunk.Choices {
		s.Text += choice.Message.Content + choice.Delta.Content
		if choice.FinishReason != nil && *choice.FinishReason != "" {
			s.FinishReason = *choice.FinishReason
		}
	}
	if chunk.Usage != nil {
		s.PromptTokens = &chunk.Usage.PromptTokens
		s.CompletionTokens = &chunk.Usage.CompletionTokens
	}
}
