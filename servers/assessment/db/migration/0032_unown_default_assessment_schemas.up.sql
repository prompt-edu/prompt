-- The default schemas the course_phase_config columns fall back to belong to no phase. Migration
-- 0020_add_source_phase_id gave the default assessment schema an owner wherever a config selected
-- it, which let that phase edit the schema every new phase starts with, and deleting that phase
-- would delete the schema the column default still points to. The defaults are looked up from the
-- column defaults rather than by name, because a schema can be renamed.
UPDATE assessment_schema
SET source_phase_id = NULL,
    updated_at      = CURRENT_TIMESTAMP
WHERE source_phase_id IS NOT NULL
  AND id IN (SELECT substring(column_default FROM '[0-9a-fA-F-]{36}')::uuid
             FROM information_schema.columns
             WHERE table_schema = current_schema()
               AND table_name = 'course_phase_config'
               AND column_name IN ('assessment_schema_id', 'self_evaluation_schema',
                                   'peer_evaluation_schema', 'tutor_evaluation_schema')
               AND column_default IS NOT NULL);
