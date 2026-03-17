package copy

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/stretchr/testify/require"
)

func TestHandlePhaseCopyReturnsNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest("POST", "/copy", nil)
	c.Params = []gin.Param{{Key: "coursePhaseID", Value: uuid.New().String()}}

	handler := SelfTeamCopyHandler{}
	err := handler.HandlePhaseCopy(c, promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.New(),
		TargetCoursePhaseID: uuid.New(),
	})

	require.NoError(t, err)
}
