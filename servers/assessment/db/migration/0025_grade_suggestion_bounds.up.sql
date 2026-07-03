BEGIN;

-- ponytail: API rejects >5.0; DB tolerates the legacy 6.0 "no grade" default/backfill from 0010 so the constraint can't fail on existing rows
ALTER TABLE assessment_completion
    ADD CONSTRAINT assessment_completion_grade_suggestion_check
        CHECK (grade_suggestion >= 1.0 AND grade_suggestion <= 6.0);

COMMIT;
