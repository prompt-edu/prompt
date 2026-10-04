package announcement

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/core/announcement/announcementDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AnnouncementRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	cleanup func()
	service *AnnouncementService
}

func (s *AnnouncementRouterTestSuite) SetupSuite() {
	s.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.ctx, "../database_dumps/announcement_test.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		log.Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup
	s.service = NewAnnouncementService(*testDB.Queries)
}

func (s *AnnouncementRouterTestSuite) TearDownSuite() {
	s.cleanup()
}

func (s *AnnouncementRouterTestSuite) routerWithRoles(roles ...string) *gin.Engine {
	router := gin.New()
	api := router.Group("/api")
	authMiddleware := func() gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(roles)
	}
	setupAnnouncementRouter(api, s.service, authMiddleware, permissionValidation.CheckAccessControlByRole)
	return router
}

func (s *AnnouncementRouterTestSuite) request(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var payload bytes.Buffer
	if body != nil {
		require.NoError(s.T(), json.NewEncoder(&payload).Encode(body))
	}
	req := httptest.NewRequest(method, path, &payload)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func (s *AnnouncementRouterTestSuite) TestActiveReturnsOnlyEnabledAnnouncementsInTheirWindow() {
	w := s.request(s.routerWithRoles(), http.MethodGet, "/api/announcements/active", nil)

	require.Equal(s.T(), http.StatusOK, w.Code)
	var announcements []announcementDTO.Announcement
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &announcements))
	titles := make([]string, 0, len(announcements))
	for _, a := range announcements {
		titles = append(titles, a.Title)
	}
	assert.Equal(s.T(), []string{"Active critical", "Active info"}, titles)
}

func (s *AnnouncementRouterTestSuite) TestAdminEndpointsRequirePromptAdmin() {
	router := s.routerWithRoles(permissionValidation.PromptLecturer)

	assert.Equal(s.T(), http.StatusForbidden, s.request(router, http.MethodGet, "/api/announcements", nil).Code)
	assert.Equal(s.T(), http.StatusForbidden, s.request(router, http.MethodPost, "/api/announcements", announcementDTO.UpsertAnnouncement{Severity: "info", Message: "Hi"}).Code)
}

func (s *AnnouncementRouterTestSuite) TestAdminListHidesExpiredUnlessRequested() {
	router := s.routerWithRoles(permissionValidation.PromptAdmin)

	var withoutExpired, withExpired []announcementDTO.Announcement
	require.NoError(s.T(), json.Unmarshal(s.request(router, http.MethodGet, "/api/announcements", nil).Body.Bytes(), &withoutExpired))
	require.NoError(s.T(), json.Unmarshal(s.request(router, http.MethodGet, "/api/announcements?includeExpired=true", nil).Body.Bytes(), &withExpired))

	assert.Len(s.T(), withExpired, len(withoutExpired)+1)
}

func (s *AnnouncementRouterTestSuite) TestCreateUpdateAndDeleteAnnouncement() {
	router := s.routerWithRoles(permissionValidation.PromptAdmin)

	w := s.request(router, http.MethodPost, "/api/announcements", announcementDTO.UpsertAnnouncement{
		Severity: "warning",
		Title:    "Maintenance",
		Message:  "PROMPT is down tonight",
		LinkURL:  "https://example.com/status",
	})
	require.Equal(s.T(), http.StatusCreated, w.Code)
	var created announcementDTO.Announcement
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &created))
	assert.False(s.T(), created.Enabled)

	w = s.request(router, http.MethodPut, "/api/announcements/"+created.ID.String(), announcementDTO.UpsertAnnouncement{
		Severity: "warning",
		Title:    "Maintenance",
		Message:  "PROMPT is down tonight",
		Enabled:  true,
	})
	require.Equal(s.T(), http.StatusOK, w.Code)
	var updated announcementDTO.Announcement
	require.NoError(s.T(), json.Unmarshal(w.Body.Bytes(), &updated))
	assert.True(s.T(), updated.Enabled)

	assert.Equal(s.T(), http.StatusNoContent, s.request(router, http.MethodDelete, "/api/announcements/"+created.ID.String(), nil).Code)
	assert.Equal(s.T(), http.StatusNotFound, s.request(router, http.MethodDelete, "/api/announcements/"+created.ID.String(), nil).Code)
}

func (s *AnnouncementRouterTestSuite) TestCreateRejectsInvalidInput() {
	router := s.routerWithRoles(permissionValidation.PromptAdmin)

	invalid := []announcementDTO.UpsertAnnouncement{
		{Severity: "urgent", Message: "Unknown severity"},
		{Severity: "info", Message: "   "},
		{Severity: "info", Message: "Bad link", LinkURL: "javascript:alert(1)"},
		{Severity: "info", Message: "Label without link", LinkLabel: "Read more"},
	}
	for _, request := range invalid {
		assert.Equal(s.T(), http.StatusBadRequest, s.request(router, http.MethodPost, "/api/announcements", request).Code, request.Message)
	}
}

func TestAnnouncementRouterTestSuite(t *testing.T) {
	suite.Run(t, new(AnnouncementRouterTestSuite))
}
