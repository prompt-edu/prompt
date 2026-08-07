package auditLog

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prompt-edu/prompt-sdk/audit"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/auditLog/auditLogDTO"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	"github.com/prompt-edu/prompt/servers/core/utils"
)

const ingestServiceContextKey = "auditServiceName"

func setupAuditLogRouter(api *gin.RouterGroup, sink *DBSink) {
	authMiddleware := keycloakTokenVerifier.KeycloakMiddleware

	// Platform-wide log (admins only).
	api.GET("/audit-log",
		authMiddleware(),
		permissionValidation.CheckAccessControlByRole(permissionValidation.PromptAdmin),
		listGlobalAuditLog)

	// Per-course log (course lecturers of that course, or admins).
	api.GET("/courses/:uuid/audit-log",
		authMiddleware(),
		permissionValidation.CheckAccessControlByID(permissionValidation.CheckCoursePermission, "uuid",
			permissionValidation.CourseLecturer, permissionValidation.PromptAdmin),
		listCourseAuditLog)

	// Ingest endpoint for phase services, authenticated by a per-service shared
	// secret (not a user/Keycloak token).
	keys := parseIngestKeys(sdkUtils.GetEnv("AUDIT_INGEST_KEYS", ""))
	api.POST("/audit", ingestAuth(keys), ingestAuditEvent(sink))
}

// listGlobalAuditLog godoc
// @Summary Global audit log
// @Description Keyset-paginated, filterable platform-wide audit log (admins only).
// @Tags auditLog
// @Produce json
// @Success 200 {object} auditLogDTO.AuditLogPage
// @Failure 500 {object} utils.ErrorResponse
// @Router /audit-log [get]
func listGlobalAuditLog(c *gin.Context) {
	page, err := AuditLogServiceSingleton.ListAuditLog(c.Request.Context(), parseListFilters(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, page)
}

// listCourseAuditLog godoc
// @Summary Course audit log
// @Description Keyset-paginated, filterable audit log scoped to one course.
// @Tags auditLog
// @Produce json
// @Param uuid path string true "Course UUID"
// @Success 200 {object} auditLogDTO.AuditLogPage
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/audit-log [get]
func listCourseAuditLog(c *gin.Context) {
	filters := parseListFilters(c)
	filters.CourseID = c.Param("uuid")
	page, err := AuditLogServiceSingleton.ListAuditLog(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, page)
}

// ingestAuditEvent stores an audit event reported by a phase service. The actor
// is taken from the (trusted) request body; source_service is stamped from the
// authenticated service identity, never the body.
func ingestAuditEvent(sink *DBSink) gin.HandlerFunc {
	return func(c *gin.Context) {
		var e audit.Event
		if err := c.ShouldBindJSON(&e); err != nil {
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
			return
		}
		e.SourceService = c.GetString(ingestServiceContextKey)
		if err := sink.Record(c.Request.Context(), e); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusCreated)
	}
}

// ingestAuth validates the per-service shared secret on the ingest endpoint.
func ingestAuth(keys map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.GetHeader("X-Audit-Service")
		token := c.GetHeader("X-Audit-Token")
		expected, ok := keys[service]
		if service == "" || !ok || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "invalid audit credentials"})
			return
		}
		c.Set(ingestServiceContextKey, service)
		c.Next()
	}
}

// parseIngestKeys parses AUDIT_INGEST_KEYS ("service1:key1,service2:key2") into
// a lookup map.
func parseIngestKeys(raw string) map[string]string {
	keys := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		service, key, ok := strings.Cut(pair, ":")
		if !ok {
			continue
		}
		keys[strings.TrimSpace(service)] = strings.TrimSpace(key)
	}
	return keys
}

func parseListFilters(c *gin.Context) auditLogDTO.ListFilters {
	f := auditLogDTO.ListFilters{
		ActorRole:     c.Query("actorRole"),
		Outcome:       c.Query("outcome"),
		ActionKey:     c.Query("actionKey"),
		EntityType:    c.Query("entityType"),
		SourceService: c.Query("sourceService"),
		CoursePhaseID: c.Query("coursePhaseID"),
		Search:        c.Query("search"),
		CursorID:      c.Query("cursorId"),
	}
	if v := c.Query("from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.From = &t
		}
	}
	if v := c.Query("to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.To = &t
		}
	}
	if v := c.Query("cursorTs"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.CursorCreatedAt = &t
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			f.Limit = n
		}
	}
	return f
}
