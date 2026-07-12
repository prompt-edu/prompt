BEGIN;

DROP INDEX IF EXISTS idx_tutor_phase_login;
ALTER TABLE tutor DROP COLUMN IF EXISTS university_login;

COMMIT;
