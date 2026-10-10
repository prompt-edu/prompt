BEGIN;

ALTER TABLE competency
    DROP CONSTRAINT competency_category_id_name_unique;

ALTER TABLE competency
    ADD CONSTRAINT competency_category_id_name_unique
        UNIQUE (category_id, name);

ALTER TABLE competency
    DROP COLUMN IF EXISTS sort_order;

ALTER TABLE category
    DROP COLUMN IF EXISTS sort_order;

COMMIT;
