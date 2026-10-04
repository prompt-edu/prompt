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
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/prompt-edu/prompt/servers/core/student/studentDTO"
	"github.com/prompt-edu/prompt/servers/core/utils"
	log "github.com/sirupsen/logrus"
)

// ResolutionReplacer rewrites the base URLs of course phase resolutions.
type ResolutionReplacer interface {
	ReplaceResolutionURLs(ctx context.Context, resolutions []resolutionDTO.Resolution) ([]resolutionDTO.Resolution, error)
}

type CoursePhaseParticipationService struct {
	queries     db.Queries
	conn        *pgxpool.Pool
	resolutions ResolutionReplacer
}

func NewCoursePhaseParticipationService(queries db.Queries, conn *pgxpool.Pool, resolutions ResolutionReplacer) *CoursePhaseParticipationService {
	return &CoursePhaseParticipationService{
		queries:     queries,
		conn:        conn,
		resolutions: resolutions,
	}
}

func (s *CoursePhaseParticipationService) GetOwnCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, matriculationNumber string, universityLogin string) (coursePhaseParticipationDTO.CoursePhaseParticipationStudent, error) {
	coursePhaseParticipation, err := s.queries.GetCoursePhaseParticipationByUniversityLoginAndCoursePhase(ctx, db.GetCoursePhaseParticipationByUniversityLoginAndCoursePhaseParams{
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

	resolutions, err := s.queries.GetResolutionsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationStudent{}, err
	}

	resolutionDTOs := resolutionDTO.GetParticipationResolutionsDTOFromDBModels(resolutions)
	resolutionDTOs, err = s.resolutions.ReplaceResolutionURLs(ctx, resolutionDTOs)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationStudent{}, errors.New("failed to replace resolution URLs")
	}
	participationDTO.Resolutions = resolutionDTOs

	return participationDTO, nil
}

func (s *CoursePhaseParticipationService) GetAllParticipationsForCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions, error) {
	coursePhaseParticipations, err := s.queries.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPrevious(ctx, coursePhaseID)
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
	resolutions, err := s.queries.GetResolutionsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, err
	}

	resolutionDTOs := resolutionDTO.GetParticipationResolutionsDTOFromDBModels(resolutions)
	resolutionDTOs, err = s.resolutions.ReplaceResolutionURLs(ctx, resolutionDTOs)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{}, errors.New("failed to replace resolution URLs")
	}

	return coursePhaseParticipationDTO.CoursePhaseParticipationsWithResolutions{
		Participations: participationDTOs,
		Resolutions:    resolutionDTOs,
	}, nil
}

func (s *CoursePhaseParticipationService) GetCoursePhaseParticipation(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationID uuid.UUID) (coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution, error) {
	coursePhaseParticipations, err := s.queries.GetAllCoursePhaseParticipationsForCoursePhaseIncludingPrevious(ctx, coursePhaseID)
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

	resolutions, err := s.queries.GetResolutionsForCoursePhase(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, err
	}

	resolutionDTOs := resolutionDTO.GetParticipationResolutionsDTOFromDBModels(resolutions)
	resolutionDTOs, err = s.resolutions.ReplaceResolutionURLs(ctx, resolutionDTOs)
	if err != nil {
		log.Error(err)
		return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{}, errors.New("failed to replace resolution URLs")
	}

	return coursePhaseParticipationDTO.CoursePhaseParticipationWithResolution{
		Participation: participationDTO,
		Resolutions:   resolutionDTOs,
	}, nil
}

func (s *CoursePhaseParticipationService) CreateOrUpdateCoursePhaseParticipation(ctx context.Context, transactionQueries *db.Queries, newCoursePhaseParticipation coursePhaseParticipationDTO.CreateCoursePhaseParticipation) (coursePhaseParticipationDTO.GetCoursePhaseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &s.queries)
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

func (s *CoursePhaseParticipationService) UpdateCoursePhaseParticipation(ctx context.Context, transactionQueries *db.Queries, updatedCoursePhaseParticipation coursePhaseParticipationDTO.UpdateCoursePhaseParticipation) error {
	queries := utils.GetQueries(transactionQueries, &s.queries)
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

func (s *CoursePhaseParticipationService) UpdateBatchCoursePhaseParticipation(ctx context.Context, createOrUpdateCoursePhaseParticipation []coursePhaseParticipationDTO.CreateCoursePhaseParticipation) ([]uuid.UUID, error) {
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer sdkUtils.DeferRollback(tx, ctx)
	qtx := s.queries.WithTx(tx)

	updatedIDs := make([]uuid.UUID, 0, len(createOrUpdateCoursePhaseParticipation))

	// Replace for loop by DB batch operation in near future
	for _, participation := range createOrUpdateCoursePhaseParticipation {
		updatedParticipation, err := s.CreateOrUpdateCoursePhaseParticipation(ctx, qtx, participation)
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

func (s *CoursePhaseParticipationService) CreateIfNotExistingPhaseParticipation(ctx context.Context, transactionQueries *db.Queries, CourseParticipationID uuid.UUID, coursePhaseID uuid.UUID) (coursePhaseParticipationDTO.GetCoursePhaseParticipation, error) {
	queries := utils.GetQueries(transactionQueries, &s.queries)
	participation, err := queries.GetCoursePhaseParticipationByCourseParticipationAndCoursePhase(ctx, db.GetCoursePhaseParticipationByCourseParticipationAndCoursePhaseParams{
		CourseParticipationID: CourseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err == nil {
		return coursePhaseParticipationDTO.GetCoursePhaseParticipationDTOFromDBModel(participation)
	} else if errors.Is(err, sql.ErrNoRows) {
		// has to be created
		passStatus := db.PassStatusNotAssessed
		return s.CreateOrUpdateCoursePhaseParticipation(ctx, &queries, coursePhaseParticipationDTO.CreateCoursePhaseParticipation{
			CourseParticipationID: CourseParticipationID,
			CoursePhaseID:         coursePhaseID,
			PassStatus:            &passStatus,
		})

	} else {
		return coursePhaseParticipationDTO.GetCoursePhaseParticipation{}, err
	}
}

func (s *CoursePhaseParticipationService) BatchUpdatePassStatus(ctx context.Context, coursePhaseID uuid.UUID, courseParticipationIDs []uuid.UUID, passStatus db.PassStatus) ([]uuid.UUID, error) {
	// passing the coursePhaseID to query ensures that only the coursePhases that are in the course are updated
	changedParticipations, err := s.queries.UpdateCoursePhasePassStatus(ctx, db.UpdateCoursePhasePassStatusParams{
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

func (s *CoursePhaseParticipationService) GetStudentsOfCoursePhase(ctx context.Context, coursePhaseID uuid.UUID) ([]studentDTO.Student, error) {
	students, err := s.queries.GetStudentsOfCoursePhase(ctx, coursePhaseID)
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
