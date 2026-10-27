BEGIN;

ALTER TABLE interview_assignment
    DROP CONSTRAINT IF EXISTS interview_assignment_participation_phase_uk,
    DROP CONSTRAINT IF EXISTS interview_assignment_slot_phase_fk,
    DROP COLUMN IF EXISTS course_phase_id;

ALTER TABLE interview_slot
    DROP CONSTRAINT IF EXISTS interview_slot_id_course_phase_uk;

COMMIT;
