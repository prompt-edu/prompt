package assessmentDTO

import (
	"time"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

// Assessment represents a simplified view of the assessment record.
type Assessment struct {
	ID                    uuid.UUID                `json:"id"`
	CourseParticipationID uuid.UUID                `json:"courseParticipationID"`
	CoursePhaseID         uuid.UUID                `json:"coursePhaseID"`
	CompetencyID          uuid.UUID                `json:"competencyID"`
	ScoreLevel            scoreLevelDTO.ScoreLevel `json:"scoreLevel"`
	Comment               string                   `json:"comment"`
	Examples              string                   `json:"examples"`
	AssessedAt            time.Time                `json:"assessedAt"`
	Author                string                   `json:"author"`
}

// GetAssessmentDTOsFromDBModels converts a slice of db.Assessment to DTOs.
func GetAssessmentDTOsFromDBModels(dbAssessments []db.Assessment) []Assessment {
	assessments := make([]Assessment, 0, len(dbAssessments))
	for _, a := range dbAssessments {
		assessments = append(assessments, Assessment{
			ID:                    a.ID,
			CourseParticipationID: a.CourseParticipationID,
			CoursePhaseID:         a.CoursePhaseID,
			CompetencyID:          a.CompetencyID,
			ScoreLevel:            scoreLevelDTO.MapDBScoreLevelToDTO(a.ScoreLevel),
			Comment:               a.Comment.String,
			Examples:              a.Examples,
			AssessedAt:            a.AssessedAt.Time,
			Author:                a.Author,
		})
	}
	return assessments
}
