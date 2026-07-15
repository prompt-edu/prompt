BEGIN;

ALTER TABLE application_assessment
    DROP CONSTRAINT IF EXISTS unique_course_phase_participation_assessment;

COMMIT;
