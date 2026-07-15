BEGIN;

-- Restores the original UNIQUE constraints from 0001. Data note: this can fail if, while
-- the partial indexes were active, more than one student was stored with an empty-string
-- matriculation_number or university_login (the partial indexes allowed that; a plain
-- UNIQUE constraint does not).
DROP INDEX IF EXISTS student_matriculation_number_unique;
DROP INDEX IF EXISTS student_university_login_unique;

ALTER TABLE student ADD CONSTRAINT student_matriculation_number_key UNIQUE (matriculation_number);
ALTER TABLE student ADD CONSTRAINT student_university_login_key UNIQUE (university_login);

COMMIT;
