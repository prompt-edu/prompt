BEGIN;

-- API rejects >5.0; DB tolerates the legacy 6.0 "no grade" default/backfill from 0010.
-- NOT VALID skips validating pre-existing rows so a stale out-of-range value can't block startup; new writes are still enforced.
ALTER TABLE assessment_completion
    ADD CONSTRAINT assessment_completion_grade_suggestion_check
        CHECK (grade_suggestion >= 1.0 AND grade_suggestion <= 6.0) NOT VALID;

COMMIT;
