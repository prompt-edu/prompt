package db

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

// CreateResourceConfigParams holds parameters for creating a resource config.
type CreateResourceConfigParams struct {
	CoursePhaseID       uuid.UUID
	ProviderType        ProviderType
	ResourceType        string
	Scope               ResourceScope
	NameTemplate        string
	PermissionMapping   json.RawMessage
	ResourceExtraConfig json.RawMessage
}

// UpdateResourceConfigParams holds parameters for updating a resource config.
type UpdateResourceConfigParams struct {
	ID                  uuid.UUID
	CoursePhaseID       uuid.UUID
	ResourceType        string
	Scope               ResourceScope
	NameTemplate        string
	PermissionMapping   json.RawMessage
	ResourceExtraConfig json.RawMessage
}

// CreateResourceConfig inserts a new resource config.
func (q *Queries) CreateResourceConfig(ctx context.Context, arg CreateResourceConfigParams) (ResourceConfig, error) {
	const sql = `
		INSERT INTO resource_config (id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
		RETURNING id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at`

	row := q.db.QueryRow(ctx, sql,
		arg.CoursePhaseID, arg.ProviderType, arg.ResourceType, arg.Scope,
		arg.NameTemplate, arg.PermissionMapping, arg.ResourceExtraConfig)
	return scanResourceConfig(row)
}

// GetResourceConfig retrieves a resource config by ID and phase ID.
func (q *Queries) GetResourceConfig(ctx context.Context, id, coursePhaseID uuid.UUID) (ResourceConfig, error) {
	const sql = `
		SELECT id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at
		FROM resource_config
		WHERE id = $1 AND course_phase_id = $2`

	row := q.db.QueryRow(ctx, sql, id, coursePhaseID)
	return scanResourceConfig(row)
}

// ListResourceConfigs retrieves all resource configs for a course phase.
func (q *Queries) ListResourceConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]ResourceConfig, error) {
	const sql = `
		SELECT id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at
		FROM resource_config
		WHERE course_phase_id = $1
		ORDER BY created_at`

	rows, err := q.db.Query(ctx, sql, coursePhaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []ResourceConfig
	for rows.Next() {
		rc, err := scanResourceConfigFromRows(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, rc)
	}
	return configs, rows.Err()
}

// UpdateResourceConfig updates an existing resource config.
func (q *Queries) UpdateResourceConfig(ctx context.Context, arg UpdateResourceConfigParams) (ResourceConfig, error) {
	const sql = `
		UPDATE resource_config
		SET resource_type = $3, scope = $4, name_template = $5, permission_mapping = $6, resource_extra_config = $7
		WHERE id = $1 AND course_phase_id = $2
		RETURNING id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config, created_at`

	row := q.db.QueryRow(ctx, sql,
		arg.ID, arg.CoursePhaseID, arg.ResourceType, arg.Scope,
		arg.NameTemplate, arg.PermissionMapping, arg.ResourceExtraConfig)
	return scanResourceConfig(row)
}

// DeleteResourceConfig removes a resource config by ID and phase ID.
func (q *Queries) DeleteResourceConfig(ctx context.Context, id, coursePhaseID uuid.UUID) error {
	const sql = `DELETE FROM resource_config WHERE id = $1 AND course_phase_id = $2`
	_, err := q.db.Exec(ctx, sql, id, coursePhaseID)
	return err
}

// CopyResourceConfigs copies resource configs from one phase to another.
func (q *Queries) CopyResourceConfigs(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	const sql = `
		INSERT INTO resource_config (id, course_phase_id, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config)
		SELECT gen_random_uuid(), $2, provider_type, resource_type, scope, name_template, permission_mapping, resource_extra_config
		FROM resource_config
		WHERE course_phase_id = $1`

	_, err := q.db.Exec(ctx, sql, sourceCoursePhaseID, targetCoursePhaseID)
	return err
}

// scanResourceConfig scans a single ResourceConfig row from a QueryRow result.
func scanResourceConfig(row interface{ Scan(dest ...any) error }) (ResourceConfig, error) {
	var rc ResourceConfig
	err := row.Scan(
		&rc.ID, &rc.CoursePhaseID, &rc.ProviderType, &rc.ResourceType,
		&rc.Scope, &rc.NameTemplate, &rc.PermissionMapping, &rc.ResourceExtraConfig, &rc.CreatedAt)
	return rc, err
}

// scanResourceConfigFromRows scans a single ResourceConfig row from pgx.Rows.
func scanResourceConfigFromRows(rows interface {
	Scan(dest ...any) error
}) (ResourceConfig, error) {
	var rc ResourceConfig
	err := rows.Scan(
		&rc.ID, &rc.CoursePhaseID, &rc.ProviderType, &rc.ResourceType,
		&rc.Scope, &rc.NameTemplate, &rc.PermissionMapping, &rc.ResourceExtraConfig, &rc.CreatedAt)
	return rc, err
}
