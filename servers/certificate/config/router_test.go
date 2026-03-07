package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ls1intum/prompt2/servers/certificate/config/configDTO"
	"github.com/ls1intum/prompt2/servers/certificate/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ConfigRouterTestSuite struct {
	suite.Suite
	router   *gin.Engine
	suiteCtx context.Context
	cleanup  func()
}

func (s *ConfigRouterTestSuite) SetupSuite() {
	s.suiteCtx = context.Background()
	testDB, cleanup, err := testutils.SetupTestDB(s.suiteCtx, "../database_dumps/certificate.sql")
	if err != nil {
		s.T().Fatalf("Failed to set up test database: %v", err)
	}
	s.cleanup = cleanup

	ConfigServiceSingleton = &ConfigService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	gin.SetMode(gin.TestMode)
	s.router = gin.Default()
	api := s.router.Group("/api/course_phase/:coursePhaseID")
	setupConfigRouter(api, testutils.MockPermissionMiddleware)
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
	_, err := ConfigServiceSingleton.queries.CreateCoursePhaseConfig(s.suiteCtx, noTemplateID)
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
