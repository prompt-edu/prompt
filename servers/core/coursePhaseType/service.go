package coursePhaseType

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/core/coursePhase/resolution"
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

	return addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

// GetCoursePhaseTypesForStudent returns the course phase types the student has been involved in
// via at least one course_phase_participation. For studentID == uuid.Nil it returns an empty slice.
func GetCoursePhaseTypesForStudent(ctx context.Context, studentID uuid.UUID) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	if studentID == uuid.Nil {
		return []coursePhaseTypeDTO.CoursePhaseType{}, nil
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseTypesForStudent(ctxWithTimeout, studentID)
	if err != nil {
		return nil, err
	}

	return addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

func GetCoursePhaseTypesForStudentCourses(ctx context.Context, studentID uuid.UUID) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	if studentID == uuid.Nil {
		return []coursePhaseTypeDTO.CoursePhaseType{}, nil
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseTypesForStudentCourses(ctxWithTimeout, studentID)
	if err != nil {
		return nil, err
	}

	return addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

func addCoursePhaseTypeInputOutput(ctx context.Context, coursePhaseTypes []db.CoursePhaseType) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	dtoCoursePhaseTypes := make([]coursePhaseTypeDTO.CoursePhaseType, 0, len(coursePhaseTypes))
	for _, phaseType := range coursePhaseTypes {
		// Participation Graph
		fetchedParticipationInputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseRequiredParticipationInputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedParticipationOutputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseProvidedParticipationOutputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}

		// Phase Data Graph
		fetchedPhaseInputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseRequiredPhaseInputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedPhaseOutputDTOs, err := CoursePhaseTypeServiceSingleton.queries.GetCoursePhaseProvidedPhaseOutputs(ctx, phaseType.ID)
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
		dtoCoursePhaseType.BaseUrl = replaceCoreHostPlaceholder(dtoCoursePhaseType.BaseUrl)
		dtoCoursePhaseTypes = append(dtoCoursePhaseTypes, dtoCoursePhaseType)
	}

	return dtoCoursePhaseTypes, nil
}

func replaceCoreHostPlaceholder(baseURL string) string {
	if !strings.Contains(baseURL, "{CORE_HOST}") {
		return baseURL
	}

	coreHost := resolution.NormaliseHost(sdkUtils.GetEnv("CORE_HOST", "http://localhost:8080"))
	return strings.ReplaceAll(baseURL, "{CORE_HOST}", coreHost)
}
