package auditLog

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
// @Param actorRole query string false "Filter by actor role"
// @Param outcome query string false "Filter by outcome (success|denied)"
// @Param actionKey query string false "Filter by action key (e.g. POST /api/.../slots)"
// @Param entityType query string false "Filter by entity type"
// @Param sourceService query string false "Filter by source service (e.g. core)"
// @Param coursePhaseID query string false "Filter by course phase UUID"
// @Param search query string false "Free-text search over actor, action and entity"
// @Param from query string false "Start of time range (RFC3339)"
// @Param to query string false "End of time range (RFC3339)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param cursorTs query string false "Keyset cursor timestamp (RFC3339) from a previous page"
// @Param cursorId query string false "Keyset cursor entry UUID from a previous page"
// @Success 200 {object} auditLogDTO.AuditLogPage
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /audit-log [get]
func listGlobalAuditLog(c *gin.Context) {
	filters, err := parseListFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}
	page, err := AuditLogServiceSingleton.ListAuditLog(c.Request.Context(), filters)
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
// @Param actorRole query string false "Filter by actor role"
// @Param outcome query string false "Filter by outcome (success|denied)"
// @Param actionKey query string false "Filter by action key (e.g. POST /api/.../slots)"
// @Param entityType query string false "Filter by entity type"
// @Param sourceService query string false "Filter by source service (e.g. core)"
// @Param coursePhaseID query string false "Filter by course phase UUID"
// @Param search query string false "Free-text search over actor, action and entity"
// @Param from query string false "Start of time range (RFC3339)"
// @Param to query string false "End of time range (RFC3339)"
// @Param limit query int false "Page size (default 50, max 200)"
// @Param cursorTs query string false "Keyset cursor timestamp (RFC3339) from a previous page"
// @Param cursorId query string false "Keyset cursor entry UUID from a previous page"
// @Success 200 {object} auditLogDTO.AuditLogPage
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /courses/{uuid}/audit-log [get]
func listCourseAuditLog(c *gin.Context) {
	filters, err := parseListFilters(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: err.Error()})
		return
	}
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
		// source_service is authoritative (from the matched key), never the body.
		e.SourceService = c.GetString(ingestServiceContextKey)
		// outcome is a closed enum; reject anything else rather than persisting a
		// forged value into the append-only log.
		switch e.Outcome {
		case "", audit.OutcomeSuccess, audit.OutcomeDenied:
		default:
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{Error: "invalid outcome"})
			return
		}
		if err := sink.Record(c.Request.Context(), e); err != nil {
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{Error: err.Error()})
			return
		}
		c.Status(http.StatusCreated)
	}
}

// ingestAuth validates the per-service shared secret on the ingest endpoint. A
// service may have more than one accepted key so a key can be rotated without
// downtime; the request authenticates if its token matches any of them.
func ingestAuth(keys map[string][]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		service := c.GetHeader("X-Audit-Service")
		token := c.GetHeader("X-Audit-Token")
		if service == "" || token == "" || !matchesAnyKey(token, keys[service]) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, utils.ErrorResponse{Error: "invalid audit credentials"})
			return
		}
		c.Set(ingestServiceContextKey, service)
		c.Next()
	}
}

// matchesAnyKey reports whether token equals one of the accepted keys using a
// constant-time compare. It does not short-circuit, so its timing does not leak
// which key matched. Empty keys are never stored (see parseIngestKeys), so an
// empty token can never authenticate.
func matchesAnyKey(token string, accepted []string) bool {
	match := false
	for _, key := range accepted {
		if subtle.ConstantTimeCompare([]byte(token), []byte(key)) == 1 {
			match = true
		}
	}
	return match
}

// parseIngestKeys parses AUDIT_INGEST_KEYS ("service1:key1,service2:key2", with
// a service repeated to list multiple accepted keys) into a per-service list of
// keys. Empty keys are dropped: subtle.ConstantTimeCompare("", "") returns 1, so
// a blank configured key would authenticate every request for that service.
func parseIngestKeys(raw string) map[string][]string {
	keys := make(map[string][]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		service, key, ok := strings.Cut(pair, ":")
		if !ok {
			continue
		}
		service = strings.TrimSpace(service)
		key = strings.TrimSpace(key)
		if service == "" || key == "" {
			continue
		}
		keys[service] = append(keys[service], key)
	}
	return keys
}

// parseListFilters reads the query parameters into ListFilters. Unparseable
// values are reported as an error (surfaced as 400) rather than silently
// dropped: a malformed "from"/"to" would otherwise return the full unfiltered
// log with a 200, and a malformed "cursorTs" would ignore the cursor and re-serve
// the first page forever, so "Load more" would loop on duplicate rows.
func parseListFilters(c *gin.Context) (auditLogDTO.ListFilters, error) {
	f := auditLogDTO.ListFilters{
		ActorRole:     c.Query("actorRole"),
		Outcome:       c.Query("outcome"),
		ActionKey:     c.Query("actionKey"),
		EntityType:    c.Query("entityType"),
		SourceService: c.Query("sourceService"),
		CoursePhaseID: c.Query("coursePhaseID"),
		Search:        c.Query("search"),
	}

	var err error
	if f.From, err = parseTimeQuery(c, "from"); err != nil {
		return f, err
	}
	if f.To, err = parseTimeQuery(c, "to"); err != nil {
		return f, err
	}
	if f.CursorCreatedAt, err = parseTimeQuery(c, "cursorTs"); err != nil {
		return f, err
	}
	if v := c.Query("cursorId"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return f, fmt.Errorf("invalid cursorId: %w", err)
		}
		f.CursorID = v
	}
	// The keyset predicate needs both cursor halves or it matches nothing.
	if (f.CursorCreatedAt != nil) != (f.CursorID != "") {
		return f, fmt.Errorf("cursorTs and cursorId must be provided together")
	}
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return f, fmt.Errorf("invalid limit: %w", err)
		}
		f.Limit = n
	}
	return f, nil
}

func parseTimeQuery(c *gin.Context, key string) (*time.Time, error) {
	v := c.Query(key)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, fmt.Errorf("invalid %s timestamp: %w", key, err)
	}
	return &t, nil
}
