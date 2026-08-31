package copy

import (
	"context"
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

func NewCopyService(queries db.Queries, conn *pgxpool.Pool) *CopyService {
	return &CopyService{
		queries: queries,
		conn:    conn,
	}
}

// HandlePhaseCopy implements promptTypes.PhaseCopyHandler.
func (s *CopyService) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
	return s.CopyPhase(c.Request.Context(), req.SourceCoursePhaseID, req.TargetCoursePhaseID)
}

func (s *CopyService) CopyPhase(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	if sourceCoursePhaseID == targetCoursePhaseID {
		return nil
	}

	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := s.queries.WithTx(tx)

	skills, err := qtx.GetSkillsByCoursePhase(ctx, sourceCoursePhaseID)
	if err != nil {
		return err
	}

	// Copy skills to the new course phase
	for _, skill := range skills {
		err := qtx.CreateSkill(ctx, db.CreateSkillParams{
			ID:            uuid.New(),
			Name:          skill.Name,
			CoursePhaseID: targetCoursePhaseID,
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
