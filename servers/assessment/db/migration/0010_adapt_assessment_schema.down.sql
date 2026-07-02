BEGIN;

DROP TABLE IF EXISTS action_item;

ALTER TABLE assessment_completion
    DROP COLUMN comment,
    DROP COLUMN grade_suggestion,
    DROP COLUMN completed;

ALTER TABLE assessment
    DROP COLUMN examples;

COMMIT;
