--
-- PostgreSQL database dump for Schema Copy Tests
--
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

-- Drop existing tables and types
DROP TABLE IF EXISTS public.assessment CASCADE;
DROP TABLE IF EXISTS public.evaluation CASCADE;
DROP TABLE IF EXISTS public.competency CASCADE;
DROP TABLE IF EXISTS public.category CASCADE;
DROP TABLE IF EXISTS public.course_phase_config CASCADE;
DROP TABLE IF EXISTS public.assessment_schema CASCADE;
DROP TABLE IF EXISTS public.schema_migrations CASCADE;
DROP TYPE IF EXISTS public.score_level;

-- Create types
CREATE TYPE public.score_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

SET default_table_access_method = HEAP;

-- Create tables
CREATE TABLE public.assessment_schema (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name text NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    source_phase_id uuid,
    CONSTRAINT assessment_schema_pkey PRIMARY KEY (id)
);

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

CREATE TABLE public.category (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    weight integer DEFAULT 1 NOT NULL,
    short_name character varying(10),
    assessment_schema_id uuid NOT NULL,
    CONSTRAINT category_pkey PRIMARY KEY (id),
    FOREIGN KEY (assessment_schema_id) REFERENCES assessment_schema (id) ON DELETE CASCADE
);

CREATE TABLE public.competency (
    id uuid NOT NULL,
    category_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    description_very_bad text NOT NULL DEFAULT '',
    description_bad text NOT NULL DEFAULT '',
    description_ok text NOT NULL DEFAULT '',
    description_good text NOT NULL DEFAULT '',
    description_very_good text NOT NULL DEFAULT '',
    weight integer DEFAULT 1 NOT NULL,
    short_name character varying(10),
    CONSTRAINT competency_pkey PRIMARY KEY (id),
    FOREIGN KEY (category_id) REFERENCES category (id) ON DELETE CASCADE
);

CREATE TABLE public.assessment (
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    assessed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text DEFAULT ''::text NOT NULL,
    author_id text DEFAULT ''::text NOT NULL,
    CONSTRAINT assessment_pkey PRIMARY KEY (id),
    FOREIGN KEY (competency_id) REFERENCES competency (id) ON DELETE CASCADE
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
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    comment text,
    evaluated_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text DEFAULT ''::text NOT NULL,
    CONSTRAINT evaluation_pkey PRIMARY KEY (id),
    FOREIGN KEY (competency_id) REFERENCES competency (id) ON DELETE CASCADE
);

CREATE TABLE public.competency_map (
    id uuid NOT NULL,
    from_competency_id uuid NOT NULL,
    to_competency_id uuid NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT competency_map_pkey PRIMARY KEY (id),
    FOREIGN KEY (from_competency_id) REFERENCES competency (id) ON DELETE CASCADE,
    FOREIGN KEY (to_competency_id) REFERENCES competency (id) ON DELETE CASCADE
);

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL,
    CONSTRAINT schema_migrations_pkey PRIMARY KEY (version)
);

-- Create view for category_course_phase
CREATE VIEW category_course_phase AS
SELECT c.id AS category_id,
       cpc.course_phase_id
FROM category c
INNER JOIN course_phase_config cpc ON c.assessment_schema_id = cpc.assessment_schema_id;

-- Insert test data

-- SCENARIO 1: Schema with no assessments in any phase (for TestUpdateCategory_NoAssessments)
-- This schema is only used in Phase 1 and has no assessments anywhere
INSERT INTO public.assessment_schema (id, name, description)
VALUES ('00000000-0000-0000-0000-000000000001', 'Test Schema 1 - No Assessments', 
        'Schema with no assessments for testing no-copy scenario');

INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES 
    ('20000000-0000-0000-0000-000000000001', 'Test Category 1', 'First test category', 1, 'TC1', 
     '00000000-0000-0000-0000-000000000001');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name, 
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES 
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Test Competency 1-1', 
     'First competency in category 1', 1, 'TC1-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

-- SCENARIO 2: Schema used in multiple phases, with assessments in "other" phase
-- (for TestUpdateCategory_WithAssessmentsInOtherPhase)
INSERT INTO public.assessment_schema (id, name, description, source_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', 'Test Schema 2 - With Other Phase Assessments', 
        'Schema used in multiple phases for testing copy scenario',
        '10000000-0000-0000-0000-000000000002'); -- Phase 2 is the owner

-- Phase 2a: Current phase (where we'll make the update)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002');

-- Phase 2b: Other phase (has assessments, should trigger copy)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000004');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES 
    ('20000000-0000-0000-0000-000000000002', 'Test Category 2', 'Second test category', 1, 'TC2', 
     '00000000-0000-0000-0000-000000000002'),
    ('20000000-0000-0000-0000-000000000003', 'Test Category 2-B', 'Another category in schema 2', 2, 'TC2-B', 
     '00000000-0000-0000-0000-000000000002');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name, 
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES 
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'Test Competency 2-1', 
     'First competency in category 2', 1, 'TC2-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'Test Competency 2-B-1', 
     'Competency in category 2-B', 1, 'TC2-B-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

-- Assessments in the "other" phase (Phase 2b/Phase 4) - this will trigger the copy
INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, author)
VALUES
    ('40000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001',
     '10000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000002',
     'good', 'Test Author');

-- Evaluation in the "other" phase (Phase 2b/Phase 4)
INSERT INTO public.evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, comment, author)
VALUES 
    ('45000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000002', 
     '10000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000003', 
     'very_good', 'Test evaluation in other phase', 'Test Evaluator');

-- SCENARIO 3: Schema with assessments only in the same phase (for TestUpdateCategory_WithAssessmentsInSamePhase)
INSERT INTO public.assessment_schema (id, name, description, source_phase_id)
VALUES ('00000000-0000-0000-0000-000000000003', 'Test Schema 3 - Same Phase Assessments', 
        'Schema with assessments only in same phase for testing no-copy scenario',
        '10000000-0000-0000-0000-000000000003');

INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES 
    ('20000000-0000-0000-0000-000000000004', 'Test Category 3', 'Third test category', 2, 'TC3', 
     '00000000-0000-0000-0000-000000000003');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name, 
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES 
    ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', 'Test Competency 3-1', 
     'Competency in category 3', 1, 'TC3-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

-- Assessments in the "same" phase (Phase 3) - should NOT trigger copy when updating from Phase 3
INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, author)
VALUES
    ('40000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000003',
     '10000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004',
     'ok', 'Test Author Same Phase');

INSERT INTO public.schema_migrations
VALUES (10, false);
