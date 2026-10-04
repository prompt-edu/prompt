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
	"github.com/stretchr/testify/require"
)

func TestCopyPhaseRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("/interview/api")
	// RegisterRoutes wires the real SDK auth middleware, which no request here
	// carries a token for, so the handler is reached through the SDK registrar
	// directly. TestRegisterRoutesRequiresAuthentication covers the wiring.
	promptTypes.RegisterCopyEndpoint(api, func(c *gin.Context) { c.Next() }, &InterviewCopyHandler{})

	payload, _ := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		TargetCoursePhaseID: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	})

	req, _ := http.NewRequest("POST", "/interview/api/copy", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
}

func TestRegisterRoutesRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	RegisterRoutes(router.Group("/interview/api"))

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest("POST", "/interview/api/copy", nil))

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
