--
-- Test database setup for privacy router tests
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';
SET default_table_access_method = heap;

-- ============================================================
-- Student table (required as FK target for privacy_export)
-- ============================================================

CREATE TYPE gender AS ENUM ('male', 'female', 'diverse', 'prefer_not_to_say');
CREATE TYPE study_degree AS ENUM ('bachelor', 'master');

CREATE TABLE student (
    id                      uuid                    NOT NULL,
    first_name              character varying(50),
    last_name               character varying(50),
    email                   character varying(255),
    matriculation_number    character varying(30),
    university_login        character varying(20),
    has_university_account  boolean,
    gender                  gender                  NOT NULL,
    nationality             VARCHAR(2),
    study_program           varchar(100),
    study_degree            study_degree            NOT NULL DEFAULT 'bachelor',
    current_semester        int,
    last_modified           timestamp               NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE ONLY student ADD CONSTRAINT student_pkey PRIMARY KEY (id);
ALTER TABLE ONLY student ADD CONSTRAINT student_email_key UNIQUE (email);
ALTER TABLE ONLY student ADD CONSTRAINT student_matriculation_number_key UNIQUE (matriculation_number);
ALTER TABLE ONLY student ADD CONSTRAINT student_university_login_key UNIQUE (university_login);

-- ============================================================
-- Privacy export tables & view
-- ============================================================

CREATE TYPE export_status AS ENUM ('pending', 'complete', 'failed', 'no_data', 'archived');

CREATE TABLE privacy_export (
    id              uuid            PRIMARY KEY,
    user_id         uuid            NOT NULL,
    student_id      uuid            REFERENCES student(id),
    status          export_status   NOT NULL DEFAULT 'pending',
    date_created    timestamptz     NOT NULL DEFAULT now(),
    valid_until     timestamptz     NOT NULL
);

CREATE TABLE privacy_export_document (
    id              uuid            PRIMARY KEY,
    export_id       uuid            REFERENCES privacy_export(id),
    date_created    timestamptz     NOT NULL DEFAULT now(),
    source_name     text            NOT NULL,
    object_key      text            NOT NULL,
    status          export_status   NOT NULL DEFAULT 'pending',
    file_size       bigint,
    downloaded_at   timestamptz
);

CREATE VIEW privacy_export_with_docs AS
SELECT
    e.id,
    e.user_id,
    e.student_id,
    e.status,
    e.date_created,
    e.valid_until,
    COALESCE(
        jsonb_agg(
            json_build_object(
                'id',           ed.id,
                'date_created', ed.date_created,
                'source_name',  ed.source_name,
                'status',       ed.status,
                'file_size',    ed.file_size,
                'downloaded_at',ed.downloaded_at
            ) ORDER BY ed.date_created ASC
        ) FILTER (WHERE ed.id IS NOT NULL),
        '[]'::jsonb
    ) AS documents
FROM privacy_export e
LEFT JOIN privacy_export_document ed ON ed.export_id = e.id
GROUP BY e.id;

-- ============================================================
-- Test data
-- ============================================================
-- Students and their corresponding export user IDs:
--   Alice  student_id: aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa  user_id: 11111111-1111-1111-1111-111111111111
--   Bob    student_id: bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb  user_id: 22222222-2222-2222-2222-222222222222
--   Carol  student_id: cccccccc-cccc-cccc-cccc-cccccccccccc  user_id: 33333333-3333-3333-3333-333333333333
--   Dave   student_id: dddddddd-dddd-dddd-dddd-dddddddddddd  user_id: 44444444-4444-4444-4444-444444444444

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Alice', 'Alpha', 'alice.alpha@tum.de', '03700001', 'aa00aaa', true,  'female'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Bob',   'Beta',  'bob.beta@tum.de',   '03700002', 'bb00bbb', true,  'male'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Carol', 'Gamma', 'carol.gamma@tum.de','03700003', 'cc00ccc', false, 'female'),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Dave',  'Delta', 'dave.delta@tum.de', '03700004', 'dd00ddd', true,  'male');

-- ---------------------------------------------------------------
-- Export 1 – all docs successful (complete or no_data)
--   owner: Alice (user_id 11111111-...)
-- ---------------------------------------------------------------
INSERT INTO privacy_export (id, user_id, student_id, status, date_created, valid_until)
VALUES (
    'e1111111-1111-1111-1111-111111111111',
    '11111111-1111-1111-1111-111111111111',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'complete',
    '2026-03-01 10:00:00+00',
    '2030-01-01 00:00:00+00'
);

INSERT INTO privacy_export_document (id, export_id, date_created, source_name, object_key, status, file_size)
VALUES
    (
        'd1111111-1111-1111-1111-100000000001',
        'e1111111-1111-1111-1111-111111111111',
        '2026-03-01 10:00:01+00',
        'Core',
        'exports/e1111111/core.json',
        'complete',
        2048
    ),
    (
        'd1111111-1111-1111-1111-100000000002',
        'e1111111-1111-1111-1111-111111111111',
        '2026-03-01 10:00:02+00',
        'MicroserviceA',
        'exports/e1111111/microservice_a.json',
        'no_data',
        NULL
    );

-- ---------------------------------------------------------------
-- Export 2 – partially failed (complete + no_data + failed docs)
--   owner: Bob (user_id 22222222-...)
-- ---------------------------------------------------------------
INSERT INTO privacy_export (id, user_id, student_id, status, date_created, valid_until)
VALUES (
    'e2222222-2222-2222-2222-222222222222',
    '22222222-2222-2222-2222-222222222222',
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'failed',
    '2026-03-05 10:00:00+00',
    '2030-01-01 00:00:00+00'
);

INSERT INTO privacy_export_document (id, export_id, date_created, source_name, object_key, status, file_size)
VALUES
    (
        'd2222222-2222-2222-2222-200000000001',
        'e2222222-2222-2222-2222-222222222222',
        '2026-03-05 10:00:01+00',
        'Core',
        'exports/e2222222/core.json',
        'complete',
        1024
    ),
    (
        'd2222222-2222-2222-2222-200000000002',
        'e2222222-2222-2222-2222-222222222222',
        '2026-03-05 10:00:02+00',
        'MicroserviceA',
        'exports/e2222222/microservice_a.json',
        'no_data',
        NULL
    ),
    (
        'd2222222-2222-2222-2222-200000000003',
        'e2222222-2222-2222-2222-222222222222',
        '2026-03-05 10:00:03+00',
        'MicroserviceB',
        'exports/e2222222/microservice_b.json',
        'failed',
        NULL
    );

-- ---------------------------------------------------------------
-- Export 3 – archived (deletion routine ran; files gone from S3)
--   owner: Carol (user_id 33333333-...)
--   valid_until is in the past (expiry triggers the deletion routine)
-- ---------------------------------------------------------------
INSERT INTO privacy_export (id, user_id, student_id, status, date_created, valid_until)
VALUES (
    'e3333333-3333-3333-3333-333333333333',
    '33333333-3333-3333-3333-333333333333',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'archived',
    '2026-02-01 10:00:00+00',
    '2026-03-01 00:00:00+00'
);

INSERT INTO privacy_export_document (id, export_id, date_created, source_name, object_key, status, file_size)
VALUES
    (
        'd3333333-3333-3333-3333-300000000001',
        'e3333333-3333-3333-3333-333333333333',
        '2026-02-01 10:00:01+00',
        'Core',
        'exports/e3333333/core.json',
        'archived',
        2048
    ),
    (
        'd3333333-3333-3333-3333-300000000002',
        'e3333333-3333-3333-3333-333333333333',
        '2026-02-01 10:00:02+00',
        'MicroserviceA',
        'exports/e3333333/microservice_a.json',
        'archived',
        512
    );

-- ---------------------------------------------------------------
-- Export 4 – rate limited: expired yesterday, created 5 days ago
--   valid_until is in the past so the export is no longer accessible,
--   but date_created + 30d is still in the future so the user is rate limited.
--   Relative timestamps keep this test case valid regardless of when it runs.
--   owner: Dave (user_id 44444444-...)
-- ---------------------------------------------------------------
INSERT INTO privacy_export (id, user_id, student_id, status, date_created, valid_until)
VALUES (
    'e4444444-4444-4444-4444-444444444444',
    '44444444-4444-4444-4444-444444444444',
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'complete',
    NOW() - INTERVAL '5 days',
    NOW() - INTERVAL '1 day'
);

INSERT INTO privacy_export_document (id, export_id, source_name, object_key, status, file_size)
VALUES (
    'd4444444-4444-4444-4444-400000000001',
    'e4444444-4444-4444-4444-444444444444',
    'Core',
    'exports/e4444444/core.json',
    'complete',
    1024
);
