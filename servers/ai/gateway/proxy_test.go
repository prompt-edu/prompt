package gateway

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/ai/calls"
	"github.com/prompt-edu/prompt/servers/ai/encryption"
	"github.com/prompt-edu/prompt/servers/ai/testutils"
	"github.com/stretchr/testify/suite"
)

const (
	testIdleTimeout  = 300 * time.Millisecond
	testTotalTimeout = 1500 * time.Millisecond
)

type ProxySuite struct {
	suite.Suite
	ctx     context.Context
	db      *testutils.TestDB
	cleanup func()
	calls   *calls.Service
}

func (s *ProxySuite) SetupSuite() {
	s.ctx = context.Background()
	s.T().Setenv("AI_ENCRYPTION_KEYS", "1:ZTJlLWFpLXRlc3Qta2V5LW5vdC1hLXJlYWwtc2VjcmU=")
	gin.SetMode(gin.TestMode)
	testDB, cleanup, err := testutils.SetupTestDB(s.ctx)
	s.Require().NoError(err)
	s.db, s.cleanup = testDB, cleanup
	box, err := encryption.New()
	s.Require().NoError(err)
	s.calls = calls.NewService(testDB.Queries, testDB.Conn, "slow.test", "test", box)
}

func (s *ProxySuite) TearDownSuite() { s.cleanup() }

func TestProxySuite(t *testing.T) { suite.Run(t, new(ProxySuite)) }

// forwardTo sends one call through forward to provider and returns the response status and the recorded error code.
func (s *ProxySuite) forwardTo(provider http.HandlerFunc, streamed bool) (int, string) {
	stub := httptest.NewServer(provider)
	defer stub.Close()
	providerURL, err := url.Parse(stub.URL + "/v1")
	s.Require().NoError(err)
	return s.forwardThrough(New(providerURL, []string{"m1"}, "", s.calls), streamed)
}

func (s *ProxySuite) forwardThrough(gateway *Gateway, streamed bool) (int, string) {
	gateway.totalTimeout, gateway.idleTimeout = testTotalTimeout, testIdleTimeout

	request := calls.Request{
		CoursePhaseID: uuid.New(), ActorID: uuid.NewString(), ActorRole: "Lecturer", Feature: "assessment.action_item_suggestions",
		Model: "m1", Streamed: streamed, Body: []byte(`{"model":"m1"}`),
	}
	callID, err := s.calls.Begin(s.ctx, request)
	s.Require().NoError(err)

	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/", func(c *gin.Context) { gateway.forward(c, callID, "logos-key-1234", request) })
	server := httptest.NewServer(router)
	defer server.Close()

	response, err := http.Post(server.URL, "application/json", strings.NewReader("{}"))
	s.Require().NoError(err)
	_, _ = io.ReadAll(response.Body)
	_ = response.Body.Close()

	var errorCode *string
	s.Require().NoError(s.db.Conn.QueryRow(s.ctx, "SELECT error_code FROM ai_call WHERE id = $1", callID).Scan(&errorCode))
	if errorCode == nil {
		return response.StatusCode, ""
	}
	return response.StatusCode, *errorCode
}

// waitForDisconnect stalls like a hung provider. The server only notices a closed connection once the body is read.
func waitForDisconnect(r *http.Request) {
	_, _ = io.ReadAll(r.Body)
	<-r.Context().Done()
}

func (s *ProxySuite) TestSlowNonStreamedAnswerIsNotCutByTheIdleLimit() {
	status, errorCode := s.forwardTo(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * testIdleTimeout)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"m1","choices":[{"message":{"content":"done"},"finish_reason":"stop"}]}`))
	}, false)

	s.Equal(http.StatusOK, status)
	s.Empty(errorCode)
}

func (s *ProxySuite) TestNonStreamedAnswerStopsAtTheTotalLimit() {
	status, errorCode := s.forwardTo(func(_ http.ResponseWriter, r *http.Request) {
		waitForDisconnect(r)
	}, false)

	s.Equal(http.StatusGatewayTimeout, status)
	s.Equal("total_timeout", errorCode)
}

func (s *ProxySuite) TestStalledStreamStopsAtTheIdleLimit() {
	_, errorCode := s.forwardTo(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Hi\"}}]}\n\n"))
		w.(http.Flusher).Flush()
		waitForDisconnect(r)
	}, true)

	s.Equal("idle_timeout", errorCode)
}

func (s *ProxySuite) TestBugIsNotBlamedOnTheProvider() {
	broken := New(nil, []string{"m1"}, "", s.calls)

	status, errorCode := s.forwardThrough(broken, false)

	s.Equal(http.StatusInternalServerError, status, "the panic reaches gin's recovery")
	s.Equal("internal_error", errorCode)
}
