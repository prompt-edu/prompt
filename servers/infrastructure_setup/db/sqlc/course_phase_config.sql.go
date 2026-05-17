package db

import (
	"context"

	"github.com/google/uuid"
)

// UpsertCoursePhaseConfigParams holds phase configuration values.
type UpsertCoursePhaseConfigParams struct {
	CoursePhaseID              uuid.UUID
	TeamSourceCoursePhaseID    *uuid.UUID
	StudentSourceCoursePhaseID *uuid.UUID
	SemesterTag                string
}

// GetCoursePhaseConfig returns the infrastructure setup config for a phase.
func (q *Queries) GetCoursePhaseConfig(ctx context.Context, coursePhaseID uuid.UUID) (CoursePhaseConfig, error) {
	const sql = `
		SELECT course_phase_id, team_source_course_phase_id, student_source_course_phase_id, semester_tag
		FROM course_phase_config
		WHERE course_phase_id = $1`

	var cfg CoursePhaseConfig
	err := q.db.QueryRow(ctx, sql, coursePhaseID).Scan(
		&cfg.CoursePhaseID,
		&cfg.TeamSourceCoursePhaseID,
		&cfg.StudentSourceCoursePhaseID,
		&cfg.SemesterTag,
	)
	return cfg, err
}

// UpsertCoursePhaseConfig creates or updates infrastructure setup config.
func (q *Queries) UpsertCoursePhaseConfig(ctx context.Context, arg UpsertCoursePhaseConfigParams) (CoursePhaseConfig, error) {
	const sql = `
		INSERT INTO course_phase_config (
			course_phase_id,
			team_source_course_phase_id,
			student_source_course_phase_id,
			semester_tag
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (course_phase_id)
		DO UPDATE SET
			team_source_course_phase_id = EXCLUDED.team_source_course_phase_id,
			student_source_course_phase_id = EXCLUDED.student_source_course_phase_id,
			semester_tag = EXCLUDED.semester_tag
		RETURNING course_phase_id, team_source_course_phase_id, student_source_course_phase_id, semester_tag`

	var cfg CoursePhaseConfig
	err := q.db.QueryRow(
		ctx,
		sql,
		arg.CoursePhaseID,
		arg.TeamSourceCoursePhaseID,
		arg.StudentSourceCoursePhaseID,
		arg.SemesterTag,
	).Scan(
		&cfg.CoursePhaseID,
		&cfg.TeamSourceCoursePhaseID,
		&cfg.StudentSourceCoursePhaseID,
		&cfg.SemesterTag,
	)
	return cfg, err
}
