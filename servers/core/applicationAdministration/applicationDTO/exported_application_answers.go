package applicationDTO

import "github.com/google/uuid"

type ExportedAnswerColumn struct {
	QuestionID uuid.UUID `json:"questionID"`
	Key        string    `json:"key"`
	Title      string    `json:"title"`
	OrderNum   int32     `json:"orderNum"`
	Type       string    `json:"type"`
}

type ExportedAnswer struct {
	QuestionID uuid.UUID `json:"questionID"`
	Answer     string    `json:"answer"`
}

type ParticipationExportedAnswers struct {
	CourseParticipationID uuid.UUID        `json:"courseParticipationID"`
	Answers               []ExportedAnswer `json:"answers"`
}

type ExportedApplicationAnswersResponse struct {
	Columns []ExportedAnswerColumn         `json:"columns"`
	Answers []ParticipationExportedAnswers `json:"answers"`
}
