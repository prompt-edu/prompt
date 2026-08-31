package copy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/stretchr/testify/require"
)

func TestCopyPhaseRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(allowedRoles)
	})

	payload, _ := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		TargetCoursePhaseID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	})

	req, _ := http.NewRequest("POST", "/api/course_phase/11111111-1111-1111-1111-111111111111/copy", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}
