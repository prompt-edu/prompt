BEGIN;

CREATE TABLE
    interview_slot (
        id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
        course_phase_id uuid NOT NULL,
        start_time timestamp
        with
            time zone NOT NULL,
            end_time timestamp
        with
            time zone NOT NULL,
            location varchar(255),
            capacity integer NOT NULL DEFAULT 1,
            created_at timestamp
        with
            time zone DEFAULT now (),
            updated_at timestamp
        with
            time zone DEFAULT now ()
    );

CREATE TABLE
    interview_assignment (
        id uuid PRIMARY KEY DEFAULT gen_random_uuid (),
        interview_slot_id uuid NOT NULL REFERENCES interview_slot (id) ON DELETE CASCADE,
        course_participation_id uuid NOT NULL,
        assigned_at timestamp
        with
            time zone DEFAULT now (),
            UNIQUE (course_participation_id, interview_slot_id)
    );

CREATE INDEX idx_interview_slot_course_phase ON interview_slot (course_phase_id);

CREATE INDEX idx_interview_assignment_slot ON interview_assignment (interview_slot_id);

CREATE INDEX idx_interview_assignment_participation ON interview_assignment (course_participation_id);

COMMIT;