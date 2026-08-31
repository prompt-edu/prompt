package config

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/prompt-edu/prompt/servers/example_server/db/sqlc"
)

// ConfigService provides phase-level configuration status.
//
// It exposes a GET endpoint used by the example component to check which
// required configurations for a phase are already set and which are still missing.
// The response helps the phase settings page indicate incomplete setup steps
// (e.g. missing survey timeframe, teams, or skills) before activation.
//
// The service implements promptTypes.PhaseConfigHandler itself, so `RegisterRoutes`
// passes it straight to the SDK and there is no separate handler type to keep in sync.
type ConfigService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

func NewConfigService(queries db.Queries, conn *pgxpool.Pool) *ConfigService {
	return &ConfigService{
		queries: queries,
		conn:    conn,
	}
}

// HandlePhaseConfig godoc
// @Summary Get phase configuration status
// @Description Get configuration status flags for a course phase.
// @Tags config
// @Produce json
// @Param coursePhaseID path string true "Course Phase UUID"
// @Success 200 {object} map[string]bool
// @Failure 500 {object} map[string]string
// @Router /course_phase/{coursePhaseID}/config [get]
// HandlePhaseConfig is a placeholder implementation demonstrating the expected
// method signature for phase config handlers. It currently returns 404 until
// the actual functionality is implemented. Read the database through the
// receiver's fields (`s.queries`, `s.conn`), never through a global.
func (s *ConfigService) HandlePhaseConfig(c *gin.Context) (map[string]bool, error) {
	c.AbortWithStatus(http.StatusNotFound)
	return nil, nil
}
