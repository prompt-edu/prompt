BEGIN;

DROP VIEW IF EXISTS category_course_phase;

ALTER TABLE category
    DROP COLUMN assessment_template_id;

DROP TABLE IF EXISTS assessment_template_course_phase;
DROP TABLE IF EXISTS assessment_template;

COMMIT;
