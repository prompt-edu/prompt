BEGIN;

ALTER TABLE course DROP CONSTRAINT IF EXISTS check_end_date_after_start_date;
ALTER TABLE course DROP CONSTRAINT IF EXISTS unique_course_identifier;

COMMIT;
