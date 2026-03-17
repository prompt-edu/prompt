package applicationDTO

import db "github.com/prompt-edu/prompt/servers/core/db/sqlc"

type Form struct {
	QuestionsText        []QuestionText        `json:"questionsText"`
	QuestionsMultiSelect []QuestionMultiSelect `json:"questionsMultiSelect"`
	QuestionsFileUpload  []QuestionFileUpload  `json:"questionsFileUpload"`
}

func GetFormDTOFromDBModel(questionsText []db.ApplicationQuestionText, questionsMultiSelect []db.ApplicationQuestionMultiSelect, questionsFileUpload []db.ApplicationQuestionFileUpload) Form {
	applicationFormDTO := Form{
		QuestionsText:        make([]QuestionText, 0, len(questionsText)),
		QuestionsMultiSelect: make([]QuestionMultiSelect, 0, len(questionsMultiSelect)),
		QuestionsFileUpload:  make([]QuestionFileUpload, 0, len(questionsFileUpload)),
	}

	for _, question := range questionsText {
		applicationFormDTO.QuestionsText = append(applicationFormDTO.QuestionsText, GetQuestionTextDTOFromDBModel(question))
	}

	for _, question := range questionsMultiSelect {
		applicationFormDTO.QuestionsMultiSelect = append(applicationFormDTO.QuestionsMultiSelect, GetQuestionMultiSelectDTOFromDBModel(question))
	}

	for _, question := range questionsFileUpload {
		applicationFormDTO.QuestionsFileUpload = append(applicationFormDTO.QuestionsFileUpload, GetQuestionFileUploadDTOFromDBModel(question))
	}

	return applicationFormDTO
}
