BEGIN;

CREATE TYPE skill_level_new AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

ALTER TABLE student_skill_response
    ALTER COLUMN skill_level TYPE skill_level_new
    USING (CASE skill_level::text
        WHEN 'novice' THEN 'bad'
        WHEN 'intermediate' THEN 'ok'
        WHEN 'advanced' THEN 'good'
        WHEN 'expert' THEN 'very_good'
        ELSE skill_level::text
    END)::skill_level_new;

DROP TYPE skill_level;
ALTER TYPE skill_level_new RENAME TO skill_level;

COMMIT;
