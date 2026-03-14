package seatPlan

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	db "github.com/prompt-edu/prompt/servers/intro_course/db/sqlc"
	"github.com/prompt-edu/prompt/servers/intro_course/seatPlan/seatPlanDTO"
	log "github.com/sirupsen/logrus"
)

type SeatPlanService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var SeatPlanServiceSingleton *SeatPlanService

func CreateSeatPlan(ctx context.Context, coursePhaseID uuid.UUID, seatNames []string) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := SeatPlanServiceSingleton.queries.CreateSeatPlan(ctxWithTimeout, db.CreateSeatPlanParams{
		CoursePhaseID: coursePhaseID,
		Seats:         seatNames,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to store seat plan")
		return errors.New("failed to create seat plan")
	}
	return nil
}

func GetSeatPlan(ctx context.Context, coursePhaseID uuid.UUID) ([]seatPlanDTO.Seat, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	seats, err := SeatPlanServiceSingleton.queries.GetSeatPlan(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to get seat plan")
		return nil, errors.New("failed to get seat plan")
	}

	return seatPlanDTO.GetSeatDTOsFromDBModels(seats), nil
}

func UpdateSeatPlan(ctx context.Context, coursePhaseID uuid.UUID, seatDTOs []seatPlanDTO.Seat) error {
	tx, err := SeatPlanServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := SeatPlanServiceSingleton.queries.WithTx(tx)

	// validate that all seatDTOs belong to the coursePhaseID
	for _, seatDTO := range seatDTOs {
		err = qtx.UpdateSeat(ctx, db.UpdateSeatParams{
			CoursePhaseID:   coursePhaseID,
			SeatName:        seatDTO.SeatName,
			HasMac:          seatDTO.HasMac,
			DeviceID:        seatDTO.DeviceID,
			AssignedStudent: seatDTO.AssignedStudent,
			AssignedTutor:   seatDTO.AssignedTutor,
		})

		if err != nil {
			log.WithFields(log.Fields{
				"coursePhaseID": coursePhaseID,
				"seatDTO":       seatDTO,
				"error":         err,
			}).Error("Failed to update seat")
			return errors.New("failed to update seat")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func DeleteSeatPlan(ctx context.Context, coursePhaseID uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := SeatPlanServiceSingleton.queries.DeleteSeatPlan(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
			"error":         err,
		}).Error("Failed to delete seat plan")
		return errors.New("failed to delete seat plan")
	}
	return nil
}

func GetOwnSeatAssignment(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (seatPlanDTO.SeatAssignment, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	seat, err := SeatPlanServiceSingleton.queries.GetOwnSeatAssignment(ctxWithTimeout, db.GetOwnSeatAssignmentParams{
		CoursePhaseID:   coursePhaseID,
		AssignedStudent: pgtype.UUID{Bytes: courseParticipationID, Valid: true},
	})
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// no seat assignment found
		return seatPlanDTO.SeatAssignment{}, nil
	} else if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID":         coursePhaseID,
			"courseParticipationID": courseParticipationID,
			"error":                 err,
		}).Error("Failed to get own seat assignment")
		return seatPlanDTO.SeatAssignment{}, errors.New("failed to get own seat assignment")
	}

	return seatPlanDTO.GetSeatAssignmentDTOFromDBModel(seat), nil
}
