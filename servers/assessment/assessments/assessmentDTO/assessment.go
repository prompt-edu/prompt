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
	AssessedAt            time.Time                `json:"assessedAt"`
	Author                string                   `json:"author"`
	AuthorID              string                   `json:"authorID"`
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
			AssessedAt:            a.AssessedAt.Time,
			Author:                a.Author,
			AuthorID:              a.AuthorID,
		})
	}
	return assessments
}
