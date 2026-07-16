BEGIN;

CREATE TABLE course_phase_config (
    course_phase_id uuid PRIMARY KEY,
    target_mode text NOT NULL DEFAULT 'individual' CHECK (target_mode IN ('individual', 'team')),
    feedback_mode text NOT NULL DEFAULT 'independent' CHECK (feedback_mode IN ('independent', 'shared')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feedback_category (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL REFERENCES course_phase_config(course_phase_id) ON DELETE CASCADE,
    name text NOT NULL CHECK (length(btrim(name)) > 0),
    description text NOT NULL DEFAULT '',
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (course_phase_id, name),
    UNIQUE (course_phase_id, position)
);

CREATE TABLE presentation_slot (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL,
    start_time timestamptz NOT NULL,
    end_time timestamptz NOT NULL,
    location text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (end_time > start_time)
);

CREATE TABLE presentation (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id uuid NOT NULL,
    slot_id uuid NOT NULL UNIQUE REFERENCES presentation_slot(id) ON DELETE RESTRICT,
    target_type text NOT NULL CHECK (target_type IN ('individual', 'team')),
    target_id uuid NOT NULL,
    target_name text NOT NULL CHECK (length(btrim(target_name)) > 0),
    feedback_release_name text,
    feedback_released_at timestamptz,
    feedback_released_by_user_id text,
    feedback_released_by_name text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (course_phase_id, target_type, target_id)
);

CREATE TABLE presentation_material (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    presentation_id uuid NOT NULL REFERENCES presentation(id) ON DELETE CASCADE,
    original_filename text NOT NULL CHECK (length(btrim(original_filename)) > 0),
    content_type text NOT NULL CHECK (length(btrim(content_type)) > 0),
    size_bytes bigint NOT NULL DEFAULT 0 CHECK (size_bytes >= 0),
    storage_key text NOT NULL UNIQUE,
    state text NOT NULL DEFAULT 'pending' CHECK (state IN ('pending', 'ready')),
    uploader_user_id text NOT NULL,
    uploader_name text NOT NULL,
    uploader_email text NOT NULL DEFAULT '',
    expires_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE feedback_form (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    presentation_id uuid NOT NULL REFERENCES presentation(id) ON DELETE CASCADE,
    scope_key text NOT NULL,
    evaluator_user_id text,
    evaluator_name text NOT NULL,
    evaluator_email text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted')),
    submitted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (presentation_id, scope_key)
);

CREATE TABLE feedback_answer (
    feedback_form_id uuid NOT NULL REFERENCES feedback_form(id) ON DELETE CASCADE,
    category_id uuid NOT NULL REFERENCES feedback_category(id) ON DELETE RESTRICT,
    value text NOT NULL DEFAULT '',
    revision bigint NOT NULL DEFAULT 1 CHECK (revision > 0),
    updated_by_user_id text NOT NULL,
    updated_by_name text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (feedback_form_id, category_id)
);

CREATE TABLE feedback_contributor (
    feedback_form_id uuid NOT NULL REFERENCES feedback_form(id) ON DELETE CASCADE,
    user_id text NOT NULL,
    name text NOT NULL,
    email text NOT NULL DEFAULT '',
    first_contributed_at timestamptz NOT NULL DEFAULT now(),
    last_contributed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (feedback_form_id, user_id)
);

CREATE TABLE feedback_presence (
    presentation_id uuid NOT NULL REFERENCES presentation(id) ON DELETE CASCADE,
    connection_id uuid NOT NULL,
    user_id text NOT NULL,
    name text NOT NULL,
    expires_at timestamptz NOT NULL,
    PRIMARY KEY (presentation_id, connection_id)
);

CREATE INDEX idx_feedback_category_phase ON feedback_category(course_phase_id, position);
CREATE INDEX idx_presentation_slot_phase_time ON presentation_slot(course_phase_id, start_time);
CREATE INDEX idx_presentation_phase ON presentation(course_phase_id);
CREATE INDEX idx_presentation_target ON presentation(course_phase_id, target_type, target_id);
CREATE INDEX idx_material_presentation_state ON presentation_material(presentation_id, state);
CREATE INDEX idx_material_uploader ON presentation_material(uploader_user_id);
CREATE INDEX idx_feedback_form_presentation_status ON feedback_form(presentation_id, status);
CREATE INDEX idx_feedback_form_evaluator ON feedback_form(evaluator_user_id);
CREATE INDEX idx_feedback_answer_category ON feedback_answer(category_id);
CREATE INDEX idx_presence_expiry ON feedback_presence(expires_at);

COMMIT;
