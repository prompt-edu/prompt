-- Presigned uploads that were handed out but not registered as a file yet. Uploads that
-- are never completed (or are rejected) would otherwise stay in the bucket forever, so a
-- reaper deletes the objects of entries that outlive the presign window.
CREATE TABLE pending_upload (
    storage_key TEXT PRIMARY KEY,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pending_upload_created_at ON pending_upload (created_at);
