BEGIN;

ALTER TABLE competency
    DROP COLUMN IF EXISTS sort_order;

ALTER TABLE category
    DROP COLUMN IF EXISTS sort_order;

COMMIT;
