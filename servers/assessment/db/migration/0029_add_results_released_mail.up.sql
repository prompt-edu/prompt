BEGIN;

CREATE TABLE IF NOT EXISTS results_released_mail
(
    course_phase_id         uuid                     NOT NULL,
    course_participation_id uuid                     NOT NULL,
    sent_at                 timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (course_phase_id, course_participation_id)
);

COMMIT;
