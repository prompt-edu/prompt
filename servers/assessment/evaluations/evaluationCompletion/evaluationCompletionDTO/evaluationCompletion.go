package evaluationCompletionDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type EvaluationCompletion struct {
	ID                          uuid.UUID                     `json:"id"`
	CourseParticipationID       uuid.UUID                     `json:"courseParticipationID"`
	CoursePhaseID               uuid.UUID                     `json:"coursePhaseID"`
	AuthorCourseParticipationID uuid.UUID                     `json:"authorCourseParticipationID"`
	CompletedAt                 pgtype.Timestamptz            `json:"completedAt"`
	Completed                   bool                          `json:"completed"`
	Type                        assessmentType.AssessmentType `json:"type"`
}

// GetEvaluationCompletionDTOsFromDBModels converts a slice of db.EvaluationCompletion to DTOs.
func GetEvaluationCompletionDTOsFromDBModels(dbEvaluationCompletions []db.EvaluationCompletion) []EvaluationCompletion {
	evaluationCompletions := make([]EvaluationCompletion, 0, len(dbEvaluationCompletions))
	for _, e := range dbEvaluationCompletions {
		evaluationCompletions = append(evaluationCompletions, MapDBEvaluationCompletionToEvaluationCompletionDTO(e))
	}
	return evaluationCompletions
}

func MapDBEvaluationCompletionToEvaluationCompletionDTO(dbEvaluationCompletion db.EvaluationCompletion) EvaluationCompletion {
	return EvaluationCompletion{
		ID:                          dbEvaluationCompletion.ID,
		CourseParticipationID:       dbEvaluationCompletion.CourseParticipationID,
		CoursePhaseID:               dbEvaluationCompletion.CoursePhaseID,
		AuthorCourseParticipationID: dbEvaluationCompletion.AuthorCourseParticipationID,
		CompletedAt:                 dbEvaluationCompletion.CompletedAt,
		Completed:                   dbEvaluationCompletion.Completed,
		Type:                        assessmentType.MapDBAssessmentTypeToDTO(dbEvaluationCompletion.Type),
	}
}
