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
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/stretchr/testify/require"
)

func TestRegisterRoutesRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/certificate/api"), NewCopyService(*db.New(nil)))

	body, err := json.Marshal(promptTypes.PhaseCopyRequest{
		SourceCoursePhaseID: uuid.New(),
		TargetCoursePhaseID: uuid.New(),
	})
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/certificate/api/copy", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(resp, request)
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
