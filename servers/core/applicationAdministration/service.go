package applicationAdministration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student"
	log "github.com/sirupsen/logrus"
)

type ApplicationService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var ApplicationServiceSingleton *ApplicationService

var ErrNotFound = errors.New("application was not found")
var ErrAlreadyApplied = errors.New("application already exists")
var ErrStudentDetailsDoNotMatch = errors.New("student details do not match")

func GetApplicationForm(ctx context.Context, coursePhaseID uuid.UUID) (applicationDTO.Form, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	isApplicationPhase, err := ApplicationServiceSingleton.queries.CheckIfCoursePhaseIsApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	if !isApplicationPhase {
		return applicationDTO.Form{}, errors.New("course phase is not an application phase")
	}

	applicationQuestionsText, err := ApplicationServiceSingleton.queries.GetApplicationQuestionsTextForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	applicationQuestionsMultiSelect, err := ApplicationServiceSingleton.queries.GetApplicationQuestionsMultiSelectForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	applicationFormDTO := applicationDTO.GetFormDTOFromDBModel(applicationQuestionsText, applicationQuestionsMultiSelect)

	return applicationFormDTO, nil
}

func UpdateApplicationForm(ctx context.Context, coursePhaseId uuid.UUID, form applicationDTO.UpdateForm) error {
	tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := ApplicationServiceSingleton.queries.WithTx(tx)

	// Check if course phase is application phase
	isApplicationPhase, err := qtx.CheckIfCoursePhaseIsApplicationPhase(ctx, coursePhaseId)
	if err != nil {
		log.Error(err)
		return err
	}

	if !isApplicationPhase {
		return errors.New("course phase is not an application phase")
	}

	// Delete all questions to be deleted
	for _, questionID := range form.DeleteQuestionsMultiSelect {
		err := qtx.DeleteApplicationQuestionMultiSelect(ctx, questionID)
		if err != nil {
			log.Error(err)
			return errors.New("could not delete question")
		}
	}

	for _, questionID := range form.DeleteQuestionsText {
		err := qtx.DeleteApplicationQuestionText(ctx, questionID)
		if err != nil {
			log.Error(err)
			return errors.New("could not delete question")
		}
	}

	// Create all questions to be created
	for _, question := range form.CreateQuestionsText {
		questionDBModel := question.GetDBModel()
		questionDBModel.ID = uuid.New()
		// force ensuring right course phase id -> but also checked in validation
		questionDBModel.CoursePhaseID = coursePhaseId

		err = qtx.CreateApplicationQuestionText(ctx, questionDBModel)
		if err != nil {
			log.Error(err)
			return errors.New("could not create question")
		}
	}

	for _, question := range form.CreateQuestionsMultiSelect {
		questionDBModel := question.GetDBModel()
		questionDBModel.ID = uuid.New()
		// force ensuring right course phase id -> but also checked in validation
		questionDBModel.CoursePhaseID = coursePhaseId

		err = qtx.CreateApplicationQuestionMultiSelect(ctx, questionDBModel)
		if err != nil {
			log.Error(err)
			return errors.New("could not create question")
		}
	}

	// Update the rest
	for _, question := range form.UpdateQuestionsMultiSelect {
		questionDBModel := question.GetDBModel()
		err = qtx.UpdateApplicationQuestionMultiSelect(ctx, questionDBModel)
		if err != nil {
			log.Error(err)
			return errors.New("could not update question")
		}
	}

	for _, question := range form.UpdateQuestionsText {
		questionDBModel := question.GetDBModel()
		err = qtx.UpdateApplicationQuestionText(ctx, questionDBModel)
		if err != nil {
			log.Error(err)
			return errors.New("could not update question")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}

func GetOpenApplicationPhases(ctx context.Context) ([]applicationDTO.OpenApplication, error) {
	applicationCoursePhases, err := ApplicationServiceSingleton.queries.GetAllOpenApplicationPhases(ctx)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get open application phases")
	}

	openApplications := make([]applicationDTO.OpenApplication, 0, len(applicationCoursePhases))
	for _, openApplication := range applicationCoursePhases {
		applicationPhase := applicationDTO.GetOpenApplicationPhaseDTO(openApplication)
		openApplications = append(openApplications, applicationPhase)
	}

	return openApplications, nil
}

func GetApplicationFormWithDetails(ctx context.Context, coursePhaseID uuid.UUID) (applicationDTO.FormWithDetails, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()
	applicationCoursePhase, err := ApplicationServiceSingleton.queries.GetOpenApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, ErrNotFound
	}

	applicationFormText, err := ApplicationServiceSingleton.queries.GetApplicationQuestionsTextForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, errors.New("could not get application form")
	}

	applicationFormMultiSelect, err := ApplicationServiceSingleton.queries.GetApplicationQuestionsMultiSelectForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, errors.New("could not get application form")
	}

	openApplicationDTO := applicationDTO.GetFormWithDetailsDTOFromDBModel(applicationCoursePhase, applicationFormText, applicationFormMultiSelect)

	return openApplicationDTO, nil
}

func PostApplicationExtern(ctx context.Context, coursePhaseID uuid.UUID, application applicationDTO.PostApplication) (uuid.UUID, error) {
	tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := ApplicationServiceSingleton.queries.WithTx(tx)

	// 1. Check if studentObj with this email already exists
	studentObj, err := student.GetStudentByEmail(ctx, application.Student.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error(err)
		return uuid.Nil, errors.New("could save the application")
	}

	// this means that a student with this email exists
	if err == nil {
		// check if student details are the same
		if studentObj.FirstName != application.Student.FirstName || studentObj.LastName != application.Student.LastName {
			return uuid.Nil, ErrStudentDetailsDoNotMatch
		}

		// check if student already applied -> External students are not allowed to apply twice
		exists, err := qtx.GetApplicationExistsForStudent(ctx, db.GetApplicationExistsForStudentParams{StudentID: studentObj.ID, ID: coursePhaseID})
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not get existing student")
		}
		if exists {
			return uuid.Nil, ErrAlreadyApplied
		}
	} else {
		// create student
		studentObj, err = student.CreateStudent(ctx, qtx, application.Student)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save student")
		}
	}

	// 2. Create Course and Course Phase Participation
	courseID, err := qtx.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not find the application")
	}

	cParticipation, err := courseParticipation.CreateCourseParticipation(ctx, qtx, courseParticipationDTO.CreateCourseParticipation{StudentID: studentObj.ID, CourseID: courseID})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not create course participation")
	}

	cPhaseParticipation, err := coursePhaseParticipation.CreateOrUpdateCoursePhaseParticipation(ctx, qtx, coursePhaseParticipationDTO.CreateCoursePhaseParticipation{CourseParticipationID: cParticipation.ID, CoursePhaseID: coursePhaseID})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not create course phase participation")
	}

	// 3. Save answers
	for _, answer := range application.AnswersText {
		answerDBModel := answer.GetDBModel()
		answerDBModel.ID = uuid.New()
		answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
		err = qtx.CreateApplicationAnswerText(ctx, answerDBModel)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could save the application answers")
		}
	}

	for _, answer := range application.AnswersMultiSelect {
		answerDBModel := answer.GetDBModel()
		answerDBModel.ID = uuid.New()
		answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
		err = qtx.CreateApplicationAnswerMultiSelect(ctx, answerDBModel)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could save the application answers")
		}
	}

	// Set Application To Passed if feature is turned on
	err = qtx.AcceptApplicationIfAutoAccept(ctx, db.AcceptApplicationIfAutoAcceptParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: cPhaseParticipation.CourseParticipationID,
	})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the application answers")
	}

	err = qtx.StoreApplicationAnswerUpdateTimestamp(ctx, db.StoreApplicationAnswerUpdateTimestampParams{
		CoursePhaseID:         cPhaseParticipation.CoursePhaseID,
		CourseParticipationID: cPhaseParticipation.CourseParticipationID,
	})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the application answers")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cPhaseParticipation.CourseParticipationID, nil
}

func GetApplicationAuthenticatedByMatriculationNumberAndUniversityLogin(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (applicationDTO.Application, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	studentObj, err := student.GetStudentByMatriculationNumberAndUniversityLogin(ctxWithTimeout, matriculationNumber, universityLogin)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return applicationDTO.Application{
			Status:             applicationDTO.StatusNewUser,
			Student:            nil,
			AnswersText:        make([]applicationDTO.AnswerText, 0),
			AnswersMultiSelect: make([]applicationDTO.AnswerMultiSelect, 0),
		}, nil
	}
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get the student")
	}

	exists, err := ApplicationServiceSingleton.queries.GetApplicationExistsForStudent(ctxWithTimeout, db.GetApplicationExistsForStudentParams{StudentID: studentObj.ID, ID: coursePhaseID})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application details")
	}

	if exists {
		// Get courseParticipation
		courseParticipation, err := ApplicationServiceSingleton.queries.GetCourseParticipationByStudentAndCoursePhaseID(ctxWithTimeout, db.GetCourseParticipationByStudentAndCoursePhaseIDParams{
			StudentID:     studentObj.ID,
			CoursePhaseID: coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get course participation")
		}

		answersText, err := ApplicationServiceSingleton.queries.GetApplicationAnswersTextForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersTextForCourseParticipationIDParams{
			CourseParticipationID: courseParticipation.ID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get application answers")
		}

		answersMultiSelect, err := ApplicationServiceSingleton.queries.GetApplicationAnswersMultiSelectForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersMultiSelectForCourseParticipationIDParams{
			CourseParticipationID: courseParticipation.ID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get application answers")
		}
		return applicationDTO.Application{
			Status:             applicationDTO.StatusApplied,
			Student:            &studentObj,
			AnswersText:        applicationDTO.GetAnswersTextDTOFromDBModels(answersText),
			AnswersMultiSelect: applicationDTO.GetAnswersMultiSelectDTOFromDBModels(answersMultiSelect),
		}, nil

	} else {
		return applicationDTO.Application{
			Status:             applicationDTO.StatusNotApplied,
			Student:            &studentObj,
			AnswersText:        make([]applicationDTO.AnswerText, 0),
			AnswersMultiSelect: make([]applicationDTO.AnswerMultiSelect, 0),
		}, nil
	}

}

func PostApplicationAuthenticatedStudent(ctx context.Context, coursePhaseID uuid.UUID, application applicationDTO.PostApplication) (uuid.UUID, error) {
	tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := ApplicationServiceSingleton.queries.WithTx(tx)

	// 1. Update student details
	studentObj, err := student.CreateOrUpdateStudent(ctx, qtx, application.Student)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the student")
	}

	// 2. Possibly Create Course and Course Phase Participation
	courseID, err := qtx.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not get the application phase")
	}

	cParticipation, err := courseParticipation.CreateIfNotExistingCourseParticipation(ctx, qtx, studentObj.ID, courseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the course participation")
	}

	cPhaseParticipation, err := coursePhaseParticipation.CreateIfNotExistingPhaseParticipation(ctx, qtx, cParticipation.ID, coursePhaseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the course phase participation")
	}

	// 3. Save answers
	for _, answer := range application.AnswersText {
		answerDBModel := answer.GetDBModel()
		answerDBModel.ID = uuid.New()
		answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
		err = qtx.CreateOrOverwriteApplicationAnswerText(ctx, db.CreateOrOverwriteApplicationAnswerTextParams(answerDBModel))
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save the application answers")
		}

	}

	for _, answer := range application.AnswersMultiSelect {
		answerDBModel := answer.GetDBModel()
		answerDBModel.ID = uuid.New()
		answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
		err = qtx.CreateOrOverwriteApplicationAnswerMultiSelect(ctx, db.CreateOrOverwriteApplicationAnswerMultiSelectParams(answerDBModel))
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save the application answers")
		}

	}

	// 4. Set Application To Passed if feature is turned on
	err = qtx.AcceptApplicationIfAutoAccept(ctx, db.AcceptApplicationIfAutoAcceptParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: cPhaseParticipation.CourseParticipationID,
	})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the application answers")
	}

	err = qtx.StoreApplicationAnswerUpdateTimestamp(ctx, db.StoreApplicationAnswerUpdateTimestampParams{
		CoursePhaseID:         cPhaseParticipation.CoursePhaseID,
		CourseParticipationID: cPhaseParticipation.CourseParticipationID,
	})
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the application answers")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return uuid.Nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return cPhaseParticipation.CourseParticipationID, nil

}

// TODO update
func GetApplicationByCPID(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (applicationDTO.Application, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	applicationExists, err := ApplicationServiceSingleton.queries.GetApplicationExists(ctxWithTimeout, db.GetApplicationExistsParams{CoursePhaseID: coursePhaseID, CourseParticipationID: courseParticipationID})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application")
	}

	if !applicationExists {
		return applicationDTO.Application{}, ErrNotFound
	}

	studentObj, err := student.GetStudentByCourseParticipationID(ctxWithTimeout, courseParticipationID)
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get student")
	}

	answersText, err := ApplicationServiceSingleton.queries.GetApplicationAnswersTextForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersTextForCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application answers")
	}

	answersMultiSelect, err := ApplicationServiceSingleton.queries.GetApplicationAnswersMultiSelectForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersMultiSelectForCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application answers")
	}

	return applicationDTO.Application{
		Status:             applicationDTO.StatusApplied,
		Student:            &studentObj,
		AnswersText:        applicationDTO.GetAnswersTextDTOFromDBModels(answersText),
		AnswersMultiSelect: applicationDTO.GetAnswersMultiSelectDTOFromDBModels(answersMultiSelect),
	}, nil
}

func GetAllApplicationParticipations(ctx context.Context, coursePhaseID uuid.UUID) ([]applicationDTO.ApplicationParticipation, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	applicationParticipations, err := ApplicationServiceSingleton.queries.GetAllApplicationParticipations(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get application participations")
	}

	applicationParticipationsDTO := make([]applicationDTO.ApplicationParticipation, 0, len(applicationParticipations))
	for _, applicationParticipation := range applicationParticipations {
		application, err := applicationDTO.GetAllCPPsForCoursePhaseDTOFromDBModel(applicationParticipation)
		if err != nil {
			log.Error(err)
			return nil, errors.New("could not get application participations")
		}
		applicationParticipationsDTO = append(applicationParticipationsDTO, application)
	}

	return applicationParticipationsDTO, nil
}

func UpdateApplicationAssessment(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID, assessment applicationDTO.PutAssessment) error {
	tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := ApplicationServiceSingleton.queries.WithTx(tx)

	if assessment.PassStatus != nil || assessment.RestrictedData.Length() > 0 {
		err := coursePhaseParticipation.UpdateCoursePhaseParticipation(ctx, qtx, coursePhaseParticipationDTO.UpdateCoursePhaseParticipation{
			CourseParticipationID: courseParticipationID,
			PassStatus:            assessment.PassStatus,
			RestrictedData:        assessment.RestrictedData,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return errors.New("could not update application assessment")
		}
	}

	if assessment.Score.Valid {
		err := qtx.UpdateApplicationAssessment(ctx, db.UpdateApplicationAssessmentParams{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: courseParticipationID,
			Score:                 assessment.Score,
		})
		if err != nil {
			log.Error(err)
			return errors.New("could not update application assessment")
		}
	}

	err = qtx.StoreApplicationAssessmentUpdateTimestamp(ctx, db.StoreApplicationAssessmentUpdateTimestampParams{
		CoursePhaseID:         coursePhaseID,
		CourseParticipationID: courseParticipationID,
	})
	if err != nil {
		log.Error(err)
		return errors.New("could not save the assessment")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func UploadAdditionalScore(ctx context.Context, coursePhaseID uuid.UUID, additionalScore applicationDTO.AdditionalScoreUpload) error {
	tx, err := ApplicationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer promptSDK.DeferDBRollback(tx, ctx)
	qtx := ApplicationServiceSingleton.queries.WithTx(tx)

	// generate batch of scores
	batchScores := make([]pgtype.Numeric, 0, len(additionalScore.Scores))
	courseParticipationIDs := make([]uuid.UUID, 0, len(additionalScore.Scores))

	for _, score := range additionalScore.Scores {
		batchScores = append(batchScores, score.Score)
		courseParticipationIDs = append(courseParticipationIDs, score.CourseParticipationID)
	}
	scoreNameArray := make([]string, 0, 1)
	scoreNameArray = append(scoreNameArray, additionalScore.Key)

	// 1.) Store the new score for each participation
	err = qtx.BatchUpdateAdditionalScores(ctx, db.BatchUpdateAdditionalScoresParams{
		CoursePhaseID:          coursePhaseID,
		CourseParticipationIds: courseParticipationIDs,
		Scores:                 batchScores,
		ScoreName:              scoreNameArray,
	})
	if err != nil {
		log.Error(err)
		return errors.New("could not update additional scores")
	}

	// 2.) Set students to failed, if under threshold
	if additionalScore.ThresholdActive && additionalScore.Threshold.Valid {
		batchSetFailed := []uuid.UUID{}
		thresholdValue, err := additionalScore.Threshold.Float64Value()
		if err != nil {
			log.Error(err)
			return errors.New("could not update additional scores")
		}

		for _, score := range additionalScore.Scores {
			scoreValue, err := score.Score.Float64Value()
			if err != nil {
				log.Error(err)
				return errors.New("could not update additional scores")
			}
			if scoreValue.Float64 < thresholdValue.Float64 {
				batchSetFailed = append(batchSetFailed, score.CourseParticipationID)
			}
		}

		// TODO MAIL: use the changed participations for mailing!
		_, err = qtx.UpdateCoursePhasePassStatus(ctx, db.UpdateCoursePhasePassStatusParams{
			CourseParticipationID: batchSetFailed,
			CoursePhaseID:         coursePhaseID,
			PassStatus:            db.PassStatusFailed,
		})
		if err != nil {
			log.Error(err)
			return errors.New("could not update additional scores")
		}
	}

	coursePhaseDTO, err := coursePhase.GetCoursePhaseByID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return errors.New("could not update additional scores")
	}

	restrictedDataUpdate, err := addScoreName(coursePhaseDTO.RestrictedData, additionalScore.Name, additionalScore.Key)
	if err != nil {
		return err
	}

	err = qtx.UpdateExistingAdditionalScores(ctx, db.UpdateExistingAdditionalScoresParams{
		ID:             coursePhaseID,
		RestrictedData: restrictedDataUpdate,
	})
	if err != nil {
		log.Error(err)
		return errors.New("could not update additional scores")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetAdditionalScores(ctx context.Context, coursePhaseID uuid.UUID) ([]applicationDTO.AdditionalScore, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseDTO, err := coursePhase.GetCoursePhaseByID(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not update additional scores")
	}

	return metaToScoresArray(coursePhaseDTO.RestrictedData)
}

func DeleteApplications(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := ApplicationServiceSingleton.queries.DeleteApplications(ctxWithTimeout, db.DeleteApplicationsParams{CoursePhaseID: coursePhaseID, CourseParticipationIds: courseParticipationIDs})
	if err != nil {
		log.Error(err)
		return errors.New("could not delete applications")
	}
	return nil
}
