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

// defaultEmptyJSONArray is used when a caller omits one of the JSON
// snapshot fields on PUT / POST requests. Matches the column default.
var defaultEmptyJSONArray = json.RawMessage("[]")

// GetTeaseWorkspace returns the persisted workspace for a course phase.
// If no row exists, it returns a zero-value response with empty default
// arrays (per §4.2 of the Phase 1 plan).
func GetTeaseWorkspace(ctx context.Context, coursePhaseID uuid.UUID) (teaseDTO.TeaseWorkspace, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	row, err := TeaseServiceSingleton.queries.GetTeaseWorkspace(ctxWithTimeout, coursePhaseID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No workspace yet for this course phase → return empty defaults.
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

// UpsertTeaseWorkspace writes the payload to the tease_workspace table
// (insert-or-update on course_phase_id) and returns the server-stamped
// row. `last_exported_at` is left untouched — only the /save flow
// stamps it.
func UpsertTeaseWorkspace(ctx context.Context, coursePhaseID uuid.UUID, req teaseDTO.TeaseWorkspaceRequest) (teaseDTO.TeaseWorkspace, error) {
	ctxWithTimeout, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	params := upsertParamsFromRequest(coursePhaseID, req)

	row, err := TeaseServiceSingleton.queries.UpsertTeaseWorkspace(ctxWithTimeout, params)
	if err != nil {
		log.Error("could not upsert tease workspace: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not upsert tease workspace: %w", err)
	}

	return workspaceFromDB(row), nil
}

// SaveTeaseWorkspaceAndAllocations performs the atomic "Save to PROMPT"
// flow: in one transaction it upserts the tease_workspace row, upserts
// every allocation row via CreateOrUpdateAllocation, and stamps
// last_exported_at. Any error rolls the whole transaction back.
func SaveTeaseWorkspaceAndAllocations(ctx context.Context, coursePhaseID uuid.UUID, req teaseDTO.TeaseSaveRequest) (teaseDTO.TeaseWorkspace, error) {
	ctx, cancel := db.GetTimeoutContext(ctx)
	defer cancel()

	tx, err := TeaseServiceSingleton.conn.Begin(ctx)
	if err != nil {
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer promptSDK.DeferDBRollback(tx, ctx)

	qtx := TeaseServiceSingleton.queries.WithTx(tx)

	// 1. Upsert the workspace row.
	upsertParams := upsertParamsFromRequest(coursePhaseID, teaseDTO.TeaseWorkspaceRequest{
		Constraints:      req.Constraints,
		LockedStudents:   req.LockedStudents,
		AllocationsDraft: req.AllocationsDraft,
		AlgorithmType:    req.AlgorithmType,
		UpdatedBy:        req.UpdatedBy,
	})
	if _, err := qtx.UpsertTeaseWorkspace(ctx, upsertParams); err != nil {
		log.Error("could not upsert tease workspace within save transaction: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not upsert tease workspace: %w", err)
	}

	// 2. Replace the allocation set for this phase with the payload.
	//    The save endpoint is the authoritative publish, so the incoming
	//    `req.Allocations` is treated as the *complete* desired state — not
	//    a partial upsert. We therefore delete every existing allocation
	//    for the phase first so that:
	//      - resetting allocations in the client and saving an empty
	//        payload actually unallocates everyone, and
	//      - removing a single student from a team in the client
	//        propagates as a delete instead of being silently retained.
	//    Both the delete and the subsequent inserts run inside the same
	//    transaction, so concurrent readers never observe an empty board.
	if err := qtx.DeleteAllocationsByPhase(ctx, coursePhaseID); err != nil {
		log.Error("could not clear existing allocations within save transaction: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not clear existing allocations: %w", err)
	}

	for _, allocation := range req.Allocations {
		if allocation.ProjectID == uuid.Nil {
			return teaseDTO.TeaseWorkspace{}, fmt.Errorf("invalid project ID in allocation")
		}
		if len(allocation.Students) == 0 {
			log.Warn("allocation with no students: ", allocation.ProjectID)
			continue
		}
		for _, studentID := range allocation.Students {
			if studentID == uuid.Nil {
				return teaseDTO.TeaseWorkspace{}, fmt.Errorf("invalid student ID in allocation")
			}
			err = qtx.CreateOrUpdateAllocation(ctx, db.CreateOrUpdateAllocationParams{
				ID:                    uuid.New(),
				CourseParticipationID: studentID,
				TeamID:                allocation.ProjectID,
				CoursePhaseID:         coursePhaseID,
			})
			if err != nil {
				log.Error("could not create or update allocation: ", err)
				return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not create or update allocation: %w", err)
			}
		}
	}

	// 3. Stamp last_exported_at = now() on the workspace row we just
	//    upserted in step 1.
	if err := qtx.StampTeaseWorkspaceExportedAt(ctx, coursePhaseID); err != nil {
		log.Error("could not stamp last_exported_at on tease workspace: ", err)
		return teaseDTO.TeaseWorkspace{}, fmt.Errorf("could not stamp last_exported_at: %w", err)
	}

	// Re-read so we return the final stamped row.
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

// workspaceFromDB maps the sqlc-style row to the over-the-wire DTO,
// substituting SQL nulls with Go nil pointers and stored jsonb nulls
// with empty JSON arrays.
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

func upsertParamsFromRequest(coursePhaseID uuid.UUID, req teaseDTO.TeaseWorkspaceRequest) db.UpsertTeaseWorkspaceParams {
	params := db.UpsertTeaseWorkspaceParams{
		CoursePhaseID:    coursePhaseID,
		Constraints:      jsonOrEmptyBytes(req.Constraints),
		LockedStudents:   jsonOrEmptyBytes(req.LockedStudents),
		AllocationsDraft: jsonOrEmptyBytes(req.AllocationsDraft),
	}
	if req.AlgorithmType != nil {
		params.AlgorithmType = pgtype.Text{String: *req.AlgorithmType, Valid: true}
	}
	if req.UpdatedBy != nil {
		params.UpdatedBy = pgtype.UUID{Bytes: *req.UpdatedBy, Valid: true}
	}
	return params
}

func jsonOrEmpty(b []byte) json.RawMessage {
	if len(b) == 0 {
		return defaultEmptyJSONArray
	}
	return json.RawMessage(b)
}

func jsonOrEmptyBytes(b []byte) []byte {
	if len(b) == 0 {
		return []byte("[]")
	}
	return b
}
