BEGIN;

DROP VIEW IF EXISTS privacy_export_with_docs;
DROP TABLE IF EXISTS privacy_export_document;
DROP TABLE IF EXISTS privacy_export;
DROP TYPE IF EXISTS export_status;

COMMIT;
