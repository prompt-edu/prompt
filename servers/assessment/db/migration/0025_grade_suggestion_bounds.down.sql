BEGIN;

ALTER TABLE assessment_completion
    DROP CONSTRAINT IF EXISTS assessment_completion_grade_suggestion_check;

COMMIT;
