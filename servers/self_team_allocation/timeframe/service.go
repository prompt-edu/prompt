package timeframe

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/timeframe/timeframeDTO"
	log "github.com/sirupsen/logrus"
)

type TimeframeService struct {
	queries db.Queries
}

func NewTimeframeService(queries db.Queries) *TimeframeService {
	return &TimeframeService{
		queries: queries,
	}
}

func (s *TimeframeService) GetTimeframe(ctx context.Context, coursePhaseID uuid.UUID) (timeframeDTO.Timeframe, error) {
	timeframe, err := s.queries.GetTimeframe(ctx, coursePhaseID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return timeframeDTO.Timeframe{TimeframeSet: false}, nil
	} else if err != nil {
		log.Error("could not get timeframe: ", err)
		return timeframeDTO.Timeframe{}, errors.New("could not get timeframe")
	}
	return timeframeDTO.GetTimeframeDTOFromDBModel(timeframe), nil
}

func (s *TimeframeService) SetTimeframe(ctx context.Context, coursePhaseID uuid.UUID, startTime, endTime time.Time) error {
	if !startTime.Before(endTime) {
		return errors.New("team allocation end date must be before start date")
	}

	var startTimestamp, deadlineTimestamp pgtype.Timestamp
	startTimestamp = pgtype.Timestamp{Time: startTime, Valid: true}
	deadlineTimestamp = pgtype.Timestamp{Time: endTime, Valid: true}

	err := s.queries.SetTimeframe(ctx, db.SetTimeframeParams{
		CoursePhaseID: coursePhaseID,
		Starttime:     startTimestamp,
		Endtime:       deadlineTimestamp,
	})
	if err != nil {
		log.Error("failed to set timeframe: ", err)
		return errors.New("failed to set timeframe")
	}

	return nil
}
