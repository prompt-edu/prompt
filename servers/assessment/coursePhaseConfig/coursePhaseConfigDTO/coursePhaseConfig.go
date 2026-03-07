package coursePhaseConfigDTO

import (
	"time"

	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
)

type CoursePhaseConfig struct {
	CoursePhaseID            uuid.UUID `json:"coursePhaseID"`
	AssessmentSchemaID       uuid.UUID `json:"assessmentSchemaID"`
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
	GradeSuggestionVisible   bool      `json:"gradeSuggestionVisible"`
	ActionItemsVisible       bool      `json:"actionItemsVisible"`
	ResultsReleased          bool      `json:"resultsReleased"`
	GradingSheetVisible      bool      `json:"gradingSheetVisible"`
}

func MapDBCoursePhaseConfigToDTOCoursePhaseConfig(dbConfig db.CoursePhaseConfig) CoursePhaseConfig {
	return CoursePhaseConfig{
		CoursePhaseID:            dbConfig.CoursePhaseID,
		AssessmentSchemaID:       dbConfig.AssessmentSchemaID,
		Start:                    dbConfig.Start.Time,
		Deadline:                 dbConfig.Deadline.Time,
		SelfEvaluationEnabled:    dbConfig.SelfEvaluationEnabled,
		SelfEvaluationSchema:     dbConfig.SelfEvaluationSchema,
		SelfEvaluationStart:      dbConfig.SelfEvaluationStart.Time,
		SelfEvaluationDeadline:   dbConfig.SelfEvaluationDeadline.Time,
		PeerEvaluationEnabled:    dbConfig.PeerEvaluationEnabled,
		PeerEvaluationSchema:     dbConfig.PeerEvaluationSchema,
		PeerEvaluationStart:      dbConfig.PeerEvaluationStart.Time,
		PeerEvaluationDeadline:   dbConfig.PeerEvaluationDeadline.Time,
		TutorEvaluationEnabled:   dbConfig.TutorEvaluationEnabled,
		TutorEvaluationSchema:    dbConfig.TutorEvaluationSchema,
		TutorEvaluationStart:     dbConfig.TutorEvaluationStart.Time,
		TutorEvaluationDeadline:  dbConfig.TutorEvaluationDeadline.Time,
		EvaluationResultsVisible: dbConfig.EvaluationResultsVisible,
		GradeSuggestionVisible:   dbConfig.GradeSuggestionVisible,
		ActionItemsVisible:       dbConfig.ActionItemsVisible,
		ResultsReleased:          dbConfig.ResultsReleased,
		GradingSheetVisible:      dbConfig.GradingSheetVisible,
	}
}
