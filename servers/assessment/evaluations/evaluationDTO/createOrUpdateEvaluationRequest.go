package evaluationDTO

import (
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/assessment/assessmentType"
	"github.com/prompt-edu/prompt/servers/assessment/assessments/scoreLevel/scoreLevelDTO"
)

type CreateOrUpdateEvaluationRequest struct {
	CourseParticipationID       uuid.UUID                     `json:"courseParticipationID" binding:"required,uuid"`
	CompetencyID                uuid.UUID                     `json:"competencyID" binding:"required,uuid"`
	ScoreLevel                  scoreLevelDTO.ScoreLevel      `json:"scoreLevel" binding:"required,oneof='veryBad' 'bad' 'ok' 'good' 'veryGood'"`
	AuthorCourseParticipationID uuid.UUID                     `json:"authorCourseParticipationID" binding:"required,uuid"`
	Type                        assessmentType.AssessmentType `json:"type" binding:"required,oneof='self' 'peer' 'tutor' 'assessment'"`
}
