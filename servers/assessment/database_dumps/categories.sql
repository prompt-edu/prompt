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

SELECT pg_catalog.set_config ('search_path', 'public', false);

SET check_function_bodies = false;

SET xmloption = content;

SET client_min_messages = warning;

SET row_security = off;

ALTER TABLE IF EXISTS ONLY public.competency DROP CONSTRAINT IF EXISTS competency_category_id_fkey;

ALTER TABLE IF EXISTS ONLY public.assessment DROP CONSTRAINT IF EXISTS assessment_competency_id_fkey;

DROP INDEX IF EXISTS public.idx_assessment_completion_participation_phase;

ALTER TABLE IF EXISTS ONLY public.schema_migrations DROP CONSTRAINT IF EXISTS schema_migrations_pkey;

ALTER TABLE IF EXISTS ONLY public.competency DROP CONSTRAINT IF EXISTS competency_pkey;

ALTER TABLE IF EXISTS ONLY public.category DROP CONSTRAINT IF EXISTS category_pkey;

ALTER TABLE IF EXISTS ONLY public.assessment DROP CONSTRAINT IF EXISTS assessment_pkey;

ALTER TABLE IF EXISTS ONLY public.assessment DROP CONSTRAINT IF EXISTS assessment_course_participation_id_course_phase_id_competen_key;

ALTER TABLE IF EXISTS ONLY public.assessment_completion DROP CONSTRAINT IF EXISTS assessment_completion_pkey;

DROP TABLE IF EXISTS public.schema_migrations;

DROP VIEW IF EXISTS public.completed_score_levels;

DROP VIEW IF EXISTS public.weighted_participant_scores;

DROP TABLE IF EXISTS public.competency;

DROP TABLE IF EXISTS public.category;

DROP TABLE IF EXISTS public.assessment_completion;

DROP TABLE IF EXISTS public.assessment;

DROP TYPE IF EXISTS public.score_level;

DROP TYPE IF EXISTS public.assessment_type;

CREATE TYPE public.score_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

CREATE TYPE public.assessment_type AS ENUM (
    'self',
    'peer',
    'tutor',
    'assessment'
);

SET default_table_access_method = HEAP;

CREATE TABLE public.category (
    id uuid NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    weight integer DEFAULT 1 NOT NULL,
    short_name character varying(10),
    assessment_schema_id uuid NOT NULL
);

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
    short_name character varying(10)
);

CREATE TABLE public.assessment (
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    assessed_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text DEFAULT ''::text NOT NULL,
    author_id text DEFAULT ''::text NOT NULL
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
    author text DEFAULT ''::text NOT NULL
);

CREATE TABLE public.schema_migrations (version bigint NOT NULL, dirty boolean NOT NULL);

-- Insert the default assessment schema
INSERT INTO public.assessment_schema (id, name, description)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Intro Course Assessment Schema', 'This is the default assessment schema.');

-- Insert the default tutor evaluation schema
INSERT INTO public.assessment_schema (id, name, description)
VALUES ('d5e6f7a8-b9c0-1234-5678-90abcdef1234', 'Tutor Evaluation Schema', 'This is the default tutor evaluation schema.');

-- Insert some sample course_phase_config records
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, tutor_evaluation_enabled, tutor_evaluation_start, tutor_evaluation_deadline, tutor_evaluation_schema, evaluation_results_visible, grade_suggestion_visible, action_items_visible, results_released, grading_sheet_visible)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', true, '2024-01-01 00:00:00+00', '2025-12-31 23:59:59+00', 'd5e6f7a8-b9c0-1234-5678-90abcdef1234', true, true, true, true, true);

INSERT INTO public.category
VALUES (
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'Version Control',
        '',
        1,
        'Git',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.category
VALUES (
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'User Interface',
        '',
        1,
        'UI',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.category
VALUES (
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Fundamentals in Software Engineering',
        '',
        1,
        'SE',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.competency
VALUES (
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'GitLab Project Management',
        'Understand GitLab’s collaboration features, including issue tracking, merge request workflows, and navigation of the issue board.',
        'Can create and manage repositories and basic issues.',
        'Can manage issues, labels, and milestones; creates and reviews MRs. ',
        'Can manage issue workflows, works together with Tutor on Reviews, and enforces best practices.',
        'Defines and optimizes issue tracking and MR processes for efficient collaboration. ',
        'Describes the GitLab project management features and how they can be used to manage the intro course app project.',
        1,
        'GitLab PM'
    );

INSERT INTO public.competency
VALUES (
        '0431b736-7fab-4333-b83e-fe3927f32475',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Requirements & Backlog Management',
        'Understand and apply structured documentation techniques such as product backlogs and requirement artifacts. Use common principles as Abbots Technique or FURPS+',
        'Writes simple user stories but lacks clarity and adherence to best practices.',
        'Creates well-structured user stories using INVEST criteria and documents them systematically.',
        'Manages a product backlog effectively, refining requirements iteratively.',
        'Ensures requirement traceability, prioritization, and alignment with long-term goals.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency
VALUES (
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Architecture and System Design',
        'Understand and apply architectural principles, including top-level architecture, subsystem decomposition, and deployment diagrams.',
        'Recognizes key architectural elements and struggles with formal documentation.',
        'Creates high-level system architecture with some guidance.',
        'Develops and evaluates scalable architectures tailored to project requirements.',
        'Designs architecture with clear subsystem decomposition, using SDD and API specifications.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency
VALUES (
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Software Engineering & Modeling',
        'Recognize why modeling is important in software engineering and how it contributes to structured development.',
        'Has a basic understanding of software engineering and modeling but struggles to phrase their significance.',
        'Understands the role of modeling in software engineering and can explain why it is useful.',
        'Applies model-based approaches to structure software development and can justify their importance.',
        'Critically evaluates modeling techniques and adapts them to project-specific needs.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency
VALUES (
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'Low-Fidelity Mockups / Prototyping',
        'Learn UI wireframing and prototyping using pen & paper (not tool-bound. Create low-fidelity mockups for the intro course app and the first SwiftUI non-functional requirement (NFR). ',
        'Sketches basic UI ideas but lacks structure and clarity.',
        'Creates structured wireframes with a focus on layout and usability.',
        'Designs clear and functional low-fidelity prototypes, considering user flow and NFRs.',
        'Rapidly iterates on mockups, ensuring usability and alignment with design principles.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency
VALUES (
        '31aea83e-407b-4428-a5da-b25dd562832b',
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'Apple''s Human Interface Guidelines',
        'Understand the Human Interface Guidelines (HIG) and how SwiftUI supports cross-platform compliance, especially for non-functional requirements of the intro course app.',
        'Recognizes that HIG exists and affects app design.',
        'Understands core principles of HIG and applies basic guidelines in SwiftUI.',
        'Integrates HIG principles effectively, ensuring usability and consistency.',
        'Applies HIG to enhance usability and consistency across platforms, aligning with Apple’s best practices.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency
VALUES (
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'Git Basics',
        'Learn and apply Git workflows, key commands, and best practices, including branching models, commit conventions, and the differences between CLI and GUI tools.',
        'Can initialize a repository and commit changes.',
        'Can create branches, merge changes, and resolve basic conflicts.',
        'Follows a structured Git workflow with a clear branching model, adopts meaningful commit messages.',
        'Optimizes Git usage, enforces best practices, mentors others in efficient Git workflows.',
        'Very Good',
        2,
        NULL
    );

INSERT INTO public.schema_migrations
VALUES (8, false);

-- Add foreign key constraint for category assessment_schema_id
ALTER TABLE ONLY public.category
ADD CONSTRAINT category_assessment_schema_id_fkey FOREIGN KEY (assessment_schema_id) REFERENCES public.assessment_schema (id) ON DELETE SET NULL;

-- Create the view
CREATE VIEW category_course_phase AS
SELECT c.id AS category_id,
       cpc.course_phase_id
FROM category c
         INNER JOIN course_phase_config cpc
                    ON c.assessment_schema_id = cpc.assessment_schema_id;

ALTER TABLE ONLY public.category
ADD CONSTRAINT category_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.competency
ADD CONSTRAINT competency_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.assessment
ADD CONSTRAINT assessment_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.evaluation
ADD CONSTRAINT evaluation_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.schema_migrations
ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);

ALTER TABLE ONLY public.competency
ADD CONSTRAINT competency_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category (id) ON DELETE CASCADE;

ALTER TABLE ONLY public.assessment
ADD CONSTRAINT assessment_competency_id_fkey FOREIGN KEY (competency_id) REFERENCES public.competency (id) ON DELETE CASCADE;

ALTER TABLE ONLY public.evaluation
ADD CONSTRAINT evaluation_competency_id_fkey FOREIGN KEY (competency_id) REFERENCES public.competency (id) ON DELETE CASCADE;

--
-- Competency Map Table (added for GetCategoriesWithCompetencies support)
--

CREATE TABLE IF NOT EXISTS public.competency_map (
    from_competency_id uuid NOT NULL,
    to_competency_id uuid NOT NULL,
    CONSTRAINT competency_map_pkey PRIMARY KEY (from_competency_id, to_competency_id),
    CONSTRAINT competency_map_from_competency_id_fkey FOREIGN KEY (from_competency_id) REFERENCES public.competency(id) ON DELETE CASCADE,
    CONSTRAINT competency_map_to_competency_id_fkey FOREIGN KEY (to_competency_id) REFERENCES public.competency(id) ON DELETE CASCADE
);

-- Sample competency map test data
-- Map some competencies to demonstrate the functionality
INSERT INTO competency_map (from_competency_id, to_competency_id)
SELECT c1.id, c2.id
FROM competency c1, competency c2
WHERE c1.id != c2.id
AND c1.category_id = c2.category_id
LIMIT 5;