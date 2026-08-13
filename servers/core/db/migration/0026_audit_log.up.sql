-- Migration: Create the central audit log

CREATE TABLE audit_log (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at      timestamptz NOT NULL DEFAULT now(),
  actor_id        uuid,                              -- always a human in practice; nullable only for the maintenance edge
  actor_name      text NOT NULL DEFAULT '',
  actor_email     text NOT NULL DEFAULT '',
  actor_roles     text[] NOT NULL DEFAULT '{}',
  actor_role      text NOT NULL DEFAULT '',          -- role used, for the actor-role filter
  action          text NOT NULL,                     -- human-readable, e.g. "Created slot"
  action_key      text NOT NULL DEFAULT '',          -- machine key: "POST /.../slots"
  outcome         text NOT NULL DEFAULT 'success',   -- 'success' (2xx) | 'denied' (403)
  entity_type     text,
  entity_id       text,
  entity_name     text,                              -- snapshotted human-readable subject (e.g. "team Alpha")
  course_id       uuid,                              -- drives instructor scoping; null => admin-only
  course_phase_id uuid,
  source_service  text NOT NULL,                     -- "core", "interview", ...
  http_method     text,
  http_path       text,
  http_status     integer,
  metadata        jsonb NOT NULL DEFAULT '{}'
);

-- Keyset-friendly ordering indexes (created_at, id) for cursor pagination.
CREATE INDEX idx_audit_log_course_created ON audit_log (course_id, created_at DESC, id DESC);
CREATE INDEX idx_audit_log_created        ON audit_log (created_at DESC, id DESC);
CREATE INDEX idx_audit_log_actor          ON audit_log (actor_id, created_at DESC);

-- Equality filters exposed in the UI.
CREATE INDEX idx_audit_log_phase          ON audit_log (course_phase_id, created_at DESC);
CREATE INDEX idx_audit_log_source         ON audit_log (source_service);
CREATE INDEX idx_audit_log_entity_type    ON audit_log (entity_type);

-- Append-only guard: updates to an audit entry are never legitimate, so reject
-- them at the database. DELETE stays permitted (only the retention pruner uses it).
CREATE FUNCTION audit_log_no_update() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'audit_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_log_no_update
  BEFORE UPDATE ON audit_log
  FOR EACH ROW EXECUTE FUNCTION audit_log_no_update();
