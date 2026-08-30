package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/prompt-edu/prompt/servers/certificate/config/configDTO"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
	service  *ConfigService
}

func (s *ConfigRouterTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup

	s.service = NewConfigService(*testDB.Queries)

	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	api := s.router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, s.service, sdkTestUtils.MockPermissionMiddleware)
}

func (s *ConfigRouterTestSuite) TearDownSuite() {
	if s.cleanup != nil {
		s.cleanup()
	}
}

func (s *ConfigRouterTestSuite) TestGetConfig() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/config", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var cfg configDTO.CoursePhaseConfig
	err := json.Unmarshal(resp.Body.Bytes(), &cfg)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), coursePhaseID, cfg.CoursePhaseID)
	assert.True(s.T(), cfg.HasTemplate)
}

func (s *ConfigRouterTestSuite) TestGetConfig_AutoCreate() {
	newID := uuid.New()
	url := fmt.Sprintf("/api/course_phase/%s/config", newID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var cfg configDTO.CoursePhaseConfig
	err := json.Unmarshal(resp.Body.Bytes(), &cfg)
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), newID, cfg.CoursePhaseID)
	assert.False(s.T(), cfg.HasTemplate)
}

func (s *ConfigRouterTestSuite) TestGetConfig_InvalidID() {
	url := "/api/course_phase/not-a-uuid/config"

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *ConfigRouterTestSuite) TestUpdateConfig() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/config", coursePhaseID)

	body := configDTO.UpdateConfigRequest{
		TemplateContent: "= Updated via Router\nNew content here",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var cfg configDTO.CoursePhaseConfig
	err := json.Unmarshal(resp.Body.Bytes(), &cfg)
	assert.NoError(s.T(), err)
	assert.True(s.T(), cfg.HasTemplate)
	assert.Equal(s.T(), body.TemplateContent, *cfg.TemplateContent)
}

func (s *ConfigRouterTestSuite) TestUpdateConfig_InvalidID() {
	url := "/api/course_phase/not-a-uuid/config"

	body := configDTO.UpdateConfigRequest{
		TemplateContent: "some content",
	}
	bodyJSON, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *ConfigRouterTestSuite) TestUpdateConfig_EmptyBody() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/config", coursePhaseID)

	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusBadRequest, resp.Code)
}

func (s *ConfigRouterTestSuite) TestGetTemplate() {
	coursePhaseID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	url := fmt.Sprintf("/api/course_phase/%s/config/template", coursePhaseID)

	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusOK, resp.Code)
	assert.Contains(s.T(), resp.Body.String(), "Certificate of Completion")
	assert.Equal(s.T(), "text/plain", resp.Header().Get("Content-Type"))
}

func (s *ConfigRouterTestSuite) TestGetTemplate_NoTemplate() {
	// Create a config without template
	noTemplateID := uuid.New()
	_, err := s.service.queries.CreateCoursePhaseConfig(s.suiteCtx, noTemplateID)
	assert.NoError(s.T(), err)

	url := fmt.Sprintf("/api/course_phase/%s/config/template", noTemplateID)
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)

	assert.Equal(s.T(), http.StatusNotFound, resp.Code)
}

func TestConfigRouterTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigRouterTestSuite))
}

func (s *ConfigRouterTestSuite) putStudentPageText(coursePhaseID uuid.UUID, body string) *httptest.ResponseRecorder {
	url := fmt.Sprintf("/api/course_phase/%s/config/student-page-text", coursePhaseID)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	s.router.ServeHTTP(resp, req)
	return resp
}

func (s *ConfigRouterTestSuite) TestUpdateStudentPageTextRoundTrip() {
	coursePhaseID := uuid.New()

	resp := s.putStudentPageText(coursePhaseID, `{"studentPageText":"<p>Congratulations!</p>"}`)
	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var config configDTO.CoursePhaseConfig
	assert.NoError(s.T(), json.Unmarshal(resp.Body.Bytes(), &config))
	assert.NotNil(s.T(), config.StudentPageText)
	assert.Equal(s.T(), "<p>Congratulations!</p>", *config.StudentPageText)

	// It is readable back through the config endpoint the settings page uses.
	readBack, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/course_phase/%s/config", coursePhaseID), nil)
	readResp := httptest.NewRecorder()
	s.router.ServeHTTP(readResp, readBack)
	assert.Equal(s.T(), http.StatusOK, readResp.Code)

	var stored configDTO.CoursePhaseConfig
	assert.NoError(s.T(), json.Unmarshal(readResp.Body.Bytes(), &stored))
	assert.NotNil(s.T(), stored.StudentPageText)

	// An explicit null clears it.
	cleared := s.putStudentPageText(coursePhaseID, `{"studentPageText":null}`)
	assert.Equal(s.T(), http.StatusOK, cleared.Code)

	var clearedConfig configDTO.CoursePhaseConfig
	assert.NoError(s.T(), json.Unmarshal(cleared.Body.Bytes(), &clearedConfig))
	assert.Nil(s.T(), clearedConfig.StudentPageText)
}

func (s *ConfigRouterTestSuite) TestUpdateStudentPageTextRejectsBadRequests() {
	coursePhaseID := uuid.New()

	// An omitted key is not the same as an explicit null, and must not clear.
	assert.Equal(s.T(), http.StatusBadRequest, s.putStudentPageText(coursePhaseID, `{}`).Code)
	assert.Equal(s.T(), http.StatusBadRequest, s.putStudentPageText(coursePhaseID, `{"studentPageText":42}`).Code)

	oversized, _ := json.Marshal(map[string]string{
		"studentPageText": strings.Repeat("ä", configDTO.MaxStudentPageTextBytes),
	})
	assert.Equal(s.T(), http.StatusBadRequest, s.putStudentPageText(coursePhaseID, string(oversized)).Code,
		"the cap counts bytes, so multibyte text hits it sooner")

	invalidPhase, _ := http.NewRequest(http.MethodPut, "/api/course_phase/not-a-uuid/config/student-page-text", bytes.NewBufferString(`{"studentPageText":"x"}`))
	invalidResp := httptest.NewRecorder()
	s.router.ServeHTTP(invalidResp, invalidPhase)
	assert.Equal(s.T(), http.StatusBadRequest, invalidResp.Code)
}

func (s *ConfigRouterTestSuite) TestUpdateStudentPageTextPreservesTemplateAndReleaseDate() {
	coursePhaseID := uuid.New()
	releaseDate := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)

	templateBody, _ := json.Marshal(configDTO.UpdateConfigRequest{TemplateContent: "= Certificate"})
	templateReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/course_phase/%s/config", coursePhaseID), bytes.NewBuffer(templateBody))
	templateReq.Header.Set("Content-Type", "application/json")
	templateResp := httptest.NewRecorder()
	s.router.ServeHTTP(templateResp, templateReq)
	assert.Equal(s.T(), http.StatusOK, templateResp.Code)

	releaseBody, _ := json.Marshal(configDTO.UpdateReleaseDateRequest{ReleaseDate: &releaseDate})
	releaseReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/course_phase/%s/config/release-date", coursePhaseID), bytes.NewBuffer(releaseBody))
	releaseReq.Header.Set("Content-Type", "application/json")
	releaseResp := httptest.NewRecorder()
	s.router.ServeHTTP(releaseResp, releaseReq)
	assert.Equal(s.T(), http.StatusOK, releaseResp.Code)

	var beforeText configDTO.CoursePhaseConfig
	assert.NoError(s.T(), json.Unmarshal(releaseResp.Body.Bytes(), &beforeText))

	resp := s.putStudentPageText(coursePhaseID, `{"studentPageText":"<p>See you at the ceremony.</p>"}`)
	assert.Equal(s.T(), http.StatusOK, resp.Code)

	var config configDTO.CoursePhaseConfig
	assert.NoError(s.T(), json.Unmarshal(resp.Body.Bytes(), &config))
	assert.True(s.T(), config.HasTemplate, "the template must survive a text-only write")
	assert.NotNil(s.T(), config.ReleaseDate, "the release date must survive a text-only write")
	assert.NotNil(s.T(), config.StudentPageText)

	// The settings page presents updatedAt/updatedBy as the template's
	// provenance, so a text-only write must not claim it was reconfigured.
	assert.Equal(s.T(), beforeText.UpdatedAt, config.UpdatedAt)
	assert.Equal(s.T(), beforeText.UpdatedBy, config.UpdatedBy)
}
