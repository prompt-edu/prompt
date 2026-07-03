BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN self_evaluation_enabled,
    DROP COLUMN self_evaluation_template,
    DROP COLUMN self_evaluation_deadline,
    DROP COLUMN peer_evaluation_enabled,
    DROP COLUMN peer_evaluation_template,
    DROP COLUMN peer_evaluation_deadline;

-- ponytail: pre-0013 assessment_template_id values are lost; reset to the 0011 default template
UPDATE course_phase_config
SET assessment_template_id = (SELECT id FROM assessment_template WHERE name = 'Intro Course Assessment Template');

ALTER TABLE course_phase_config
    ALTER COLUMN assessment_template_id DROP DEFAULT;

DELETE FROM assessment_template
WHERE name IN ('Assessment Template', 'Self Evaluation Template', 'Peer Evaluation Template');

DROP TABLE IF EXISTS feedback_items;
DROP TYPE IF EXISTS feedback_type;
DROP TABLE IF EXISTS evaluation_completion;
DROP TABLE IF EXISTS evaluation;
DROP TABLE IF EXISTS competency_map;

COMMIT;
