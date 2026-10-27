BEGIN;

-- Records which student created a team so per-student creation limits can be enforced.
-- Nullable because lecturers create teams without a course participation.
ALTER TABLE team
    ADD COLUMN created_by_course_participation_id uuid;

CREATE INDEX team_course_phase_creator_idx
    ON team (course_phase_id, created_by_course_participation_id);

COMMIT;
