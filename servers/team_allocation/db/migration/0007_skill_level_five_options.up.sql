BEGIN;

-- Replace 4-level skill_level enum with 5-level one matching the assessment score levels.
-- Old: novice | intermediate | advanced | expert
-- New: very_bad | bad | ok | good | very_good
-- Data migration: novice→bad, intermediate→ok, advanced→good, expert→very_good
--
-- Idempotent: checks whether the old enum values are still present before acting,
-- so re-running after a partial/manual apply is safe.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_enum
        WHERE enumtypid = 'skill_level'::regtype
        AND enumlabel = 'novice'
    ) THEN
        CREATE TYPE skill_level_new AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

        ALTER TABLE student_skill_response
            ADD COLUMN skill_level_new skill_level_new;

        UPDATE student_skill_response
        SET skill_level_new = CASE skill_level
            WHEN 'novice'::skill_level       THEN 'bad'::skill_level_new
            WHEN 'intermediate'::skill_level THEN 'ok'::skill_level_new
            WHEN 'advanced'::skill_level     THEN 'good'::skill_level_new
            WHEN 'expert'::skill_level       THEN 'very_good'::skill_level_new
            ELSE 'bad'::skill_level_new
        END;

        ALTER TABLE student_skill_response DROP COLUMN skill_level;
        ALTER TABLE student_skill_response RENAME COLUMN skill_level_new TO skill_level;
        ALTER TABLE student_skill_response ALTER COLUMN skill_level SET NOT NULL;

        DROP TYPE skill_level;
        ALTER TYPE skill_level_new RENAME TO skill_level;
    END IF;
END $$;

COMMIT;
