package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateResourceInstanceParams holds parameters for creating a resource instance.
type CreateResourceInstanceParams struct {
	ResourceConfigID      uuid.UUID
	CoursePhaseID         uuid.UUID
	TeamID                *uuid.UUID
	CourseParticipationID *uuid.UUID
}

// UpdateResourceInstanceStatusParams holds parameters for updating instance status.
type UpdateResourceInstanceStatusParams struct {
	ID           uuid.UUID
	Status       ResourceStatus
	ExternalID   *string
	ExternalURL  *string
	ErrorMessage *string
	RetryCount   int32
}

// CreateResourceInstance inserts a pending resource instance.
func (q *Queries) CreateResourceInstance(ctx context.Context, arg CreateResourceInstanceParams) (ResourceInstance, error) {
	const sql = `
		INSERT INTO resource_instance (id, resource_config_id, course_phase_id, team_id, course_participation_id)
		VALUES (gen_random_uuid(), $1, $2, $3, $4)
		RETURNING id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, retry_count, created_at, updated_at`

	row := q.db.QueryRow(ctx, sql, arg.ResourceConfigID, arg.CoursePhaseID, arg.TeamID, arg.CourseParticipationID)
	return scanResourceInstance(row)
}

// GetResourceInstance retrieves a resource instance by ID.
func (q *Queries) GetResourceInstance(ctx context.Context, id uuid.UUID) (ResourceInstance, error) {
	const sql = `
		SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, retry_count, created_at, updated_at
		FROM resource_instance
		WHERE id = $1`

	row := q.db.QueryRow(ctx, sql, id)
	return scanResourceInstance(row)
}

// ListResourceInstances retrieves all instances for a course phase, optionally filtered by config ID.
func (q *Queries) ListResourceInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]ResourceInstance, error) {
	const sql = `
		SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, retry_count, created_at, updated_at
		FROM resource_instance
		WHERE course_phase_id = $1
		ORDER BY created_at DESC`

	rows, err := q.db.Query(ctx, sql, coursePhaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []ResourceInstance
	for rows.Next() {
		ri, err := scanResourceInstanceFromRows(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, ri)
	}
	return instances, rows.Err()
}

// UpdateResourceInstanceStatus updates the status and result fields of an instance.
func (q *Queries) UpdateResourceInstanceStatus(ctx context.Context, arg UpdateResourceInstanceStatusParams) error {
	const sql = `
		UPDATE resource_instance
		SET status = $2, external_id = $3, external_url = $4, error_message = $5, retry_count = $6, updated_at = NOW()
		WHERE id = $1`

	_, err := q.db.Exec(ctx, sql, arg.ID, arg.Status, arg.ExternalID, arg.ExternalURL, arg.ErrorMessage, arg.RetryCount)
	return err
}

// DeleteResourceInstance removes a resource instance by ID.
func (q *Queries) DeleteResourceInstance(ctx context.Context, id uuid.UUID) error {
	const sql = `DELETE FROM resource_instance WHERE id = $1`
	_, err := q.db.Exec(ctx, sql, id)
	return err
}

// CountNonTerminalInstances counts instances that are pending or in_progress for a phase.
func (q *Queries) CountNonTerminalInstances(ctx context.Context, coursePhaseID uuid.UUID) (int64, error) {
	const sql = `
		SELECT COUNT(*)
		FROM resource_instance
		WHERE course_phase_id = $1 AND status IN ('pending', 'in_progress')`

	var count int64
	err := q.db.QueryRow(ctx, sql, coursePhaseID).Scan(&count)
	return count, err
}

// ListPendingInstances retrieves all pending instances for a course phase.
func (q *Queries) ListPendingInstances(ctx context.Context, coursePhaseID uuid.UUID) ([]ResourceInstance, error) {
	const sql = `
		SELECT id, resource_config_id, course_phase_id, team_id, course_participation_id, status, external_id, external_url, error_message, retry_count, created_at, updated_at
		FROM resource_instance
		WHERE course_phase_id = $1 AND status = 'pending'`

	rows, err := q.db.Query(ctx, sql, coursePhaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var instances []ResourceInstance
	for rows.Next() {
		ri, err := scanResourceInstanceFromRows(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, ri)
	}
	return instances, rows.Err()
}

// ResetInProgressTosPending resets instances stuck in 'in_progress' back to 'pending'.
// This is called at startup to recover from server restarts mid-execution.
func (q *Queries) ResetInProgressToPending(ctx context.Context) error {
	const sql = `UPDATE resource_instance SET status = 'pending', updated_at = NOW() WHERE status = 'in_progress'`
	_, err := q.db.Exec(ctx, sql)
	return err
}

// ResetFailedInstanceToPending resets a single failed instance to pending for retry.
func (q *Queries) ResetFailedInstanceToPending(ctx context.Context, id uuid.UUID) error {
	const sql = `
		UPDATE resource_instance
		SET status = 'pending', error_message = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'failed'`
	_, err := q.db.Exec(ctx, sql, id)
	return err
}

// scanResourceInstance scans a ResourceInstance from a QueryRow result.
func scanResourceInstance(row interface{ Scan(dest ...any) error }) (ResourceInstance, error) {
	var ri ResourceInstance
	var updatedAt time.Time
	err := row.Scan(
		&ri.ID, &ri.ResourceConfigID, &ri.CoursePhaseID,
		&ri.TeamID, &ri.CourseParticipationID,
		&ri.Status, &ri.ExternalID, &ri.ExternalURL, &ri.ErrorMessage,
		&ri.RetryCount, &ri.CreatedAt, &updatedAt)
	ri.UpdatedAt = updatedAt
	return ri, err
}

// scanResourceInstanceFromRows scans a ResourceInstance from pgx.Rows.
func scanResourceInstanceFromRows(rows pgx.Rows) (ResourceInstance, error) {
	var ri ResourceInstance
	var updatedAt time.Time
	err := rows.Scan(
		&ri.ID, &ri.ResourceConfigID, &ri.CoursePhaseID,
		&ri.TeamID, &ri.CourseParticipationID,
		&ri.Status, &ri.ExternalID, &ri.ExternalURL, &ri.ErrorMessage,
		&ri.RetryCount, &ri.CreatedAt, &updatedAt)
	ri.UpdatedAt = updatedAt
	return ri, err
}
