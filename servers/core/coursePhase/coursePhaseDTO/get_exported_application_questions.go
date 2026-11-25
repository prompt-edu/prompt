package coursePhaseDTO

import (
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// ExportedApplicationQuestion is an application question whose answers reach a later phase
// in prevData.applicationAnswers, under the same questionID.
type ExportedApplicationQuestion struct {
	QuestionID uuid.UUID `json:"questionID"`
	Title      string    `json:"title"`
	Type       string    `json:"type" enums:"text,multiselect"`
	OrderNum   int32     `json:"orderNum"`
}

// ExportedApplicationQuestions lists the application questions a phase can read answers from.
type ExportedApplicationQuestions struct {
	// ApplicationCoursePhaseID is the Application phase that passes its answers to this phase,
	// or null when the course has no such connection.
	ApplicationCoursePhaseID *uuid.UUID                    `json:"applicationCoursePhaseID"`
	Questions                []ExportedApplicationQuestion `json:"questions"`
}

func GetExportedApplicationQuestionsDTO(applicationCoursePhaseID *uuid.UUID, questions []db.GetExportedApplicationQuestionsForCoursePhaseRow) ExportedApplicationQuestions {
	questionDTOs := make([]ExportedApplicationQuestion, 0, len(questions))
	for _, question := range questions {
		questionDTOs = append(questionDTOs, ExportedApplicationQuestion{
			QuestionID: question.ID,
			Title:      question.Title,
			Type:       question.QuestionType,
			OrderNum:   question.OrderNum,
		})
	}

	return ExportedApplicationQuestions{
		ApplicationCoursePhaseID: applicationCoursePhaseID,
		Questions:                questionDTOs,
	}
}
