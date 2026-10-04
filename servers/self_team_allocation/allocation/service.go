package allocation

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/prompt-edu/prompt/servers/self_team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/self_team_allocation/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// ErrAssignmentNotFound is returned when an assignment is not found in the database
var ErrAssignmentNotFound = errors.New("assignment not found")

type AllocationService struct {
	queries db.Queries
}

func NewAllocationService(queries db.Queries) *AllocationService {
	return &AllocationService{
		queries: queries,
	}
}

func (s *AllocationService) GetAllAllocations(ctx context.Context, coursePhaseID uuid.UUID) ([]allocationDTO.AllocationWithParticipation, error) {
	dbAssignments, err := s.queries.GetAssignmentsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("Error fetching assignments from database: ", err)
		return []allocationDTO.AllocationWithParticipation{}, err
	}
	allocations := allocationDTO.GetAllocationsFromDBModels(dbAssignments)

	return allocations, nil
}

func (s *AllocationService) GetAllocationByCourseParticipationID(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (uuid.UUID, error) {
	assignment, err := s.queries.GetAssignmentForStudent(ctx, db.GetAssignmentForStudentParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Debug("No assignment found for student: ", err)
		} else {
			log.Error("Error fetching assignment from database: ", err)
		}
		return uuid.Nil, err
	}

	return assignment.TeamID, nil
}
