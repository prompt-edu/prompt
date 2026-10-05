BEGIN;

ALTER TABLE course_phase_config
    ADD COLUMN IF NOT EXISTS independent_assessment_enabled BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS independent_assessment
(
    id                      uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    course_participation_id uuid        NOT NULL,
    course_phase_id         uuid        NOT NULL,
    competency_id           uuid        NOT NULL REFERENCES competency (id) ON DELETE CASCADE,
    score_level             score_level NOT NULL,
    author                  text        NOT NULL,
    author_id               text        NOT NULL,
    assessed_at             timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (course_participation_id, course_phase_id, competency_id, author_id)
);

COMMIT;
