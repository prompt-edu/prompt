--
-- PostgreSQL database dump
--
-- Evaluations test data dump
-- This file contains test data for evaluations functionality

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
CREATE TYPE public.score_level AS ENUM (
    'very_bad',
    'bad',
    'ok',
    'good',
    'very_good'
);

CREATE TYPE public.assessment_type AS ENUM (
    'self',
    'peer',
    'tutor',
    'assessment'
);

-- Create tables
CREATE TABLE public.course_phase (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    course_id uuid NOT NULL,
    start_date date NOT NULL,
    end_date date NOT NULL,
    CONSTRAINT course_phase_pkey PRIMARY KEY (id)
);

CREATE TABLE public.category (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    short_name character varying NOT NULL,
    description text,
    weight double precision DEFAULT 1.0 NOT NULL,
    course_phase_id uuid NOT NULL,
    CONSTRAINT category_pkey PRIMARY KEY (id),
    CONSTRAINT category_course_phase_id_fkey FOREIGN KEY (course_phase_id) REFERENCES public.course_phase(id) ON DELETE CASCADE
);

CREATE TABLE public.competency (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description text,
    category_id uuid NOT NULL,
    CONSTRAINT competency_pkey PRIMARY KEY (id),
    CONSTRAINT competency_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE
);

CREATE TABLE public.course_participation (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL,
    role character varying DEFAULT 'student' NOT NULL,
    CONSTRAINT course_participation_pkey PRIMARY KEY (id)
);

CREATE TABLE public.evaluation (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    author_course_participation_id uuid NOT NULL,
    evaluated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    type public.assessment_type NOT NULL DEFAULT 'self',
    CONSTRAINT evaluation_pkey PRIMARY KEY (id),
    CONSTRAINT evaluation_course_participation_id_fkey FOREIGN KEY (course_participation_id) REFERENCES public.course_participation(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_course_phase_id_fkey FOREIGN KEY (course_phase_id) REFERENCES public.course_phase(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_competency_id_fkey FOREIGN KEY (competency_id) REFERENCES public.competency(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_author_course_participation_id_fkey FOREIGN KEY (author_course_participation_id) REFERENCES public.course_participation(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_unique_constraint UNIQUE (course_participation_id, course_phase_id, competency_id, author_course_participation_id)
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
    CONSTRAINT evaluation_completion_course_participation_id_fkey FOREIGN KEY (course_participation_id) REFERENCES public.course_participation(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_completion_course_phase_id_fkey FOREIGN KEY (course_phase_id) REFERENCES public.course_phase(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_completion_author_course_participation_id_fkey FOREIGN KEY (author_course_participation_id) REFERENCES public.course_participation(id) ON DELETE CASCADE,
    CONSTRAINT evaluation_completion_unique_constraint UNIQUE (course_participation_id, course_phase_id, author_course_participation_id)
);

CREATE TABLE public.assessment_schema (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
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
    CONSTRAINT course_phase_config_assessment_schema_id_fkey FOREIGN KEY (assessment_schema_id) REFERENCES public.assessment_schema(id) ON DELETE CASCADE,
    CONSTRAINT course_phase_config_self_evaluation_schema_fkey FOREIGN KEY (self_evaluation_schema) REFERENCES public.assessment_schema(id) ON DELETE RESTRICT,
    CONSTRAINT course_phase_config_peer_evaluation_schema_fkey FOREIGN KEY (peer_evaluation_schema) REFERENCES public.assessment_schema(id) ON DELETE RESTRICT
);

-- Insert test course phases
INSERT INTO course_phase (id, name, course_id, start_date, end_date) VALUES
    ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Test Phase 1', '12345678-1234-1234-1234-123456789012', '2024-01-01', '2024-12-31'),
    ('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'Test Phase 2', '12345678-1234-1234-1234-123456789012', '2024-01-01', '2024-12-31');

-- Insert test categories
INSERT INTO category (id, name, short_name, description, weight, course_phase_id) VALUES
    ('12345678-1234-1234-1234-123456789012', 'Test Category 1', 'TC1', 'First test category', 1.0, '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
    ('22345678-1234-1234-1234-123456789012', 'Test Category 2', 'TC2', 'Second test category', 2.0, '4179d58a-d00d-4fa7-94a5-397bc69fab02');

-- Insert test competencies
INSERT INTO competency (id, name, description, category_id) VALUES
    ('c1234567-1234-1234-1234-123456789012', 'Test Competency 1', 'First test competency', '12345678-1234-1234-1234-123456789012'),
    ('c2234567-1234-1234-1234-123456789012', 'Test Competency 2', 'Second test competency', '12345678-1234-1234-1234-123456789012'),
    ('c3234567-1234-1234-1234-123456789012', 'Test Competency 3', 'Third test competency', '12345678-1234-1234-1234-123456789012');

-- Insert test course participations
INSERT INTO course_participation (id, course_id, student_id, role) VALUES
    ('01234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '51234567-1234-1234-1234-123456789012', 'student'),
    ('02234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '52234567-1234-1234-1234-123456789012', 'student'),
    ('03234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '53234567-1234-1234-1234-123456789012', 'student'),
    ('03711111-1111-1111-1111-111111111111', '12345678-1234-1234-1234-123456789012', '03711111-1111-1111-1111-111111111111', 'student');

-- Insert test evaluations
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type, evaluated_at) VALUES
    -- Self evaluations
    ('e1234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', 'self', '2024-01-15 10:00:00+00'),
    ('e2234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'very_good', '01234567-1234-1234-1234-123456789012', 'self', '2024-01-15 11:00:00+00'),
    
    -- Peer evaluations
    ('e3234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'ok', '02234567-1234-1234-1234-123456789012', 'peer', '2024-01-16 10:00:00+00'),
    ('e4234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', 'peer', '2024-01-16 11:00:00+00'),
    
    -- More test data for different scenarios
    ('e5234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'bad', '02234567-1234-1234-1234-123456789012', 'self', '2024-01-17 10:00:00+00'),
    ('e6234567-1234-1234-1234-123456789012', '03234567-1234-1234-1234-123456789012', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'c3234567-1234-1234-1234-123456789012', 'very_bad', '03234567-1234-1234-1234-123456789012', 'self', '2024-01-18 10:00:00+00');

-- Insert test assessment schemas
INSERT INTO assessment_schema (id, name, description) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests'),
    ('550e8400-e29b-41d4-a716-446655440001', 'Self Evaluation Schema', 'This is the default self evaluation schema.'),
    ('550e8400-e29b-41d4-a716-446655440002', 'Peer Evaluation Schema', 'This is the default peer evaluation schema.');

-- Insert test course phase configurations  
INSERT INTO course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440002', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00'),
    ('550e8400-e29b-41d4-a716-446655440000', '5179d58a-d00d-4fa7-94a5-397bc69fab03', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440002', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00');
