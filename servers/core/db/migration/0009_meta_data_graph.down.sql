BEGIN;

ALTER TABLE application_question_multi_select
    DROP COLUMN IF EXISTS accessible_for_other_phases,
    DROP COLUMN IF EXISTS access_key;

ALTER TABLE application_question_text
    DROP COLUMN IF EXISTS accessible_for_other_phases,
    DROP COLUMN IF EXISTS access_key;

DROP TABLE IF EXISTS meta_data_dependency_graph;

COMMIT;
