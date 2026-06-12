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
  user_id               uuid                              NOT NULL,
  student_id            uuid                              REFERENCES student(id) ON DELETE SET NULL,
  requested_at          timestamptz                       NOT NULL DEFAULT now(),
  status                privacy_deletion_request_status   NOT NULL DEFAULT 'pending_approval',
  auditor_id            uuid,
  auditor_name          text                              NOT NULL DEFAULT '',
  auditor_email         text                              NOT NULL DEFAULT '',
  auditor_responded_at  timestamptz,
  auditor_note          text                              NOT NULL DEFAULT '',
  completed_at          timestamptz
);

-- Enforce: never more than one request in  in_progress  or  pending_approval  status.
CREATE UNIQUE INDEX privacy_deletion_request_one_open_per_user
  ON privacy_deletion_request (user_id)
  WHERE status IN ('pending_approval', 'in_progress');

CREATE TABLE privacy_deletion_subrequest (
  id                    uuid                                PRIMARY KEY,
  deletion_request_id   uuid                                NOT NULL REFERENCES privacy_deletion_request(id) ON DELETE CASCADE,
  source_name           text                                NOT NULL,
  status                privacy_deletion_subrequest_status  NOT NULL DEFAULT 'pending',
  created_at            timestamptz                         NOT NULL DEFAULT now(),
  completed_at          timestamptz,
  error_message         text                                NOT NULL DEFAULT ''
);

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
