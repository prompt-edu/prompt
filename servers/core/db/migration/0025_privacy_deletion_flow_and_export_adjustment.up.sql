-- Migration: Add Privacy Deletion Tables

CREATE TYPE privacy_deletion_request_status AS ENUM (
  'pending_approval',
  'rejected',
  'in_progress',
  'succeeded',
  'failed'
);

CREATE TYPE privacy_deletion_subrequest_status AS ENUM (
  'pending',
  'in_progress',
  'succeeded',
  'failed'
);

CREATE TABLE privacy_deletion_request (
  id                    uuid                              PRIMARY KEY,
  user_id               uuid,
  student_id            uuid                              REFERENCES student(id) ON DELETE SET NULL,
  requested_at          timestamptz                       NOT NULL DEFAULT now(),
  status                privacy_deletion_request_status   NOT NULL DEFAULT 'pending_approval',
  auditor_id            uuid,
  auditor_name          text                              NOT NULL DEFAULT '',
  auditor_email         text                              NOT NULL DEFAULT '',
  auditor_responded_at  timestamptz,
  auditor_note          text                              NOT NULL DEFAULT '',
  recipient_email       text                              NOT NULL DEFAULT '',
  completed_at          timestamptz
);

CREATE TABLE privacy_deletion_subrequest (
  id                    uuid                                PRIMARY KEY,
  deletion_request_id   uuid                                NOT NULL REFERENCES privacy_deletion_request(id) ON DELETE CASCADE,
  source_name           text                                NOT NULL,
  status                privacy_deletion_subrequest_status  NOT NULL DEFAULT 'pending',
  created_at            timestamptz                         NOT NULL DEFAULT now(),
  completed_at          timestamptz,
  error_message         text                                NOT NULL DEFAULT ''
);

-- privacy_export: user_id is nullable too (admin-initiated exports may have no Keycloak user).
ALTER TABLE privacy_export
  ALTER COLUMN user_id DROP NOT NULL;

-- The privacy_export_with_docs view was created in 0024 when user_id was still NOT NULL,
-- so its column-nullability metadata is stale. Drop and recreate so sqlc (and any other
-- introspection tool) sees user_id as nullable.
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

-- instructor notes: on delete cascade
ALTER TABLE note
  DROP CONSTRAINT note_for_student_fkey,
  ADD CONSTRAINT note_for_student_fkey
    FOREIGN KEY (for_student) REFERENCES student(id) ON DELETE CASCADE;

ALTER TABLE note_version
  DROP CONSTRAINT note_version_for_note_fkey,
  ADD CONSTRAINT note_version_for_note_fkey
    FOREIGN KEY (for_note) REFERENCES note(id) ON DELETE CASCADE;


CREATE VIEW privacy_deletion_request_with_subrequests AS
SELECT
  r.id,
  r.user_id,
  r.student_id,
  r.requested_at,
  r.status,
  r.auditor_id,
  r.auditor_name,
  r.auditor_email,
  r.auditor_responded_at,
  r.auditor_note,
  r.recipient_email,
  r.completed_at,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'id', sr.id,
        'source_name', sr.source_name,
        'status', sr.status,
        'created_at', sr.created_at,
        'completed_at', sr.completed_at,
        'error_message', sr.error_message
      ) ORDER BY sr.created_at ASC
    ) FILTER (WHERE sr.id IS NOT NULL),
    '[]'::jsonb
  )::jsonb AS subrequests
FROM privacy_deletion_request r
LEFT JOIN privacy_deletion_subrequest sr ON sr.deletion_request_id = r.id
GROUP BY r.id;
