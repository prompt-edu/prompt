package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CopyServiceSingleton *CopyService

type TeamAllocationCopyHandler struct{}

func (h *TeamAllocationCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	if req.SourceCoursePhaseID == req.TargetCoursePhaseID {
		return nil
	}

	ctx := c.Request.Context()

	tx, err := CopyServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := CopyServiceSingleton.queries.WithTx(tx)

	skills, err := qtx.GetSkillsByCoursePhase(ctx, req.SourceCoursePhaseID)
	if err != nil {
		return err
	}

	// Copy skills to the new course phase
	for _, skill := range skills {
		err := qtx.CreateSkill(ctx, db.CreateSkillParams{
			ID:            uuid.New(),
			Name:          skill.Name,
			CoursePhaseID: req.TargetCoursePhaseID,
		})
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit phase copy: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}
