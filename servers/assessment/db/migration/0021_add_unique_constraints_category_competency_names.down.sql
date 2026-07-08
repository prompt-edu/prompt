BEGIN;

ALTER TABLE competency
    DROP CONSTRAINT IF EXISTS competency_category_id_name_unique;

ALTER TABLE category
    DROP CONSTRAINT IF EXISTS category_assessment_schema_id_name_unique;

COMMIT;
