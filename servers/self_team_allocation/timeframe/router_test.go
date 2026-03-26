package timeframe

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
	"github.com/jackc/pgx/v5/pgxpool"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe/timeframeDTO"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TimeframeRouterTestSuite struct {
	suite.Suite
	ctx     context.Context
	router  *gin.Engine
	testDB  *sdkTestUtils.TestDB[*db.Queries]
	cleanup func()
}

func (suite *TimeframeRouterTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.ctx = context.Background()

	testDB, cleanup, err := sdkTestUtils.SetupTestDB(suite.ctx, "../database_dumps/base.sql", func(conn *pgxpool.Pool) *db.Queries { return db.New(conn) })
	require.NoError(suite.T(), err)
	suite.testDB = testDB
	suite.cleanup = cleanup

	TimeframeServiceSingleton = NewTimeframeService(*testDB.Queries, testDB.Conn)

	suite.router = gin.Default()
	api := suite.router.Group("/api/course_phase/:coursePhaseID")
	authMiddleware := func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.DefaultMockAuthMiddleware()
	}
	setupTimeframeRouter(api, authMiddleware)
}

func (suite *TimeframeRouterTestSuite) TearDownSuite() {
	if suite.cleanup != nil {
		suite.cleanup()
	}
}

func (suite *TimeframeRouterTestSuite) TestGetTimeframe() {
	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/timeframe", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusOK, resp.Code)

	var tf timeframeDTO.Timeframe
	err := json.Unmarshal(resp.Body.Bytes(), &tf)
	require.NoError(suite.T(), err)
	require.True(suite.T(), tf.TimeframeSet)
}

func (suite *TimeframeRouterTestSuite) TestGetTimeframeInvalidUUID() {
	req, _ := http.NewRequest("GET", "/api/course_phase/not-a-uuid/timeframe", nil)
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TimeframeRouterTestSuite) TestSetTimeframe() {
	coursePhaseID := uuid.New()
	body := timeframeDTO.Timeframe{
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(2 * time.Hour),
	}
	payload, _ := json.Marshal(body)

	url := "/api/course_phase/" + coursePhaseID.String() + "/timeframe"
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusNoContent, resp.Code)

	record, err := suite.testDB.Queries.GetTimeframe(suite.ctx, coursePhaseID)
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), body.StartTime.UTC().Truncate(time.Second), record.Starttime.Time.UTC().Truncate(time.Second))
}

func (suite *TimeframeRouterTestSuite) TestSetTimeframeInvalidUUID() {
	body := timeframeDTO.Timeframe{
		StartTime: time.Now().UTC(),
		EndTime:   time.Now().UTC().Add(time.Hour),
	}
	payload, _ := json.Marshal(body)

	req, _ := http.NewRequest("PUT", "/api/course_phase/not-a-uuid/timeframe", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TimeframeRouterTestSuite) TestSetTimeframeInvalidPayload() {
	coursePhaseID := uuid.New()
	req, _ := http.NewRequest("PUT", "/api/course_phase/"+coursePhaseID.String()+"/timeframe", bytes.NewBuffer([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	suite.router.ServeHTTP(resp, req)

	require.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func TestTimeframeRouterTestSuite(t *testing.T) {
	suite.Run(t, new(TimeframeRouterTestSuite))
}
