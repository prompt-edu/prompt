package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// copyApplicationForm copies the application form—including all questions—from
// the source course phase to the target course phase.
func copyApplicationForm(c *gin.Context, qtx *db.Queries, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	applicationForm, err := getApplicationFormHelper(c, qtx, sourceCoursePhaseID)
	if err != nil {
		return fmt.Errorf("failed to retrieve source application form: %w", err)
	}

	createQuestionsText := make([]applicationDTO.CreateQuestionText, 0, len(applicationForm.QuestionsText))
	for _, question := range applicationForm.QuestionsText {
		createQuestionsText = append(createQuestionsText, applicationDTO.CreateQuestionText{
			CoursePhaseID:            targetCoursePhaseID,
			Title:                    question.Title,
			Description:              question.Description,
			Placeholder:              question.Placeholder,
			ValidationRegex:          question.ValidationRegex,
			ErrorMessage:             question.ErrorMessage,
			IsRequired:               question.IsRequired,
			AllowedLength:            question.AllowedLength,
			OrderNum:                 question.OrderNum,
			AccessibleForOtherPhases: question.AccessibleForOtherPhases,
			AccessKey:                question.AccessKey,
		})
	}

	createQuestionsMultiSelect := make([]applicationDTO.CreateQuestionMultiSelect, 0, len(applicationForm.QuestionsMultiSelect))
	for _, question := range applicationForm.QuestionsMultiSelect {
		createQuestionsMultiSelect = append(createQuestionsMultiSelect, applicationDTO.CreateQuestionMultiSelect{
			CoursePhaseID:            targetCoursePhaseID,
			Title:                    question.Title,
			Description:              question.Description,
			Placeholder:              question.Placeholder,
			ErrorMessage:             question.ErrorMessage,
			IsRequired:               question.IsRequired,
			MinSelect:                question.MinSelect,
			MaxSelect:                question.MaxSelect,
			Options:                  question.Options,
			OrderNum:                 question.OrderNum,
			AccessibleForOtherPhases: question.AccessibleForOtherPhases,
			AccessKey:                question.AccessKey,
		})
	}

	createQuestionsFileUpload := make([]applicationDTO.CreateQuestionFileUpload, 0, len(applicationForm.QuestionsFileUpload))
	for _, question := range applicationForm.QuestionsFileUpload {
		createQuestionsFileUpload = append(createQuestionsFileUpload, applicationDTO.CreateQuestionFileUpload{
			CoursePhaseID:            targetCoursePhaseID,
			Title:                    question.Title,
			Description:              question.Description,
			IsRequired:               question.IsRequired,
			AllowedFileTypes:         question.AllowedFileTypes,
			MaxFileSizeMB:            question.MaxFileSizeMB,
			OrderNum:                 question.OrderNum,
			AccessibleForOtherPhases: question.AccessibleForOtherPhases,
			AccessKey:                question.AccessKey,
		})
	}

	form := applicationDTO.UpdateForm{
		DeleteQuestionsText:        []uuid.UUID{},
		DeleteQuestionsMultiSelect: []uuid.UUID{},
		DeleteQuestionsFileUpload:  []uuid.UUID{},
		CreateQuestionsText:        createQuestionsText,
		CreateQuestionsMultiSelect: createQuestionsMultiSelect,
		CreateQuestionsFileUpload:  createQuestionsFileUpload,
		UpdateQuestionsText:        []applicationDTO.QuestionText{},
		UpdateQuestionsMultiSelect: []applicationDTO.QuestionMultiSelect{},
		UpdateQuestionsFileUpload:  []applicationDTO.QuestionFileUpload{},
	}

	if err := updateApplicationFormHelper(c, qtx, targetCoursePhaseID, form); err != nil {
		return fmt.Errorf("failed to update application form: %w", err)
	}
	return nil
}

// updateApplicationFormHelper applies updates to a course phase's application form.
// It handles creation, deletion, and updating of text and multi-select questions.
func updateApplicationFormHelper(c *gin.Context, qtx *db.Queries, coursePhaseId uuid.UUID, form applicationDTO.UpdateForm) error {
	isApplicationPhase, err := qtx.CheckIfCoursePhaseIsApplicationPhase(c, coursePhaseId)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("failed to check if course phase is application phase: %w", err)
	}

	if !isApplicationPhase {
		return fmt.Errorf("course phase is not an application phase")
	}

	for _, questionID := range form.DeleteQuestionsMultiSelect {
		if err := qtx.DeleteApplicationQuestionMultiSelect(c, questionID); err != nil {
			log.Error(err)
			return fmt.Errorf("could not delete multi-select question: %w", err)
		}
	}
	for _, questionID := range form.DeleteQuestionsText {
		if err := qtx.DeleteApplicationQuestionText(c, questionID); err != nil {
			log.Error(err)
			return fmt.Errorf("could not delete text question: %w", err)
		}
	}
	for _, questionID := range form.DeleteQuestionsFileUpload {
		if err := qtx.DeleteApplicationQuestionFileUpload(c, questionID); err != nil {
			log.Error(err)
			return fmt.Errorf("could not delete file upload question: %w", err)
		}
	}

	for _, question := range form.CreateQuestionsText {
		model := question.GetDBModel()
		model.ID = uuid.New()
		model.CoursePhaseID = coursePhaseId
		if err := qtx.CreateApplicationQuestionText(c, model); err != nil {
			log.Error(err)
			return fmt.Errorf("could not create text question: %w", err)
		}
	}
	for _, question := range form.CreateQuestionsMultiSelect {
		model := question.GetDBModel()
		model.ID = uuid.New()
		model.CoursePhaseID = coursePhaseId
		if err := qtx.CreateApplicationQuestionMultiSelect(c, model); err != nil {
			log.Error(err)
			return fmt.Errorf("could not create multi-select question: %w", err)
		}
	}
	for _, question := range form.CreateQuestionsFileUpload {
		model := question.GetDBModel()
		model.ID = uuid.New()
		model.CoursePhaseID = coursePhaseId
		if err := qtx.CreateApplicationQuestionFileUpload(c, model); err != nil {
			log.Error(err)
			return fmt.Errorf("could not create file upload question: %w", err)
		}
	}

	for _, question := range form.UpdateQuestionsMultiSelect {
		if err := qtx.UpdateApplicationQuestionMultiSelect(c, question.GetDBModel()); err != nil {
			log.Error(err)
			return fmt.Errorf("could not update multi-select question: %w", err)
		}
	}
	for _, question := range form.UpdateQuestionsText {
		if err := qtx.UpdateApplicationQuestionText(c, question.GetDBModel()); err != nil {
			log.Error(err)
			return fmt.Errorf("could not update text question: %w", err)
		}
	}
	for _, question := range form.UpdateQuestionsFileUpload {
		if err := qtx.UpdateApplicationQuestionFileUpload(c, question.GetDBModel()); err != nil {
			log.Error(err)
			return fmt.Errorf("could not update file upload question: %w", err)
		}
	}

	return nil
}

// getApplicationFormHelper retrieves the application form for the given course phase,
// including all associated questions. Returns an error if the phase is not an application phase.
func getApplicationFormHelper(c *gin.Context, qtx *db.Queries, coursePhaseID uuid.UUID) (applicationDTO.Form, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(c)
	defer cancel()

	isApplicationPhase, err := qtx.CheckIfCoursePhaseIsApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, fmt.Errorf("failed to check application phase: %w", err)
	}
	if !isApplicationPhase {
		return applicationDTO.Form{}, fmt.Errorf("course phase is not an application phase")
	}

	questionsText, err := qtx.GetApplicationQuestionsTextForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, fmt.Errorf("failed to get text questions: %w", err)
	}

	questionsMultiSelect, err := qtx.GetApplicationQuestionsMultiSelectForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, fmt.Errorf("failed to get multi-select questions: %w", err)
	}

	questionsFileUpload, err := qtx.GetApplicationQuestionsFileUploadForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, fmt.Errorf("failed to get file upload questions: %w", err)
	}

	return applicationDTO.GetFormDTOFromDBModel(questionsText, questionsMultiSelect, questionsFileUpload), nil
}
