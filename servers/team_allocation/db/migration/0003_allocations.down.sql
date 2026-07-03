BEGIN;

DROP TABLE IF EXISTS allocations;

ALTER TABLE team DROP CONSTRAINT IF EXISTS team_id_course_phase_uk;

COMMIT;
