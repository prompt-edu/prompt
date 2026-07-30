BEGIN;

-- Best-effort reverse. Data note: backfilled next_request_allowed_at values are dropped.
DROP VIEW IF EXISTS privacy_export_with_docs;

ALTER TABLE privacy_export DROP COLUMN IF EXISTS next_request_allowed_at;

-- Recreate the 0022-version view (without next_request_allowed_at).
CREATE VIEW privacy_export_with_docs AS
SELECT
  e.id,
  e.user_id,
  e.student_id,
  e.status,
  e.date_created,
  e.valid_until,
  COALESCE(
    jsonb_agg(
      json_build_object(
        'id', ed.id,
        'date_created', ed.date_created,
        'source_name', ed.source_name,
        'status', ed.status,
        'file_size', ed.file_size,
        'downloaded_at', ed.downloaded_at
      ) ORDER BY ed.date_created ASC
    ) FILTER (WHERE ed.id IS NOT NULL),
    '[]'::jsonb
  ) AS documents
FROM privacy_export e
LEFT JOIN privacy_export_document ed ON ed.export_id = e.id
GROUP BY e.id;

COMMIT;
