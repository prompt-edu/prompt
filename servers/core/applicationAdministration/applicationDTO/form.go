package applicationDTO

import db "github.com/prompt-edu/prompt/servers/core/db/sqlc"

type Form struct {
	QuestionsText        []QuestionText        `json:"questionsText"`
	QuestionsMultiSelect []QuestionMultiSelect `json:"questionsMultiSelect"`
}

func GetFormDTOFromDBModel(questionsText []db.ApplicationQuestionText, questionsMultiSelect []db.ApplicationQuestionMultiSelect) Form {
	applicationFormDTO := Form{
		QuestionsText:        make([]QuestionText, 0, len(questionsText)),
		QuestionsMultiSelect: make([]QuestionMultiSelect, 0, len(questionsMultiSelect)),
	}

	for _, question := range questionsText {
		applicationFormDTO.QuestionsText = append(applicationFormDTO.QuestionsText, GetQuestionTextDTOFromDBModel(question))
	}

	for _, question := range questionsMultiSelect {
		applicationFormDTO.QuestionsMultiSelect = append(applicationFormDTO.QuestionsMultiSelect, GetQuestionMultiSelectDTOFromDBModel(question))
	}

	return applicationFormDTO
}
