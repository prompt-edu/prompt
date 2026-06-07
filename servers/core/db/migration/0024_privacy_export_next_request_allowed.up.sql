ALTER TABLE privacy_export
  ADD COLUMN IF NOT EXISTS next_request_allowed_at timestamptz;

UPDATE privacy_export
  SET next_request_allowed_at = date_created + INTERVAL '30 days'
  WHERE next_request_allowed_at IS NULL;

ALTER TABLE privacy_export
  ALTER COLUMN next_request_allowed_at SET NOT NULL;

DROP VIEW IF EXISTS privacy_export_with_docs;

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
