BEGIN;

-- Best-effort reverse. Data note: 'not_assessed' maps back to NULL, so the explicit
-- not-assessed state is not distinguishable from a missing value after rollback.
ALTER TABLE course_phase_participation ADD COLUMN passed boolean;

UPDATE course_phase_participation
SET passed = CASE pass_status
    WHEN 'passed' THEN true
    WHEN 'failed' THEN false
    ELSE NULL
END;

ALTER TABLE course_phase_participation DROP COLUMN pass_status;

DROP TYPE pass_status;

COMMIT;
