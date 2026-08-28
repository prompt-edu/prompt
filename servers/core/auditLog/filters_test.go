package auditLog

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseListFilters_RejectsUnparseableValues covers the parser directly: a
// value it cannot parse must surface as a 400 rather than fall through as "no
// filter", which would answer a typo with the full unfiltered log.
func TestParseListFilters_RejectsUnparseableValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/audit-log", func(c *gin.Context) {
		filters, err := parseListFilters(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"coursePhaseID": filters.CoursePhaseID, "limit": filters.Limit})
	})

	get := func(query string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/audit-log?"+query, nil))
		return w
	}

	rejected := []string{
		"coursePhaseID=not-a-uuid",
		"from=yesterday",
		"to=yesterday",
		"cursorTs=yesterday",
		"cursorId=not-a-uuid",
		"limit=many",
		"cursorTs=2026-08-28T12:00:00Z",
		"cursorId=11111111-1111-1111-1111-111111111111",
	}
	for _, query := range rejected {
		assert.Equal(t, http.StatusBadRequest, get(query).Code, query)
	}

	accepted := get("coursePhaseID=11111111-1111-1111-1111-111111111111&limit=10")
	require.Equal(t, http.StatusOK, accepted.Code)
	assert.Contains(t, accepted.Body.String(), "11111111-1111-1111-1111-111111111111")

	require.Equal(t, http.StatusOK, get("").Code)
}
