BEGIN;

ALTER TABLE tutor ADD COLUMN university_login TEXT;

CREATE UNIQUE INDEX idx_tutor_phase_login
    ON tutor (course_phase_id, university_login)
    WHERE university_login IS NOT NULL;

COMMIT;
