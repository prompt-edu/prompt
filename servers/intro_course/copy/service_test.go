package copy

import (
	"context"
	"testing"

	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/assert"
)

func TestHandlePhaseCopyCopiesSeatPlan(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	testDB, cleanup, err := testutils.SetupTestDB(ctx, "../database_dumps/intro_course.sql")
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	defer cleanup()

	CopyServiceSingleton = &CopyService{
		queries: *testDB.Queries,
		conn:    testDB.Conn,
	}

	sourcePhase := uuid.MustParse("4179d58a-d00d-4fa7-94a5-397bc69fab02")
	targetPhase := uuid.New()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	req := httptest.NewRequest("POST", "/intro-course/api/copy", nil)
	c.Request = req
	c.Params = []gin.Param{{Key: "coursePhaseID", Value: targetPhase.String()}}
	handler := IntroCourseCopyHandler{}

	err = handler.HandlePhaseCopy(c, promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: sourcePhase,
		TargetCoursePhaseID: targetPhase,
	})
	assert.NoError(t, err)

	seats, err := CopyServiceSingleton.queries.GetSeatPlan(ctx, targetPhase)
	assert.NoError(t, err)
	assert.Len(t, seats, 2)
}
