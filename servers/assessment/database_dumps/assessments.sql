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

CREATE TYPE public.score_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

SET default_table_access_method = HEAP;

CREATE TABLE public.assessment (
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    competency_id uuid NOT NULL,
    score_level public.score_level NOT NULL,
    assessed_at timestamp WITH time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
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

CREATE TABLE public.assessment_completion (
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    completed_at timestamp WITH time zone NOT NULL,
    author text NOT NULL,
    comment text DEFAULT '' NOT NULL,
    grade_suggestion numeric(2, 1) DEFAULT 4.0 NOT NULL,
    completed boolean DEFAULT false NOT NULL
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
    short_name character varying(10)
);

CREATE VIEW public.weighted_participant_scores AS WITH score_values AS (
    SELECT a.course_phase_id,
        a.course_participation_id,
        CASE
            a.score_level
            WHEN 'very_bad' THEN 5::integer
            WHEN 'bad' THEN 4::integer
            WHEN 'ok' THEN 3::integer
            WHEN 'good' THEN 2::integer
            WHEN 'very_good' THEN 1::integer
            ELSE NULL::integer
        END AS score_numeric,
        comp.weight AS competency_weight,
        cat.weight AS category_weight
    FROM (
            (
                public.assessment a
                JOIN public.competency comp ON ((a.competency_id = comp.id))
            )
            JOIN public.category cat ON ((comp.category_id = cat.id))
        )
),
weighted_scores AS (
    SELECT score_values.course_phase_id,
        score_values.course_participation_id,
        (
            sum(
                (
                    (
                        score_values.score_numeric * score_values.competency_weight
                    ) * score_values.category_weight
                )
            )
        )::double precision AS weighted_score_sum,
        (
            sum(
                (
                    score_values.competency_weight * score_values.category_weight
                )
            )
        )::double precision AS total_weight
    FROM score_values
    GROUP BY score_values.course_phase_id,
        score_values.course_participation_id
)
SELECT weighted_scores.course_phase_id,
    weighted_scores.course_participation_id,
    round(
        (
            (
                weighted_scores.weighted_score_sum / weighted_scores.total_weight
            )
        )::numeric,
        2
    ) AS score_numeric,
    CASE
        WHEN round(
            (
                (
                    weighted_scores.weighted_score_sum / weighted_scores.total_weight
                )
            )::numeric,
            2
        ) < 1.25 THEN 'very_good'
        WHEN round(
            (
                (
                    weighted_scores.weighted_score_sum / weighted_scores.total_weight
                )
            )::numeric,
            2
        ) < 1.75 THEN 'good'
        WHEN round(
            (
                (
                    weighted_scores.weighted_score_sum / weighted_scores.total_weight
                )
            )::numeric,
            2
        ) < 2.5 THEN 'ok'
        WHEN round(
            (
                (
                    weighted_scores.weighted_score_sum / weighted_scores.total_weight
                )
            )::numeric,
            2
        ) < 3.25 THEN 'bad'
        ELSE 'very_bad'
    END AS score_level
FROM weighted_scores;

CREATE VIEW public.completed_score_levels AS
SELECT slc.course_phase_id,
    slc.course_participation_id,
    slc.score_level
FROM public.weighted_participant_scores slc
WHERE (
        EXISTS (
            SELECT 1
            FROM public.assessment_completion ac
            WHERE (
                    (
                        ac.course_participation_id = slc.course_participation_id
                    )
                    AND (ac.course_phase_id = slc.course_phase_id)
                )
        )
    );

CREATE TABLE public.action_item (
    id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    action text NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    author text NOT NULL
);

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);

INSERT INTO public.assessment
VALUES (
        '1950fdb7-d736-4fe6-81f9-b8b1cf7c85df',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 20:16:50.331851+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '9f1a5b41-f1de-43a8-853b-9b5f944742d4',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'bad',
        '2025-05-11 21:17:43.757443+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'f9ed6b56-a323-4b8d-9e7c-a921eca13fb2',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 21:17:55.486197+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '4422e68c-d042-43f0-a725-a2b2b7e387d8',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'ok',
        '2025-05-11 21:41:57.148125+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'dc760efd-9210-40d3-86b7-7364acc5b185',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:42:03.794931+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'f968004e-a653-4063-a6e1-3fa5941a1ab0',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:42:11.555711+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e77d4512-1060-4a05-90e3-83cb85f06d4e',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'bad',
        '2025-05-11 21:42:25.550212+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '15d0fcc3-044f-4a4e-9ba6-4ffa169b2dff',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:42:35.764131+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '5f14cb4c-bacf-42b2-8442-a87fe61b83b6',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:42:42.223527+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e209b2e8-3b7d-4c83-aba3-87427fcee39b',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:43:24.702805+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'bd3c570b-db28-456d-ae30-882e01b243cc',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 21:46:32.204333+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '387310ca-1aba-4e8d-a0af-f04a251d251a',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:46:32.599828+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e91ed28f-bac0-4468-afa4-c10215bad335',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 21:47:33.825038+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '40fbba73-b816-4495-9fad-586a860bd4a4',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-11 21:47:34.273177+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '395a0ba0-9588-4133-8e0f-35424cefacd9',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'bad',
        '2025-05-12 13:42:39.845865+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '59417303-3d6f-4803-94d6-373cb6daafa8',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:28:51.732717+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b6b759c3-976e-4224-bb19-7a11fee6deed',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'ok',
        '2025-05-13 16:27:17.8223+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '4f2db1a3-3aee-4807-ab01-e19de3a541e4',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'ok',
        '2025-05-13 16:27:19.05869+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '386b0d55-0f5e-4157-8497-7d4e75316cf9',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:27:20.274728+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b3fbbe08-eed2-44bd-b55a-48e59f3ec0ba',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:27:22.189391+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'af4e9948-2387-4112-8ca1-9d8bf6e443f5',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'ok',
        '2025-05-06 16:12:10.131988+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'cee0b023-af6e-4fdb-8266-9dfae3f054a3',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-06 16:12:18.934186+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'eaf1dd73-1fa7-4f28-9363-836c5fca4684',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'ok',
        '2025-05-13 16:29:05.638874+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'ba5dbcbf-3496-4de0-bff7-d0b668201123',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'bad',
        '2025-05-13 16:27:33.380423+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e8139c0d-130a-440c-8b8f-98b2c3be3d90',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'bad',
        '2025-05-13 16:27:33.811249+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '64e97c29-7221-47e4-a802-59638b7b872a',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'bad',
        '2025-05-13 16:27:34.383677+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '7db6d645-69f7-4bbf-85d8-03f486149b37',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:27:35.807434+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '9de5b596-431f-4145-9ba9-31b4565cc73d',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:27:36.73803+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b74484b8-bdbb-4fa1-8067-f099bd3869f0',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'bad',
        '2025-05-13 16:27:44.200275+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '58062eb5-b876-4fc7-9628-0fa819f19f7c',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'good',
        '2025-05-13 16:27:46.799316+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '4a81a6f9-38c9-49d5-91ba-96bef5bb5dda',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'ok',
        '2025-05-13 16:27:47.134331+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '93cd13de-bb13-4c31-bb02-f0715ddc4b6d',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:27:47.720613+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'a0a2a6bb-cb12-46d3-a26c-81f8d2dc7cc1',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'ok',
        '2025-05-13 16:28:11.562981+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '92466e88-338f-411b-8e16-dd57c0dd3d0a',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:28:12.261692+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'aa7e756a-5557-4082-aa6e-5e8875e3a5d9',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'good',
        '2025-05-13 16:28:14.606244+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '5c893007-8b86-469a-a1a7-e7f9008c7671',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:28:15.734002+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'dfc55587-ad5a-472b-b2fe-6dcf06c06c27',
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'ok',
        '2025-05-13 16:28:17.20885+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '98197ed7-8667-40d5-b599-5176bc579329',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'good',
        '2025-05-13 16:28:34.446321+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '586f5489-d8bc-48b1-94c8-4d0cfbbb2fd8',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:28:34.808017+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '40c26233-f9a6-43e3-adf3-ca9e64da1d4e',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'bad',
        '2025-05-13 16:29:06.98053+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '335caf05-dae0-4ca9-9416-a46a67b5f715',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:28:35.837071+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'd56aa109-7335-41a9-99ab-be2c56d91e32',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'good',
        '2025-05-13 16:28:37.773968+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '0d1b0842-c567-4bdd-822b-62e7fb7b67fe',
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'good',
        '2025-05-13 16:28:38.502108+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '0893b03f-5706-48a4-88ee-385cd18747c3',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'bad',
        '2025-05-13 16:28:46.408643+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '091297d0-046d-420e-bc27-787c6e33e083',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'bad',
        '2025-05-13 16:28:46.824251+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '14f12900-86fe-4e16-8789-57cdc80e2c98',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'bad',
        '2025-05-13 16:28:49.086161+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '76002aa9-7b4c-4a61-9c65-da14c30c57f2',
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:28:50.357166+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '97e0961f-f913-4b3a-aba7-2efa6039b5fe',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'bad',
        '2025-05-13 16:29:08.313012+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'bad73ab3-dc8d-4c29-95b1-41bfbf549593',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:29:09.671308+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '75800429-b672-4ad2-9621-34b4914b3c34',
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:29:10.739365+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b0c852bf-fd5f-4b63-862d-7a2a2e14c9dd',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'good',
        '2025-05-13 16:29:35.286414+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '54669159-67d7-4d07-bda3-74c45cbebece',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:29:36.020599+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '794097c3-87f8-4a25-aa0c-d8112832e7d9',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'ok',
        '2025-05-13 16:29:37.211371+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'be26b4f8-422f-47c8-8697-8eedaf314f9e',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:29:47.837351+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'a04fa8e6-7189-45a1-861d-c30e53127eb0',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:29:48.973228+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e31bbefe-6acb-4931-aacc-1ab862e8aaa2',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:29:50.129126+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'fa0123a3-ec2e-460b-9c47-8e6cfb04a5b1',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'good',
        '2025-05-13 16:29:51.085875+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '29cf45f0-9c92-4510-8700-d70343962344',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:36:22.891252+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '7f8e4f00-adcc-4623-b4d3-1a44ab56ec9d',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:36:33.336817+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '70cb7d68-6e44-423a-aa96-24d876fd6124',
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:36:34.015844+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '4eb84105-3168-46ba-9831-77d3d97166d0',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:36:52.263122+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '924af27b-9aa1-4a19-8268-2c2de543a985',
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:36:52.524766+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e009ee77-37bd-4961-9d01-a77876c5684b',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-17 22:32:12.771476+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'ee6277a9-3983-4d26-8246-1ddd6f943cfb',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'ok',
        '2025-05-26 23:38:21.912757+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'cbc6628b-1f5c-4772-b1c0-e5dfc036e1ad',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'good',
        '2025-05-13 16:29:59.005625+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'c56e7798-0a81-4950-858a-bdc1f34609ce',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:29:59.27572+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'cd423830-b5bb-45ed-98c6-b0e22b79b811',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:30:00.445641+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '3a30110f-a70d-40c1-bf31-0f059fc362aa',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'good',
        '2025-05-13 16:30:01.799219+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '9333d189-757a-491a-8766-32246695eca0',
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'good',
        '2025-05-13 16:30:02.890921+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '9a5197b6-41cd-492c-a276-59b5a3109565',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-13 16:30:11.705307+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '957be29e-9d9c-4142-b3da-6609deb904b2',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-13 16:30:12.36149+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'ca051fd4-487c-4a66-8653-c7bb17de1e45',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'good',
        '2025-05-13 16:30:13.468073+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '34300d4c-47d0-4081-aa60-54baab9b1b51',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'good',
        '2025-05-13 16:30:15.016173+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'bf464af5-5554-491d-8068-e6a676685eed',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-13 16:30:16.198924+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b4779ed5-dfb4-4fd9-a0a0-180888b61d0c',
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:36:22.573663+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'e7621b44-e6d8-484c-8d54-dc15ed157a9d',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'good',
        '2025-05-13 20:49:51.396398+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b77dcc15-f1f2-4e7d-80bc-315d2072198f',
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'good',
        '2025-05-13 20:49:53.190794+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '6d2cd0a2-23b5-45d7-a430-b16d2a265f70',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'bad',
        '2025-05-26 15:08:03.779092+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '6133637c-4d2f-40a7-b14d-455e087606bf',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'ok',
        '2025-05-26 15:08:19.233445+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '3f95ed6d-a086-47e2-9c31-97a95f581c60',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'good',
        '2025-05-26 15:08:19.941767+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '71a4691c-026e-4bbc-892a-013511296882',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'good',
        '2025-05-26 15:08:21.274871+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'b73a0f53-01bc-48f4-86de-e7934a539216',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'good',
        '2025-05-26 15:08:22.575322+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        '196a9333-6f50-45ce-aa11-449d42a194a4',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-26 15:08:23.818293+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment
VALUES (
        'fc049717-e1f4-4706-a362-b2b9e53ecfed',
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'good',
        '2025-05-26 15:08:24.951111+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment_completion
VALUES (
        'e482ab63-c1c0-4943-9221-989b0c257559',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:28:21.311469+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '577dca3b-62cb-4fc5-92c0-9f1effa23e74',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:28:41.067543+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '336978b4-de3b-4537-afd4-5c7d85e47f0d',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:28:59.358812+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        'f1017827-0ac5-433c-afbb-b02512c0254f',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:29:14.684664+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '8d3e9a5b-96bd-4378-a412-fd01ad4b110a',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:30:06.493293+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '2a73bae5-f7d8-46e8-9cc6-dfd8a8368294',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:36:36.49778+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '0d92c377-bf5c-49f3-b5f9-ce2faed4215b',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:36:54.860814+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '1e59241a-0917-475b-be6b-79739a757657',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 20:49:55.480085+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        'a1b17b88-ce07-431d-840b-1b37f2058cee',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-17 22:32:47.265095+02',
        'Maximilian Rapp'
    );

INSERT INTO public.assessment_completion
VALUES (
        '413caaf6-e0ec-43af-9c5c-27550958ee2c',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-26 15:08:27.83436+02',
        'Maximilian Rapp'
    );

-- Insert test assessment schemas
INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests'),
('550e8400-e29b-41d4-a716-446655440001', 'Self Evaluation Schema', 'This is the default self evaluation schema.'),
('550e8400-e29b-41d4-a716-446655440002', 'Peer Evaluation Schema', 'This is the default peer evaluation schema.'),
('550e8400-e29b-41d4-a716-446655440003', 'Tutor Evaluation Schema', 'This is the default tutor evaluation schema.');

-- Insert test course_phase_config entries for visibility tests
-- Course phase config for visible scenario (both grade suggestions and action items visible, deadline passed)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2025-01-01 00:00:00+00', true, true, true);

-- Course phase config for not visible scenario (both grade suggestions and action items hidden, deadline passed)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', '2025-01-01 00:00:00+00', false, false, false);

-- Course phase config for before deadline scenario (deadline in the future)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '2099-12-31 23:59:59+00', true, true, false);

-- Insert test assessment_completion entries for visibility tests
-- Using different course_participation_id to avoid conflicts with other tests
INSERT INTO public.assessment_completion VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2025-05-13 16:36:46.782123+02', 'Visibility Test Author', 'Test comment for visible scenario', 4.5, true),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', '2025-05-13 16:36:46.782123+02', 'Visibility Test Author', 'Test comment for not visible scenario', 3.5, true);

-- Insert test action_item entries for visibility tests
INSERT INTO public.action_item (id, course_phase_id, course_participation_id, action, author) VALUES
('a1111111-1111-1111-1111-111111111111', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for visible scenario', 'tester'),
('a2222222-2222-2222-2222-222222222222', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for not visible scenario', 'tester');

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
VALUES (10, false);

ALTER TABLE ONLY public.assessment_completion
ADD CONSTRAINT assessment_completion_pkey PRIMARY KEY (course_participation_id, course_phase_id);

ALTER TABLE ONLY public.action_item
ADD CONSTRAINT action_item_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.assessment
ADD CONSTRAINT assessment_course_participation_id_course_phase_id_competen_key UNIQUE (
        course_participation_id,
        course_phase_id,
        competency_id
    );

ALTER TABLE ONLY public.assessment
ADD CONSTRAINT assessment_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.category
ADD CONSTRAINT category_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.competency
ADD CONSTRAINT competency_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.schema_migrations
ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);

CREATE INDEX idx_assessment_completion_participation_phase ON public.assessment_completion USING btree (course_participation_id, course_phase_id);

ALTER TABLE ONLY public.assessment
ADD CONSTRAINT assessment_competency_id_fkey FOREIGN KEY (competency_id) REFERENCES public.competency(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.competency
ADD CONSTRAINT competency_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.category(id) ON DELETE CASCADE;
