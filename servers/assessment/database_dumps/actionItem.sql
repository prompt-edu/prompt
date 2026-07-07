--
-- PostgreSQL database dump
--
-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.13 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET row_security = off;

DROP TABLE IF EXISTS public.action_item CASCADE;
DROP TABLE IF EXISTS public.assessment_completion CASCADE;
DROP TABLE IF EXISTS public.course_phase_config CASCADE;
DROP TABLE IF EXISTS public.assessment_schema CASCADE;

--
-- Name: action_item; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.action_item (
    id uuid NOT NULL PRIMARY KEY,
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    action text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text NOT NULL
);

--
-- Data for Name: action_item; Type: TABLE DATA; Schema: public; Owner: -
--

--
-- Name: assessment_schema; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assessment_schema (
    id uuid PRIMARY KEY,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    source_phase_id uuid
);

--
-- Name: course_phase_config; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.course_phase_config (
    assessment_schema_id uuid NOT NULL,
    course_phase_id uuid PRIMARY KEY NOT NULL,
    deadline timestamp with time zone DEFAULT NULL,
    self_evaluation_enabled boolean NOT NULL DEFAULT false,
    self_evaluation_schema uuid,
    self_evaluation_deadline timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    peer_evaluation_enabled boolean NOT NULL DEFAULT false,
    peer_evaluation_schema uuid,
    peer_evaluation_deadline timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    start timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    self_evaluation_start timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    peer_evaluation_start timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
    tutor_evaluation_enabled boolean NOT NULL DEFAULT false,
    tutor_evaluation_start timestamp with time zone,
    tutor_evaluation_deadline timestamp with time zone,
    tutor_evaluation_schema uuid,
    evaluation_results_visible boolean NOT NULL DEFAULT true,
    grade_suggestion_visible boolean NOT NULL DEFAULT true,
    action_items_visible boolean NOT NULL DEFAULT true,
    results_released boolean NOT NULL DEFAULT false,
    grading_sheet_visible boolean NOT NULL DEFAULT false,
    FOREIGN KEY (assessment_schema_id) REFERENCES assessment_schema (id) ON DELETE CASCADE,
    FOREIGN KEY (self_evaluation_schema) REFERENCES assessment_schema (id) ON DELETE RESTRICT,
    FOREIGN KEY (peer_evaluation_schema) REFERENCES assessment_schema (id) ON DELETE RESTRICT,
    FOREIGN KEY (tutor_evaluation_schema) REFERENCES assessment_schema (id) ON DELETE RESTRICT
);

CREATE TABLE public.assessment_completion (
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    completed_at timestamp WITH time zone NOT NULL,
    author text NOT NULL,
    comment text DEFAULT '' NOT NULL,
    grade_suggestion numeric(2, 1) DEFAULT 4.0 NOT NULL,
    completed boolean DEFAULT false NOT NULL,
    CONSTRAINT assessment_completion_grade_suggestion_check CHECK (grade_suggestion >= 1.0 AND grade_suggestion <= 6.0)
);

--
-- Data for Name: assessment_schema; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests');

--
-- Data for Name: course_phase_config; Type: TABLE DATA; Schema: public; Owner: -
--

-- Config for phase where action items should be visible (results_released = true)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, true);

--
-- Data for Name: assessment_completion; Type: TABLE DATA; Schema: public; Owner: -
--

-- Completed assessment for the visible action items test
INSERT INTO public.assessment_completion (course_participation_id, course_phase_id, completed_at, author, comment, grade_suggestion, completed) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2025-01-15 10:00:00+00', 'test_author', 'Test completion', 4.5, true);

-- Config for phase where action items should not be visible (results_released = false)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, false);

-- Config for service tests (assessment is open, allows creating/editing action items)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, false);

-- Test data for visibility tests
INSERT INTO public.action_item (id, course_phase_id, course_participation_id, action, author) VALUES
    ('a1111111-1111-1111-1111-111111111111', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for visible scenario', 'tester'),
    ('a2222222-2222-2222-2222-222222222222', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for not visible scenario', 'tester');

--
-- PostgreSQL database dump complete
--
