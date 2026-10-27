BEGIN;

DROP INDEX IF EXISTS team_course_phase_creator_idx;

ALTER TABLE team
    DROP COLUMN IF EXISTS created_by_course_participation_id;

COMMIT;
