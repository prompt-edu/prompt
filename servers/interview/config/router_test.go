package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	sdkTestUtils "github.com/prompt-edu/prompt-sdk/testutils"
	"github.com/stretchr/testify/require"
)

func TestGetPhaseConfigRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	RegisterRoutes(api, func(allowedRoles ...string) gin.HandlerFunc {
		return sdkTestUtils.MockAuthMiddleware(allowedRoles)
	})

	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var config map[string]bool
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &config))
	require.NotNil(t, config)
}
