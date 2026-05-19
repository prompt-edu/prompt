SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET row_security = off;

CREATE TYPE public.assessment_type AS ENUM ('self', 'peer', 'tutor', 'assessment');

CREATE TABLE public.assessment_schema (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    description text
);

CREATE TABLE public.course_phase_config (
    assessment_schema_id uuid NOT NULL,
    course_phase_id uuid PRIMARY KEY NOT NULL,
    deadline timestamp with time zone DEFAULT NULL,
    self_evaluation_enabled boolean NOT NULL DEFAULT false,
    self_evaluation_schema uuid NOT NULL,
    self_evaluation_deadline timestamp with time zone NOT NULL,
    peer_evaluation_enabled boolean NOT NULL DEFAULT false,
    peer_evaluation_schema uuid NOT NULL,
    peer_evaluation_deadline timestamp with time zone NOT NULL,
    start timestamp with time zone NOT NULL,
    self_evaluation_start timestamp with time zone NOT NULL,
    peer_evaluation_start timestamp with time zone NOT NULL,
    tutor_evaluation_enabled boolean NOT NULL DEFAULT false,
    tutor_evaluation_start timestamp with time zone NOT NULL,
    tutor_evaluation_deadline timestamp with time zone NOT NULL,
    tutor_evaluation_schema uuid NOT NULL,
    evaluation_results_visible boolean NOT NULL DEFAULT true,
    grade_suggestion_visible boolean NOT NULL DEFAULT true,
    action_items_visible boolean NOT NULL DEFAULT true,
    results_released boolean NOT NULL DEFAULT false,
    grading_sheet_visible boolean NOT NULL DEFAULT false
);

CREATE TABLE public.evaluation_completion (
    id uuid PRIMARY KEY,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    author_course_participation_id uuid NOT NULL,
    completed_at timestamp with time zone NOT NULL,
    completed boolean NOT NULL DEFAULT false,
    type public.assessment_type NOT NULL
);

INSERT INTO public.assessment_schema (id, name, description)
VALUES
  ('90000000-0000-0000-0000-000000000001', 'Reminder Schema', 'Schema for reminder recipient tests');

INSERT INTO public.course_phase_config (
    assessment_schema_id,
    course_phase_id,
    deadline,
    self_evaluation_enabled,
    self_evaluation_schema,
    self_evaluation_deadline,
    peer_evaluation_enabled,
    peer_evaluation_schema,
    peer_evaluation_deadline,
    start,
    self_evaluation_start,
    peer_evaluation_start,
    tutor_evaluation_enabled,
    tutor_evaluation_start,
    tutor_evaluation_deadline,
    tutor_evaluation_schema,
    evaluation_results_visible,
    grade_suggestion_visible,
    action_items_visible,
    results_released,
    grading_sheet_visible
)
VALUES
  (
    '90000000-0000-0000-0000-000000000001',
    '88888888-8888-8888-8888-888888888888',
    '2025-05-30T12:00:00Z',
    true,
    '90000000-0000-0000-0000-000000000001',
    '2025-05-30T12:00:00Z',
    true,
    '90000000-0000-0000-0000-000000000001',
    '2025-05-30T12:00:00Z',
    '2025-04-01T08:00:00Z',
    '2025-04-01T08:00:00Z',
    '2025-04-01T08:00:00Z',
    true,
    '2025-04-01T08:00:00Z',
    '2025-05-30T12:00:00Z',
    '90000000-0000-0000-0000-000000000001',
    true,
    true,
    true,
    false,
    false
  );
