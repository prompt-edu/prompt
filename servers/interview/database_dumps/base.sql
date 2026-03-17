--
-- Minimal dataset for interview module tests
--
SET
    statement_timeout = 0;

SET
    lock_timeout = 0;

SET
    client_encoding = 'UTF8';

SET
    standard_conforming_strings = ON;

SET
    check_function_bodies = false;

SET
    xmloption = content;

SET
    client_min_messages = warning;

SET
    row_security = off;

DROP TABLE IF EXISTS public.interview_assignment CASCADE;

DROP TABLE IF EXISTS public.interview_slot CASCADE;

CREATE TABLE
    public.interview_slot (
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
    public.interview_assignment (
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

-- Seed data for testing
-- Active course phase with interview slots
INSERT INTO
    public.interview_slot (
        id,
        course_phase_id,
        start_time,
        end_time,
        location,
        capacity,
        created_at,
        updated_at
    )
VALUES
    (
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        '11111111-1111-1111-1111-111111111111',
        '2026-03-01 09:00:00+00',
        '2026-03-01 10:00:00+00',
        'Room 101',
        2,
        '2026-01-01 10:00:00+00',
        '2026-01-01 10:00:00+00'
    ),
    (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '11111111-1111-1111-1111-111111111111',
        '2026-03-01 10:00:00+00',
        '2026-03-01 11:00:00+00',
        'Room 102',
        3,
        '2026-01-02 10:00:00+00',
        '2026-01-02 10:00:00+00'
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        '11111111-1111-1111-1111-111111111111',
        '2026-03-01 11:00:00+00',
        '2026-03-01 12:00:00+00',
        'Room 103',
        1,
        '2026-01-03 10:00:00+00',
        '2026-01-03 10:00:00+00'
    ),
    (
        'dddddddd-dddd-dddd-dddd-dddddddddddd',
        '22222222-2222-2222-2222-222222222222',
        '2026-04-01 09:00:00+00',
        '2026-04-01 10:00:00+00',
        'Room 201',
        2,
        '2026-01-04 10:00:00+00',
        '2026-01-04 10:00:00+00'
    );

-- Some existing assignments
INSERT INTO
    public.interview_assignment (
        id,
        interview_slot_id,
        course_participation_id,
        assigned_at
    )
VALUES
    (
        '11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'aaaa1111-1111-1111-1111-111111111111',
        '2026-01-05 10:00:00+00'
    ),
    (
        '22222222-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'bbbb1111-1111-1111-1111-111111111111',
        '2026-01-06 11:00:00+00'
    ),
    (
        '33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        'cccc1111-1111-1111-1111-111111111111',
        '2026-01-07 12:00:00+00'
    ),
    (
        '44444444-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'ffffffff-ffff-ffff-ffff-ffffffffffff',
        '2026-01-08 13:00:00+00'
    );