BEGIN;

DROP INDEX IF EXISTS uq_resource_instance_team;
DROP INDEX IF EXISTS uq_resource_instance_student;

CREATE UNIQUE INDEX uq_resource_instance_team
    ON resource_instance (resource_config_id, team_id)
    WHERE team_id IS NOT NULL AND status != 'failed';

CREATE UNIQUE INDEX uq_resource_instance_student
    ON resource_instance (resource_config_id, course_participation_id)
    WHERE course_participation_id IS NOT NULL AND status != 'failed';

COMMIT;
