package auditLog

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt-sdk/audit"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/keycloakTokenVerifier"
	"github.com/prompt-edu/prompt/servers/core/permissionValidation"
	log "github.com/sirupsen/logrus"
)

// DBSink persists audit events directly into core's database. It is core's
// implementation of the shared audit.Sink interface.
type DBSink struct {
	queries db.Queries
}

// NewDBSink builds a database-backed audit sink.
func NewDBSink(queries db.Queries) *DBSink {
	return &DBSink{queries: queries}
}

// Record implements audit.Sink.
func (s *DBSink) Record(ctx context.Context, e audit.Event) error {
	params := eventToParams(ctx, &s.queries, e)
	_, err := s.queries.CreateAuditEntry(ctx, params)
	return err
}

// courseResolver resolves a course from a course phase; satisfied by both
// db.Queries (via pointer) and a transaction-bound *db.Queries.
type courseResolver interface {
	GetCourseIDByCoursePhaseID(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
}

// eventToParams maps a neutral audit.Event onto the sqlc insert params,
// resolving course_id from the course phase when it is not already set.
func eventToParams(ctx context.Context, r courseResolver, e audit.Event) db.CreateAuditEntryParams {
	courseID := pgUUID(e.CourseID)
	if !courseID.Valid && e.CoursePhaseID != "" {
		if phaseID, err := uuid.Parse(e.CoursePhaseID); err == nil {
			if cid, err := r.GetCourseIDByCoursePhaseID(ctx, phaseID); err == nil {
				courseID = pgtype.UUID{Bytes: cid, Valid: true}
			} else {
				log.WithError(err).Debug("audit: could not resolve course from phase")
			}
		}
	}

	// Core's course-level routes carry the course in the ":uuid" param (e.g.
	// PUT/DELETE /api/courses/:uuid, .../archive, POST .../participations). The
	// SDK only fills CourseID from ":courseId"/":courseID", so backfill from the
	// entity id on these routes; otherwise course renames, archival and deletion
	// would be invisible in the course log. Scoped to the courses path so the
	// ":uuid" of unrelated routes (course_phases, students, …) is never mistaken
	// for a course.
	if !courseID.Valid && strings.HasPrefix(e.HTTPPath, "/api/courses/:uuid") && e.EntityID != "" {
		if cid, err := uuid.Parse(e.EntityID); err == nil {
			courseID = pgtype.UUID{Bytes: cid, Valid: true}
		}
	}

	roles := e.ActorRoles
	if roles == nil {
		roles = []string{}
	}

	metadata := []byte("{}")
	if len(e.Metadata) > 0 {
		if b, err := json.Marshal(e.Metadata); err == nil {
			metadata = b
		}
	}

	outcome := e.Outcome
	if outcome == "" {
		outcome = audit.OutcomeSuccess
	}

	return db.CreateAuditEntryParams{
		ActorID:       pgUUID(e.ActorID),
		ActorName:     e.ActorName,
		ActorEmail:    e.ActorEmail,
		ActorRoles:    roles,
		ActorRole:     e.ActorRole,
		Action:        e.Action,
		ActionKey:     e.ActionKey,
		Outcome:       outcome,
		EntityType:    pgText(e.EntityType),
		EntityID:      pgText(e.EntityID),
		EntityName:    pgText(e.EntityName),
		CourseID:      courseID,
		CoursePhaseID: pgUUID(e.CoursePhaseID),
		SourceService: e.SourceService,
		HttpMethod:    pgText(e.HTTPMethod),
		HttpPath:      pgText(e.HTTPPath),
		HttpStatus:    pgInt4(e.HTTPStatus),
		Metadata:      metadata,
	}
}

// CoreActorExtractor reads the authenticated actor from core's flat context
// keys (core's own verifier does not populate the SDK TokenUser struct).
func CoreActorExtractor(c *gin.Context) (audit.Actor, bool) {
	id := c.GetString(keycloakTokenVerifier.CtxUserID)
	if id == "" {
		return audit.Actor{}, false
	}
	name := strings.TrimSpace(c.GetString(keycloakTokenVerifier.CtxFirstName) + " " + c.GetString(keycloakTokenVerifier.CtxLastName))

	var roles []string
	var role string
	if v, ok := c.Get(keycloakTokenVerifier.CtxUserRoles); ok {
		if roleMap, ok := v.(map[string]bool); ok {
			for r := range roleMap {
				roles = append(roles, r)
			}
			sort.Strings(roles) // stable order (map iteration is randomized)
			role = primaryRole(roleMap)
		}
	}

	return audit.Actor{
		ID:    id,
		Name:  name,
		Email: c.GetString(keycloakTokenVerifier.CtxUserEmail),
		Role:  role,
		Roles: roles,
	}, true
}

// primaryRole picks the highest-privilege role present, used as the filterable
// actor_role. Course roles are matched by suffix (they carry a course prefix).
func primaryRole(roles map[string]bool) string {
	if roles[permissionValidation.PromptAdmin] {
		return permissionValidation.PromptAdmin
	}
	if roles[permissionValidation.PromptLecturer] {
		return permissionValidation.PromptLecturer
	}
	for _, courseRole := range []string{permissionValidation.CourseLecturer, permissionValidation.CourseEditor, permissionValidation.CourseStudent} {
		for userRole := range roles {
			if strings.HasSuffix(userRole, courseRole) {
				return courseRole
			}
		}
	}
	// No known role: pick the lexicographically smallest so the value is stable
	// across requests (map iteration order is randomized).
	fallback := ""
	for r := range roles {
		if fallback == "" || r < fallback {
			fallback = r
		}
	}
	return fallback
}
