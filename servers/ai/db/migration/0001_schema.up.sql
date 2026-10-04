BEGIN;

CREATE TABLE ai_phase_key (
    course_phase_id uuid PRIMARY KEY,
    phase_type text NOT NULL,
    encrypted_key bytea NOT NULL,
    last4 text NOT NULL,
    set_by text NOT NULL,
    set_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ai_call (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL,
    actor_id text NOT NULL,
    actor_role text NOT NULL,
    feature text NOT NULL,
    template text,
    template_version text,
    requested_model text,
    served_model text,
    provider text NOT NULL,
    params jsonb NOT NULL DEFAULT '{}',
    context_hash text,
    response_hash text,
    outcome text NOT NULL CHECK (outcome IN ('pending', 'success', 'error', 'timeout', 'cancelled', 'denied')),
    http_status integer,
    finish_reason text,
    error_code text,
    prompt_tokens integer,
    completion_tokens integer,
    streamed boolean NOT NULL,
    server_version text NOT NULL,
    requested_at timestamptz NOT NULL DEFAULT now(),
    first_token_at timestamptz,
    completed_at timestamptz
);

CREATE INDEX ai_call_phase_requested_idx ON ai_call (course_phase_id, requested_at DESC, id DESC);
CREATE INDEX ai_call_actor_idx ON ai_call (actor_id);
CREATE INDEX ai_call_feature_requested_idx ON ai_call (feature, requested_at);
CREATE INDEX ai_call_requested_idx ON ai_call (requested_at);

CREATE TABLE ai_call_content (
    call_id uuid PRIMARY KEY REFERENCES ai_call (id) ON DELETE CASCADE,
    request bytea NOT NULL,
    response bytea
);

CREATE TABLE ai_call_subject (
    call_id uuid NOT NULL REFERENCES ai_call (id) ON DELETE CASCADE,
    course_participation_id uuid NOT NULL,
    PRIMARY KEY (call_id, course_participation_id)
);

CREATE INDEX ai_call_subject_participation_idx ON ai_call_subject (course_participation_id);

CREATE TABLE ai_call_content_restriction (
    call_id uuid PRIMARY KEY REFERENCES ai_call (id) ON DELETE CASCADE,
    restricted_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ai_call_event (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    call_id uuid NOT NULL REFERENCES ai_call (id) ON DELETE CASCADE,
    actor_id text NOT NULL,
    type text NOT NULL CHECK (type IN ('shown', 'accepted', 'edited', 'rejected', 'content_viewed')),
    data jsonb NOT NULL DEFAULT '{}',
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ai_call_event_call_idx ON ai_call_event (call_id, created_at);

-- Append-only: the one permitted update completes a pending call once; purge and erasure delete.
CREATE FUNCTION ai_call_complete_only() RETURNS trigger AS $$
BEGIN
  IF OLD.completed_at IS NOT NULL
    OR (NEW.id, NEW.course_phase_id, NEW.actor_id, NEW.actor_role, NEW.feature, NEW.template,
        NEW.template_version, NEW.requested_model, NEW.provider, NEW.params, NEW.context_hash,
        NEW.streamed, NEW.server_version, NEW.requested_at)
      IS DISTINCT FROM
       (OLD.id, OLD.course_phase_id, OLD.actor_id, OLD.actor_role, OLD.feature, OLD.template,
        OLD.template_version, OLD.requested_model, OLD.provider, OLD.params, OLD.context_hash,
        OLD.streamed, OLD.server_version, OLD.requested_at)
  THEN
    RAISE EXCEPTION 'ai_call is append-only except for completing a pending call';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ai_call_complete_only
  BEFORE UPDATE ON ai_call
  FOR EACH ROW EXECUTE FUNCTION ai_call_complete_only();

CREATE FUNCTION ai_call_content_response_only() RETURNS trigger AS $$
BEGIN
  IF OLD.response IS NOT NULL
    OR (NEW.call_id, NEW.request) IS DISTINCT FROM (OLD.call_id, OLD.request)
  THEN
    RAISE EXCEPTION 'ai_call_content is append-only except for setting the response once';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ai_call_content_response_only
  BEFORE UPDATE ON ai_call_content
  FOR EACH ROW EXECUTE FUNCTION ai_call_content_response_only();

CREATE FUNCTION ai_audit_no_update() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION '% is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_ai_call_subject_no_update
  BEFORE UPDATE ON ai_call_subject
  FOR EACH ROW EXECUTE FUNCTION ai_audit_no_update();

CREATE TRIGGER trg_ai_call_content_restriction_no_update
  BEFORE UPDATE ON ai_call_content_restriction
  FOR EACH ROW EXECUTE FUNCTION ai_audit_no_update();

CREATE TRIGGER trg_ai_call_event_no_update
  BEFORE UPDATE ON ai_call_event
  FOR EACH ROW EXECUTE FUNCTION ai_audit_no_update();

COMMIT;
