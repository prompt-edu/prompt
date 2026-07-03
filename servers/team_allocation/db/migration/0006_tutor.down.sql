BEGIN;

DROP TABLE IF EXISTS tutor;

ALTER TABLE allocations ADD COLUMN student_full_name TEXT NOT NULL DEFAULT '';

UPDATE allocations
SET student_full_name = TRIM(COALESCE(student_first_name, '') || ' ' || COALESCE(student_last_name, ''))
WHERE student_first_name IS NOT NULL OR student_last_name IS NOT NULL;

ALTER TABLE allocations
    DROP COLUMN student_first_name,
    DROP COLUMN student_last_name;

COMMIT;
