package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/skills/skillDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SkillRouterTestSuite struct {
	suite.Suite
	router       *gin.Engine
	suiteCtx     context.Context
	cleanup      func()
	skillService SkillsService
}

func (suite *SkillRouterTestSuite) SetupSuite() {
	suite.suiteCtx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.suiteCtx, "../database_dumps/skills.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.skillService = SkillsService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	SkillsServiceSingleton = &suite.skillService
	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	testMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithEmail(allowedRoles, "lecturer@example.com", "03711111", "ab12cde")
	}
	setupSkillRouter(api, testMiddleware)
}

func (suite *SkillRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *SkillRouterTestSuite) TestGetAllSkills() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("GET", "/api/course_phase/"+coursePhaseID+"/skill", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var skills []skillDTO.Skill
	err := json.Unmarshal(resp.Body.Bytes(), &skills)
	assert.NoError(suite.T(), err)
	assert.Greater(suite.T(), len(skills), 0, "Should return a list of skills")
}

func (suite *SkillRouterTestSuite) TestGetAllSkillsInvalidCoursePhaseID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/invalid-uuid/skill", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *SkillRouterTestSuite) TestCreateSkills() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	createReq := skillDTO.CreateSkillsRequest{
		SkillNames: []string{"React", "Node.js", "Docker"},
	}
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/skill", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *SkillRouterTestSuite) TestCreateSkillsInvalidJSON() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	req, _ := http.NewRequest("POST", "/api/course_phase/"+coursePhaseID+"/skill", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *SkillRouterTestSuite) TestUpdateSkill() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	skillID := "11111111-1111-1111-1111-111111111111"

	updateReq := skillDTO.UpdateSkillRequest{
		NewSkillName: "Advanced Java",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/skill/"+skillID, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *SkillRouterTestSuite) TestUpdateSkillInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	updateReq := skillDTO.UpdateSkillRequest{
		NewSkillName: "Advanced Java",
	}
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID+"/skill/invalid-uuid", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *SkillRouterTestSuite) TestDeleteSkill() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"
	skillID := "22222222-2222-2222-2222-222222222222"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/skill/"+skillID, nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *SkillRouterTestSuite) TestDeleteSkillInvalidID() {
	coursePhaseID := "4179d58a-d00d-4fa7-94a5-397bc69fab02"

	req, _ := http.NewRequest("DELETE", "/api/course_phase/"+coursePhaseID+"/skill/invalid-uuid", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestSkillRouterTestSuite(t *testing.T) {
	suite.Run(t, new(SkillRouterTestSuite))
}
