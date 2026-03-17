package seatPlan

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/seatPlan/seatPlanDTO"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SeatPlanRouterTestSuite struct {
	suite.Suite
	ctx           context.Context
	router        *gin.Engine
	cleanup       func()
	coursePhaseID uuid.UUID
	studentID     uuid.UUID
}

func (suite *SeatPlanRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()
	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/intro_course.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	if err != nil {
		suite.T().Fatalf("Failed to set up test database: %v", err)
	}
	suite.cleanup = cleanup
	suite.coursePhaseID = uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	suite.studentID = uuid.MustParse("33333333-3333-3333-3333-333333333333")

	seatPlanService := SeatPlanService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
	SeatPlanServiceSingleton = &seatPlanService

	suite.router = gin.Default()
	api := suite.router.Group("/intro-course/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddlewareWithParticipation(allowedRoles, suite.studentID)
	}
	setupSeatPlanRouter(api, authMiddleware)
}

func (suite *SeatPlanRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func TestSeatPlanRouterTestSuite(t *testing.T) {
	suite.Run(t, new(SeatPlanRouterTestSuite))
}

func (suite *SeatPlanRouterTestSuite) TestGetSeatPlanSuccess() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/seat_plan", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var seats []seatPlanDTO.Seat
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &seats))
	assert.Len(suite.T(), seats, 2)
}

func (suite *SeatPlanRouterTestSuite) TestCreateSeatPlan() {
	newCoursePhaseID := uuid.New().String()
	body, _ := json.Marshal([]string{"New-1", "New-2"})
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/"+newCoursePhaseID+"/seat_plan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)
}

func (suite *SeatPlanRouterTestSuite) TestCreateSeatPlanDuplicateNames() {
	body, _ := json.Marshal([]string{"Dup", "Dup"})
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/seat_plan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *SeatPlanRouterTestSuite) TestCreateSeatPlanInvalidCoursePhase() {
	req, _ := http.NewRequest("POST", "/intro-course/api/course_phase/invalid-uuid/seat_plan", bytes.NewBuffer([]byte("[]")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *SeatPlanRouterTestSuite) TestUpdateSeatPlan() {
	seatUpdate := []seatPlanDTO.Seat{
		{
			SeatName:        "Seat-1",
			HasMac:          true,
			DeviceID:        pgtype.Text{String: "Updated", Valid: true},
			AssignedStudent: pgtype.UUID{Bytes: suite.studentID, Valid: true},
			AssignedTutor:   pgtype.UUID{Bytes: uuid.MustParse("11111111-1111-1111-1111-111111111111"), Valid: true},
		},
	}
	body, _ := json.Marshal(seatUpdate)
	req, _ := http.NewRequest("PUT", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/seat_plan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *SeatPlanRouterTestSuite) TestDeleteSeatPlan() {
	deleteCoursePhaseID := uuid.New()
	_ = CreateSeatPlan(suite.ctx, deleteCoursePhaseID, []string{"Temp"})

	req, _ := http.NewRequest("DELETE", "/intro-course/api/course_phase/"+deleteCoursePhaseID.String()+"/seat_plan", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}

func (suite *SeatPlanRouterTestSuite) TestGetOwnSeatAssignment() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+suite.coursePhaseID.String()+"/seat_plan/own-assignment", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)
	var assignment seatPlanDTO.SeatAssignment
	assert.NoError(suite.T(), json.Unmarshal(resp.Body.Bytes(), &assignment))
	assert.Equal(suite.T(), "Seat-1", assignment.SeatName)
}

func (suite *SeatPlanRouterTestSuite) TestGetOwnSeatAssignmentInvalidUUID() {
	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/not-a-uuid/seat_plan/own-assignment", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}
