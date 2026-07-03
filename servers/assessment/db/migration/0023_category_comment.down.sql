BEGIN;

-- ponytail: pre-0023 comment/examples were aggregated into category_assessment; not restorable, re-add empty
ALTER TABLE assessment
    ADD COLUMN comment TEXT,
    ADD COLUMN examples text DEFAULT '' NOT NULL;

ALTER TABLE assessment
    DROP COLUMN author_id;

DROP TABLE IF EXISTS category_assessment;

COMMIT;
