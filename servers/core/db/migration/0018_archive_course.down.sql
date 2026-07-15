ALTER TABLE course
    DROP COLUMN IF EXISTS archived,
    DROP COLUMN IF EXISTS archived_on;
