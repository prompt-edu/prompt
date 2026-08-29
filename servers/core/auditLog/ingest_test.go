package auditLog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	"github.com/prompt-edu/prompt/servers/core/auditLog/auditLogDTO"
	"github.com/stretchr/testify/require"
)

const ingestService = "interview"

// ingestRouter mounts only the ingest route: the read routes of
// setupAuditLogRouter need the Keycloak and permission-validation singletons,
// which are irrelevant to the ingest contract.
func (s *AuditLogTestSuite) ingestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	keys := parseIngestKeys(ingestService + ":sekret")
	router.POST("/api/audit", ingestAuth(keys), ingestAuditEvent(s.sink))
	return router
}

func (s *AuditLogTestSuite) postEvent(router *gin.Engine, e audit.Event, service, token string) *httptest.ResponseRecorder {
	body, err := json.Marshal(e)
	require.NoError(s.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/api/audit", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	if service != "" {
		req.Header.Set("X-Audit-Service", service)
	}
	if token != "" {
		req.Header.Set("X-Audit-Token", token)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func (s *AuditLogTestSuite) TestIngest_AcceptsErrorOutcome() {
	// The SDK stamps "error" on an explicitly recorded event whose request ends
	// on an unclassified status, so core must store it rather than 400 and drop
	// the entry (the HTTP sink never retries a 4xx).
	router := s.ingestRouter()

	res := s.postEvent(router, audit.Event{
		Action:  "Published grades",
		Outcome: audit.OutcomeError,
	}, ingestService, "sekret")
	require.Equal(s.T(), http.StatusCreated, res.Code)

	page, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, auditLogDTO.ListFilters{})
	require.NoError(s.T(), err)
	require.Len(s.T(), page.Entries, 1)
	require.Equal(s.T(), audit.OutcomeError, page.Entries[0].Outcome)
}

func (s *AuditLogTestSuite) TestIngest_StampsSourceServiceFromCredentialsNotBody() {
	router := s.ingestRouter()

	res := s.postEvent(router, audit.Event{
		Action:        "Created slot",
		SourceService: "spoofed",
	}, ingestService, "sekret")
	require.Equal(s.T(), http.StatusCreated, res.Code)

	page, err := AuditLogServiceSingleton.ListAuditLog(s.ctx, auditLogDTO.ListFilters{})
	require.NoError(s.T(), err)
	require.Len(s.T(), page.Entries, 1)
	require.Equal(s.T(), ingestService, page.Entries[0].SourceService)
}

func (s *AuditLogTestSuite) TestIngest_RejectsUnstorableEvents() {
	router := s.ingestRouter()

	testCases := []struct {
		name  string
		event audit.Event
	}{
		{"bogus outcome", audit.Event{Action: "Created slot", Outcome: "totally-fine"}},
		{"empty action", audit.Event{Action: "   "}},
		{"malformed actorID", audit.Event{Action: "Created slot", ActorID: "not-a-uuid"}},
		{"malformed courseID", audit.Event{Action: "Created slot", CourseID: "not-a-uuid"}},
		{"malformed coursePhaseID", audit.Event{Action: "Created slot", CoursePhaseID: "not-a-uuid"}},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			res := s.postEvent(router, tc.event, ingestService, "sekret")
			require.Equal(s.T(), http.StatusBadRequest, res.Code)
		})
	}
	require.Equal(s.T(), 0, s.countRows())
}

func (s *AuditLogTestSuite) TestIngest_RejectsBadCredentials() {
	router := s.ingestRouter()
	event := audit.Event{Action: "Created slot"}

	require.Equal(s.T(), http.StatusUnauthorized, s.postEvent(router, event, ingestService, "wrong").Code)
	require.Equal(s.T(), http.StatusUnauthorized, s.postEvent(router, event, "unknown-service", "sekret").Code)
	require.Equal(s.T(), http.StatusUnauthorized, s.postEvent(router, event, "", "sekret").Code)
	require.Equal(s.T(), http.StatusUnauthorized, s.postEvent(router, event, ingestService, "").Code)
	require.Equal(s.T(), 0, s.countRows())
}
