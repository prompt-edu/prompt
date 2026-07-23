BEGIN;

-- Restores the stricter 0002-version check. Data note: this can fail if template courses
-- with a null end_date (allowed by the relaxed check) exist at rollback time.
ALTER TABLE course DROP CONSTRAINT IF EXISTS check_end_date_after_start_date;
ALTER TABLE course
    ADD CONSTRAINT check_end_date_after_start_date CHECK (end_date > start_date);

COMMIT;
