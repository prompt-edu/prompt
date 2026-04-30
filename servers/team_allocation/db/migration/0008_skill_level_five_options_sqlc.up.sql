-- Rewrites the skill_level enum from 4-level (novice/intermediate/advanced/expert)
-- to 5-level (very_bad/bad/ok/good/very_good).
--
-- On databases where migration 0007 already ran (PL/pgSQL), the column values are
-- already 5-level, so the text cast in the USING clause is an identity operation.
-- This plain-SQL equivalent lets sqlc correctly infer the final enum type.

BEGIN;

CREATE TYPE skill_level_new AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

ALTER TABLE student_skill_response
    ALTER COLUMN skill_level TYPE skill_level_new
    USING skill_level::text::skill_level_new;

DROP TYPE skill_level;
ALTER TYPE skill_level_new RENAME TO skill_level;

COMMIT;
