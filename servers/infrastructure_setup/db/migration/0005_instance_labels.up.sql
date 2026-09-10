BEGIN;

-- What the row is about, so the execution list can name a team and its resource instead
-- of showing two UUID prefixes. Both are labels PROMPT resolved at run time: the target
-- name comes from core, the resolved name is what the provider was asked to create.
ALTER TABLE resource_instance
    ADD COLUMN target_name   text NOT NULL DEFAULT '',
    ADD COLUMN resolved_name text NOT NULL DEFAULT '';

COMMIT;
