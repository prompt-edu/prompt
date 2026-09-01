package allocation

import (
	"context"

	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/team_allocation/allocation/allocationDTO"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	log "github.com/sirupsen/logrus"
)

type AllocationService struct {
	queries db.Queries
}

func NewAllocationService(queries db.Queries) *AllocationService {
	return &AllocationService{
		queries: queries,
	}
}

func (s *AllocationService) GetAllAllocations(ctx context.Context, coursePhaseID uuid.UUID) ([]allocationDTO.AllocationWithParticipation, error) {
	dbAllocations, err := s.queries.GetAllocationsByCoursePhase(ctx, coursePhaseID)
	if err != nil {
		log.Error("Error fetching allocations from database: ", err)
		return []allocationDTO.AllocationWithParticipation{}, err
	}
	allocations := allocationDTO.GetAllocationsFromDBModels(dbAllocations)

	return allocations, nil
}

func (s *AllocationService) GetAllocationByCourseParticipationID(ctx context.Context, courseParticipationID, coursePhaseID uuid.UUID) (uuid.UUID, error) {
	allocation, err := s.queries.GetAllocationForStudent(ctx, db.GetAllocationForStudentParams{
		CourseParticipationID: courseParticipationID,
		CoursePhaseID:         coursePhaseID,
	})
	if err != nil {
		log.Error("Error fetching allocation from database: ", err)
		return uuid.Nil, err
	}

	return allocation.TeamID, nil
}
