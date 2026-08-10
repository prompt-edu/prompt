BEGIN;

DROP VIEW IF EXISTS note_with_versions;
DROP TABLE IF EXISTS note_tag_relation;
DROP TABLE IF EXISTS note_tag;
DROP TABLE IF EXISTS note_version;
DROP TABLE IF EXISTS note;
DROP TYPE IF EXISTS note_tag_color;

COMMIT;
