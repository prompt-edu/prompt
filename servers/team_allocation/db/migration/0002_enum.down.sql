BEGIN;

ALTER TABLE student_skill_response ADD COLUMN rating INT;

UPDATE student_skill_response
SET rating = CASE skill_level
    WHEN 'novice'       THEN 1
    WHEN 'intermediate' THEN 2
    WHEN 'advanced'     THEN 3
    WHEN 'expert'       THEN 4
END;

ALTER TABLE student_skill_response ALTER COLUMN rating SET NOT NULL;
ALTER TABLE student_skill_response DROP COLUMN skill_level;

DROP TYPE skill_level;

COMMIT;
