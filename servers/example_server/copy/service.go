package copy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt-sdk/audit"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
)

// auditCopyAction labels the copy route on both the denied and the completed
// path, so filtering the audit log by action finds every attempt.
const auditCopyAction = "Copied course phase"

// CopyService handles phase-level data duplication.
//
// It implements the `/copy` endpoint for a phase server, used during course
// deep copy operations to replicate all phase-specific data from a source
// course phase to a new target phase. Each phase server defines its own
// copy handler to duplicate relevant entities (e.g. skills, teams, or configs)
// and persist them under the target phase ID.
// It is also the functionality called when a course is templated to set up
// a new phase based on an existing one.
//
// The service implements promptTypes.PhaseCopyHandler itself, so `RegisterRoutes`
// passes it straight to the SDK and there is no separate handler type to keep in sync.
type CopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewCopyService(queries db.Queries, conn *pgxpool.Pool) *CopyService {
	return &CopyService{
		queries: queries,
		conn:    conn,
	}
}

// HandlePhaseCopy godoc
// @Summary Copy course phase data
// @Description Copy example server data from a source course phase to a target course phase.
// @Tags copy
// @Accept json
// @Produce json
// @Param payload body promptTypes.PhaseCopyRequest true "Copy request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /copy [post]
// HandlePhaseCopy is a placeholder implementation demonstrating the expected
// method signature for phase copy handlers. It currently returns 404 until
// the actual functionality is implemented. Copy inside a transaction taken from
// the receiver's pool (`s.conn`), never through a global.
func (s *CopyService) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	recordCopyAudit(c, req)

	c.AbortWithStatus(http.StatusNotFound)
	return nil
}

// recordCopyAudit scopes the event to the target phase. The route sits outside
// :coursePhaseID, so an automatically captured event would carry no phase and
// never reach the course audit log. A blank target is left to that automatic
// entry rather than pinning the log to the nil phase.
func recordCopyAudit(c *gin.Context, req promptTypes.PhaseCopyRequest) {
	if req.TargetCoursePhaseID == uuid.Nil {
		return
	}
	audit.Record(c, audit.Event{
		Action:        auditCopyAction,
		EntityType:    "coursePhase",
		EntityID:      req.TargetCoursePhaseID.String(),
		CoursePhaseID: req.TargetCoursePhaseID.String(),
		Metadata:      map[string]any{"sourceCoursePhaseID": req.SourceCoursePhaseID.String()},
	})
}
