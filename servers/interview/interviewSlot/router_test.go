package interviewSlot

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/interview/db/sqlc"
	interviewSlotDTO "github.com/prompt-edu/prompt/servers/interview/interviewSlot/interviewSlotDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InterviewSlotRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *InterviewSlotRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	InterviewSlotServiceSingleton = &InterviewSlotService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}
}

func (suite *InterviewSlotRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *InterviewSlotRouterTestSuite) newRouter() *gin.Engine {
	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	setupInterviewSlotRouter(api, func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(allowedRoles)
	})
	return router
}

func (suite *InterviewSlotRouterTestSuite) TestCreateInterviewSlotRoute() {
	router := suite.newRouter()
	location := "Router Test Room"
	body := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  2,
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-slots", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusCreated, resp.Code)

	var slot db.InterviewSlot
	err := json.Unmarshal(resp.Body.Bytes(), &slot)
	require.NoError(suite.T(), err)
	require.NotEqual(suite.T(), uuid.Nil, slot.ID)
}

func (suite *InterviewSlotRouterTestSuite) TestGetAllInterviewSlotsRoute() {
	router := suite.newRouter()
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-slots", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var slots []interviewSlotDTO.InterviewSlotResponse
	err := json.Unmarshal(resp.Body.Bytes(), &slots)
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), slots)
}

func (suite *InterviewSlotRouterTestSuite) TestGetInterviewSlotRoute() {
	router := suite.newRouter()
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-slots/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	req, _ := http.NewRequest("GET", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var slot interviewSlotDTO.InterviewSlotResponse
	err := json.Unmarshal(resp.Body.Bytes(), &slot)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"), slot.ID)
}

func (suite *InterviewSlotRouterTestSuite) TestUpdateInterviewSlotRoute() {
	location := "Original Location"
	createBody := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  2,
	}
	slot, err := CreateInterviewSlot(suite.ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), createBody)
	require.NoError(suite.T(), err)

	router := suite.newRouter()
	newLocation := "Updated Location"
	updateBody := interviewSlotDTO.UpdateInterviewSlotRequest{
		StartTime: time.Now().Add(26 * time.Hour),
		EndTime:   time.Now().Add(27 * time.Hour),
		Location:  &newLocation,
		Capacity:  3,
	}
	payload, _ := json.Marshal(updateBody)

	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-slots/" + slot.ID.String()
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var updated db.InterviewSlot
	err = json.Unmarshal(resp.Body.Bytes(), &updated)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), int32(3), updated.Capacity)
}

func (suite *InterviewSlotRouterTestSuite) TestDeleteInterviewSlotRoute() {
	location := "Delete Test"
	createBody := interviewSlotDTO.CreateInterviewSlotRequest{
		StartTime: time.Now().Add(24 * time.Hour),
		EndTime:   time.Now().Add(25 * time.Hour),
		Location:  &location,
		Capacity:  1,
	}
	slot, err := CreateInterviewSlot(suite.ctx, uuid.MustParse("11111111-1111-1111-1111-111111111111"), createBody)
	require.NoError(suite.T(), err)

	router := suite.newRouter()
	url := "/api/course_phase/11111111-1111-1111-1111-111111111111/interview-slots/" + slot.ID.String()
	req, _ := http.NewRequest("DELETE", url, nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusNoContent, resp.Code)

	_, err = suite.testDB.Queries.GetInterviewSlot(suite.ctx, slot.ID)
	require.Error(suite.T(), err)
}

func TestInterviewSlotRouterTestSuite(t *testing.T) {
	suite.Run(t, new(InterviewSlotRouterTestSuite))
}
