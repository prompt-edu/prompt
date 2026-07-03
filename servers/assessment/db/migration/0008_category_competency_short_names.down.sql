BEGIN;

ALTER TABLE competency
DROP COLUMN short_name;

ALTER TABLE category
DROP COLUMN short_name;

COMMIT;
