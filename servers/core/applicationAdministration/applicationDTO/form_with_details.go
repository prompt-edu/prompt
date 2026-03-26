package applicationDTO

import db "github.com/prompt-edu/prompt/servers/core/db/sqlc"

type FormWithDetails struct {
	ApplicationPhase     OpenApplication       `json:"applicationPhase"`
	QuestionsText        []QuestionText        `json:"questionsText"`
	QuestionsMultiSelect []QuestionMultiSelect `json:"questionsMultiSelect"`
	QuestionsFileUpload  []QuestionFileUpload  `json:"questionsFileUpload"`
}

func GetFormWithDetailsDTOFromDBModel(applicationPhase db.GetOpenApplicationPhaseRow, questionsText []db.ApplicationQuestionText, questionsMultiSelect []db.ApplicationQuestionMultiSelect, questionsFileUpload []db.ApplicationQuestionFileUpload) FormWithDetails {
	var shortDesc *string
	if applicationPhase.ShortDescription.Valid {
		shortDesc = &applicationPhase.ShortDescription.String
	}

	var longDesc *string
	if applicationPhase.LongDescription.Valid {
		longDesc = &applicationPhase.LongDescription.String
	}

	applicationPhaseDTO := OpenApplication{
		CourseName:               applicationPhase.CourseName,
		CoursePhaseID:            applicationPhase.CoursePhaseID,
		CourseType:               string(applicationPhase.CourseType),
		ECTS:                     int(applicationPhase.Ects.Int32),
		StartDate:                applicationPhase.StartDate,
		EndDate:                  applicationPhase.EndDate,
		ApplicationDeadline:      applicationPhase.ApplicationEndDate,
		ExternalStudentsAllowed:  applicationPhase.ExternalStudentsAllowed,
		UniversityLoginAvailable: applicationPhase.UniversityLoginAvailable,
		ShortDescription:         shortDesc,
		LongDescription:          longDesc,
	}

	applicationFormDTO := FormWithDetails{
		ApplicationPhase:     applicationPhaseDTO,
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
