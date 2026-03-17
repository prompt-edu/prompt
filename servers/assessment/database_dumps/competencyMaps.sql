--
-- PostgreSQL database dump
--
-- Competency Maps test data dump
-- This file contains test data for competency mapping functionality

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

CREATE VIEW public.category_course_phase AS
SELECT
    id AS category_id,
    course_phase_id
FROM public.category;

CREATE TABLE public.competency (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying NOT NULL,
    description text,
    category_id uuid NOT NULL,
    short_name character varying,
    description_very_bad text,
    description_bad text,
    description_ok text,
    description_good text,
    description_very_good text,
    weight double precision DEFAULT 1.0 NOT NULL,
    CONSTRAINT competency_pkey PRIMARY KEY (id),
    CONSTRAINT competency_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE
);

CREATE TABLE public.competency_map (
    from_competency_id uuid NOT NULL,
    to_competency_id uuid NOT NULL,
    CONSTRAINT competency_map_pkey PRIMARY KEY (from_competency_id, to_competency_id),
    CONSTRAINT competency_map_from_competency_id_fkey FOREIGN KEY (from_competency_id) REFERENCES public.competency(id) ON DELETE CASCADE,
    CONSTRAINT competency_map_to_competency_id_fkey FOREIGN KEY (to_competency_id) REFERENCES public.competency(id) ON DELETE CASCADE
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

-- Insert test course phases
INSERT INTO course_phase (id, name, course_id, start_date, end_date) VALUES
    ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Test Phase 1', '12345678-1234-1234-1234-123456789012', '2024-01-01', '2024-12-31'),
    ('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'Test Phase 2', '12345678-1234-1234-1234-123456789012', '2024-01-01', '2024-12-31');

-- Insert test categories
INSERT INTO category (id, name, short_name, description, weight, course_phase_id) VALUES
    ('12345678-1234-1234-1234-123456789012', 'Technical Skills', 'TECH', 'Technical competency category', 1.0, '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
    ('22345678-1234-1234-1234-123456789012', 'Soft Skills', 'SOFT', 'Soft skills competency category', 1.5, '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
    ('32345678-1234-1234-1234-123456789012', 'Leadership', 'LEAD', 'Leadership competency category', 2.0, '4179d58a-d00d-4fa7-94a5-397bc69fab02');

-- Insert test competencies
INSERT INTO competency (id, name, short_name, description, category_id, weight, description_very_bad, description_bad, description_ok, description_good, description_very_good) VALUES
    -- Technical Skills
    ('c1234567-1234-1234-1234-123456789012', 'Programming Skills', 'PROG', 'Ability to write clean, efficient code', '12345678-1234-1234-1234-123456789012', 1.0, 'Cannot write basic code', 'Writes buggy code', 'Writes working code', 'Writes clean code', 'Writes excellent code'),
    ('c2234567-1234-1234-1234-123456789012', 'System Design', 'SYS', 'Ability to design scalable systems', '12345678-1234-1234-1234-123456789012', 1.2, 'No design skills', 'Poor design choices', 'Basic design skills', 'Good design skills', 'Excellent design skills'),
    ('c3234567-1234-1234-1234-123456789012', 'Testing', 'TEST', 'Ability to write comprehensive tests', '12345678-1234-1234-1234-123456789012', 0.8, 'No testing', 'Basic testing', 'Good testing', 'Comprehensive testing', 'Excellent testing practices'),
    
    -- Soft Skills
    ('c4234567-1234-1234-1234-123456789012', 'Communication', 'COMM', 'Ability to communicate effectively', '22345678-1234-1234-1234-123456789012', 1.0, 'Poor communication', 'Basic communication', 'Clear communication', 'Effective communication', 'Excellent communication'),
    ('c5234567-1234-1234-1234-123456789012', 'Teamwork', 'TEAM', 'Ability to work in teams', '22345678-1234-1234-1234-123456789012', 1.1, 'Poor teamwork', 'Basic teamwork', 'Good teamwork', 'Great teamwork', 'Exceptional teamwork'),
    
    -- Leadership
    ('c6234567-1234-1234-1234-123456789012', 'Project Management', 'PM', 'Ability to manage projects', '32345678-1234-1234-1234-123456789012', 1.5, 'No PM skills', 'Basic PM skills', 'Good PM skills', 'Strong PM skills', 'Excellent PM skills'),
    ('c7234567-1234-1234-1234-123456789012', 'Mentoring', 'MENTOR', 'Ability to mentor others', '32345678-1234-1234-1234-123456789012', 1.3, 'No mentoring', 'Basic mentoring', 'Good mentoring', 'Strong mentoring', 'Excellent mentoring');

-- Insert test course participations
INSERT INTO course_participation (id, course_id, student_id, role) VALUES
    ('01234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '51234567-1234-1234-1234-123456789012', 'student'),
    ('02234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '52234567-1234-1234-1234-123456789012', 'student'),
    ('03234567-1234-1234-1234-123456789012', '12345678-1234-1234-1234-123456789012', '53234567-1234-1234-1234-123456789012', 'student');

-- Insert test competency mappings
INSERT INTO competency_map (from_competency_id, to_competency_id) VALUES
    -- Programming skills map to system design (technical progression)
    ('c1234567-1234-1234-1234-123456789012', 'c2234567-1234-1234-1234-123456789012'),
    -- Programming skills map to testing (technical complementary)
    ('c1234567-1234-1234-1234-123456789012', 'c3234567-1234-1234-1234-123456789012'),
    -- Communication maps to teamwork (soft skills progression)
    ('c4234567-1234-1234-1234-123456789012', 'c5234567-1234-1234-1234-123456789012'),
    -- Teamwork maps to project management (leadership progression)
    ('c5234567-1234-1234-1234-123456789012', 'c6234567-1234-1234-1234-123456789012'),
    -- Project management maps to mentoring (leadership complementary)
    ('c6234567-1234-1234-1234-123456789012', 'c7234567-1234-1234-1234-123456789012'),
    -- Cross-category mapping: System design to project management
    ('c2234567-1234-1234-1234-123456789012', 'c6234567-1234-1234-1234-123456789012'),
    -- Add some mappings TO Programming Skills for testing
    ('c3234567-1234-1234-1234-123456789012', 'c1234567-1234-1234-1234-123456789012'), -- Testing to Programming
    ('c4234567-1234-1234-1234-123456789012', 'c1234567-1234-1234-1234-123456789012'); -- Communication to Programming

-- Insert test evaluations for mapped competencies
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, evaluated_at) VALUES
    -- Evaluations for Programming Skills (c1234567-1234-1234-1234-123456789012)
    ('e1234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', '2024-01-15 10:00:00+00'),
    ('e2234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'very_good', '02234567-1234-1234-1234-123456789012', '2024-01-15 11:00:00+00'),
    
    -- Evaluations for Communication (c4234567-1234-1234-1234-123456789012)
    ('e3234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c4234567-1234-1234-1234-123456789012', 'ok', '01234567-1234-1234-1234-123456789012', '2024-01-16 10:00:00+00'),
    ('e4234567-1234-1234-1234-123456789012', '03234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c4234567-1234-1234-1234-123456789012', 'good', '03234567-1234-1234-1234-123456789012', '2024-01-16 11:00:00+00'),
    
    -- Evaluations for System Design (c2234567-1234-1234-1234-123456789012) - mapped from Programming
    ('e5234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'bad', '01234567-1234-1234-1234-123456789012', '2024-01-17 10:00:00+00'),
    ('e6234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'good', '02234567-1234-1234-1234-123456789012', '2024-01-17 11:00:00+00'),
    
    -- Evaluations for Testing (c3234567-1234-1234-1234-123456789012) - also mapped from Programming
    ('e7234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c3234567-1234-1234-1234-123456789012', 'ok', '01234567-1234-1234-1234-123456789012', '2024-01-18 10:00:00+00'),
    ('e8234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c3234567-1234-1234-1234-123456789012', 'very_good', '02234567-1234-1234-1234-123456789012', '2024-01-18 11:00:00+00');
