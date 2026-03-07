package coursePhaseType

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prompt-edu/prompt/servers/core/coursePhaseType/coursePhaseTypeDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

type CoursePhaseTypeService struct {
	queries          db.Queries
	conn             *pgxpool.Pool
	isDevEnvironment bool
}

var CoursePhaseTypeServiceSingleton *CoursePhaseTypeService

func GetAllCoursePhaseTypes(ctx context.Context) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := CoursePhaseTypeServiceSingleton.queries.GetAllCoursePhaseTypes(ctxWithTimeout)
	if err != nil {
		return nil, err
	}

	dtoCoursePhaseTypes := make([]coursePhaseTypeDTO.CoursePhaseType, 0, len(coursePhaseTypes))
	for _, phaseType := range coursePhaseTypes {
		// Participation Graph
		fetchedParticipationInputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseRequiredParticipationInputs(ctxWithTimeout, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedParticipationOutputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseProvidedParticipationOutputs(ctxWithTimeout, phaseType.ID)
		if err != nil {
			return nil, err
		}

		// Phase Data Graph
		fetchedPhaseInputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseRequiredPhaseInputs(ctxWithTimeout, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedPhaseOutputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseProvidedPhaseOutputs(ctxWithTimeout, phaseType.ID)
		if err != nil {
			return nil, err
		}

		participationInputDTOs, err := coursePhaseTypeDTO.GetParticipationInputDTOsFromDBModel(fetchedParticipationInputDTOs)
		if err != nil {
			return nil, err
		}

		participationOutputDTOs, err := coursePhaseTypeDTO.GetParticipationOutputDTOsFromDBModel(fetchedParticipationOutputDTOs)
		if err != nil {
			return nil, err
		}

		phaseInputDTOs, err := coursePhaseTypeDTO.GetPhaseInputDTOsFromDBModel(fetchedPhaseInputDTOs)
		if err != nil {
			return nil, err
		}

		phaseOutputDTOs, err := coursePhaseTypeDTO.GetPhaseOutputDTOsFromDBModel(fetchedPhaseOutputDTOs)
		if err != nil {
			return nil, err
		}

		dtoCoursePhaseType, err := coursePhaseTypeDTO.GetCoursePhaseTypeDTOFromDBModel(phaseType, participationInputDTOs, participationOutputDTOs, phaseInputDTOs, phaseOutputDTOs)
		if err != nil {
			return nil, err
		}
		dtoCoursePhaseTypes = append(dtoCoursePhaseTypes, dtoCoursePhaseType)
	}

	return dtoCoursePhaseTypes, nil
}
