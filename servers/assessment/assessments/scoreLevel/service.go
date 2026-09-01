package scoreLevel

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type ScoreLevelService struct {
	queries db.Queries
}

func NewScoreLevelService(queries db.Queries) *ScoreLevelService {
	return &ScoreLevelService{
		queries: queries,
	}
}

func (s *ScoreLevelService) GetAllScoreLevels(ctx context.Context, coursePhaseID uuid.UUID) ([]scoreLevelDTO.ScoreLevelWithParticipation, error) {
	dbScoreLevels, err := s.queries.GetAllScoreLevels(ctx, coursePhaseID)
	if err != nil {
		log.Error("Error fetching score levels from database: ", err)
		return []scoreLevelDTO.ScoreLevelWithParticipation{}, err
	}
	scoreLevels := scoreLevelDTO.GetScoreLevelsFromDBScoreLevels(dbScoreLevels)

	return scoreLevels, nil
}

func (s *ScoreLevelService) GetScoreLevelByCourseParticipationID(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (db.ScoreLevel, error) {
	dbScoreLevel, err := s.queries.GetScoreLevelByCourseParticipationID(ctx, db.GetScoreLevelByCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Error fetching score level from database: ", err)
		return db.ScoreLevelVeryBad, err
	}

	return db.ScoreLevel(dbScoreLevel), nil
}

func (s *ScoreLevelService) GetStudentScore(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (scoreLevelDTO.StudentScore, error) {
	studentScoreWithLevel, err := s.queries.GetStudentScoreWithLevel(ctx, db.GetStudentScoreWithLevelParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Error fetching student score with level from database: ", err)
		return scoreLevelDTO.StudentScore{}, err
	}
	scoreNumeric, err := studentScoreWithLevel.ScoreNumeric.Float64Value()
	if err != nil {
		log.Error("Error converting score to float64: ", err)
		return scoreLevelDTO.StudentScore{}, err
	}

	return scoreLevelDTO.StudentScore{
		ScoreLevel:   scoreLevelDTO.MapDBScoreLevelToDTO(db.ScoreLevel(studentScoreWithLevel.ScoreLevel)),
		ScoreNumeric: scoreNumeric,
	}, nil
}
