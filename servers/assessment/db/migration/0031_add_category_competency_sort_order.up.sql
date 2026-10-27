BEGIN;

ALTER TABLE category
    ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

ALTER TABLE competency
    ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- Keep the alphabetical order existing schemas were displayed in
UPDATE category c
SET sort_order = ordered.position
FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY assessment_schema_id ORDER BY name) - 1 AS position
      FROM category) ordered
WHERE c.id = ordered.id;

UPDATE competency cmp
SET sort_order = ordered.position
FROM (SELECT id, ROW_NUMBER() OVER (PARTITION BY category_id ORDER BY name) - 1 AS position
      FROM competency) ordered
WHERE cmp.id = ordered.id;

-- Check competency name uniqueness at the end of each statement rather than per row, so a single
-- statement can move competencies between categories in any order (e.g. swap two same-named ones)
ALTER TABLE competency
    DROP CONSTRAINT competency_category_id_name_unique;

ALTER TABLE competency
    ADD CONSTRAINT competency_category_id_name_unique
        UNIQUE (category_id, name) DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
