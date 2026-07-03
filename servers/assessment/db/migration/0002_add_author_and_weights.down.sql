BEGIN;

ALTER TABLE assessment
DROP COLUMN author;

ALTER TABLE competency
DROP COLUMN weight;

ALTER TABLE category
DROP COLUMN weight;

COMMIT;
