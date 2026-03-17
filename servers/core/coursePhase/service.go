package coursePhase

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/coursePhaseDTO"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type CoursePhaseService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CoursePhaseServiceSingleton *CoursePhaseService

func GetCoursePhaseByID(ctx context.Context, id uuid.UUID) (coursePhaseDTO.CoursePhase, error) {
	coursePhase, err := CoursePhaseServiceSingleton.queries.GetCoursePhase(ctx, id)
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	return coursePhaseDTO.GetCoursePhaseDTOFromDBModel(coursePhase)
}

func UpdateCoursePhase(ctx context.Context, coursePhase coursePhaseDTO.UpdateCoursePhase) error {
	dbModel, err := coursePhase.GetDBModel()
	if err != nil {
		return err
	}

	dbModel.ID = coursePhase.ID
	return CoursePhaseServiceSingleton.queries.UpdateCoursePhase(ctx, dbModel)
}

func CreateCoursePhase(ctx context.Context, coursePhase coursePhaseDTO.CreateCoursePhase) (coursePhaseDTO.CoursePhase, error) {
	dbModel, err := coursePhase.GetDBModel()
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	dbModel.ID = uuid.New()
	createdCoursePhase, err := CoursePhaseServiceSingleton.queries.CreateCoursePhase(ctx, dbModel)
	if err != nil {
		return coursePhaseDTO.CoursePhase{}, err
	}

	return GetCoursePhaseByID(ctx, createdCoursePhase.ID)
}

func DeleteCoursePhase(ctx context.Context, id uuid.UUID) error {
	return CoursePhaseServiceSingleton.queries.DeleteCoursePhase(ctx, id)
}

func CheckCoursePhasesBelongToCourse(ctx context.Context, courseId uuid.UUID, coursePhaseIds []uuid.UUID) (bool, error) {
	ok, err := CoursePhaseServiceSingleton.queries.CheckCoursePhasesBelongToCourse(ctx, db.CheckCoursePhasesBelongToCourseParams{
		CourseID: courseId,
		Column1:  coursePhaseIds,
	})

	if err != nil {
		log.Error(err)
		return false, errors.New("error checking course phases")
	}

	return ok, nil
}

func GetPrevPhaseDataByCoursePhaseID(ctx context.Context, coursePhaseID uuid.UUID) (coursePhaseDTO.PrevCoursePhaseData, error) {
	dataFromCore, err := CoursePhaseServiceSingleton.queries.GetPrevCoursePhaseDataFromCore(ctx, coursePhaseID)
	if err != nil {
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	resolutions, err := CoursePhaseServiceSingleton.queries.GetPrevCoursePhaseDataResolution(ctx, coursePhaseID)
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
	prevCoursePhaseDataDTO.Resolutions, err = resolution.ReplaceResolutionURLs(ctx, prevCoursePhaseDataDTO.Resolutions)
	if err != nil {
		log.WithFields(log.Fields{
			"coursePhaseID": coursePhaseID,
		}).Error("failed to replace resolution URLs: ", err)
		return coursePhaseDTO.PrevCoursePhaseData{}, err
	}

	return prevCoursePhaseDataDTO, nil
}

func GetCoursePhaseParticipationStatusCounts(ctx context.Context, coursePhaseID uuid.UUID) (map[string]int, error) {
	counts, err := CoursePhaseServiceSingleton.queries.GetCoursePhaseParticipationStatusCounts(ctx, coursePhaseID)

	// Convert the slice of structs to a map
	countsMap := make(map[string]int)
	for _, count := range counts {
		status := string(count.PassStatus.PassStatus)
		countsMap[status] = int(count.Count)
	}

	return countsMap, err
}
