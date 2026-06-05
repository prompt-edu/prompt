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

--
-- Data for Name: assessment_schema; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests'),
('550e8400-e29b-41d4-a716-446655440001', 'Self Evaluation Schema', 'This is the default self evaluation schema.'),
('550e8400-e29b-41d4-a716-446655440002', 'Peer Evaluation Schema', 'This is the default peer evaluation schema.'),
('550e8400-e29b-41d4-a716-446655440003', 'Tutor Evaluation Schema', 'This is the default tutor evaluation schema.');

--
-- Data for Name: course_phase_config; Type: TABLE DATA; Schema: public; Owner: -
--

-- Sample data can be inserted here if needed for tests
-- INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline) VALUES
-- ('550e8400-e29b-41d4-a716-446655440000', '123e4567-e89b-12d3-a456-426614174000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440002', '2025-12-31 23:59:59+00');

--
-- Additional tables for schema change validation tests
--

-- Create types
CREATE TYPE public.score_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

-- Competency tables
CREATE TABLE public.category (
    id uuid PRIMARY KEY,
    name varchar(255) NOT NULL,
    description text,
    weight integer DEFAULT 1 NOT NULL,
    short_name varchar(10),
    assessment_schema_id uuid NOT NULL,
    FOREIGN KEY (assessment_schema_id) REFERENCES assessment_schema (id) ON DELETE CASCADE
);

CREATE TABLE public.competency (
    id uuid PRIMARY KEY,
    category_id uuid NOT NULL,
    name varchar(255) NOT NULL,
    description text,
    description_very_bad text NOT NULL DEFAULT '',
    description_bad text NOT NULL DEFAULT '',
    description_ok text NOT NULL DEFAULT '',
    description_good text NOT NULL DEFAULT '',
    description_very_good text NOT NULL DEFAULT '',
    weight integer DEFAULT 1 NOT NULL,
    short_name varchar(10),
    FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE
);

-- Assessment and evaluation tables
CREATE TABLE public.assessment (
    id uuid PRIMARY KEY,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    assessed_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text DEFAULT ''::text NOT NULL,
    author_id text DEFAULT ''::text NOT NULL,
    FOREIGN KEY (competency_id) REFERENCES competency (id) ON DELETE RESTRICT
);

CREATE TABLE public.category_assessment (
    id uuid NOT NULL PRIMARY KEY,
    category_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    comment text DEFAULT '' NOT NULL,
    author text DEFAULT '' NOT NULL,
    author_id text DEFAULT '' NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    UNIQUE (category_id, course_phase_id, course_participation_id)
);

CREATE TABLE public.evaluation (
    id uuid PRIMARY KEY,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    comment text,
    evaluation_type varchar(50) NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (competency_id) REFERENCES competency (id) ON DELETE RESTRICT
);

--
-- PostgreSQL database dump complete
--
