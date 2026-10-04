package gateway

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareStripsUserAndForcesUsage(t *testing.T) {
	request, err := prepare([]byte(`{"model":"m1","stream":true,"stream_options":{"include_usage":false,"extra":1},
		"user":"student@tum.de","temperature":0.2,"messages":[{"role":"user","content":"hi"}]}`))
	require.NoError(t, err)

	var forwarded map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(request.body, &forwarded))
	assert.NotContains(t, forwarded, "user")
	assert.JSONEq(t, `{"include_usage":true,"extra":1}`, string(forwarded["stream_options"]))
	assert.JSONEq(t, `[{"role":"user","content":"hi"}]`, string(forwarded["messages"]))
	assert.Equal(t, "m1", request.model)
	assert.True(t, request.streamed)
	assert.JSONEq(t, `{"temperature":0.2}`, string(request.params))
}

func TestPrepareLeavesNonStreamedRequestWithoutStreamOptions(t *testing.T) {
	request, err := prepare([]byte(`{"model":"m1","stream_options":null,"messages":[]}`))
	require.NoError(t, err)
	assert.False(t, request.streamed)
	assert.JSONEq(t, `{"model":"m1","stream_options":null,"messages":[]}`, string(request.body))
}

func TestPrepareRejectsNonObjects(t *testing.T) {
	for _, body := range []string{``, `null`, `[]`, `"x"`, `{"model":`} {
		_, err := prepare([]byte(body))
		assert.ErrorIs(t, err, errInvalidBody, body)
	}
}

func TestParseSubjects(t *testing.T) {
	first, second := uuid.New(), uuid.New()

	subjects, err := parseSubjects(first.String() + ", " + second.String() + "," + first.String())
	require.NoError(t, err)
	assert.Equal(t, []uuid.UUID{first, second}, subjects)

	subjects, err = parseSubjects("")
	require.NoError(t, err)
	assert.Empty(t, subjects)

	_, err = parseSubjects(first.String() + ",not-a-uuid")
	assert.Error(t, err)

	_, err = parseSubjects(strings.Repeat(first.String()+",", maxSubjects) + second.String())
	assert.ErrorIs(t, err, errTooManySubjects)
}
