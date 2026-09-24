package coursePhaseDeletion

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// The SDK builds the endpoint's auth middleware itself, so the production registration must deny a
// request without a token before the handler, which would need a database, ever runs.
func TestRegisterRoutesRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/presentation/api/course_phase/:coursePhaseID"), NewCoursePhaseDeletionService(nil, nil, nil))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodDelete,
		"/presentation/api/course_phase/"+deletedCoursePhaseID.String(), nil))

	assert.Equal(t, http.StatusUnauthorized, response.Code)
}
