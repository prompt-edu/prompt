package coursePhaseConfigDTO

import (
	"time"

	"github.com/google/uuid"
)

type CreateOrUpdateCoursePhaseConfigRequest struct {
	AssessmentEnabled        *bool     `json:"assessmentEnabled,omitempty"`
	AssessmentSchemaID       uuid.UUID `json:"assessmentSchemaId"`
	CoursePhaseID            uuid.UUID `json:"coursePhaseId"`
	Start                    time.Time `json:"start"`
	Deadline                 time.Time `json:"deadline"`
	SelfEvaluationEnabled    bool      `json:"selfEvaluationEnabled"`
	SelfEvaluationSchema     uuid.UUID `json:"selfEvaluationSchema"`
	SelfEvaluationStart      time.Time `json:"selfEvaluationStart"`
	SelfEvaluationDeadline   time.Time `json:"selfEvaluationDeadline"`
	PeerEvaluationEnabled    bool      `json:"peerEvaluationEnabled"`
	PeerEvaluationSchema     uuid.UUID `json:"peerEvaluationSchema"`
	PeerEvaluationStart      time.Time `json:"peerEvaluationStart"`
	PeerEvaluationDeadline   time.Time `json:"peerEvaluationDeadline"`
	TutorEvaluationEnabled   bool      `json:"tutorEvaluationEnabled"`
	TutorEvaluationSchema    uuid.UUID `json:"tutorEvaluationSchema"`
	TutorEvaluationStart     time.Time `json:"tutorEvaluationStart"`
	TutorEvaluationDeadline  time.Time `json:"tutorEvaluationDeadline"`
	EvaluationResultsVisible bool      `json:"evaluationResultsVisible"`
	GradeSuggestionVisible   *bool     `json:"gradeSuggestionVisible,omitempty"`
	ActionItemsVisible       *bool     `json:"actionItemsVisible,omitempty"`
	GradingSheetVisible      *bool     `json:"gradingSheetVisible,omitempty"`
}
