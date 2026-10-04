--
-- Minimal dataset for self-team allocation module tests
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

DROP TABLE IF EXISTS public.tutor CASCADE;
DROP TABLE IF EXISTS public.assignments CASCADE;
DROP TABLE IF EXISTS public.team CASCADE;
DROP TABLE IF EXISTS public.timeframe CASCADE;

CREATE TABLE public.team (
    id uuid PRIMARY KEY,
    name varchar(255) NOT NULL,
    course_phase_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name),
    CONSTRAINT team_id_course_phase_uk UNIQUE (id, course_phase_id)
);

CREATE TABLE public.assignments (
    id uuid PRIMARY KEY,
    course_participation_id uuid NOT NULL,
    team_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    student_first_name text NOT NULL,
    student_last_name text NOT NULL,
    FOREIGN KEY (team_id) REFERENCES public.team (id) ON DELETE CASCADE,
    FOREIGN KEY (team_id, course_phase_id) REFERENCES public.team (id, course_phase_id) ON DELETE CASCADE,
    CONSTRAINT assignments_participation_phase_uk UNIQUE (course_participation_id, course_phase_id)
);

CREATE TABLE public.timeframe (
    course_phase_id uuid PRIMARY KEY,
    starttime timestamptz NOT NULL,
    endtime timestamptz NOT NULL
);

CREATE TABLE public.tutor (
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    team_id uuid NOT NULL,
    PRIMARY KEY (course_phase_id, course_participation_id),
    FOREIGN KEY (team_id, course_phase_id) REFERENCES public.team (id, course_phase_id) ON DELETE CASCADE
);

INSERT INTO public.team (id, name, course_phase_id, created_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Alpha Team', '11111111-1111-1111-1111-111111111111', '2024-01-01 10:00:00'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Beta Team', '11111111-1111-1111-1111-111111111111', '2024-01-02 10:00:00'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Gamma Team', '22222222-2222-2222-2222-222222222222', '2024-01-03 10:00:00');

INSERT INTO public.assignments (
    id,
    course_participation_id,
    team_id,
    course_phase_id,
    created_at,
    updated_at,
    student_first_name,
    student_last_name
) VALUES
    ('11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaa1111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '2024-01-05 10:00:00', '2024-01-05 10:00:00', 'Alice', 'Anderson'),
    ('22222222-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbb1111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '2024-01-06 11:00:00', '2024-01-06 11:00:00', 'Bob', 'Brown'),
    ('33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'cccc1111-1111-1111-1111-111111111111', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '2024-01-07 12:00:00', '2024-01-07 12:00:00', 'Charlie', 'Clark'),
    ('44444444-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'dddd2222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', '2024-02-01 09:00:00', '2024-02-01 09:00:00', 'Dana', 'Davis');

INSERT INTO public.timeframe (course_phase_id, starttime, endtime) VALUES
    ('11111111-1111-1111-1111-111111111111', '2020-01-01 00:00:00+00', '2035-01-01 00:00:00+00'),
    ('22222222-2222-2222-2222-222222222222', '2099-01-01 00:00:00+00', '2100-01-01 00:00:00+00');

INSERT INTO public.tutor (course_phase_id, course_participation_id, first_name, last_name, team_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'eeee1111-1111-1111-1111-111111111111', 'Tara', 'Tutor', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');
