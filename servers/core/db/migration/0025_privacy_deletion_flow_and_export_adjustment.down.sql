BEGIN;

-- Best-effort reverse. Data note: restoring privacy_export.user_id to NOT NULL below can
-- fail if admin-initiated exports with a NULL user_id exist at rollback time.

DROP VIEW IF EXISTS privacy_deletion_request_with_subrequests;

-- Restore the instructor-note foreign keys to their pre-0025 (non-cascading) form.
ALTER TABLE note_version
    DROP CONSTRAINT note_version_for_note_fkey,
    ADD CONSTRAINT note_version_for_note_fkey
        FOREIGN KEY (for_note) REFERENCES note(id);

ALTER TABLE note
    DROP CONSTRAINT note_for_student_fkey,
    ADD CONSTRAINT note_for_student_fkey
        FOREIGN KEY (for_student) REFERENCES student(id);

-- Restore privacy_export.user_id NOT NULL and recreate the 0024-version view (the up
-- recreated the view after changing nullability, so the down mirrors that order).
DROP VIEW IF EXISTS privacy_export_with_docs;

ALTER TABLE privacy_export ALTER COLUMN user_id SET NOT NULL;

CREATE VIEW privacy_export_with_docs AS
SELECT
  e.id,
  e.user_id,
  e.student_id,
  e.status,
  e.date_created,
  e.valid_until,
  e.next_request_allowed_at,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'id', ed.id,
        'date_created', ed.date_created,
        'source_name', ed.source_name,
        'status', ed.status,
        'file_size', ed.file_size,
        'downloaded_at', ed.downloaded_at
      ) ORDER BY ed.date_created ASC
    ) FILTER (WHERE ed.id IS NOT NULL),
    '[]'::jsonb
  )::jsonb AS documents
FROM privacy_export e
LEFT JOIN privacy_export_document ed ON ed.export_id = e.id
GROUP BY e.id;

DROP TABLE IF EXISTS privacy_deletion_subrequest;
DROP TABLE IF EXISTS privacy_deletion_request;

DROP TYPE IF EXISTS privacy_deletion_subrequest_status;
DROP TYPE IF EXISTS privacy_deletion_request_status;

COMMIT;
