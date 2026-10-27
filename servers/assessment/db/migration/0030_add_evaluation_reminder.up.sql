BEGIN;

CREATE TABLE IF NOT EXISTS evaluation_reminder
(
    course_phase_id uuid                     NOT NULL,
    evaluation_type assessment_type          NOT NULL,
    last_sent_at    timestamp with time zone NOT NULL,
    PRIMARY KEY (course_phase_id, evaluation_type)
);

COMMIT;
