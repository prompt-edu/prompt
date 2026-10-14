package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	"github.com/prompt-edu/prompt/servers/ai/calls/callDTO"
	"github.com/prompt-edu/prompt/servers/ai/gateway"
	"github.com/prompt-edu/prompt/servers/ai/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testEncryptionKey = "ZTJlLWFpLXRlc3Qta2V5LW5vdC1hLXJlYWwtc2VjcmU="
	testLogosKey      = "logos-test-key-7f3a"
	testFeature       = "assessment.action_item_suggestions"
	summaryPrompt     = "Summarize the peer feedback"
	summaryAnswer     = "The team communicates well and should record its decisions earlier."
)

type AIServerSuite struct {
	suite.Suite
	ctx      context.Context
	db       *testutils.TestDB
	aimock   *testutils.AIMock
	identity *testutils.Identity
	server   *httptest.Server
	cleanups []func()

	adminID    string
	lecturerID string
}

func (s *AIServerSuite) SetupSuite() {
	s.ctx = context.Background()
	s.T().Setenv("AI_ENCRYPTION_KEY", testEncryptionKey)
	gin.SetMode(gin.TestMode)

	testDB, stopDB, err := testutils.SetupTestDB(s.ctx)
	s.Require().NoError(err)
	s.cleanups = append(s.cleanups, stopDB)
	s.db = testDB

	aimock, stopAIMock, err := testutils.StartAIMock(s.ctx)
	s.Require().NoError(err)
	s.cleanups = append(s.cleanups, stopAIMock)
	s.aimock = aimock

	identity, stopIdentity, err := testutils.StartIdentity()
	s.Require().NoError(err)
	s.cleanups = append(s.cleanups, stopIdentity)
	s.identity = identity

	providerURL, err := url.Parse(aimock.BaseURL + "/v1")
	s.Require().NoError(err)
	router := gin.New()
	setupRouter(router, testDB.Conn, config{
		providerURL:           providerURL,
		allowedModels:         []string{testutils.TestModel, "unlisted-local-model"},
		metadataRetentionDays: 1825,
		serverVersion:         "test",
		coreURL:               identity.CoreURL,
		clientHost:            "http://localhost:3000",
	})
	s.server = httptest.NewServer(router)
	s.cleanups = append(s.cleanups, s.server.Close)

	s.adminID = uuid.NewString()
	s.lecturerID = uuid.NewString()
}

func (s *AIServerSuite) TearDownSuite() {
	for i := len(s.cleanups) - 1; i >= 0; i-- {
		s.cleanups[i]()
	}
}

func (s *AIServerSuite) SetupTest() {
	s.Require().NoError(s.aimock.ResetJournal())
}

func TestAIServerSuite(t *testing.T) {
	suite.Run(t, new(AIServerSuite))
}

func (s *AIServerSuite) admin() string { return s.identity.Token(s.adminID, promptSDK.PromptAdmin) }

func (s *AIServerSuite) lecturer(coursePhaseID string) string {
	return s.identity.Token(s.lecturerID, testutils.LecturerRole(coursePhaseID))
}

func (s *AIServerSuite) editor(coursePhaseID string) string {
	return s.identity.Token(uuid.NewString(), testutils.EditorRole(coursePhaseID))
}

func phasePath(coursePhaseID string) string { return "/ai/api/course_phase/" + coursePhaseID }

func (s *AIServerSuite) send(method, path, token string, body any, headers map[string]string) (*http.Response, []byte) {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		s.Require().NoError(err)
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, s.server.URL+path, reader)
	s.Require().NoError(err)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", token)
	}
	for name, value := range headers {
		if value != "" {
			request.Header.Set(name, value)
		}
	}
	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)
	defer func() { _ = response.Body.Close() }()
	content, err := io.ReadAll(response.Body)
	s.Require().NoError(err)
	return response, content
}

// newPhase creates a course phase of the given type in the fake core.
func (s *AIServerSuite) newPhase(phaseType string) string {
	coursePhaseID := uuid.NewString()
	s.identity.PhaseTypes[coursePhaseID] = phaseType
	return coursePhaseID
}

// phaseHeaders are the headers a phase server sends: its own Logos key and the feature. An empty
// override leaves that header out.
func phaseHeaders(overrides map[string]string) map[string]string {
	headers := map[string]string{gateway.ProviderKeyHeader: testLogosKey, gateway.FeatureHeader: testFeature}
	for name, value := range overrides {
		headers[name] = value
	}
	return headers
}

func (s *AIServerSuite) chat(coursePhaseID, token string, body any, overrides map[string]string) (*http.Response, []byte) {
	return s.send(http.MethodPost, phasePath(coursePhaseID)+"/v1/chat/completions", token, body, phaseHeaders(overrides))
}

func completion(model, prompt string, stream bool) map[string]any {
	return map[string]any{
		"model":       model,
		"stream":      stream,
		"user":        "student@example.com",
		"temperature": 0.2,
		"messages":    []map[string]string{{"role": "user", "content": prompt}},
	}
}

type callRow struct {
	Outcome          string
	HTTPStatus       *int32
	ErrorCode        *string
	Feature          string
	ActorID          string
	ActorRole        string
	Issuer           string
	Template         *string
	ServedModel      *string
	FinishReason     *string
	PromptTokens     *int32
	CompletionTokens *int32
	Streamed         bool
	Completed        bool
	Subjects         int
	HasContent       bool
}

func (s *AIServerSuite) call(callID string) callRow {
	var row callRow
	err := s.db.Conn.QueryRow(s.ctx, `
		SELECT outcome, http_status, error_code, feature, actor_id, actor_role, issuer, template, served_model,
		       finish_reason, prompt_tokens, completion_tokens, streamed, completed_at IS NOT NULL,
		       (SELECT count(*) FROM ai_call_subject WHERE call_id = ai_call.id),
		       EXISTS (SELECT 1 FROM ai_call_content WHERE call_id = ai_call.id)
		FROM ai_call WHERE id = $1`, callID).Scan(
		&row.Outcome, &row.HTTPStatus, &row.ErrorCode, &row.Feature, &row.ActorID, &row.ActorRole, &row.Issuer, &row.Template,
		&row.ServedModel, &row.FinishReason, &row.PromptTokens, &row.CompletionTokens, &row.Streamed, &row.Completed,
		&row.Subjects, &row.HasContent)
	s.Require().NoError(err)
	return row
}

func (s *AIServerSuite) providerRequests() []testutils.JournalEntry {
	entries, err := s.aimock.Journal()
	s.Require().NoError(err)
	return entries
}

func (s *AIServerSuite) TestAuthorizationMatrix() {
	coursePhaseID := s.newPhase("Assessment")
	otherPhaseID := uuid.NewString()
	chat := completion(testutils.TestModel, summaryPrompt, false)

	unauthorized := map[string]string{
		"no token":                           "",
		"student":                            s.identity.Token(uuid.NewString(), coursePhaseID+"-Student"),
		"lecturer of another phase":          s.lecturer(otherPhaseID),
		"PromptLecturer without course role": s.identity.Token(uuid.NewString(), promptSDK.PromptLecturer),
	}
	for name, token := range unauthorized {
		expected := http.StatusForbidden
		if token == "" {
			expected = http.StatusUnauthorized
		}
		response, _ := s.chat(coursePhaseID, token, chat, nil)
		s.Equal(expected, response.StatusCode, "chat completions: %s", name)
		response, _ = s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls", token, nil, nil)
		s.Equal(expected, response.StatusCode, "calls: %s", name)
	}
	s.Empty(s.providerRequests(), "no unauthorized request may reach the provider")

	response, _ := s.chat(coursePhaseID, s.editor(coursePhaseID), chat, nil)
	s.Equal(http.StatusOK, response.StatusCode, "editors may call the model")
	response, _ = s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls", s.lecturer(coursePhaseID), nil, nil)
	s.Equal(http.StatusForbidden, response.StatusCode, "audit reads are admin-only")
	response, _ = s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls", s.admin(), nil, nil)
	s.Equal(http.StatusOK, response.StatusCode)
}

func (s *AIServerSuite) TestProviderKeyIsForwardedButNeverStored() {
	coursePhaseID := s.newPhase("Assessment")

	response, body := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, false), nil)
	s.Require().Equal(http.StatusOK, response.StatusCode, string(body))
	requests := s.providerRequests()
	s.Require().Len(requests, 1)
	s.Equal("[REDACTED]", requests[0].Headers["authorization"], "the phase's key authenticates the call")

	var stored string
	s.Require().NoError(s.db.Conn.QueryRow(s.ctx, `
		SELECT row_to_json(ai_call)::text || coalesce(row_to_json(content)::text, '')
		FROM ai_call LEFT JOIN ai_call_content content ON content.call_id = ai_call.id
		WHERE ai_call.id = $1`, response.Header.Get(gateway.CallIDHeader)).Scan(&stored))
	s.NotContains(stored, testLogosKey, "the AI server keeps no key")
}

func (s *AIServerSuite) TestCallWithoutProviderKeyIsDenied() {
	coursePhaseID := s.newPhase("Assessment")

	for _, key := range []string{"", "short", "with a space in it"} {
		response, body := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, false),
			map[string]string{gateway.ProviderKeyHeader: key})
		s.Equal(http.StatusBadRequest, response.StatusCode, "key %q", key)
		s.Contains(string(body), gateway.ProviderKeyHeader)
		s.Equal("provider_key_missing", *s.call(response.Header.Get(gateway.CallIDHeader)).ErrorCode)
	}
	response, _ := s.send(http.MethodGet, phasePath(coursePhaseID)+"/v1/models", s.lecturer(coursePhaseID), nil, nil)
	s.Equal(http.StatusBadRequest, response.StatusCode)
	s.Empty(s.providerRequests(), "without a key nothing reaches the provider")
}

func (s *AIServerSuite) TestNonStreamedCompletionIsRecorded() {
	coursePhaseID := s.newPhase("Assessment")
	subject := uuid.NewString()

	response, body := s.chat(coursePhaseID, s.lecturer(coursePhaseID),
		completion(testutils.TestModel, summaryPrompt, false), map[string]string{
			"X-Prompt-Template":         "action-items",
			"X-Prompt-Template-Version": "3",
			"X-Prompt-Subjects":         subject,
		})
	s.Require().Equal(http.StatusOK, response.StatusCode, string(body))
	s.Contains(string(body), summaryAnswer)

	row := s.call(response.Header.Get(gateway.CallIDHeader))
	s.Equal(calls.OutcomeSuccess, row.Outcome)
	s.Equal(testFeature, row.Feature)
	s.Equal(s.lecturerID, row.ActorID)
	s.Equal(promptSDK.CourseLecturer, row.ActorRole)
	s.Equal(s.identity.IssuerURL(), row.Issuer, "the issuer is taken from the token")
	s.Equal("action-items", *row.Template)
	s.Equal(testutils.TestModel, *row.ServedModel)
	s.Equal("stop", *row.FinishReason)
	s.EqualValues(42, *row.PromptTokens)
	s.EqualValues(12, *row.CompletionTokens)
	s.False(row.Streamed)
	s.Equal(1, row.Subjects)
	s.True(row.HasContent)

	requests := s.providerRequests()
	s.Require().Len(requests, 1)
	s.NotContains(requests[0].Body, "user", "no identifier may reach the provider")
	for name := range requests[0].Headers {
		s.NotContains(strings.ToLower(name), "x-prompt", "PROMPT context stays with the AI server")
	}
}

func (s *AIServerSuite) TestStreamedCompletionIsRecorded() {
	coursePhaseID := s.newPhase("Assessment")

	response, body := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, true), nil)
	s.Require().Equal(http.StatusOK, response.StatusCode, string(body))
	s.Contains(response.Header.Get("Content-Type"), "text/event-stream")
	s.Contains(string(body), "data: [DONE]")

	row := s.call(response.Header.Get(gateway.CallIDHeader))
	s.Equal(calls.OutcomeSuccess, row.Outcome)
	s.True(row.Streamed)
	s.Equal("stop", *row.FinishReason)
	s.EqualValues(42, *row.PromptTokens, "the usage chunk is recorded")
	s.EqualValues(12, *row.CompletionTokens)

	requests := s.providerRequests()
	s.Require().Len(requests, 1)
	s.NotContains(requests[0].Body, "user")
	s.Equal(map[string]any{"include_usage": true}, requests[0].Body["stream_options"])
}

func (s *AIServerSuite) TestCallerDisconnectStillCompletesTheCall() {
	coursePhaseID := s.newPhase("Assessment")
	ctx, cancel := context.WithCancel(s.ctx)
	defer cancel()

	encoded, err := json.Marshal(completion(testutils.TestModel, "Stream slowly", true))
	s.Require().NoError(err)
	request, err := http.NewRequestWithContext(ctx, http.MethodPost,
		s.server.URL+phasePath(coursePhaseID)+"/v1/chat/completions", bytes.NewReader(encoded))
	s.Require().NoError(err)
	request.Header.Set("Authorization", s.lecturer(coursePhaseID))
	for name, value := range phaseHeaders(nil) {
		request.Header.Set(name, value)
	}
	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)
	callID := response.Header.Get(gateway.CallIDHeader)

	stream := bufio.NewReader(response.Body)
	for {
		line, err := stream.ReadString('\n')
		s.Require().NoError(err)
		if strings.Contains(line, `"delta":{"content":"`) {
			break
		}
	}
	cancel()
	_ = response.Body.Close()

	s.Eventually(func() bool { return s.call(callID).Completed }, 10*time.Second, 100*time.Millisecond)
	row := s.call(callID)
	s.Equal(calls.OutcomeCancelled, row.Outcome)
	s.Equal("caller_disconnected", *row.ErrorCode)

	_, body := s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls/"+callID, s.admin(), nil, nil)
	var detail callDTO.Detail
	s.Require().NoError(json.Unmarshal(body, &detail))
	s.Require().NotNil(detail.Content)
	s.NotEmpty(detail.Content.ResponseText, "the partial output is kept")
	s.Less(len(detail.Content.ResponseText), len("This answer arrives in many small chunks, so a caller can disconnect long before it ends."))
}

func (s *AIServerSuite) TestProviderBreakingOffAbortsTheCallersStream() {
	coursePhaseID := s.newPhase("Assessment")
	encoded, err := json.Marshal(completion(testutils.TestModel, "Break off", true))
	s.Require().NoError(err)
	request, err := http.NewRequest(http.MethodPost, s.server.URL+phasePath(coursePhaseID)+"/v1/chat/completions", bytes.NewReader(encoded))
	s.Require().NoError(err)
	request.Header.Set("Authorization", s.lecturer(coursePhaseID))
	for name, value := range phaseHeaders(nil) {
		request.Header.Set(name, value)
	}

	response, err := http.DefaultClient.Do(request)
	s.Require().NoError(err)
	defer func() { _ = response.Body.Close() }()
	_, err = io.ReadAll(response.Body)
	s.Error(err, "a broken stream must not look complete to the caller")

	row := s.call(response.Header.Get(gateway.CallIDHeader))
	s.Equal(calls.OutcomeError, row.Outcome)
	s.Equal("stream_interrupted", *row.ErrorCode)
}

func (s *AIServerSuite) TestDisallowedModelIsDeniedBeforeTheProvider() {
	coursePhaseID := s.newPhase("Assessment")

	response, _ := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion("cloud-only-model", "Only offered by the cloud", false), nil)
	s.Equal(http.StatusBadRequest, response.StatusCode)

	row := s.call(response.Header.Get(gateway.CallIDHeader))
	s.Equal(calls.OutcomeDenied, row.Outcome)
	s.Equal("model_not_allowed", *row.ErrorCode)
	s.False(row.HasContent, "a denied call keeps no content")
	s.Empty(s.providerRequests())
}

func (s *AIServerSuite) TestFeatureMustBelongToThePhaseType() {
	coursePhaseID := s.newPhase("Interview")
	chat := completion(testutils.TestModel, summaryPrompt, false)

	response, _ := s.chat(coursePhaseID, s.lecturer(coursePhaseID), chat, nil)
	s.Equal(http.StatusBadRequest, response.StatusCode, "a feature of another phase type would get its retention")
	s.Equal("feature_not_allowed", *s.call(response.Header.Get(gateway.CallIDHeader)).ErrorCode)

	for _, name := range []string{"interview.summary", "adhoc", ""} {
		response, _ = s.chat(coursePhaseID, s.lecturer(coursePhaseID), chat, map[string]string{gateway.FeatureHeader: name})
		s.Equal(http.StatusBadRequest, response.StatusCode, "feature %q has no retention", name)
		s.Equal("feature_not_allowed", *s.call(response.Header.Get(gateway.CallIDHeader)).ErrorCode)
	}
	s.Empty(s.providerRequests())
}

func (s *AIServerSuite) TestFailedRecordBlocksTheProviderCall() {
	coursePhaseID := s.newPhase("Assessment")
	_, err := s.db.Conn.Exec(s.ctx, "ALTER TABLE ai_call_content ADD CONSTRAINT block_inserts CHECK (false) NOT VALID")
	s.Require().NoError(err)
	defer func() {
		_, err := s.db.Conn.Exec(s.ctx, "ALTER TABLE ai_call_content DROP CONSTRAINT block_inserts")
		s.Require().NoError(err)
	}()

	response, body := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, false), nil)
	s.Equal(http.StatusServiceUnavailable, response.StatusCode, string(body))
	s.Empty(s.providerRequests(), "no record, no model call")

	var calls int
	s.Require().NoError(s.db.Conn.QueryRow(s.ctx, "SELECT count(*) FROM ai_call WHERE course_phase_id = $1", coursePhaseID).Scan(&calls))
	s.Zero(calls, "the call row rolls back with its content")
}

func (s *AIServerSuite) TestModelsAreAllowedAndAvailable() {
	coursePhaseID := s.newPhase("Assessment")

	response, body := s.send(http.MethodGet, phasePath(coursePhaseID)+"/v1/models", s.editor(coursePhaseID), nil, phaseHeaders(nil))
	s.Require().Equal(http.StatusOK, response.StatusCode, string(body))
	var listing struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	s.Require().NoError(json.Unmarshal(body, &listing))
	s.Require().Len(listing.Data, 1, "cloud-only-model is not allowed and unlisted-local-model is not available")
	s.Equal(testutils.TestModel, listing.Data[0].ID)
}

func (s *AIServerSuite) TestAuditReadsAndEvents() {
	coursePhaseID := s.newPhase("Assessment")
	var callIDs []string
	for range 3 {
		response, _ := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, false), nil)
		s.Require().Equal(http.StatusOK, response.StatusCode)
		callIDs = append(callIDs, response.Header.Get(gateway.CallIDHeader))
	}

	_, body := s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls?limit=2", s.admin(), nil, nil)
	var page callDTO.Page
	s.Require().NoError(json.Unmarshal(body, &page))
	s.Require().Len(page.Calls, 2)
	s.Require().NotNil(page.NextCursor)
	s.Equal(callIDs[2], page.Calls[0].ID.String(), "newest first")
	_, body = s.send(http.MethodGet, phasePath(coursePhaseID)+"/calls?limit=2&cursorRequestedAt="+
		url.QueryEscape(page.NextCursor.RequestedAt.Format(time.RFC3339Nano))+"&cursorId="+page.NextCursor.ID.String(), s.admin(), nil, nil)
	var secondPage callDTO.Page
	s.Require().NoError(json.Unmarshal(body, &secondPage))
	s.Require().Len(secondPage.Calls, 1)
	s.Equal(callIDs[0], secondPage.Calls[0].ID.String())
	s.Nil(secondPage.NextCursor)

	callPath := phasePath(coursePhaseID) + "/calls/" + callIDs[0]
	response, _ := s.send(http.MethodPost, callPath+"/events", s.editor(coursePhaseID), map[string]any{"type": "shown"}, nil)
	s.Equal(http.StatusNotFound, response.StatusCode, "only the caller reports events on their call")
	response, _ = s.send(http.MethodPost, callPath+"/events", s.lecturer(coursePhaseID), map[string]any{"type": "shown"}, nil)
	s.Equal(http.StatusCreated, response.StatusCode)
	response, _ = s.send(http.MethodPost, callPath+"/events", s.lecturer(coursePhaseID), map[string]any{"type": "edited"}, nil)
	s.Equal(http.StatusBadRequest, response.StatusCode, "an edit needs its edit distance")
	response, _ = s.send(http.MethodPost, callPath+"/events", s.lecturer(coursePhaseID),
		map[string]any{"type": "edited", "editDistance": 17, "actionItemId": uuid.NewString()}, nil)
	s.Equal(http.StatusCreated, response.StatusCode)
	response, _ = s.send(http.MethodPost, callPath+"/events", s.lecturer(coursePhaseID), map[string]any{"type": "content_viewed"}, nil)
	s.Equal(http.StatusBadRequest, response.StatusCode, "content views are only recorded by the server")

	otherPhaseID := s.newPhase("Assessment")
	response, _ = s.send(http.MethodPost, phasePath(otherPhaseID)+"/calls/"+callIDs[0]+"/events", s.lecturer(otherPhaseID),
		map[string]any{"type": "shown"}, nil)
	s.Equal(http.StatusNotFound, response.StatusCode, "a call is only reachable through its own phase")
	response, _ = s.send(http.MethodGet, phasePath(otherPhaseID)+"/calls/"+callIDs[0], s.admin(), nil, nil)
	s.Equal(http.StatusNotFound, response.StatusCode)

	denied, _ := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion("cloud-only-model", "Only offered by the cloud", false), nil)
	response, _ = s.send(http.MethodPost, phasePath(coursePhaseID)+"/calls/"+denied.Header.Get(gateway.CallIDHeader)+"/events",
		s.lecturer(coursePhaseID), map[string]any{"type": "shown"}, nil)
	s.Equal(http.StatusNotFound, response.StatusCode, "a denied call produced no output to oversee")

	response, body = s.send(http.MethodGet, callPath, s.admin(), nil, nil)
	s.Require().Equal(http.StatusOK, response.StatusCode)
	s.Contains(string(body), `"subjects":[]`, "a call without subjects lists none rather than null")
	var detail callDTO.Detail
	s.Require().NoError(json.Unmarshal(body, &detail))
	s.Equal(callDTO.ContentAvailable, detail.ContentState)
	s.Equal(summaryAnswer, detail.Content.ResponseText)
	s.Contains(string(detail.Content.Request), summaryPrompt)
	s.Require().Len(detail.Events, 3)
	s.Equal([]string{"shown", "edited", "content_viewed"}, []string{detail.Events[0].Type, detail.Events[1].Type, detail.Events[2].Type})
	s.Equal(s.adminID, detail.Events[2].ActorID, "reading the content is itself recorded")
}

func (s *AIServerSuite) TestAuditTablesAreAppendOnly() {
	coursePhaseID := s.newPhase("Assessment")
	response, _ := s.chat(coursePhaseID, s.lecturer(coursePhaseID), completion(testutils.TestModel, summaryPrompt, false),
		map[string]string{"X-Prompt-Subjects": uuid.NewString()})
	callID := response.Header.Get(gateway.CallIDHeader)
	eventResponse, _ := s.send(http.MethodPost, phasePath(coursePhaseID)+"/calls/"+callID+"/events", s.lecturer(coursePhaseID),
		map[string]any{"type": "shown"}, nil)
	s.Require().Equal(http.StatusCreated, eventResponse.StatusCode)

	for _, statement := range []string{
		"UPDATE ai_call SET outcome = 'error' WHERE id = $1",
		"UPDATE ai_call SET feature = 'assessment.other' WHERE id = $1",
		"UPDATE ai_call SET issuer = 'https://elsewhere.example' WHERE id = $1",
		"UPDATE ai_call_content SET response = NULL WHERE call_id = $1",
		"UPDATE ai_call_subject SET course_participation_id = gen_random_uuid() WHERE call_id = $1",
		"UPDATE ai_call_event SET type = 'accepted' WHERE call_id = $1",
	} {
		_, err := s.db.Conn.Exec(s.ctx, statement, callID)
		s.Error(err, statement)
	}
}
