-- Migration: Add Privacy Export Tables

CREATE TYPE export_status AS ENUM ('pending', 'complete', 'failed');

CREATE TABLE privacy_export (
  id           uuid          PRIMARY KEY,
  userID       uuid          NOT NULL,
  studentID    uuid          REFERENCES student(id),
  status       export_status NOT NULL DEFAULT 'pending',
  date_created timestamptz   NOT NULL DEFAULT now(),
  valid_until  timestamptz   NOT NULL
);

CREATE TABLE privacy_export_document (
  id           uuid          PRIMARY KEY,
  exportID     uuid          REFERENCES privacy_export(id),
  date_created timestamptz   NOT NULL DEFAULT now(),
  source_name  text          NOT NULL,
  object_key   text          NOT NULL,
  status       export_status NOT NULL DEFAULT 'pending',
  file_size    bigint
);

CREATE VIEW privacy_export_with_docs AS
SELECT
  e.id,
  e.userID,
  e.studentID,
  e.status,
  e.date_created,
  e.valid_until,
  COALESCE(
    jsonb_agg(
      json_build_object(
        'id', ed.id,
        'date_created', ed.date_created,
        'source_name', ed.source_name,
        'object_key', ed.object_key,
        'status', ed.status,
        'file_size', ed.file_size
      ) ORDER BY ed.date_created ASC
    ) FILTER (WHERE ed.id IS NOT NULL),
    '[]'::jsonb
  ) AS documents
FROM privacy_export e
LEFT JOIN privacy_export_document ed ON ed.exportID = e.id
GROUP BY e.id;
