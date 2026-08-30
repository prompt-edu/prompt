package allocation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	sdkUtils "github.com/prompt-edu/prompt-sdk/utils"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	"github.com/prompt-edu/prompt/servers/team_allocation/coreRequests"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// foreignKeyViolationCode is the PostgreSQL SQLSTATE raised when the target team
// disappears between the phase check and the write.
const foreignKeyViolationCode = "23503"

var (
	ErrAllocationNotFound    = errors.New("allocation not found")
	ErrParticipantNotInPhase = errors.New("course participation is not part of this course phase")
	ErrInvalidTeamForPhase   = errors.New("team does not belong to this course phase")
	ErrTeamWriteDenied       = errors.New("access restricted to assigned team")
	ErrParticipantLookup     = errors.New("could not verify course participation")
)

// participantResolver returns the participants of a course phase keyed by course
// participation ID. It is a field on the service so tests can stub the core call.
type participantResolver func(authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]coreRequests.Participant, error)

// DefaultParticipantResolver resolves participants against the core service.
func DefaultParticipantResolver() participantResolver {
	return func(authHeader string, coursePhaseID uuid.UUID) (map[uuid.UUID]coreRequests.Participant, error) {
		return coreRequests.GetCoursePhaseParticipants(sdkUtils.GetCoreUrl(), authHeader, coursePhaseID)
	}
}

type AllocationService struct {
	queries             db.Queries
	conn                *pgxpool.Pool
	resolveParticipants participantResolver
}

var AllocationServiceSingleton *AllocationService

func GetAllAllocations(ctx context.Context, coursePhaseID uuid.UUID) ([]allocationDTO.AllocationWithParticipation, error) {
	dbAllocations, err := AllocationServiceSingleton.queries.GetAllocationsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("Error fetching allocations from database: ", err)
		return []allocationDTO.AllocationWithParticipation{}, err
	}
	allocations := allocationDTO.GetAllocationsFromDBModels(dbAllocations)

	return allocations, nil
}

func GetAllocationByCourseParticipationID(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (uuid.UUID, error) {
	allocation, err := AllocationServiceSingleton.queries.GetAllocationForStudent(ctx, db.GetAllocationForStudentParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Error fetching allocation from database: ", err)
		return uuid.Nil, err
	}

	return allocation.TeamID, nil
}

// UpsertAllocation assigns a participant to a team. expectedTeamID scopes the
// write to a source team; an unset value writes unconditionally.
func UpsertAllocation(ctx context.Context, authHeader string, coursePhaseID, courseParticipationID, teamID uuid.UUID, expectedTeamID pgtype.UUID) error {
	participants, err := AllocationServiceSingleton.resolveParticipants(authHeader, coursePhaseID)
	if err != nil {
		log.Error("could not fetch course phase participants from core: ", err)
		return ErrParticipantLookup
	}

	participant, isParticipant := participants[courseParticipationID]
	if !isParticipant {
		return ErrParticipantNotInPhase
	}

	if _, err := AllocationServiceSingleton.queries.GetTeamByCoursePhaseAndTeamID(ctx, db.GetTeamByCoursePhaseAndTeamIDParams{
		ID:            teamID,
		CoursePhaseID: coursePhaseID,
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidTeamForPhase
		}
		return fmt.Errorf("could not verify the team of the course phase: %w", err)
	}

	rows, err := AllocationServiceSingleton.queries.UpsertAllocationForParticipant(ctx, db.UpsertAllocationForParticipantParams{
		ID:                    uuid.New(),
		CourseParticipationID: courseParticipationID,
		TeamID:                teamID,
		CoursePhaseID:         coursePhaseID,
		StudentFirstName:      participant.FirstName,
		StudentLastName:       participant.LastName,
		ExpectedTeamID:        expectedTeamID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationCode {
			return ErrInvalidTeamForPhase
		}
		return fmt.Errorf("could not store the allocation: %w", err)
	}
	if rows == 0 {
		return ErrTeamWriteDenied
	}
	return nil
}

// DeleteAllocation removes a participant's allocation. expectedTeamID scopes the
// delete to a source team; an unset value deletes unconditionally.
func DeleteAllocation(ctx context.Context, coursePhaseID, courseParticipationID uuid.UUID, expectedTeamID pgtype.UUID) error {
	rows, err := AllocationServiceSingleton.queries.DeleteAllocationForParticipant(ctx, db.DeleteAllocationForParticipantParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
		ExpectedTeamID:        expectedTeamID,
	})
	if err != nil {
		return fmt.Errorf("could not delete the allocation: %w", err)
	}
	if rows == 0 {
		return ErrAllocationNotFound
	}
	return nil
}
