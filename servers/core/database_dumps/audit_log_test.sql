-- Minimal, self-contained schema for auditLog package tests.

CREATE TABLE audit_log (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at      timestamptz NOT NULL DEFAULT now(),
  actor_id        uuid,
  actor_name      text NOT NULL DEFAULT '',
  actor_email     text NOT NULL DEFAULT '',
  actor_roles     text[] NOT NULL DEFAULT '{}',
  actor_role      text NOT NULL DEFAULT '',
  action          text NOT NULL,
  action_key      text NOT NULL DEFAULT '',
  outcome         text NOT NULL DEFAULT 'success',
  entity_type     text,
  entity_id       text,
  entity_name     text,
  course_id       uuid,
  course_phase_id uuid,
  source_service  text NOT NULL,
  http_method     text,
  http_path       text,
  http_status     integer,
  metadata        jsonb NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_audit_log_course_created ON audit_log (course_id, created_at DESC, id DESC);
CREATE INDEX idx_audit_log_created        ON audit_log (created_at DESC, id DESC);

CREATE FUNCTION audit_log_no_update() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'audit_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_audit_log_no_update
  BEFORE UPDATE ON audit_log
  FOR EACH ROW EXECUTE FUNCTION audit_log_no_update();

-- Minimal course_phase for course_id resolution (GetCourseIDByCoursePhaseID).
CREATE TABLE course_phase (
  id        uuid PRIMARY KEY,
  course_id uuid NOT NULL
);

INSERT INTO course_phase (id, course_id) VALUES
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222');
