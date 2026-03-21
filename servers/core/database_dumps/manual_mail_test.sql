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
    PRIMARY KEY (course_participation_id, course_phase_id),
    CONSTRAINT fk_cpp_participation FOREIGN KEY (course_participation_id) REFERENCES public.course_participation(id),
    CONSTRAINT fk_cpp_phase FOREIGN KEY (course_phase_id) REFERENCES public.course_phase(id)
);

INSERT INTO public.course (id, name, start_date, end_date, restricted_data)
VALUES
  (
    '11111111-1111-1111-1111-111111111111',
    'Reminder Test Course',
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
VALUES
  (
    '22222222-2222-2222-2222-222222222222',
    'Assessment',
    false,
    'http://assessment.test/assessment/api',
    'Assessment phase'
  );

INSERT INTO public.course_phase (
    id,
    course_id,
    name,
    restricted_data,
    is_initial_phase,
    course_phase_type_id,
    student_readable_data
)
VALUES
  (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    'Assessment Phase',
    '{
      "mailingSettings": {
        "assessmentReminder": {
          "subject": "Reminder {{evaluationType}}",
          "content": "Hi {{firstName}}, finish {{evaluationType}} for {{coursePhaseName}} before {{evaluationDeadline}}.",
          "lastSentAtByType": {}
        }
      }
    }'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
  );

INSERT INTO public.student (
    id,
    first_name,
    last_name,
    email,
    matriculation_number,
    university_login,
    study_degree,
    current_semester,
    study_program
)
VALUES
  (
    '66666666-6666-6666-6666-666666666666',
    'Alice',
    'Anderson',
    'alice@example.com',
    '100001',
    'alice.a',
    'bachelor',
    3,
    'Informatics'
  ),
  (
    '77777777-7777-7777-7777-777777777777',
    'Bob',
    'Brown',
    'bob@example.com',
    '100002',
    'bob.b',
    'master',
    2,
    'Informatics'
  );

INSERT INTO public.course_participation (id, course_id, student_id)
VALUES
  (
    '44444444-4444-4444-4444-444444444444',
    '11111111-1111-1111-1111-111111111111',
    '66666666-6666-6666-6666-666666666666'
  ),
  (
    '55555555-5555-5555-5555-555555555555',
    '11111111-1111-1111-1111-111111111111',
    '77777777-7777-7777-7777-777777777777'
  );

INSERT INTO public.course_phase_participation (course_participation_id, course_phase_id)
VALUES
  (
    '44444444-4444-4444-4444-444444444444',
    '33333333-3333-3333-3333-333333333333'
  ),
  (
    '55555555-5555-5555-5555-555555555555',
    '33333333-3333-3333-3333-333333333333'
  );
