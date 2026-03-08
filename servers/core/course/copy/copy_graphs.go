package copy

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prompt-edu/prompt/servers/core/course/courseDTO"
	db "github.com/prompt-edu/prompt/servers/core/db/sqlc"
	log "github.com/sirupsen/logrus"
)

// copyCoursePhaseGraph replicates the course phase dependency graph from the source course
// to the target course using the provided phase ID mapping.
func copyCoursePhaseGraph(c *gin.Context, qtx *db.Queries, sourceID, targetID uuid.UUID, phaseMap map[uuid.UUID]uuid.UUID) error {
	graph, err := qtx.GetCoursePhaseGraph(c, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get course phase graph: %w", err)
	}
	for _, item := range graph {
		fromID, ok1 := phaseMap[item.FromCoursePhaseID]
		toID, ok2 := phaseMap[item.ToCoursePhaseID]
		if !ok1 || !ok2 {
			return fmt.Errorf("missing phase mapping for graph edge from %s to %s", item.FromCoursePhaseID, item.ToCoursePhaseID)
		}
		if err := qtx.CreateCourseGraphConnection(c, db.CreateCourseGraphConnectionParams{
			FromCoursePhaseID: fromID,
			ToCoursePhaseID:   toID,
		}); err != nil {
			return fmt.Errorf("failed to create course graph connection: %w", err)
		}
	}
	return nil
}

// copyMetaGraphs recreates the phase and participation data graphs for the target course
// using the given mappings for phases and DTOs.
func copyMetaGraphs(c *gin.Context, qtx *db.Queries, sourceID, targetID uuid.UUID, phaseMap, dtoMap map[uuid.UUID]uuid.UUID) error {
	// Phase Data Graph
	phaseGraph, err := qtx.GetPhaseDataGraph(c, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get phase data graph: %w", err)
	}
	converted := []courseDTO.MetaDataGraphItem{}
	for _, i := range phaseGraph {
		fromP, ok1 := phaseMap[i.FromCoursePhaseID]
		toP, ok2 := phaseMap[i.ToCoursePhaseID]
		fromD, ok3 := dtoMap[i.FromCoursePhaseDtoID]
		toD, ok4 := dtoMap[i.ToCoursePhaseDtoID]
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return fmt.Errorf("missing mapping in phase data graph - phases: %s->%v, %s->%v; dtos: %s->%v, %s->%v",
				i.FromCoursePhaseID, ok1, i.ToCoursePhaseID, ok2,
				i.FromCoursePhaseDtoID, ok3, i.ToCoursePhaseDtoID, ok4)
		}
		converted = append(converted, courseDTO.MetaDataGraphItem{
			FromCoursePhaseID:    fromP,
			ToCoursePhaseID:      toP,
			FromCoursePhaseDtoID: fromD,
			ToCoursePhaseDtoID:   toD,
		})
	}
	if err := updatePhaseDataGraphHelper(c, qtx, targetID, converted); err != nil {
		return fmt.Errorf("failed to update phase data graph: %w", err)
	}

	// Participation Data Graph
	participationGraph, err := qtx.GetParticipationDataGraph(c, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get participation data graph: %w", err)
	}
	converted = []courseDTO.MetaDataGraphItem{}
	for _, i := range participationGraph {
		fromP, ok1 := phaseMap[i.FromCoursePhaseID]
		toP, ok2 := phaseMap[i.ToCoursePhaseID]
		fromD, ok3 := dtoMap[i.FromCoursePhaseDtoID]
		toD, ok4 := dtoMap[i.ToCoursePhaseDtoID]
		if !ok1 || !ok2 || !ok3 || !ok4 {
			return fmt.Errorf("missing mapping in participation data graph - phases: %s->%v, %s->%v; dtos: %s->%v, %s->%v",
				i.FromCoursePhaseID, ok1, i.ToCoursePhaseID, ok2,
				i.FromCoursePhaseDtoID, ok3, i.ToCoursePhaseDtoID, ok4)
		}
		converted = append(converted, courseDTO.MetaDataGraphItem{
			FromCoursePhaseID:    fromP,
			ToCoursePhaseID:      toP,
			FromCoursePhaseDtoID: fromD,
			ToCoursePhaseDtoID:   toD,
		})
	}

	if err := updateParticipationDataGraphHelper(c, qtx, targetID, converted); err != nil {
		return fmt.Errorf("failed to update participation data graph: %w", err)
	}
	return nil
}

// updatePhaseDataGraphHelper deletes and recreates all phase data graph connections
// for the given course using the provided metadata graph items.
func updatePhaseDataGraphHelper(c *gin.Context, qtx *db.Queries, courseID uuid.UUID, graphUpdate []courseDTO.MetaDataGraphItem) error {
	if err := qtx.DeletePhaseDataGraphConnections(c, courseID); err != nil {
		return fmt.Errorf("failed to delete old phase data graph connections: %w", err)
	}

	for _, item := range graphUpdate {
		err := qtx.CreatePhaseDataConnection(c, db.CreatePhaseDataConnectionParams{
			FromCoursePhaseID:    item.FromCoursePhaseID,
			ToCoursePhaseID:      item.ToCoursePhaseID,
			FromCoursePhaseDtoID: item.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   item.ToCoursePhaseDtoID,
		})
		if err != nil {
			log.Error("Error creating phase data connection: ", err)
			return fmt.Errorf("failed to create phase data connection: %w", err)
		}
	}
	return nil
}

// updateParticipationDataGraphHelper deletes and recreates all participation data graph connections
// for the given course using the provided metadata graph items.
func updateParticipationDataGraphHelper(c *gin.Context, qtx *db.Queries, courseID uuid.UUID, graphUpdate []courseDTO.MetaDataGraphItem) error {
	if err := qtx.DeleteParticipationDataGraphConnections(c, courseID); err != nil {
		return fmt.Errorf("failed to delete old participation data graph connections: %w", err)
	}

	for _, item := range graphUpdate {
		err := qtx.CreateParticipationDataConnection(c, db.CreateParticipationDataConnectionParams{
			FromCoursePhaseID:    item.FromCoursePhaseID,
			ToCoursePhaseID:      item.ToCoursePhaseID,
			FromCoursePhaseDtoID: item.FromCoursePhaseDtoID,
			ToCoursePhaseDtoID:   item.ToCoursePhaseDtoID,
		})
		if err != nil {
			log.Error("Error creating participation data connection: ", err)
			return fmt.Errorf("failed to create participation data connection: %w", err)
		}
	}
	return nil
}
