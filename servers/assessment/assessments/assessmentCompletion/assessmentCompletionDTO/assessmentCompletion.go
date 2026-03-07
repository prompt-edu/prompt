package assessmentCompletionDTO

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	"github.com/prompt-edu/prompt/servers/assessment/utils"
)

type AssessmentCompletion struct {
	CourseParticipationID uuid.UUID          `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID          `json:"coursePhaseID"`
	CompletedAt           pgtype.Timestamptz `json:"completedAt"`
	Author                string             `json:"author"`
	Completed             bool               `json:"completed"`
	Comment               string             `json:"comment"`
	GradeSuggestion       float64            `json:"gradeSuggestion"`
}

func GetAssessmentCompletionDTOsFromDBModels(dbAssessments []db.AssessmentCompletion) []AssessmentCompletion {
	assessmentCompletions := make([]AssessmentCompletion, 0, len(dbAssessments))
	for _, a := range dbAssessments {
		assessmentCompletions = append(assessmentCompletions, MapDBAssessmentCompletionToAssessmentCompletionDTO(a))
	}
	return assessmentCompletions
}

func MapDBAssessmentCompletionToAssessmentCompletionDTO(dbAssessment db.AssessmentCompletion) AssessmentCompletion {
	return AssessmentCompletion{
		CourseParticipationID: dbAssessment.CourseParticipationID,
		CoursePhaseID:         dbAssessment.CoursePhaseID,
		CompletedAt:           dbAssessment.CompletedAt,
		Author:                dbAssessment.Author,
		Completed:             dbAssessment.Completed,
		Comment:               dbAssessment.Comment,
		GradeSuggestion:       utils.MapNumericToFloat64(dbAssessment.GradeSuggestion),
	}
}
