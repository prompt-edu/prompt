BEGIN;

-- PostgreSQL cannot drop a single enum value, so the type is rebuilt without 'partial'.
-- Any instance still in that state becomes 'failed', the retryable terminal state it
-- would otherwise have had.
UPDATE resource_instance SET status = 'failed' WHERE status = 'partial';

-- The partial indexes carry a predicate bound to the old type, so they are dropped and
-- recreated around the swap rather than left for the rewrite to reinterpret.
DROP INDEX IF EXISTS uq_resource_instance_team;
DROP INDEX IF EXISTS uq_resource_instance_student;

ALTER TYPE resource_status RENAME TO resource_status_old;
CREATE TYPE resource_status AS ENUM ('pending', 'in_progress', 'created', 'failed');

ALTER TABLE resource_instance ALTER COLUMN status DROP DEFAULT;
ALTER TABLE resource_instance
    ALTER COLUMN status TYPE resource_status USING status::text::resource_status;
ALTER TABLE resource_instance ALTER COLUMN status SET DEFAULT 'pending'::resource_status;

DROP TYPE resource_status_old;

CREATE UNIQUE INDEX uq_resource_instance_team
    ON resource_instance (resource_config_id, team_id)
    WHERE team_id IS NOT NULL AND status != 'failed';

CREATE UNIQUE INDEX uq_resource_instance_student
    ON resource_instance (resource_config_id, course_participation_id)
    WHERE course_participation_id IS NOT NULL AND status != 'failed';

COMMIT;
