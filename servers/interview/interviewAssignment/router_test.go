package interviewAssignment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	keycloakTokenVerifier "github.com/prompt-edu/prompt-sdk/keycloakTokenVerifier"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewAssignmentDTO "github.com/prompt-edu/prompt/servers/interview/interviewAssignment/interviewAssignmentDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InterviewAssignmentRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *InterviewAssignmentRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	InterviewAssignmentServiceSingleton = &InterviewAssignmentService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *InterviewAssignmentRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *InterviewAssignmentRouterTestSuite) newRouter(courseParticipationID uuid.UUID) *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupInterviewAssignmentRouter(api, func(allowedRoles ...string) gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Set("courseParticipationID", courseParticipationID)
			keycloakTokenVerifier.SetTokenUser(c, keycloakTokenVerifier.TokenUser{
				Roles:                 map[string]bool{},
				Email:                 "test@example.com",
				IsStudentOfCourse:     true,
				CourseParticipationID: courseParticipationID,
			})
			c.Next()
		}
	})
	return router
}

func (suite *InterviewAssignmentRouterTestSuite) TestCreateInterviewAssignmentRoute() {
	participationID := uuid.New()
	router := suite.newRouter(participationID)

	body := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-assignments", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusCreated, resp.Code)

	var assignment interviewAssignmentDTO.InterviewAssignmentResponse
	err := json.Unmarshal(resp.Body.Bytes(), &assignment)
	require.NoError(suite.T(), err)
	require.NotEqual(suite.T(), uuid.Nil, assignment.ID)
}

func (suite *InterviewAssignmentRouterTestSuite) TestGetMyInterviewAssignmentRoute() {
	participationID := uuid.MustParse("ffffffff-ffff-ffff-ffff-ffffffffffff")
	router := suite.newRouter(participationID)

	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-assignments/my-assignment", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var assignment interviewAssignmentDTO.InterviewAssignmentResponse
	err := json.Unmarshal(resp.Body.Bytes(), &assignment)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), participationID, assignment.CourseParticipationID)
}

func (suite *InterviewAssignmentRouterTestSuite) TestDeleteInterviewAssignmentRoute() {
	// Create a dedicated slot for this test with large capacity
	slot, err := suite.testDB.Queries.CreateInterviewSlot(suite.ctx, db.CreateInterviewSlotParams{
		CoursePhaseID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		StartTime:     pgtype.Timestamptz{Time: time.Now().Add(102 * time.Hour), Valid: true},
		EndTime:       pgtype.Timestamptz{Time: time.Now().Add(103 * time.Hour), Valid: true},
		Capacity:      10,
	})
	require.NoError(suite.T(), err)

	participationID := uuid.New()
	createBody := interviewAssignmentDTO.CreateInterviewAssignmentRequest{
		InterviewSlotID: slot.ID,
	}
	assignment, err := CreateInterviewAssignment(suite.ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), participationID, createBody)
	require.NoError(suite.T(), err)

	router := suite.newRouter(participationID)
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-assignments/" + assignment.ID.String()
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusNoContent, resp.Code)

	_, err = suite.testDB.Queries.GetInterviewAssignment(suite.ctx, assignment.ID)
	require.Error(suite.T(), err)
}

func (suite *InterviewAssignmentRouterTestSuite) TestDeleteInterviewAssignmentUnauthorizedRoute() {
	existingAssignmentID := uuid.MustParse("11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	wrongParticipationID := uuid.New()

	router := suite.newRouter(wrongParticipationID)
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-assignments/" + existingAssignmentID.String()
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusForbidden, resp.Code)
}

func TestInterviewAssignmentRouterTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewAssignmentRouterTestSuite))
}
