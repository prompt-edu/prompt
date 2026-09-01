package coursePhase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution/resolutionDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// ResolutionReplacer rewrites the base URLs of course phase resolutions.
type ResolutionReplacer interface {
	ReplaceResolutionURLs(ctx context.Context, resolutions []resolutionDTO.Resolution) ([]resolutionDTO.Resolution, error)
}

type CoursePhaseService struct {
	queries     db.Queries
	conn        *pgxpool.Pool
	resolutions ResolutionReplacer
}

func NewCoursePhaseService(queries db.Queries, conn *pgxpool.Pool, resolutions ResolutionReplacer) *CoursePhaseService {
	return &CoursePhaseService{
		queries:     queries,
		conn:        conn,
		resolutions: resolutions,
	}
}

func (s *CoursePhaseService) GetCoursePhaseByID(ctx context.Context, id uuid.UUID) (coursePhaseDTO.CoursePhase, error) {
	coursePhase, err := s.queries.GetCoursePhase(ctx, id)
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	return coursePhaseDTO.GetCoursePhaseDTOFromDBModel(coursePhase)
}

func (s *CoursePhaseService) UpdateCoursePhase(ctx context.Context, coursePhase coursePhaseDTO.UpdateCoursePhase) error {
	dbModel, err := coursePhase.GetDBModel()
	if err != nil {
		return err
	}

	dbModel.ID = coursePhase.ID
	return s.queries.UpdateCoursePhase(ctx, dbModel)
}

func (s *CoursePhaseService) CreateCoursePhase(ctx context.Context, coursePhase coursePhaseDTO.CreateCoursePhase) (coursePhaseDTO.CoursePhase, error) {
	dbModel, err := coursePhase.GetDBModel()
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	dbModel.ID = uuid.New()
	createdCoursePhase, err := s.queries.CreateCoursePhase(ctx, dbModel)
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	return s.GetCoursePhaseByID(ctx, createdCoursePhase.ID)
}

func (s *CoursePhaseService) DeleteCoursePhase(ctx context.Context, id uuid.UUID) error {
	return s.queries.DeleteCoursePhase(ctx, id)
}

func (s *CoursePhaseService) CheckCoursePhasesBelongToCourse(ctx context.Context, courseId uuid.UUID, coursePhaseIds []uuid.UUID) (bool, error) {
	ok, err := s.queries.CheckCoursePhasesBelongToCourse(ctx, db.CheckCoursePhasesBelongToCourseParams{
		CourseID: courseId,
		Column1:  coursePhaseIds,
	})

	if err != nil {
		log.Error(err)
		return false, errors.New("error checking course phases")
	}

	return ok, nil
}

func (s *CoursePhaseService) GetPrevPhaseDataByCoursePhaseID(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseDTO.PrevCoursePhaseData, error) {
	dataFromCore, err := s.queries.GetPrevCoursePhaseDataFromCore(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	resolutions, err := s.queries.GetPrevCoursePhaseDataResolution(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	prevCoursePhaseDataDTO, err := coursePhaseDTO.GetPrevCoursePhaseDataDTO(dataFromCore, resolutions)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
		}).Error("failed to create previous course phase data DTO: ", err)
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	// Replace resolution URLs with the correct host
	prevCoursePhaseDataDTO.Resolutions, err = s.resolutions.ReplaceResolutionURLs(ctx, prevCoursePhaseDataDTO.Resolutions)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
		}).Error("failed to replace resolution URLs: ", err)
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	return prevCoursePhaseDataDTO, nil
}

func (s *CoursePhaseService) GetCoursePhaseParticipationStatusCounts(ctx context.Context, coursePhaseID uuid.UUID) (map[string]int, error) {
	counts, err := s.queries.GetCoursePhaseParticipationStatusCounts(ctx, coursePhaseID)

	// Convert the slice of structs to a map
	countsMap := make(map[string]int)
	for _, count := range counts {
		status := string(count.PassStatus.PassStatus)
		countsMap[status] = int(count.Count)
	}

	return countsMap, err
}
