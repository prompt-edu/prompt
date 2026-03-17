package copy

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/template_server/db/sqlc"
)

// CopyService handles phase-level data duplication.
//
// It implements the `/copy` endpoint for a phase server, used during course
// deep copy operations to replicate all phase-specific data from a source
// course phase to a new target phase. Each phase server defines its own
// copy handler to duplicate relevant entities (e.g. skills, teams, or configs)
// and persist them under the target phase ID.
// It is also the functionality called when a course is templated to set up
// a new phase based on an existing one.
type CopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CopyServiceSingleton *CopyService

type TemplateServerCopyHandler struct{}

// HandlePhaseCopy godoc
// @Summary Copy course phase data
// @Description Copy template server data from a source course phase to a target course phase.
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
// the actual functionality is implemented.
func (h *TemplateServerCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	c.AbortWithStatus(http.StatusNotFound)
	return nil
}
