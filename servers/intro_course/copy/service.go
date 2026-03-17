package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	promptTypes "github.com/prompt-edu/prompt-sdk/promptTypes"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CopyService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CopyServiceSingleton *CopyService

type IntroCourseCopyHandler struct{}

func (h *IntroCourseCopyHandler) HandlePhaseCopy(c *gin.Context, req promptTypes.PhaseCopyRequest) error {
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

	seats, err := qtx.GetSeatPlan(ctx, req.SourceCoursePhaseID)
	if err != nil {
		return err
	}

	seatsList := make([]string, 0, len(seats))
	for _, seat := range seats {
		seatsList = append(seatsList, seat.SeatName)
	}

	// Copy the seat plan
	err = qtx.CreateSeatPlan(ctx, db.CreateSeatPlanParams{
		CoursePhaseID: req.TargetCoursePhaseID,
		Seats:         seatsList,
	})
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("could not commit phase copy: ", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
