package coursePhaseParticipation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseParticipation/coursePhaseParticipationDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

type CoursePhaseParticipationService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CoursePhaseParticipationServiceSingleton *CoursePhaseParticipationService

func GetOwnCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (coursePhaseParticipationDTO.CoursePhaseParticipationStudent, error) {
	coursePhaseParticipation, err := CoursePhaseParticipationServiceSingleton.queries.GetCoursePhaseParticipationByUniversityLoginAndCoursePhase(ctx, db.GetCoursePhaseParticipationByUniversityLoginAndCoursePhaseParams{
		ToCoursePhaseID:     coursePhaseID,
		MatriculationNumber: pgtype.Text{String: matriculationNumber, Valid: true},
		UniversityLogin:     pgtype.Text{String: universityLogin, Valid: true},
	})

	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationStudent{}, err
	}

	participationDTO, err := coursePhaseParticipationDTO.GetCoursePhaseParticipationStudent(coursePhaseParticipation)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationStudent{}, err
	}

	return participationDTO, nil
}

func GetAllParticipationsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions, error) {
	coursePhaseParticipations, err := CoursePhaseParticipationServiceSingleton.queries.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPrevious(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, err
	}

	participationDTOs := make([]coursePhaseParticipationDTO.GetAllCPPsForCoursePhase, 0, len(coursePhaseParticipations))
	for _, coursePhaseParticipation := range coursePhaseParticipations {
		dto, err := coursePhaseParticipationDTO.GetAllCPPsForCoursePhaseDTOFromDBModel(coursePhaseParticipation)
		if err != nil {
			return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, err
		}
		participationDTOs = append(participationDTOs, dto)
	}

	// Get required resolutions
	resolutions, err := CoursePhaseParticipationServiceSingleton.queries.GetResolutionsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, err
	}

	resolutionDTOs := resolutionDTO.GetParticipationResolutionsDTOFromDBModels(resolutions)
	resolutionDTOs, err = resolution.ReplaceResolutionURLs(ctx, resolutionDTOs)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, errors.New("failed to replace resolution URLs")
	}

	return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{
		Participations: participationDTOs,
		Resolutions:    resolutionDTOs,
	}, nil
}

func GetCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution, error) {
	coursePhaseParticipations, err := CoursePhaseParticipationServiceSingleton.queries.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPrevious(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, err
	}

	found := false
	coursePhaseParticipation := db.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPreviousRow{}
	for _, participation := range coursePhaseParticipations {
		if participation.CourseParticipationID == courseParticipationID {
			coursePhaseParticipation = participation
			found = true
			break
		}
	}
	if !found {
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, errors.New("course phase participation not found")
	}

	participationDTO, err := coursePhaseParticipationDTO.GetAllCPPsForCoursePhaseDTOFromDBModel(coursePhaseParticipation)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, err
	}

	resolutions, err := CoursePhaseParticipationServiceSingleton.queries.GetResolutionsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, err
	}

	resolutionDTOs := resolutionDTO.GetParticipationResolutionsDTOFromDBModels(resolutions)
	resolutionDTOs, err = resolution.ReplaceResolutionURLs(ctx, resolutionDTOs)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, errors.New("failed to replace resolution URLs")
	}

	return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{
		Participation: participationDTO,
		Resolutions:   resolutionDTOs,
	}, nil
}

func CreateOrUpdateCoursePhaseParticipation(ctx context.Context, transactionQueries *db.Queries, newCoursePhaseParticipation coursePhaseParticipationDTO.CreateCoursePhaseParticipation) (coursePhaseParticipationDTO.GetCoursePhaseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &CoursePhaseParticipationServiceSingleton.queries)
	participation, err := newCoursePhaseParticipation.GetDBModel()
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.GetCoursePhaseParticipation{}, errors.New("failed to create DB model from DTO")
	}

	updatedParticipation, err := queries.CreateOrUpdateCoursePhaseParticipation(ctx, participation)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.GetCoursePhaseParticipation{}, errors.New("failed to create or update course phase participation")
	}

	return coursePhaseParticipationDTO.GetCoursePhaseParticipationDTOFromDBModel(updatedParticipation)
}

func UpdateCoursePhaseParticipation(ctx context.Context, transactionQueries *db.Queries, updatedCoursePhaseParticipation coursePhaseParticipationDTO.UpdateCoursePhaseParticipation) error {
	queries := utils.GetQueries(transactionQueries, &CoursePhaseParticipationServiceSingleton.queries)
	participation, err := updatedCoursePhaseParticipation.GetDBModel()
	if err != nil {
		log.Error(err)
		return errors.New("failed to create DB model from DTO")
	}

	_, err = queries.UpdateCoursePhaseParticipation(ctx, participation)
	if err != nil {
		log.Error(err)
		return errors.New("failed to update course phase participation")
	}

	return nil
}

func UpdateBatchCoursePhaseParticipation(ctx context.Context, createOrUpdateCoursePhaseParticipation []coursePhaseParticipationDTO.CreateCoursePhaseParticipation) ([]uuid.UUID, error) {
	tx, err := CoursePhaseParticipationServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := CoursePhaseParticipationServiceSingleton.queries.WithTx(tx)

	updatedIDs := make([]uuid.UUID, 0, len(createOrUpdateCoursePhaseParticipation))

	// Replace for loop by DB batch operation in near future
	for _, participation := range createOrUpdateCoursePhaseParticipation {
		updatedParticipation, err := CreateOrUpdateCoursePhaseParticipation(ctx, qtx, participation)
		if err != nil {
			log.Error(err)
			return nil, errors.New("failed to update course phase participation")
		}
		updatedIDs = append(updatedIDs, updatedParticipation.CourseParticipationID)
	}

	// commit if all updates were successful
	if err := tx.Commit(ctx); err != nil {
		log.Error(err)
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return updatedIDs, nil
}

func CreateIfNotExistingPhaseParticipation(ctx context.Context, transactionQueries *db.Queries, CourseParticipationID uuid.UUID, coursePhaseID uuid.UUID) (coursePhaseParticipationDTO.GetCoursePhaseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &CoursePhaseParticipationServiceSingleton.queries)
	participation, err := queries.GetCoursePhaseParticipationByCourseParticipationAndCoursePhase(ctx, db.GetCoursePhaseParticipationByCourseParticipationAndCoursePhaseParams{
		CourseParticipationID: CourseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err == nil {
		return coursePhaseParticipationDTO.GetCoursePhaseParticipationDTOFromDBModel(participation)
	} else if errors.Is(err, sql.ErrNoRows) {
		// has to be created
		passStatus := db.PassStatusNotAssessed
		return CreateOrUpdateCoursePhaseParticipation(ctx, &queries, coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
			CourseParticipationID: CourseParticipationID,
			CoursePhaseID:         coursePhaseID,
			PassStatus:            &passStatus,
		})

	} else {
		return coursePhaseParticipationDTO.GetCoursePhaseParticipation{}, err
	}
}

func BatchUpdatePassStatus(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID, passStatus db.PassStatus) ([]uuid.UUID, error) {
	// passing the coursePhaseID to query ensures that only the coursePhases that are in the course are updated
	changedParticipations, err := CoursePhaseParticipationServiceSingleton.queries.UpdateCoursePhasePassStatus(ctx, db.UpdateCoursePhasePassStatusParams{
		CourseParticipationID: courseParticipationIDs,
		CoursePhaseID:         coursePhaseID,
		PassStatus:            passStatus,
	})
	if err != nil {
		log.Error(err)
		return nil, errors.New("failed to update pass status")
	}

	return changedParticipations, nil
}

func GetStudentsOfCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]studentDTO.Student, error) {
	students, err := CoursePhaseParticipationServiceSingleton.queries.GetStudentsOfCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error(err)
		return nil, errors.New("failed to get participations")
	}

	studentDTOs := make([]studentDTO.Student, 0, len(students))
	for _, student := range students {
		dto := studentDTO.GetStudentDTOFromCourseParticipation(student)
		studentDTOs = append(studentDTOs, dto)
	}

	return studentDTOs, nil
}
