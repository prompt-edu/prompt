package applicationDTO

import (
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
)

type PostApplication struct {
	Student            studentDTO.CreateStudent    `json:"student"` // should be able to handle either a new student or an existing dependent on ID
	AnswersText        []CreateAnswerText          `json:"answersText"`
	AnswersMultiSelect []CreateAnswerMultiSelect   `json:"answersMultiSelect"`
	AnswersFileUpload  []CreateAnswerFileUpload    `json:"answersFileUpload"`
}
