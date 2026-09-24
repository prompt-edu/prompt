BEGIN;

-- A certificate is now released only once a release date is set and has passed. Phases that
-- already have a template but no release date were released under the previous rule, so they
-- keep their access with the migration time as their release date. Their certificates print that
-- date from now on.
UPDATE course_phase_config
SET
    release_date = NOW()
WHERE
    release_date IS NULL
    AND template_content IS NOT NULL
    AND template_content <> '';

COMMIT;
