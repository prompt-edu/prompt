BEGIN;

-- One row per (config, target), whatever its status. The previous indexes excluded
-- failed rows, so a re-trigger inserted a second row for a target whose first attempt
-- had failed instead of converging on the existing one, and resetting the older row to
-- pending then collided with the newer one.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY resource_config_id, team_id
               ORDER BY CASE status
                            WHEN 'created' THEN 0
                            WHEN 'partial' THEN 1
                            WHEN 'in_progress' THEN 2
                            WHEN 'pending' THEN 3
                            ELSE 4
                        END,
                        updated_at DESC,
                        id
           ) AS duplicate_rank
    FROM resource_instance
    WHERE team_id IS NOT NULL
)
DELETE FROM resource_instance
WHERE id IN (SELECT id FROM ranked WHERE duplicate_rank > 1);

WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (
               PARTITION BY resource_config_id, course_participation_id
               ORDER BY CASE status
                            WHEN 'created' THEN 0
                            WHEN 'partial' THEN 1
                            WHEN 'in_progress' THEN 2
                            WHEN 'pending' THEN 3
                            ELSE 4
                        END,
                        updated_at DESC,
                        id
           ) AS duplicate_rank
    FROM resource_instance
    WHERE course_participation_id IS NOT NULL
)
DELETE FROM resource_instance
WHERE id IN (SELECT id FROM ranked WHERE duplicate_rank > 1);

DROP INDEX IF EXISTS uq_resource_instance_team;
DROP INDEX IF EXISTS uq_resource_instance_student;

CREATE UNIQUE INDEX uq_resource_instance_team
    ON resource_instance (resource_config_id, team_id)
    WHERE team_id IS NOT NULL;

CREATE UNIQUE INDEX uq_resource_instance_student
    ON resource_instance (resource_config_id, course_participation_id)
    WHERE course_participation_id IS NOT NULL;

COMMIT;
