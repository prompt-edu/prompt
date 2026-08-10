--
-- PostgreSQL database dump
--
-- Feedback Items test data dump
-- This file contains test data for feedback items functionality

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = ON;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create ENUM types
CREATE TYPE public.feedback_type AS ENUM (
    'positive',
    'negative'
);

CREATE TYPE public.assessment_type AS ENUM (
    'self',
    'peer',
    'tutor',
    'assessment'
);

-- Create minimal tables needed for feedback items tests
CREATE TABLE public.course_phase (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    course_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    CONSTRAINT course_phase_pkey PRIMARY KEY (id)
);

CREATE TABLE public.course_participation (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    course_id uuid NOT NULL,
    user_email character varying NOT NULL,
    role character varying NOT NULL,
    CONSTRAINT course_participation_pkey PRIMARY KEY (id)
);

CREATE TABLE public.assessment_schema (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    description text,
    CONSTRAINT assessment_schema_pkey PRIMARY KEY (id)
);

CREATE TABLE public.course_phase_config (
    assessment_schema_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
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
    CONSTRAINT course_phase_config_pkey PRIMARY KEY (course_phase_id),
    CONSTRAINT course_phase_config_assessment_schema_id_fkey FOREIGN KEY (assessment_schema_id) REFERENCES public.assessment_schema(id) ON DELETE CASCADE
);

CREATE TABLE public.evaluation_completion (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    author_course_participation_id uuid NOT NULL,
    completed_at timestamp with time zone NOT NULL,
    completed boolean DEFAULT false NOT NULL,
    type public.assessment_type NOT NULL DEFAULT 'self',
    CONSTRAINT evaluation_completion_pkey PRIMARY KEY (id),
    CONSTRAINT evaluation_completion_unique_constraint UNIQUE (course_participation_id, course_phase_id, author_course_participation_id)
);

-- Main feedback items table
CREATE TABLE public.feedback_items (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    feedback_type feedback_type NOT NULL,
    feedback_text text NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    author_course_participation_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type public.assessment_type NOT NULL DEFAULT 'self'
);

-- Insert test data for course phases
INSERT INTO public.course_phase (id, name, course_id, start_date, end_date) VALUES
('24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'Test Phase 1', '12345678-1234-1234-1234-123456789abc', '2024-01-01', '2024-06-30'),
('34561b6b-3c3a-4bc6-ba42-69eeb1514da9', 'Test Phase 2', '12345678-1234-1234-1234-123456789abc', '2024-07-01', '2024-12-31');

-- Insert test data for course participations
INSERT INTO public.course_participation (id, course_id, user_email, role) VALUES
('ca42e447-60f9-4fe0-b297-2dae3f924fd7', '12345678-1234-1234-1234-123456789abc', 'student1@example.com', 'student'),
('da42e447-60f9-4fe0-b297-2dae3f924fd7', '12345678-1234-1234-1234-123456789abc', 'student2@example.com', 'student'),
('ea42e447-60f9-4fe0-b297-2dae3f924fd7', '12345678-1234-1234-1234-123456789abc', 'lecturer@example.com', 'lecturer');

-- Insert test data for assessment schemas
INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests');

-- Insert test course phase configurations with all evaluation types open
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, start,
                                        self_evaluation_enabled, self_evaluation_start, self_evaluation_deadline,
                                        peer_evaluation_enabled, peer_evaluation_start, peer_evaluation_deadline,
                                        tutor_evaluation_enabled, tutor_evaluation_start, tutor_evaluation_deadline) VALUES
('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2099-12-31 23:59:59+00', '2024-01-01 00:00:00+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00'),
('550e8400-e29b-41d4-a716-446655440000', '34561b6b-3c3a-4bc6-ba42-69eeb1514da9', '2099-12-31 23:59:59+00', '2024-01-01 00:00:00+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00');

-- Insert test data for feedback items
INSERT INTO public.feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id) VALUES
('11111111-1111-1111-1111-111111111111', 'positive', 'Great teamwork and communication skills!', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'da42e447-60f9-4fe0-b297-2dae3f924fd7'),
('22222222-2222-2222-2222-222222222222', 'negative', 'Need to improve time management', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ea42e447-60f9-4fe0-b297-2dae3f924fd7'),
('33333333-3333-3333-3333-333333333333', 'positive', 'Excellent problem-solving abilities', 'da42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7'),
('44444444-4444-4444-4444-444444444444', 'negative', 'Could be more active in discussions', 'da42e447-60f9-4fe0-b297-2dae3f924fd7', '34561b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ea42e447-60f9-4fe0-b297-2dae3f924fd7');

--
-- PostgreSQL database dump complete
--
