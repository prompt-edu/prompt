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

COMMIT;
