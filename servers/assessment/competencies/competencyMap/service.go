package competencyMap

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/assessment/competencies/competencyMap/competencyMapDTO"
	db "github.com/prompt-edu/prompt/servers/assessment/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type CompetencyMapService struct {
	queries db.Queries
	conn    *pgxpool.Pool
}

var CompetencyMapServiceSingleton *CompetencyMapService
var ErrCompetencyNotInCoursePhase = errors.New("competency not part of course phase")
var ErrCompetencyScopeCheckFailed = errors.New("could not validate competency scope")

func CreateCompetencyMapping(ctx context.Context, coursePhaseID uuid.UUID, req competencyMapDTO.CompetencyMapping) error {
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, req.FromCompetencyID); err != nil {
		return err
	}
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, req.ToCompetencyID); err != nil {
		return err
	}

	err := CompetencyMapServiceSingleton.queries.CreateCompetencyMap(ctx, db.CreateCompetencyMapParams{
		FromCompetencyID: req.FromCompetencyID,
		ToCompetencyID:   req.ToCompetencyID,
	})
	if err != nil {
		log.Error("could not create competency mapping: ", err)
		return errors.New("could not create competency mapping")
	}
	return nil
}

func DeleteCompetencyMapping(ctx context.Context, coursePhaseID uuid.UUID, req competencyMapDTO.CompetencyMapping) error {
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, req.FromCompetencyID); err != nil {
		return err
	}
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, req.ToCompetencyID); err != nil {
		return err
	}

	err := CompetencyMapServiceSingleton.queries.DeleteCompetencyMap(ctx, db.DeleteCompetencyMapParams{
		FromCompetencyID: req.FromCompetencyID,
		ToCompetencyID:   req.ToCompetencyID,
	})
	if err != nil {
		log.Error("could not delete competency mapping: ", err)
		return errors.New("could not delete competency mapping")
	}
	return nil
}

func GetCompetencyMappings(ctx context.Context, coursePhaseID uuid.UUID, fromCompetencyID uuid.UUID) ([]db.CompetencyMap, error) {
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, fromCompetencyID); err != nil {
		return nil, err
	}

	mappings, err := CompetencyMapServiceSingleton.queries.GetCompetencyMappings(ctx, fromCompetencyID)
	if err != nil {
		log.Error("could not get competency mappings: ", err)
		return nil, errors.New("could not get competency mappings")
	}
	return filterCompetencyMappingsByCoursePhase(ctx, coursePhaseID, mappings)
}

func GetAllCompetencyMappings(ctx context.Context, coursePhaseID uuid.UUID) ([]db.CompetencyMap, error) {
	mappings, err := CompetencyMapServiceSingleton.queries.GetAllCompetencyMappings(ctx)
	if err != nil {
		log.Error("could not get all competency mappings: ", err)
		return nil, errors.New("could not get all competency mappings")
	}
	return filterCompetencyMappingsByCoursePhase(ctx, coursePhaseID, mappings)
}

func GetReverseCompetencyMappings(ctx context.Context, coursePhaseID uuid.UUID, toCompetencyID uuid.UUID) ([]db.CompetencyMap, error) {
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, toCompetencyID); err != nil {
		return nil, err
	}

	mappings, err := CompetencyMapServiceSingleton.queries.GetReverseCompetencyMappings(ctx, toCompetencyID)
	if err != nil {
		log.Error("could not get reverse competency mappings: ", err)
		return nil, errors.New("could not get reverse competency mappings")
	}
	return filterCompetencyMappingsByCoursePhase(ctx, coursePhaseID, mappings)
}

func GetEvaluationsByMappedCompetency(ctx context.Context, coursePhaseID uuid.UUID, fromCompetencyID uuid.UUID) ([]db.Evaluation, error) {
	if err := validateCompetencyInCoursePhase(ctx, coursePhaseID, fromCompetencyID); err != nil {
		return nil, err
	}

	evaluations, err := CompetencyMapServiceSingleton.queries.GetEvaluationsByMappedCompetency(ctx, fromCompetencyID)
	if err != nil {
		log.Error("could not get evaluations by mapped competency: ", err)
		return nil, errors.New("could not get evaluations by mapped competency")
	}
	filtered := evaluations[:0]
	for _, evaluation := range evaluations {
		if evaluation.CoursePhaseID == coursePhaseID {
			filtered = append(filtered, evaluation)
		}
	}
	return filtered, nil
}

const competencyInCoursePhaseQuery = `
SELECT EXISTS(
	SELECT 1
	FROM competency comp
	JOIN category_course_phase ccp ON ccp.category_id = comp.category_id
	WHERE comp.id = $1 AND ccp.course_phase_id = $2
);
`

func validateCompetencyInCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, competencyID uuid.UUID) error {
	exists, err := competencyInCoursePhase(ctx, coursePhaseID, competencyID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrCompetencyNotInCoursePhase
	}
	return nil
}

func competencyInCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, competencyID uuid.UUID) (bool, error) {
	var exists bool
	err := CompetencyMapServiceSingleton.conn.QueryRow(ctx, competencyInCoursePhaseQuery, competencyID, coursePhaseID).Scan(&exists)
	if err != nil {
		log.Error("could not validate competency scope: ", err)
		return false, ErrCompetencyScopeCheckFailed
	}
	return exists, nil
}

func filterCompetencyMappingsByCoursePhase(ctx context.Context, coursePhaseID uuid.UUID, mappings []db.CompetencyMap) ([]db.CompetencyMap, error) {
	seen := make(map[uuid.UUID]bool)
	isInPhase := func(competencyID uuid.UUID) (bool, error) {
		if exists, ok := seen[competencyID]; ok {
			return exists, nil
		}
		exists, err := competencyInCoursePhase(ctx, coursePhaseID, competencyID)
		if err != nil {
			return false, err
		}
		seen[competencyID] = exists
		return exists, nil
	}

	filtered := make([]db.CompetencyMap, 0, len(mappings))
	for _, mapping := range mappings {
		fromOk, err := isInPhase(mapping.FromCompetencyID)
		if err != nil {
			return nil, err
		}
		if !fromOk {
			continue
		}
		toOk, err := isInPhase(mapping.ToCompetencyID)
		if err != nil {
			return nil, err
		}
		if !toOk {
			continue
		}
		filtered = append(filtered, mapping)
	}
	return filtered, nil
}
