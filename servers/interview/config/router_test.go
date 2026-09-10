package config

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/promptTypes"
	"github.com/stretchr/testify/require"
)

func TestGetPhaseConfigRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("/api/course_phase/:coursePhaseID")
	// RegisterRoutes wires the real SDK auth middleware, which no request here
	// carries a token for, so the handler is reached through the SDK registrar
	// directly. TestRegisterRoutesRequiresAuthentication covers the wiring.
	promptTypes.RegisterConfigEndpoint(api, func(c *gin.Context) { c.Next() }, &configHandler{})

	req, _ := http.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)

	var config map[string]bool
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &config))
	require.NotNil(t, config)
}

func TestRegisterRoutesRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	RegisterRoutes(router.Group("/api/course_phase/:coursePhaseID"))

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest("GET", "/api/course_phase/11111111-1111-1111-1111-111111111111/config", nil))

	require.Equal(t, http.StatusUnauthorized, resp.Code)
}
