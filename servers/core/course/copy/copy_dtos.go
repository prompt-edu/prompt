package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
)

// copyDTOs collects all participation DTOs (inputs and outputs) used by the source course phases.
// It returns a map from source DTO IDs to themselves, to support mapping-based operations.
func copyDTOs(c *gin.Context, qtx *db.Queries, sourceID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	dtoIDMap := make(map[uuid.UUID]uuid.UUID)

	// Collect all DTOs from course phase types
	unordered, err := qtx.GetNotOrderedCoursePhases(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get unordered course phases: %w", err)
	}
	sequence, err := qtx.GetCoursePhaseSequence(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course phase sequence: %w", err)
	}

	// Collect all unique CoursePhaseTypeIDs
	uniqueTypes := make(map[uuid.UUID]struct{})
	for _, p := range unordered {
		uniqueTypes[p.CoursePhaseTypeID] = struct{}{}
	}
	for _, p := range sequence {
		uniqueTypes[p.CoursePhaseTypeID] = struct{}{}
	}
	for tID := range uniqueTypes {
		outputs, err := qtx.GetCoursePhaseProvidedParticipationOutputs(c, tID)
		if err != nil {
			return nil, fmt.Errorf("failed to get participation outputs for type %s: %w", tID, err)
		}
		for _, o := range outputs {
			dtoIDMap[o.ID] = o.ID
		}

		// Get inputs
		inputs, err := qtx.GetCoursePhaseRequiredParticipationInputs(c, tID)
		if err != nil {
			return nil, fmt.Errorf("failed to get participation inputs for type %s: %w", tID, err)
		}
		for _, i := range inputs {
			dtoIDMap[i.ID] = i.ID
		}
	}

	// Collect all DTOs from graphs
	phaseGraph, err := qtx.GetPhaseDataGraph(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get phase data graph: %w", err)
	}
	for _, edge := range phaseGraph {
		dtoIDMap[edge.FromCoursePhaseDtoID] = edge.FromCoursePhaseDtoID
		dtoIDMap[edge.ToCoursePhaseDtoID] = edge.ToCoursePhaseDtoID
	}

	participationGraph, err := qtx.GetParticipationDataGraph(c, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participation data graph: %w", err)
	}
	for _, edge := range participationGraph {
		dtoIDMap[edge.FromCoursePhaseDtoID] = edge.FromCoursePhaseDtoID
		dtoIDMap[edge.ToCoursePhaseDtoID] = edge.ToCoursePhaseDtoID
	}

	return dtoIDMap, nil
}
