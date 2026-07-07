BEGIN;

ALTER TABLE category
    DROP CONSTRAINT IF EXISTS category_assessment_schema_id_fkey;

ALTER TABLE course_phase_config
    DROP CONSTRAINT IF EXISTS course_phase_config_assessment_schema_id_fkey,
    DROP CONSTRAINT IF EXISTS course_phase_config_self_evaluation_schema_fkey,
    DROP CONSTRAINT IF EXISTS course_phase_config_peer_evaluation_schema_fkey,
    DROP CONSTRAINT IF EXISTS course_phase_config_tutor_evaluation_schema_fkey;

ALTER TABLE assessment_schema
    RENAME TO assessment_template;

ALTER TABLE category
    RENAME COLUMN assessment_schema_id TO assessment_template_id;

ALTER TABLE course_phase_config
    RENAME COLUMN assessment_schema_id TO assessment_template_id;
ALTER TABLE course_phase_config
    RENAME COLUMN self_evaluation_schema TO self_evaluation_template;
ALTER TABLE course_phase_config
    RENAME COLUMN peer_evaluation_schema TO peer_evaluation_template;
ALTER TABLE course_phase_config
    RENAME COLUMN tutor_evaluation_schema TO tutor_evaluation_template;

ALTER TABLE category
    ADD CONSTRAINT category_assessment_template_id_fkey
        FOREIGN KEY (assessment_template_id) REFERENCES assessment_template (id) ON DELETE CASCADE;

ALTER TABLE course_phase_config
    ADD CONSTRAINT course_phase_config_assessment_template_id_fkey
        FOREIGN KEY (assessment_template_id) REFERENCES assessment_template (id) ON DELETE RESTRICT,
    ADD CONSTRAINT course_phase_config_self_evaluation_template_fkey
        FOREIGN KEY (self_evaluation_template) REFERENCES assessment_template (id) ON DELETE RESTRICT,
    ADD CONSTRAINT course_phase_config_peer_evaluation_template_fkey
        FOREIGN KEY (peer_evaluation_template) REFERENCES assessment_template (id) ON DELETE RESTRICT,
    ADD CONSTRAINT course_phase_config_tutor_evaluation_template_fkey
        FOREIGN KEY (tutor_evaluation_template) REFERENCES assessment_template (id) ON DELETE RESTRICT;

DROP VIEW IF EXISTS category_course_phase;
CREATE VIEW category_course_phase AS
SELECT c.id AS category_id,
       cpc.course_phase_id
FROM category c
         INNER JOIN course_phase_config cpc
                    ON c.assessment_template_id = cpc.assessment_template_id;

COMMIT;
