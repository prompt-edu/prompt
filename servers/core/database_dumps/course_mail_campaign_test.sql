SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET row_security = off;

CREATE TYPE public.study_degree AS ENUM ('bachelor', 'master');
CREATE TYPE public.pass_status AS ENUM ('passed', 'failed', 'not_assessed');
CREATE TYPE public.mail_campaign_status AS ENUM ('draft', 'sending', 'sent', 'partially_failed', 'failed');
CREATE TYPE public.mail_recipient_status AS ENUM ('pending', 'sent', 'failed');

CREATE TABLE public.course (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    restricted_data jsonb
);

CREATE TABLE public.course_phase_type (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    initial_phase boolean NOT NULL DEFAULT false,
    base_url text NOT NULL,
    description text
);

CREATE TABLE public.course_phase (
    id uuid PRIMARY KEY,
    course_id uuid NOT NULL,
    name text,
    restricted_data jsonb,
    is_initial_phase boolean NOT NULL DEFAULT false,
    course_phase_type_id uuid NOT NULL,
    student_readable_data jsonb,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES public.course(id),
    CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES public.course_phase_type(id)
);

CREATE TABLE public.student (
    id uuid PRIMARY KEY,
    first_name text,
    last_name text,
    email text,
    matriculation_number text,
    university_login text,
    study_degree public.study_degree NOT NULL,
    current_semester integer,
    study_program text
);

CREATE TABLE public.course_participation (
    id uuid PRIMARY KEY,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL,
    CONSTRAINT fk_course_participation_course FOREIGN KEY (course_id) REFERENCES public.course(id),
    CONSTRAINT fk_course_participation_student FOREIGN KEY (student_id) REFERENCES public.student(id)
);

CREATE TABLE public.course_phase_participation (
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    pass_status public.pass_status,
    PRIMARY KEY (course_participation_id, course_phase_id),
    CONSTRAINT fk_cpp_participation FOREIGN KEY (course_participation_id) REFERENCES public.course_participation(id),
    CONSTRAINT fk_cpp_phase FOREIGN KEY (course_phase_id) REFERENCES public.course_phase(id)
);

CREATE TABLE public.mail_campaign (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id uuid NOT NULL REFERENCES public.course(id) ON DELETE CASCADE,
    name text NOT NULL,
    subject text NOT NULL DEFAULT '',
    body text NOT NULL DEFAULT '',
    target_course_phase_id uuid REFERENCES public.course_phase(id) ON DELETE SET NULL,
    target_pass_statuses text[] NOT NULL DEFAULT '{}',
    reply_to_override jsonb,
    cc_override jsonb,
    bcc_override jsonb,
    status public.mail_campaign_status NOT NULL DEFAULT 'draft',
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by_id text NOT NULL,
    created_by_email text NOT NULL DEFAULT '',
    created_by_name text NOT NULL DEFAULT '',
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_id text NOT NULL DEFAULT '',
    updated_by_email text NOT NULL DEFAULT '',
    updated_by_name text NOT NULL DEFAULT '',
    sent_at timestamptz,
    sent_by_id text,
    sent_by_email text,
    sent_by_name text
);

CREATE TABLE public.mail_campaign_recipient (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id uuid NOT NULL REFERENCES public.mail_campaign(id) ON DELETE CASCADE,
    course_participation_id uuid NOT NULL,
    email text NOT NULL,
    status public.mail_recipient_status NOT NULL DEFAULT 'pending',
    error_message text NOT NULL DEFAULT '',
    sent_at timestamptz,
    UNIQUE (campaign_id, course_participation_id)
);

INSERT INTO public.course (id, name, start_date, end_date, restricted_data)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Campaign Test Course',
    '2025-04-01',
    '2025-09-30',
    '{
      "mailingSettings": {
        "replyToEmail": "replyto@example.com",
        "replyToName": "Course Team",
        "ccAddresses": [],
        "bccAddresses": []
      }
    }'::jsonb
);

INSERT INTO public.course_phase_type (id, name, initial_phase, base_url, description)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'Assessment',
    false,
    'http://assessment.test/assessment/api',
    'Assessment phase'
);

-- A separate course + phase used to verify a campaign cannot target a phase in
-- an unrelated course.
INSERT INTO public.course (id, name, start_date, end_date, restricted_data)
VALUES (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'Other Course',
    '2025-04-01',
    '2025-09-30',
    '{}'::jsonb
);

INSERT INTO public.course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
VALUES (
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'Other Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
);

INSERT INTO public.course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
VALUES
  (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    'Assessment Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
  ),
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '11111111-1111-1111-1111-111111111111',
    'Empty Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
  );

INSERT INTO public.student (id, first_name, last_name, email, matriculation_number, university_login, study_degree, current_semester, study_program)
VALUES
  ('66666666-6666-6666-6666-666666666666', 'Alice', 'Anderson', 'alice@example.com', '00100001', 'ab12cde', 'bachelor', 3, 'Informatics'),
  ('77777777-7777-7777-7777-777777777777', 'Bob',   'Brown',    'bob@example.com',   '00100002', 'cd34efg', 'master',   2, 'Informatics'),
  ('88888888-8888-8888-8888-888888888888', 'Carol', 'Clark',    'carol@example.com', '00100003', 'ef56ghi', 'bachelor', 1, 'Informatics'),
  ('99999999-9999-9999-9999-999999999999', 'Dan',   'Doe',      NULL,                '00100004', 'gh78ijk', 'master',   4, 'Informatics');

INSERT INTO public.course_participation (id, course_id, student_id)
VALUES
  ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', '66666666-6666-6666-6666-666666666666'),
  ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', '77777777-7777-7777-7777-777777777777'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888888'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '99999999-9999-9999-9999-999999999999');

INSERT INTO public.course_phase_participation (course_participation_id, course_phase_id, pass_status)
VALUES
  ('44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333', 'passed'),
  ('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', 'failed'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '33333333-3333-3333-3333-333333333333', 'not_assessed'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', 'passed');
