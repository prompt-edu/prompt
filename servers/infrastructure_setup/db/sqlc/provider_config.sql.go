package db

import (
	"context"

	"github.com/google/uuid"
)

// UpsertProviderConfigParams holds the parameters for creating/updating a provider config.
type UpsertProviderConfigParams struct {
	CoursePhaseID uuid.UUID
	ProviderType  ProviderType
	Credentials   []byte
}

// UpsertProviderConfig inserts or updates a provider config row.
func (q *Queries) UpsertProviderConfig(ctx context.Context, arg UpsertProviderConfigParams) (ProviderConfig, error) {
	const sql = `
		INSERT INTO provider_config (id, course_phase_id, provider_type, credentials)
		VALUES (gen_random_uuid(), $1, $2, $3)
		ON CONFLICT (course_phase_id, provider_type)
		DO UPDATE SET credentials = EXCLUDED.credentials
		RETURNING id, course_phase_id, provider_type, credentials`

	row := q.db.QueryRow(ctx, sql, arg.CoursePhaseID, arg.ProviderType, arg.Credentials)
	var pc ProviderConfig
	err := row.Scan(&pc.ID, &pc.CoursePhaseID, &pc.ProviderType, &pc.Credentials)
	return pc, err
}

// GetProviderConfig retrieves a provider config by phase and provider type.
func (q *Queries) GetProviderConfig(ctx context.Context, coursePhaseID uuid.UUID, providerType ProviderType) (ProviderConfig, error) {
	const sql = `
		SELECT id, course_phase_id, provider_type, credentials
		FROM provider_config
		WHERE course_phase_id = $1 AND provider_type = $2`

	row := q.db.QueryRow(ctx, sql, coursePhaseID, providerType)
	var pc ProviderConfig
	err := row.Scan(&pc.ID, &pc.CoursePhaseID, &pc.ProviderType, &pc.Credentials)
	return pc, err
}

// ListProviderConfigs retrieves all provider configs for a course phase.
func (q *Queries) ListProviderConfigs(ctx context.Context, coursePhaseID uuid.UUID) ([]ProviderConfig, error) {
	const sql = `
		SELECT id, course_phase_id, provider_type, credentials
		FROM provider_config
		WHERE course_phase_id = $1
		ORDER BY provider_type`

	rows, err := q.db.Query(ctx, sql, coursePhaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []ProviderConfig
	for rows.Next() {
		var pc ProviderConfig
		if err := rows.Scan(&pc.ID, &pc.CoursePhaseID, &pc.ProviderType, &pc.Credentials); err != nil {
			return nil, err
		}
		configs = append(configs, pc)
	}
	return configs, rows.Err()
}

// CopyProviderConfigsWithEmptyCredentials copies provider configs to a new phase with empty credentials.
func (q *Queries) CopyProviderConfigsWithEmptyCredentials(ctx context.Context, sourceCoursePhaseID, targetCoursePhaseID uuid.UUID) error {
	const sql = `
		INSERT INTO provider_config (id, course_phase_id, provider_type, credentials)
		SELECT gen_random_uuid(), $2, provider_type, ''::bytea
		FROM provider_config
		WHERE course_phase_id = $1
		ON CONFLICT (course_phase_id, provider_type) DO NOTHING`

	_, err := q.db.Exec(ctx, sql, sourceCoursePhaseID, targetCoursePhaseID)
	return err
}
