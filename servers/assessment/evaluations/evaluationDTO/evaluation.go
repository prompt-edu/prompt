package evaluationDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type Evaluation struct {
	ID                          uuid.UUID                     `json:"id"`
	CourseParticipationID       uuid.UUID                     `json:"courseParticipationID"`
	CoursePhaseID               uuid.UUID                     `json:"coursePhaseID"`
	CompetencyID                uuid.UUID                     `json:"competencyID"`
	ScoreLevel                  scoreLevelDTO.ScoreLevel      `json:"scoreLevel"`
	AuthorCourseParticipationID uuid.UUID                     `json:"authorCourseParticipationID"`
	EvaluatedAt                 pgtype.Timestamptz            `json:"evaluatedAt"`
	Type                        assessmentType.AssessmentType `json:"type"`
}

func MapDBToEvaluationDTO(evaluation db.Evaluation) Evaluation {
	return Evaluation{
		ID:                          evaluation.ID,
		CourseParticipationID:       evaluation.CourseParticipationID,
		CoursePhaseID:               evaluation.CoursePhaseID,
		CompetencyID:                evaluation.CompetencyID,
		ScoreLevel:                  scoreLevelDTO.MapDBScoreLevelToDTO(evaluation.ScoreLevel),
		AuthorCourseParticipationID: evaluation.AuthorCourseParticipationID,
		EvaluatedAt:                 evaluation.EvaluatedAt,
		Type:                        assessmentType.MapDBAssessmentTypeToDTO(evaluation.Type),
	}
}

func MapToEvaluationDTOs(evaluations []db.Evaluation) []Evaluation {
	result := make([]Evaluation, len(evaluations))
	for i, evaluation := range evaluations {
		result[i] = MapDBToEvaluationDTO(evaluation)
	}
	return result
}
