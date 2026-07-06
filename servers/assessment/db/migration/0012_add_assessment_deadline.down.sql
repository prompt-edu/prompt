BEGIN;

ALTER TABLE course_phase_config
    DROP COLUMN deadline;

ALTER TABLE course_phase_config
    RENAME TO assessment_template_course_phase;

CREATE OR REPLACE VIEW category_course_phase AS
SELECT c.id AS category_id,
       atcp.course_phase_id
FROM category c
         INNER JOIN assessment_template_course_phase atcp
                    ON c.assessment_template_id = atcp.assessment_template_id;

COMMIT;
