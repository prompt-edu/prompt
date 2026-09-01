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

func NewCoursePhaseTypeService(queries db.Queries, conn *pgxpool.Pool, isDevEnvironment bool) *CoursePhaseTypeService {
	return &CoursePhaseTypeService{
		queries:          queries,
		conn:             conn,
		isDevEnvironment: isDevEnvironment,
	}
}

func (s *CoursePhaseTypeService) GetAllCoursePhaseTypes(ctx context.Context) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := s.queries.GetAllCoursePhaseTypes(ctxWithTimeout)
	if err != nil {
		return nil, err
	}

	return s.addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

// GetCoursePhaseTypesForStudent returns the course phase types the student has been involved in
// via at least one course_phase_participation. For studentID == uuid.Nil it returns an empty slice.
func (s *CoursePhaseTypeService) GetCoursePhaseTypesForStudent(ctx context.Context, studentID uuid.UUID) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	if studentID == uuid.Nil {
		return []coursePhaseTypeDTO.CoursePhaseType{}, nil
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := s.queries.GetCoursePhaseTypesForStudent(ctxWithTimeout, studentID)
	if err != nil {
		return nil, err
	}

	return s.addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

func (s *CoursePhaseTypeService) GetCoursePhaseTypesForStudentCourses(ctx context.Context, studentID uuid.UUID) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	if studentID == uuid.Nil {
		return []coursePhaseTypeDTO.CoursePhaseType{}, nil
	}

	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	coursePhaseTypes, err := s.queries.GetCoursePhaseTypesForStudentCourses(ctxWithTimeout, studentID)
	if err != nil {
		return nil, err
	}

	return s.addCoursePhaseTypeInputOutput(ctxWithTimeout, coursePhaseTypes)
}

func (s *CoursePhaseTypeService) addCoursePhaseTypeInputOutput(ctx context.Context, coursePhaseTypes []db.CoursePhaseType) ([]coursePhaseTypeDTO.CoursePhaseType, error) {
	dtoCoursePhaseTypes := make([]coursePhaseTypeDTO.CoursePhaseType, 0, len(coursePhaseTypes))
	for _, phaseType := range coursePhaseTypes {
		// Participation Graph
		fetchedParticipationInputDTOs, err := s.queries.GetCoursePhaseRequiredParticipationInputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedParticipationOutputDTOs, err := s.queries.GetCoursePhaseProvidedParticipationOutputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}

		// Phase Data Graph
		fetchedPhaseInputDTOs, err := s.queries.GetCoursePhaseRequiredPhaseInputs(ctx, phaseType.ID)
		if err != nil {
			return nil, err
		}
		fetchedPhaseOutputDTOs, err := s.queries.GetCoursePhaseProvidedPhaseOutputs(ctx, phaseType.ID)
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
