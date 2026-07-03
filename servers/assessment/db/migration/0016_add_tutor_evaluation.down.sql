BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN tutor_evaluation_enabled,
    DROP COLUMN tutor_evaluation_start,
    DROP COLUMN tutor_evaluation_deadline,
    DROP COLUMN tutor_evaluation_template;

DELETE FROM assessment_template WHERE name = 'Tutor Evaluation Template';

ALTER TABLE evaluation DROP COLUMN type;
ALTER TABLE evaluation_completion DROP COLUMN type;
ALTER TABLE feedback_items DROP COLUMN type;

DROP TYPE IF EXISTS assessment_type;

COMMIT;
