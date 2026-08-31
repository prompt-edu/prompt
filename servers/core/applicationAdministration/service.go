package applicationAdministration

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/applicationAdministration/applicationDTO"
	"github.com/prompt-edu/prompt/servers/core/course/courseParticipation/courseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/meta"
	"github.com/prompt-edu/prompt/servers/core/storage/files"
	"github.com/prompt-edu/prompt/servers/core/student"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// CoursePhaseProvider reads the course phase an application belongs to.
type CoursePhaseProvider interface {
	GetCoursePhaseByID(ctx context.Context, id uuid.UUID) (coursePhaseDTO.CoursePhase, error)
}

// CoursePhaseParticipationProvider creates and updates the participations an application produces.
type CoursePhaseParticipationProvider interface {
	CreateIfNotExistingPhaseParticipation(ctx context.Context, transactionQueries *db.Queries, courseParticipationID uuid.UUID, coursePhaseID uuid.UUID) (coursePhaseParticipationDTO.GetCoursePhaseParticipation, error)
	UpdateCoursePhaseParticipation(ctx context.Context, transactionQueries *db.Queries, updatedCoursePhaseParticipation coursePhaseParticipationDTO.UpdateCoursePhaseParticipation) error
	BatchUpdatePassStatus(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID, passStatus db.PassStatus) ([]uuid.UUID, error)
}

// StudentProvider writes the student records an application creates or updates.
type StudentProvider interface {
	CreateStudent(ctx context.Context, transactionQueries *db.Queries, student studentDTO.CreateStudent) (studentDTO.Student, error)
	UpdateStudent(ctx context.Context, transactionQueries *db.Queries, id uuid.UUID, student studentDTO.CreateStudent) (studentDTO.Student, error)
	CreateOrUpdateStudent(ctx context.Context, transactionQueries *db.Queries, studentInput studentDTO.CreateStudent) (studentDTO.Student, error)
	GetStudentByCourseParticipationID(ctx context.Context, courseParticipationID uuid.UUID) (studentDTO.Student, error)
}

// CourseParticipationProvider enrolls an applicant into the course of the application phase.
type CourseParticipationProvider interface {
	CreateIfNotExistingCourseParticipation(ctx context.Context, transactionQueries *db.Queries, studentID uuid.UUID, courseID uuid.UUID) (courseParticipationDTO.GetCourseParticipation, error)
}

// FileStorageProvider serves the files uploaded as answers to file upload questions.
type FileStorageProvider interface {
	PresignUpload(ctx context.Context, req files.PresignUploadRequest) (*files.PresignUploadResponse, error)
	CreateFileFromStorageKey(ctx context.Context, req files.CreateFileFromStorageKeyRequest, uploaderUserID, uploaderEmail string) (*files.FileResponse, error)
	GetFileByID(ctx context.Context, fileID uuid.UUID) (*files.FileResponse, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID, hardDelete bool) error
}

// ConfirmationMailer sends the confirmation mail of a submitted application.
type ConfirmationMailer interface {
	SendApplicationConfirmationMail(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID) (bool, error)
}

type ApplicationService struct {
	queries              db.Queries
	conn                 *pgxpool.Pool
	coursePhases         CoursePhaseProvider
	participations       CoursePhaseParticipationProvider
	students             StudentProvider
	courseParticipations CourseParticipationProvider
	files                FileStorageProvider
	mailer               ConfirmationMailer
}

func NewApplicationService(
	queries db.Queries,
	conn *pgxpool.Pool,
	coursePhases CoursePhaseProvider,
	participations CoursePhaseParticipationProvider,
	students StudentProvider,
	courseParticipations CourseParticipationProvider,
	files FileStorageProvider,
	mailer ConfirmationMailer,
) *ApplicationService {
	return &ApplicationService{
		queries:              queries,
		conn:                 conn,
		coursePhases:         coursePhases,
		participations:       participations,
		students:             students,
		courseParticipations: courseParticipations,
		files:                files,
		mailer:               mailer,
	}
}

var ErrNotFound = errors.New("application was not found")
var ErrAlreadyApplied = errors.New("application already exists")
var ErrStudentDetailsDoNotMatch = errors.New("student details do not match")
var ErrEmailAlreadyInUse = errors.New("email already in use")
var ErrImportAnswerTooLong = errors.New("import answer exceeds the allowed length")
var ErrUniversityLoginConflict = errors.New("university login already belongs to a different student")

func (s *ApplicationService) buildFileUploadAnswerDTOs(ctx context.Context, answers []db.ApplicationAnswerFileUpload, includeDownloadURL bool) []applicationDTO.AnswerFileUpload {
	answerDTOs := make([]applicationDTO.AnswerFileUpload, 0, len(answers))
	for _, answer := range answers {
		dto := applicationDTO.AnswerFileUpload{
			ID:                    answer.ID,
			ApplicationQuestionID: answer.ApplicationQuestionID,
			CourseParticipationID: answer.CourseParticipationID,
			FileID:                answer.FileID,
		}

		if includeDownloadURL {
			file, err := s.files.GetFileByID(ctx, answer.FileID)
			if err != nil {
				log.WithError(err).WithField("fileId", answer.FileID).Warn("Failed to load file metadata for answer")
			} else {
				dto.FileName = file.OriginalFilename
				dto.FileSize = file.SizeBytes
				dto.UploadedAt = file.CreatedAt
				dto.DownloadURL = file.DownloadURL
			}
		} else {
			file, err := s.queries.GetFileByID(ctx, answer.FileID)
			if err != nil {
				log.WithError(err).WithField("fileId", answer.FileID).Warn("Failed to load file metadata for answer")
			} else {
				dto.FileName = file.OriginalFilename
				dto.FileSize = file.SizeBytes
				dto.UploadedAt = file.CreatedAt.Time
			}
		}

		answerDTOs = append(answerDTOs, dto)
	}

	return answerDTOs
}

// createOrReplaceFileUploadAnswer creates or updates a file upload answer and returns a stale file id for cleanup after commit.
func createOrReplaceFileUploadAnswer(ctx context.Context, qtx *db.Queries, answer applicationDTO.CreateAnswerFileUpload, courseParticipationID uuid.UUID) (*uuid.UUID, error) {
	return upsertFileUploadAnswer(ctx, qtx, answer, courseParticipationID)
}

// createOrOverwriteFileUploadAnswer creates or updates a file upload answer and returns a stale file id for cleanup after commit.
func createOrOverwriteFileUploadAnswer(ctx context.Context, qtx *db.Queries, answer applicationDTO.CreateAnswerFileUpload, courseParticipationID uuid.UUID) (*uuid.UUID, error) {
	return upsertFileUploadAnswer(ctx, qtx, answer, courseParticipationID)
}

func upsertFileUploadAnswer(ctx context.Context, qtx *db.Queries, answer applicationDTO.CreateAnswerFileUpload, courseParticipationID uuid.UUID) (*uuid.UUID, error) {
	if answer.FileID == uuid.Nil {
		return nil, nil
	}

	// Check if there's an existing file upload answer for this question
	existingAnswer, err := qtx.GetApplicationAnswerFileUploadByQuestionAndParticipation(ctx, db.GetApplicationAnswerFileUploadByQuestionAndParticipationParams{
		ApplicationQuestionID: answer.ApplicationQuestionID,
		CourseParticipationID: courseParticipationID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	var oldFileID *uuid.UUID
	if err == nil {
		if existingAnswer.FileID != answer.FileID {
			existingFileID := existingAnswer.FileID
			oldFileID = &existingFileID
		}
	}

	// Create or overwrite the answer in the same transaction.
	answerDBModel := answer.GetDBModel()
	answerDBModel.ID = uuid.New()
	answerDBModel.CourseParticipationID = courseParticipationID
	if err := qtx.CreateOrOverwriteApplicationAnswerFileUpload(ctx, db.CreateOrOverwriteApplicationAnswerFileUploadParams(answerDBModel)); err != nil {
		return nil, err
	}

	return oldFileID, nil
}

func (s *ApplicationService) cleanupReplacedFiles(ctx context.Context, fileIDs []uuid.UUID) {
	if len(fileIDs) == 0 {
		return
	}

	seenFileIDs := make(map[uuid.UUID]struct{}, len(fileIDs))
	for _, fileID := range fileIDs {
		if fileID == uuid.Nil {
			continue
		}
		if _, seen := seenFileIDs[fileID]; seen {
			continue
		}
		seenFileIDs[fileID] = struct{}{}

		if err := s.files.DeleteFile(ctx, fileID, true); err != nil {
			log.WithError(err).WithField("fileId", fileID).Warn("Failed to delete replaced file after transaction commit")
		}
	}
}

func (s *ApplicationService) GetApplicationForm(ctx context.Context, coursePhaseID uuid.UUID) (applicationDTO.Form, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	isApplicationPhase, err := s.queries.CheckIfCoursePhaseIsApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	if !isApplicationPhase {
		return applicationDTO.Form{}, errors.New("course phase is not an application phase")
	}

	applicationQuestionsText, err := s.queries.GetApplicationQuestionsTextForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	applicationQuestionsMultiSelect, err := s.queries.GetApplicationQuestionsMultiSelectForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	applicationQuestionsFileUpload, err := s.queries.GetApplicationQuestionsFileUploadForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		return applicationDTO.Form{}, err
	}

	applicationFormDTO := applicationDTO.GetFormDTOFromDBModel(applicationQuestionsText, applicationQuestionsMultiSelect, applicationQuestionsFileUpload)

	return applicationFormDTO, nil
}

func (s *ApplicationService) UpdateApplicationForm(ctx context.Context, coursePhaseId uuid.UUID, form applicationDTO.UpdateForm) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

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

	for _, questionID := range form.DeleteQuestionsFileUpload {
		err := qtx.DeleteApplicationQuestionFileUpload(ctx, questionID)
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

	for _, question := range form.CreateQuestionsFileUpload {
		questionDBModel := question.GetDBModel()
		questionDBModel.ID = uuid.New()
		// force ensuring right course phase id -> but also checked in validation
		questionDBModel.CoursePhaseID = coursePhaseId

		err = qtx.CreateApplicationQuestionFileUpload(ctx, questionDBModel)
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

	for _, question := range form.UpdateQuestionsFileUpload {
		questionDBModel := question.GetDBModel()
		err = qtx.UpdateApplicationQuestionFileUpload(ctx, questionDBModel)
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

func (s *ApplicationService) GetOpenApplicationPhases(ctx context.Context) ([]applicationDTO.OpenApplication, error) {
	applicationCoursePhases, err := s.queries.GetAllOpenApplicationPhases(ctx)
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

func (s *ApplicationService) GetApplicationFormWithDetails(ctx context.Context, coursePhaseID uuid.UUID) (applicationDTO.FormWithDetails, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()
	applicationCoursePhase, err := s.queries.GetOpenApplicationPhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, ErrNotFound
	}

	applicationFormText, err := s.queries.GetApplicationQuestionsTextForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, errors.New("could not get application form")
	}

	applicationFormMultiSelect, err := s.queries.GetApplicationQuestionsMultiSelectForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, errors.New("could not get application form")
	}

	applicationFormFileUpload, err := s.queries.GetApplicationQuestionsFileUploadForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.FormWithDetails{}, errors.New("could not get application form")
	}

	openApplicationDTO := applicationDTO.GetFormWithDetailsDTOFromDBModel(applicationCoursePhase, applicationFormText, applicationFormMultiSelect, applicationFormFileUpload)

	return openApplicationDTO, nil
}

func (s *ApplicationService) PostApplicationExtern(ctx context.Context, coursePhaseID uuid.UUID, application applicationDTO.PostApplication) (uuid.UUID, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	queries := utils.GetQueries(qtx, &s.queries)

	// 1. Check if studentObj with this email already exists
	studentObj, err := student.GetStudentByEmail(ctx, &queries, application.Student.Email)
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
		studentObj, err = s.students.CreateStudent(ctx, qtx, application.Student)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save student")
		}
	}
	// 2. Possibly Create Course and Course Phase Participation
	courseID, err := qtx.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not get the application phase")
	}

	cParticipation, err := s.courseParticipations.CreateIfNotExistingCourseParticipation(ctx, qtx, studentObj.ID, courseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the course participation")
	}

	cPhaseParticipation, err := s.participations.CreateIfNotExistingPhaseParticipation(ctx, qtx, cParticipation.ID, coursePhaseID)
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
			return uuid.Nil, errors.New("could not save the application answers")
		}
	}

	for _, answer := range application.AnswersMultiSelect {
		answerDBModel := answer.GetDBModel()
		answerDBModel.ID = uuid.New()
		answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
		err = qtx.CreateApplicationAnswerMultiSelect(ctx, answerDBModel)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save the application answers")
		}
	}

	replacedFileIDs := make([]uuid.UUID, 0, len(application.AnswersFileUpload))
	for _, answer := range application.AnswersFileUpload {
		var oldFileID *uuid.UUID
		oldFileID, err = createOrReplaceFileUploadAnswer(ctx, qtx, answer, cPhaseParticipation.CourseParticipationID)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save the application answers")
		}
		if oldFileID != nil {
			replacedFileIDs = append(replacedFileIDs, *oldFileID)
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

	s.cleanupReplacedFiles(ctx, replacedFileIDs)

	return cPhaseParticipation.CourseParticipationID, nil
}

func (s *ApplicationService) SyncStudentDetailsFromToken(ctx context.Context, studentObj studentDTO.Student) error {
	_, err := s.students.UpdateStudent(ctx, nil, studentObj.ID, studentDTO.CreateStudent(studentObj))
	return err
}

func (s *ApplicationService) GetApplicationAuthenticatedByMatriculationNumberAndUniversityLogin(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (applicationDTO.Application, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	studentObj, err := student.ResolveStudentByUniversityCredentials(ctxWithTimeout, &s.queries, matriculationNumber, universityLogin)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return applicationDTO.Application{
			Status:             applicationDTO.StatusNewUser,
			Student:            nil,
			AnswersText:        make([]applicationDTO.AnswerText, 0),
			AnswersMultiSelect: make([]applicationDTO.AnswerMultiSelect, 0),
			AnswersFileUpload:  make([]applicationDTO.AnswerFileUpload, 0),
		}, nil
	}
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get the student")
	}

	exists, err := s.queries.GetApplicationExistsForStudent(ctxWithTimeout, db.GetApplicationExistsForStudentParams{StudentID: studentObj.ID, ID: coursePhaseID})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application details")
	}

	if exists {
		// Get courseParticipation
		courseParticipation, err := s.queries.GetCourseParticipationByStudentAndCoursePhaseID(ctxWithTimeout, db.GetCourseParticipationByStudentAndCoursePhaseIDParams{
			StudentID:     studentObj.ID,
			CoursePhaseID: coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get course participation")
		}

		answersText, err := s.queries.GetApplicationAnswersTextForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersTextForCourseParticipationIDParams{
			CourseParticipationID: courseParticipation.ID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get application answers")
		}

		answersMultiSelect, err := s.queries.GetApplicationAnswersMultiSelectForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersMultiSelectForCourseParticipationIDParams{
			CourseParticipationID: courseParticipation.ID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get application answers")
		}

		answersFileUpload, err := s.queries.GetApplicationAnswersFileUploadForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersFileUploadForCourseParticipationIDParams{
			CourseParticipationID: courseParticipation.ID,
			CoursePhaseID:         coursePhaseID,
		})
		if err != nil {
			log.Error(err)
			return applicationDTO.Application{}, errors.New("could not get application answers")
		}

		return applicationDTO.Application{
			ID:                 courseParticipation.ID,
			Status:             applicationDTO.StatusApplied,
			Student:            &studentObj,
			AnswersText:        applicationDTO.GetAnswersTextDTOFromDBModels(answersText),
			AnswersMultiSelect: applicationDTO.GetAnswersMultiSelectDTOFromDBModels(answersMultiSelect),
			AnswersFileUpload:  s.buildFileUploadAnswerDTOs(ctxWithTimeout, answersFileUpload, true),
		}, nil

	} else {
		return applicationDTO.Application{
			ID:                 uuid.Nil,
			Status:             applicationDTO.StatusNotApplied,
			Student:            &studentObj,
			AnswersText:        make([]applicationDTO.AnswerText, 0),
			AnswersMultiSelect: make([]applicationDTO.AnswerMultiSelect, 0),
			AnswersFileUpload:  make([]applicationDTO.AnswerFileUpload, 0),
		}, nil
	}

}

func (s *ApplicationService) PostApplicationAuthenticatedStudent(ctx context.Context, coursePhaseID uuid.UUID, application applicationDTO.PostApplication) (uuid.UUID, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	// 1. Update student details
	studentObj, err := s.students.CreateOrUpdateStudent(ctx, qtx, application.Student)
	if err != nil {
		log.Error(err)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "student_email_key" {
			return uuid.Nil, ErrEmailAlreadyInUse
		}
		return uuid.Nil, errors.New("could not save the student")
	}

	// 2. Possibly Create Course and Course Phase Participation
	courseID, err := qtx.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not get the application phase")
	}

	cParticipation, err := s.courseParticipations.CreateIfNotExistingCourseParticipation(ctx, qtx, studentObj.ID, courseID)
	if err != nil {
		log.Error(err)
		return uuid.Nil, errors.New("could not save the course participation")
	}

	cPhaseParticipation, err := s.participations.CreateIfNotExistingPhaseParticipation(ctx, qtx, cParticipation.ID, coursePhaseID)
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

	replacedFileIDs := make([]uuid.UUID, 0, len(application.AnswersFileUpload))
	for _, answer := range application.AnswersFileUpload {
		var oldFileID *uuid.UUID
		oldFileID, err = createOrOverwriteFileUploadAnswer(ctx, qtx, answer, cPhaseParticipation.CourseParticipationID)
		if err != nil {
			log.Error(err)
			return uuid.Nil, errors.New("could not save the application answers")
		}
		if oldFileID != nil {
			replacedFileIDs = append(replacedFileIDs, *oldFileID)
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

	s.cleanupReplacedFiles(ctx, replacedFileIDs)

	return cPhaseParticipation.CourseParticipationID, nil

}

// TODO update
func (s *ApplicationService) GetApplicationByCPID(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (applicationDTO.Application, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	applicationExists, err := s.queries.GetApplicationExists(ctxWithTimeout, db.GetApplicationExistsParams{CoursePhaseID: coursePhaseID, CourseParticipationID: courseParticipationID})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application")
	}

	if !applicationExists {
		return applicationDTO.Application{}, ErrNotFound
	}

	studentObj, err := s.students.GetStudentByCourseParticipationID(ctxWithTimeout, courseParticipationID)
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get student")
	}

	answersText, err := s.queries.GetApplicationAnswersTextForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersTextForCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application answers")
	}

	answersMultiSelect, err := s.queries.GetApplicationAnswersMultiSelectForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersMultiSelectForCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application answers")
	}

	answersFileUpload, err := s.queries.GetApplicationAnswersFileUploadForCourseParticipationID(ctxWithTimeout, db.GetApplicationAnswersFileUploadForCourseParticipationIDParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error(err)
		return applicationDTO.Application{}, errors.New("could not get application answers")
	}

	return applicationDTO.Application{
		ID:                 courseParticipationID,
		Status:             applicationDTO.StatusApplied,
		Student:            &studentObj,
		AnswersText:        applicationDTO.GetAnswersTextDTOFromDBModels(answersText),
		AnswersMultiSelect: applicationDTO.GetAnswersMultiSelectDTOFromDBModels(answersMultiSelect),
		AnswersFileUpload:  s.buildFileUploadAnswerDTOs(ctxWithTimeout, answersFileUpload, false),
	}, nil
}

type ApplicationDataExportPerCourseParticipation struct {
	CourseParticipationID uuid.UUID                                                           `json:"courseParticipationId"`
	AnswersText           []db.GetAllApplicationAnswersTextByCourseParticipationIDsRow        `json:"answersText"`
	AnswersMultiSelect    []db.GetAllApplicationAnswersMultiSelectByCourseParticipationIDsRow `json:"answersMultiSelect"`
	AnswersFileUpload     []db.GetAllApplicationAnswersFileUploadByCourseParticipationIDsRow  `json:"answersFileUpload"`
	Assessments           []db.ApplicationAssessment                                          `json:"assessments"`
}

func (s *ApplicationService) GetAllApplicationAnswers(ctx context.Context, courseParticipationIDs []uuid.UUID) ([]ApplicationDataExportPerCourseParticipation, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	answersText, err := s.queries.GetAllApplicationAnswersTextByCourseParticipationIDs(ctxWithTimeout, courseParticipationIDs)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get application text answers")
	}

	answersMultiSelect, err := s.queries.GetAllApplicationAnswersMultiSelectByCourseParticipationIDs(ctxWithTimeout, courseParticipationIDs)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get application multi-select answers")
	}

	answersFileUpload, err := s.queries.GetAllApplicationAnswersFileUploadByCourseParticipationIDs(ctxWithTimeout, courseParticipationIDs)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get application file upload answers")
	}

	assessments, err := s.queries.GetAllApplicationAssessmentsByCourseParticipationIDs(ctxWithTimeout, courseParticipationIDs)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not get application assessments")
	}

	// Group by course participation ID
	grouped := make(map[uuid.UUID]*ApplicationDataExportPerCourseParticipation)
	getOrCreate := func(cpID uuid.UUID) *ApplicationDataExportPerCourseParticipation {
		if entry, ok := grouped[cpID]; ok {
			return entry
		}
		entry := &ApplicationDataExportPerCourseParticipation{CourseParticipationID: cpID}
		grouped[cpID] = entry
		return entry
	}

	for _, a := range answersText {
		entry := getOrCreate(a.CourseParticipationID)
		entry.AnswersText = append(entry.AnswersText, a)
	}
	for _, a := range answersMultiSelect {
		entry := getOrCreate(a.CourseParticipationID)
		entry.AnswersMultiSelect = append(entry.AnswersMultiSelect, a)
	}
	for _, a := range answersFileUpload {
		entry := getOrCreate(a.CourseParticipationID)
		entry.AnswersFileUpload = append(entry.AnswersFileUpload, a)
	}
	for _, a := range assessments {
		entry := getOrCreate(a.CourseParticipationID)
		entry.Assessments = append(entry.Assessments, a)
	}

	result := make([]ApplicationDataExportPerCourseParticipation, 0, len(grouped))
	for _, entry := range grouped {
		result = append(result, *entry)
	}
	return result, nil
}

func (s *ApplicationService) GetApplicationFileUploadAnswers(ctx context.Context, coursecourseParticipationIDS []uuid.UUID) []db.GetAllApplicationAnswersFileUploadByCourseParticipationIDsRow {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	answers, err := s.queries.GetAllApplicationAnswersFileUploadByCourseParticipationIDs(ctxWithTimeout, coursecourseParticipationIDS)
	if err != nil {
		log.Error(err)
		return nil
	}
	return answers
}

func (s *ApplicationService) GetApplicationFileUploadAnswersWithFileRecord(ctx context.Context, coursecourseParticipationIDS []uuid.UUID) []db.GetAllApplicationAnswersFileUploadWithFileRecordByCourseParticipationIDsRow {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	answersWithFileRecords, err := s.queries.GetAllApplicationAnswersFileUploadWithFileRecordByCourseParticipationIDs(ctxWithTimeout, coursecourseParticipationIDS)
	if err != nil {
		log.Error(err)
		return nil
	}
	return answersWithFileRecords
}

func (s *ApplicationService) GetAllApplicationParticipations(ctx context.Context, coursePhaseID uuid.UUID) ([]applicationDTO.ApplicationParticipation, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	applicationParticipations, err := s.queries.GetAllApplicationParticipations(ctxWithTimeout, coursePhaseID)
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

func (s *ApplicationService) GetExportedApplicationAnswers(ctx context.Context, coursePhaseID uuid.UUID) (applicationDTO.ExportedApplicationAnswersResponse, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	questions, err := s.queries.GetExportedApplicationQuestionsForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.ExportedApplicationAnswersResponse{}, errors.New("could not get exported application questions")
	}

	columns := make([]applicationDTO.ExportedAnswerColumn, 0, len(questions))
	for _, question := range questions {
		columns = append(columns, applicationDTO.ExportedAnswerColumn{
			QuestionID: question.ID,
			Key:        question.AccessKey,
			Title:      question.Title,
			OrderNum:   question.OrderNum,
			Type:       question.QuestionType,
		})
	}

	if len(columns) == 0 {
		return applicationDTO.ExportedApplicationAnswersResponse{
			Columns: columns,
			Answers: make([]applicationDTO.ParticipationExportedAnswers, 0),
		}, nil
	}

	answers, err := s.queries.GetExportedApplicationAnswersForCoursePhase(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.ExportedApplicationAnswersResponse{}, errors.New("could not get exported application answers")
	}

	// Only participations with at least one answer are emitted; the client falls back to
	// a placeholder for the ones it does not find.
	answersByParticipation := make(map[uuid.UUID][]applicationDTO.ExportedAnswer)
	participationOrder := make([]uuid.UUID, 0)
	for _, answer := range answers {
		if _, seen := answersByParticipation[answer.CourseParticipationID]; !seen {
			participationOrder = append(participationOrder, answer.CourseParticipationID)
		}
		answersByParticipation[answer.CourseParticipationID] = append(
			answersByParticipation[answer.CourseParticipationID],
			applicationDTO.ExportedAnswer{
				QuestionID: answer.ApplicationQuestionID,
				Answer:     answer.Answer,
			},
		)
	}

	exportedAnswers := make([]applicationDTO.ParticipationExportedAnswers, 0, len(participationOrder))
	for _, courseParticipationID := range participationOrder {
		exportedAnswers = append(exportedAnswers, applicationDTO.ParticipationExportedAnswers{
			CourseParticipationID: courseParticipationID,
			Answers:               answersByParticipation[courseParticipationID],
		})
	}

	return applicationDTO.ExportedApplicationAnswersResponse{
		Columns: columns,
		Answers: exportedAnswers,
	}, nil
}

func (s *ApplicationService) UpdateApplicationAssessment(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID, assessment applicationDTO.PutAssessment) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	if assessment.PassStatus != nil || assessment.RestrictedData.Length() > 0 {
		err := s.participations.UpdateCoursePhaseParticipation(ctx, qtx, coursePhaseParticipationDTO.UpdateCoursePhaseParticipation{
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

func (s *ApplicationService) UploadAdditionalScore(ctx context.Context, coursePhaseID uuid.UUID, additionalScore applicationDTO.AdditionalScoreUpload) error {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}

	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

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

	coursePhaseDTO, err := s.coursePhases.GetCoursePhaseByID(ctx, coursePhaseID)
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

func (s *ApplicationService) GetAdditionalScores(ctx context.Context, coursePhaseID uuid.UUID) ([]applicationDTO.AdditionalScore, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseDTO, err := s.coursePhases.GetCoursePhaseByID(ctxWithTimeout, coursePhaseID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("could not update additional scores")
	}

	return metaToScoresArray(coursePhaseDTO.RestrictedData)
}

// IsImportModePhase reports whether the application phase is configured for CSV import instead of
// the public application form (restricted_data.applicationMode == "import").
func (s *ApplicationService) IsImportModePhase(ctx context.Context, coursePhaseID uuid.UUID) (bool, error) {
	mode, err := s.queries.GetApplicationModeForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return false, err
	}
	return mode == "import", nil
}

// PostApplicationImport imports a batch of students into an import-mode application phase.
// It creates the text questions for the mapped columns (reusing existing questions with the same
// title on re-import), upserts each student and its participations, stores the answers and applies
// the chosen pass status to the whole batch. The whole import runs in a single transaction, so a
// failing row rolls back everything.
func (s *ApplicationService) PostApplicationImport(ctx context.Context, coursePhaseID uuid.UUID, req applicationDTO.ImportApplicationRequest) (applicationDTO.ImportResult, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return applicationDTO.ImportResult{}, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)
	queries := utils.GetQueries(qtx, &s.queries)

	courseID, err := qtx.GetCourseIDByCoursePhaseID(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.ImportResult{}, errors.New("could not get the application phase")
	}

	// 1. Resolve the question IDs for the imported columns, reusing existing questions by title.
	existingQuestions, err := qtx.GetApplicationQuestionsTextForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return applicationDTO.ImportResult{}, errors.New("could not load existing questions")
	}
	existingIDByTitle := make(map[string]uuid.UUID, len(existingQuestions))
	existingLengthByTitle := make(map[string]int, len(existingQuestions))
	nextOrder := 1
	for _, q := range existingQuestions {
		existingIDByTitle[q.Title.String] = q.ID
		existingLengthByTitle[q.Title.String] = int(q.AllowedLength.Int32)
		if int(q.OrderNum.Int32) >= nextOrder {
			nextOrder = int(q.OrderNum.Int32) + 1
		}
	}

	questionIDByColumn := make(map[string]uuid.UUID, len(req.NewQuestions))
	allowedLengthByColumn := make(map[string]int, len(req.NewQuestions))
	for _, nq := range req.NewQuestions {
		if existingID, ok := existingIDByTitle[nq.Title]; ok {
			questionIDByColumn[nq.ColumnKey] = existingID
			allowedLengthByColumn[nq.ColumnKey] = existingLengthByTitle[nq.Title]
			continue
		}
		questionDBModel := applicationDTO.CreateQuestionText{
			CoursePhaseID: coursePhaseID,
			Title:         nq.Title,
			IsRequired:    false,
			AllowedLength: nq.AllowedLength,
			OrderNum:      nextOrder,
			// Set both explicitly so the insert writes the migration-0009 defaults instead of NULL,
			// matching the non-import creation path. NULL would break the Questions editor, which
			// validates accessibleForOtherPhases as a boolean.
			AccessibleForOtherPhases: pgtype.Bool{Bool: false, Valid: true},
			AccessKey:                pgtype.Text{String: "", Valid: true},
		}.GetDBModel()
		questionDBModel.ID = uuid.New()
		if err := qtx.CreateApplicationQuestionText(ctx, questionDBModel); err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not create import question")
		}
		questionIDByColumn[nq.ColumnKey] = questionDBModel.ID
		allowedLengthByColumn[nq.ColumnKey] = nq.AllowedLength
		existingIDByTitle[nq.Title] = questionDBModel.ID
		existingLengthByTitle[nq.Title] = nq.AllowedLength
		nextOrder++
	}

	// 2. Upsert each student, participation and answers.
	result := applicationDTO.ImportResult{Rows: make([]applicationDTO.ImportRowResult, 0, len(req.Rows))}
	// Only participations this import newly creates receive the chosen pass status, so a re-import
	// of an existing roster cannot silently overwrite statuses a lecturer set by hand.
	createdParticipationIDs := make([]uuid.UUID, 0, len(req.Rows))
	for i, row := range req.Rows {
		studentInput := row.Student
		studentInput.HasUniversityAccount = true
		studentInput.UniversityLogin = strings.ToLower(strings.TrimSpace(studentInput.UniversityLogin))
		studentInput.MatriculationNumber = strings.TrimSpace(studentInput.MatriculationNumber)

		// Preserve any attributes the CSV omits so a partial re-import (or a student known from
		// another course) does not have their gender/nationality/degree/etc. reset to defaults.
		if existing, resolveErr := student.ResolveStudentByUniversityCredentials(ctx, &queries, studentInput.MatriculationNumber, studentInput.UniversityLogin); resolveErr == nil {
			if studentInput.Gender == "" {
				studentInput.Gender = existing.Gender
			}
			if studentInput.Nationality == "" {
				studentInput.Nationality = existing.Nationality
			}
			if studentInput.StudyDegree == "" {
				studentInput.StudyDegree = existing.StudyDegree
			}
			if studentInput.StudyProgram == "" {
				studentInput.StudyProgram = existing.StudyProgram
			}
			if !studentInput.CurrentSemester.Valid {
				studentInput.CurrentSemester = existing.CurrentSemester
			}
		}

		// Default the enum fields for genuinely new students, since an empty string is not a valid
		// Go/Postgres enum constant.
		if studentInput.Gender == "" {
			studentInput.Gender = db.GenderPreferNotToSay
		}
		if studentInput.StudyDegree == "" {
			studentInput.StudyDegree = db.StudyDegreeBachelor
		}

		studentObj, err := s.students.CreateOrUpdateStudent(ctx, qtx, studentInput)
		if err != nil {
			log.Error(err)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "student_email_key" {
				return applicationDTO.ImportResult{}, ErrEmailAlreadyInUse
			}
			// The university login is already taken by a different student (e.g. the same login with
			// a different matriculation number). Surface a clear conflict instead of a generic 500.
			if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "student_university_login_unique" {
				return applicationDTO.ImportResult{}, fmt.Errorf("university login %q already belongs to a different student: %w", studentInput.UniversityLogin, ErrUniversityLoginConflict)
			}
			return applicationDTO.ImportResult{}, fmt.Errorf("could not save student %s: %w", studentInput.UniversityLogin, err)
		}

		cParticipation, err := s.courseParticipations.CreateIfNotExistingCourseParticipation(ctx, qtx, studentObj.ID, courseID)
		if err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not save the course participation")
		}

		// Classify created vs. updated on the phase-scoped participation (checked before we create
		// it), not on course membership, so the reported outcome matches what this phase gains.
		alreadyInPhase, err := qtx.GetApplicationExists(ctx, db.GetApplicationExistsParams{CoursePhaseID: coursePhaseID, CourseParticipationID: cParticipation.ID})
		if err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not check for existing participation")
		}
		outcome := "created"
		if alreadyInPhase {
			outcome = "updated"
		}

		cPhaseParticipation, err := s.participations.CreateIfNotExistingPhaseParticipation(ctx, qtx, cParticipation.ID, coursePhaseID)
		if err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not save the course phase participation")
		}
		if !alreadyInPhase {
			createdParticipationIDs = append(createdParticipationIDs, cPhaseParticipation.CourseParticipationID)
		}

		for _, ans := range row.Answers {
			questionID, ok := questionIDByColumn[ans.ColumnKey]
			if !ok || strings.TrimSpace(ans.Answer) == "" {
				continue
			}
			// Enforce the question's allowed length, mirroring the public-form answer validation
			// so an oversized CSV cell cannot bypass the question configuration.
			if maxLen := allowedLengthByColumn[ans.ColumnKey]; maxLen > 0 && utf8.RuneCountInString(ans.Answer) > maxLen {
				return applicationDTO.ImportResult{}, fmt.Errorf("answer for %q in column %q exceeds the allowed length of %d: %w", studentInput.UniversityLogin, ans.ColumnKey, maxLen, ErrImportAnswerTooLong)
			}
			answerDBModel := applicationDTO.CreateAnswerText{
				ApplicationQuestionID: questionID,
				Answer:                ans.Answer,
			}.GetDBModel()
			answerDBModel.ID = uuid.New()
			answerDBModel.CourseParticipationID = cPhaseParticipation.CourseParticipationID
			if err := qtx.CreateOrOverwriteApplicationAnswerText(ctx, db.CreateOrOverwriteApplicationAnswerTextParams(answerDBModel)); err != nil {
				log.Error(err)
				return applicationDTO.ImportResult{}, errors.New("could not save the application answers")
			}
		}

		if err := qtx.StoreApplicationAnswerUpdateTimestamp(ctx, db.StoreApplicationAnswerUpdateTimestampParams{
			CoursePhaseID:         coursePhaseID,
			CourseParticipationID: cPhaseParticipation.CourseParticipationID,
		}); err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not save the application answers")
		}

		cpID := cPhaseParticipation.CourseParticipationID
		result.Rows = append(result.Rows, applicationDTO.ImportRowResult{
			Index:                 i,
			UniversityLogin:       studentInput.UniversityLogin,
			Outcome:               outcome,
			CourseParticipationID: &cpID,
		})
		if outcome == "updated" {
			result.Updated++
		} else {
			result.Created++
		}
	}

	// 3. Apply the chosen pass status only to the participations this import newly created. Pre-existing
	// participations keep whatever status they already have, so a re-import cannot flip a manual decision.
	if len(createdParticipationIDs) > 0 {
		if _, err := qtx.UpdateCoursePhasePassStatus(ctx, db.UpdateCoursePhasePassStatusParams{
			CourseParticipationID: createdParticipationIDs,
			CoursePhaseID:         coursePhaseID,
			PassStatus:            req.PassStatus,
		}); err != nil {
			log.Error(err)
			return applicationDTO.ImportResult{}, errors.New("could not set the pass status")
		}
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return applicationDTO.ImportResult{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

func (s *ApplicationService) DeleteApplications(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID) error {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	err := s.queries.DeleteApplications(ctxWithTimeout, db.DeleteApplicationsParams{CoursePhaseID: coursePhaseID, CourseParticipationIds: courseParticipationIDs})
	if err != nil {
		log.Error(err)
		return errors.New("could not delete applications")
	}
	return nil
}

// InitializeApplicationCoursePhaseType creates the application course phase type together with its
// provided outputs. It is a no-op once the type exists.
func (s *ApplicationService) InitializeApplicationCoursePhaseType(ctx context.Context) error {
	// check if the application module exists in the types
	exists, err := s.queries.TestApplicationPhaseTypeExists(ctx)
	if err != nil {
		log.Error("failed to check if application phase type exists: ", err)
		return err
	}

	if !exists {
		// 1.) start transaction
		tx, err := s.conn.Begin(ctx)
		if err != nil {
			return err
		}
		defer sdkUtils.DeferRollback(tx, ctx)
		qtx := s.queries.WithTx(tx)

		// 2.) create the application module
		newApplicationPhaseType := db.CreateCoursePhaseTypeParams{
			ID:           uuid.New(),
			Name:         "Application",
			InitialPhase: true,
			BaseUrl:      "core",
		}
		err = qtx.CreateCoursePhaseType(ctx, newApplicationPhaseType)
		if err != nil {
			log.Error("failed to create application module: ", err)
		}

		// 3.) create the provided output meta data
		// 3.1 Application Score
		scoreSpecificationJson := meta.MetaData{}
		scoreSpecificationJson["type"] = "integer"
		scoreSpecificationBytes, err := scoreSpecificationJson.GetDBModel()
		if err != nil {
			log.Error("failed to parse score specification")
			return err
		}

		scoreLevelSpecificationBytes, err := getScoreLevelSpecificationBytes()
		if err != nil {
			log.Error("failed to parse score level specification")
			return err
		}

		newProvidedOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newApplicationPhaseType.ID,
			DtoName:           "score",
			Specification:     scoreSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "core",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedOutput)
		if err != nil {
			log.Error("failed to create required score input: ", err)
			return err
		}

		newProvidedScoreLevelOutput := db.CreateCoursePhaseTypeProvidedOutputParams{
			ID:                uuid.New(),
			CoursePhaseTypeID: newApplicationPhaseType.ID,
			DtoName:           "scoreLevel",
			Specification:     scoreLevelSpecificationBytes,
			VersionNumber:     1,
			EndpointPath:      "core",
		}
		err = qtx.CreateCoursePhaseTypeProvidedOutput(ctx, newProvidedScoreLevelOutput)
		if err != nil {
			log.Error("failed to create score level output: ", err)
			return err
		}

		// 3.2 Application Answers
		err = qtx.InsertCourseProvidedApplicationAnswers(ctx, newApplicationPhaseType.ID)
		if err != nil {
			log.Error("failed to create required application answers: ", err)
			return err
		}

		// 3.3 Additional Scores
		err = qtx.InsertCourseProvidedAdditionalScores(ctx, newApplicationPhaseType.ID)
		if err != nil {
			log.Error("failed to create required additional scores: ", err)
			return err
		}

		// 4.) commit the transaction
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

	} else {
		log.Debug("application module already exists")
	}

	return nil
}
