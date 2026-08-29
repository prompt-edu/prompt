BEGIN;

CREATE TABLE interview_review (
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    score integer,
    interviewer varchar(255),
    interview_answers jsonb NOT NULL DEFAULT '[]'::jsonb,
    created_at timestamp with time zone DEFAULT now(),
    updated_at timestamp with time zone DEFAULT now(),
    PRIMARY KEY (course_phase_id, course_participation_id)
);

CREATE INDEX idx_interview_review_course_phase ON interview_review(course_phase_id);

COMMIT;
