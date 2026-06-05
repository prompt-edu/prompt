package tease

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	promptSDK "github.com/prompt-edu/prompt-sdk"
	db "github.com/prompt-edu/prompt/servers/team_allocation/db/sqlc"
	"github.com/prompt-edu/prompt/servers/team_allocation/tease/teaseDTO"
	log "github.com/sirupsen/logrus"
)

var defaultEmptyJSONArray = json.RawMessage("[]")

var errInvalidAllocation = errors.New("invalid allocation")

// GetTeaseWorkspace returns the persisted workspace for a course phase.
func GetTeaseWorkspace(ctx context.Context, coursePhaseID uuid.UUID) (teaseDTO.TeaseWorkspace, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	row, err := TeaseServiceSingleton.queries.GetTeaseWorkspace(ctxWithTimeout, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return teaseDTO.TeaseWorkspace{
				CoursePhaseID:    coursePhaseID,
				Constraints:      defaultEmptyJSONArray,
				LockedStudents:   defaultEmptyJSONArray,
				AllocationsDraft: defaultEmptyJSONArray,
			}, nil
		}
		log.Error("could not get tease workspace: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not get tease workspace: %w", err)
	}

	return workspaceFromDB(row), nil
}

// UpsertTeaseWorkspace saves a draft workspace without publishing allocations.
func UpsertTeaseWorkspace(ctx context.Context, coursePhaseID uuid.UUID, req teaseDTO.TeaseWorkspaceRequest, updatedBy uuid.UUID) (teaseDTO.TeaseWorkspace, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := upsertParamsFromRequest(coursePhaseID, req, updatedBy)

	row, err := TeaseServiceSingleton.queries.UpsertTeaseWorkspace(ctxWithTimeout, params)
	if err != nil {
		log.Error("could not upsert tease workspace: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not upsert tease workspace: %w", err)
	}

	return workspaceFromDB(row), nil
}

// SaveTeaseWorkspaceAndAllocations publishes a workspace and replaces allocations.
func SaveTeaseWorkspaceAndAllocations(ctx context.Context, coursePhaseID uuid.UUID, req teaseDTO.TeaseSaveRequest, updatedBy uuid.UUID) (teaseDTO.TeaseWorkspace, error) {
	if err := validateAllocations(req.Allocations); err != nil {
		return teaseDTO.TeaseWorkspace{}, err
	}

	ctx, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	tx, err := TeaseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := TeaseServiceSingleton.queries.WithTx(tx)

	upsertParams := upsertParamsFromRequest(coursePhaseID, req.TeaseWorkspaceRequest, updatedBy)
	if _, err := qtx.UpsertTeaseWorkspace(ctx, upsertParams); err != nil {
		log.Error("could not upsert tease workspace within save transaction: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not upsert tease workspace: %w", err)
	}

	if err := qtx.DeleteAllocationsByPhase(ctx, coursePhaseID); err != nil {
		log.WithField("course_phase_id", coursePhaseID).Error("could not clear existing allocations within save transaction: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not clear existing allocations for course phase %s: %w", coursePhaseID, err)
	}

	for _, allocation := range req.Allocations {
		if len(allocation.Students) == 0 {
			log.Warn("allocation with no students: ", allocation.ProjectID)
			continue
		}
		for _, studentID := range allocation.Students {
			err = qtx.CreateOrUpdateAllocation(ctx, db.CreateOrUpdateAllocationParams{
				ID:                    uuid.New(),
				CourseParticipationID: studentID,
				TeamID:                allocation.ProjectID,
				CoursePhaseID:         coursePhaseID,
			})
			if err != nil {
				log.WithFields(log.Fields{
					"course_phase_id":         coursePhaseID,
					"team_id":                 allocation.ProjectID,
					"course_participation_id": studentID,
				}).Error("could not create or update allocation: ", err)
				return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not create or update allocation for student %s in team %s: %w", studentID, allocation.ProjectID, err)
			}
		}
	}

	if err := qtx.StampTeaseWorkspaceExportedAt(ctx, coursePhaseID); err != nil {
		log.Error("could not stamp last_exported_at on tease workspace: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not stamp last_exported_at: %w", err)
	}

	finalRow, err := qtx.GetTeaseWorkspace(ctx, coursePhaseID)
	if err != nil {
		log.Error("could not read back tease workspace after save: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not read back tease workspace: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("tease save transaction commit failed: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("transaction commit failed: %w", err)
	}

	return workspaceFromDB(finalRow), nil
}

func validateAllocations(allocations []teaseDTO.Allocation) error {
	seen := make(map[uuid.UUID]uuid.UUID)
	for i, allocation := range allocations {
		if allocation.ProjectID == uuid.Nil {
			return fmt.Errorf("%w: invalid project ID at allocation %d", errInvalidAllocation, i)
		}
		for j, studentID := range allocation.Students {
			if studentID == uuid.Nil {
				return fmt.Errorf("%w: invalid student ID at allocation %d student %d", errInvalidAllocation, i, j)
			}
			if previousProjectID, exists := seen[studentID]; exists && previousProjectID != allocation.ProjectID {
				return fmt.Errorf("%w: student %s assigned to multiple teams (%s and %s)", errInvalidAllocation, studentID, previousProjectID, allocation.ProjectID)
			}
			seen[studentID] = allocation.ProjectID
		}
	}
	return nil
}

func workspaceFromDB(row db.TeaseWorkspace) teaseDTO.TeaseWorkspace {
	ws := teaseDTO.TeaseWorkspace{
		CoursePhaseID:    row.CoursePhaseID,
		Constraints:      jsonOrEmpty(row.Constraints),
		LockedStudents:   jsonOrEmpty(row.LockedStudents),
		AllocationsDraft: jsonOrEmpty(row.AllocationsDraft),
	}
	if row.AlgorithmType.Valid {
		s := row.AlgorithmType.String
		ws.AlgorithmType = &s
	}
	if row.UpdatedBy.Valid {
		u := uuid.UUID(row.UpdatedBy.Bytes)
		ws.UpdatedBy = &u
	}
	if row.LastSavedAt.Valid {
		t := row.LastSavedAt.Time
		ws.LastSavedAt = &t
	}
	if row.LastExportedAt.Valid {
		t := row.LastExportedAt.Time
		ws.LastExportedAt = &t
	}
	return ws
}

func upsertParamsFromRequest(coursePhaseID uuid.UUID, req teaseDTO.TeaseWorkspaceRequest, updatedBy uuid.UUID) db.UpsertTeaseWorkspaceParams {
	params := db.UpsertTeaseWorkspaceParams{
		CoursePhaseID:    coursePhaseID,
		Constraints:      []byte(jsonOrEmpty(req.Constraints)),
		LockedStudents:   []byte(jsonOrEmpty(req.LockedStudents)),
		AllocationsDraft: []byte(jsonOrEmpty(req.AllocationsDraft)),
		UpdatedBy:        pgtype.UUID{Bytes: updatedBy, Valid: true},
	}
	if req.AlgorithmType != nil {
		params.AlgorithmType = pgtype.Text{String: *req.AlgorithmType, Valid: true}
	}
	return params
}

func jsonOrEmpty(b []byte) json.RawMessage {
	if len(b) == 0 {
		return defaultEmptyJSONArray
	}
	return json.RawMessage(b)
}
