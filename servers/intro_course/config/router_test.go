package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/intro_course/testutils"
	"github.com/stretchr/testify/assert"
)

func TestConfigRouterReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.Default()
	api := router.Group("/intro-course/api/course_phase/:coursePhaseID")
	setupConfigRouter(api, func(allowedRoles ...string) gin.HandlerFunc {
		return testutils.DefaultMockAuthMiddleware()
	})

	req, _ := http.NewRequest("GET", "/intro-course/api/course_phase/"+uuid.New().String()+"/config", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}
