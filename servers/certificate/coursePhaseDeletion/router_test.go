package coursePhaseDeletion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/certificate/db/sqlc"
	"github.com/stretchr/testify/require"
)

// The endpoint is destructive, so it must never be reachable without a token.
func TestRegisterRoutesRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/certificate/api/course_phase/:coursePhaseID"), NewCoursePhaseDeletionService(*db.New(nil), nil))

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodDelete, "/certificate/api/course_phase/"+uuid.NewString(), nil))
	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
