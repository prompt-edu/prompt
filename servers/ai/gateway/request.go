package gateway

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	errInvalidBody     = errors.New("the request body must be a JSON object")
	errTooManySubjects = errors.New("too many subjects")
)

type preparedRequest struct {
	body     []byte
	model    string
	params   []byte
	streamed bool
}

var recordedParams = []string{"temperature", "max_tokens", "seed", "response_format"}

func prepare(body []byte) (preparedRequest, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil || fields == nil {
		return preparedRequest{}, errInvalidBody
	}
	var request preparedRequest
	_ = json.Unmarshal(fields["model"], &request.model)
	_ = json.Unmarshal(fields["stream"], &request.streamed)

	delete(fields, "user")
	if request.streamed {
		var options map[string]json.RawMessage
		_ = json.Unmarshal(fields["stream_options"], &options)
		if options == nil {
			options = map[string]json.RawMessage{}
		}
		options["include_usage"] = json.RawMessage("true")
		fields["stream_options"], _ = json.Marshal(options)
	}

	params := map[string]json.RawMessage{}
	for _, name := range recordedParams {
		if value, ok := fields[name]; ok {
			params[name] = value
		}
	}
	request.params, _ = json.Marshal(params)

	var err error
	request.body, err = json.Marshal(fields)
	return request, err
}

func parseSubjects(header string) ([]uuid.UUID, error) {
	if strings.TrimSpace(header) == "" {
		return nil, nil
	}
	raws := strings.Split(header, ",")
	if len(raws) > maxSubjects {
		return nil, errTooManySubjects
	}
	seen := map[uuid.UUID]bool{}
	var subjects []uuid.UUID
	for _, raw := range raws {
		id, err := uuid.Parse(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		if !seen[id] {
			seen[id] = true
			subjects = append(subjects, id)
		}
	}
	return subjects, nil
}
