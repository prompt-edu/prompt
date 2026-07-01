--
-- PostgreSQL database dump
--
-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.8 (Homebrew)
SET
    statement_timeout = 0;

SET
    lock_timeout = 0;

SET
    idle_in_transaction_session_timeout = 0;

SET
    client_encoding = 'UTF8';

SET
    standard_conforming_strings = on;

SELECT
    pg_catalog.set_config ('search_path', 'public', false);

SET
    check_function_bodies = false;

SET
    xmloption = content;

SET
    client_min_messages = warning;

SET
    row_security = off;

SET
    default_tablespace = '';

SET
    default_table_access_method = heap;

--
-- Name: course; Type: TABLE; Schema: public; Owner: prompt-postgres
--
create type course_type as enum ('lecture', 'seminar', 'practical course');

CREATE TABLE
    course (
        id uuid NOT NULL,
        name text NOT NULL,
        start_date date,
        end_date date,
        semester_tag text,
        course_type course_type NOT NULL,
        ects integer,
        restricted_data jsonb,
        student_readable_data jsonb DEFAULT '{}',
        template boolean NOT NULL DEFAULT FALSE,
        archived boolean NOT NULL DEFAULT FALSE,
        archived_on timestamptz
    );

--
-- Data for Name: course; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--
INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
        'iPraktikum',
        '2024-10-01',
        '2030-01-01',
        'ios24245',
        'practical course',
        10,
        '{
          "icon": "apple",
          "bg-color": "bg-orange-100"
                }',
                '{
                    "icon": "apple",
                    "bg-color": "bg-orange-100"
                }',
                TRUE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
        restricted_data,
        student_readable_data,
        template,
        archived,
        archived_on
    )
VALUES
    (
        '918977e1-2d27-4b55-9064-8504ff027a1a',
        'New fancy course',
        '2024-10-01',
        '2030-01-01',
        'ios24245',
        'practical course',
        10,
        '{}',
        '{}',
        FALSE,
        FALSE,
        NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        'fe672868-3d07-4bdd-af41-121fd05e2d0d',
        'iPraktikum',
        '2024-10-01',
        '2030-01-01',
        'ios24245',
        'lecture',
        5,
        '{
          "icon": "home",
          "color": "green"
                }',
                '{
                                        "icon": "home",
                                        "bg-color": null
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '0bb5c8dc-a4df-4d64-a9fd-fe8840760d6b',
        'Test5',
        '2025-01-13',
        '2030-01-18',
        'ios2425',
        'seminar',
        5,
        '{
          "icon": "smartphone",
          "bg-color": "bg-blue-100"
                }',
                '{
                    "icon": "smartphone",
                    "bg-color": "bg-blue-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '55856fdc-fc2f-456a-a5a5-726d60aaae7c',
        'iPraktikum3',
        '2025-01-07',
        '2030-01-24',
        'ios2425',
        'practical course',
        10,
        '{
          "icon": "smartphone",
          "bg-color": "bg-green-100"
                }',
                '{
                    "icon": "smartphone",
                    "bg-color": "bg-green-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '64a12e61-a238-4cea-a36a-5eaf89d7a940',
        'Another TEst',
        '2024-12-15',
        '2030-01-17',
        'ios2425',
        'seminar',
        5,
        '{
          "icon": "folder",
          "bg-color": "bg-red-100",
          "some-secret-data": "secret"
                }',
                '{
                                        "icon": "folder",
                                        "bg-color": "bg-red-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '07d0664c-6116-4897-97c9-521c8d73dd9f',
        'Further Testing',
        '2024-12-17',
        '2030-01-15',
        'ios24',
        'practical course',
        10,
        '{
          "icon": "monitor",
          "bg-color": "bg-cyan-100"
                }',
                '{
                    "icon": "monitor",
                    "bg-color": "bg-cyan-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '00f6d242-9716-487c-a8de-5e02112ea131',
        'Test150',
        '2024-12-17',
        '2030-01-17',
        'test',
        'practical course',
        10,
        '{
          "icon": "book-open-text",
          "bg-color": "bg-orange-100"
                }',
                '{
                    "icon": "book-open-text",
                    "bg-color": "bg-orange-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        '894cb6fc-9407-4642-b4de-2e0b4e893126',
        'iPraktikum-Test',
        '2025-03-10',
        '2030-08-01',
        'ios2425',
        'practical course',
        10,
        '{
          "icon": "gamepad-2",
          "bg-color": "bg-green-100"
                }',
                '{
                    "icon": "gamepad-2",
                    "bg-color": "bg-green-100"
                }',
                FALSE,
                FALSE,
                NULL
    );

INSERT INTO
    course (
        id,
        name,
        start_date,
        end_date,
        semester_tag,
        course_type,
        ects,
                restricted_data,
                student_readable_data,
                template,
                archived,
                archived_on
    )
VALUES
    (
        'b2c3d4e5-1111-4000-8000-000000000001',
        'Archived Course',
        '2023-10-01',
        '2024-03-01',
        'ios2425',
        'practical course',
        10,
        '{
          "icon": "archive",
          "bg-color": "bg-gray-100"
                }',
                '{
                    "icon": "archive",
                    "bg-color": "bg-gray-100"
                }',
                FALSE,
                TRUE,
                '2024-03-02 00:00:00+00'
    );

--
-- Name: course course_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--
ALTER TABLE ONLY course ADD CONSTRAINT course_pkey PRIMARY KEY (id);

--
-- PostgreSQL database dump complete
--
CREATE TABLE
    course_phase_type (id uuid NOT NULL, name text NOT NULL);

CREATE TABLE
    course_phase (
        id uuid NOT NULL,
        course_id uuid NOT NULL,
        name text,
        restricted_data jsonb,
        student_readable_data jsonb DEFAULT '{}',
        is_initial_phase boolean NOT NULL,
        course_phase_type_id uuid NOT NULL
    );

INSERT INTO
    course_phase_type (id, name)
VALUES
    (
        '7dc1c4e8-4255-4874-80a0-0c12b958744b',
        'application'
    );

INSERT INTO
    course_phase_type (id, name)
VALUES
    (
        '7dc1c4e8-4255-4874-80a0-0c12b958744c',
        'template_component'
    );

--
-- Data for Name: course_phase; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--
INSERT INTO
    course_phase (
        id,
        course_id,
        name,
                restricted_data,
                student_readable_data,
        is_initial_phase,
        course_phase_type_id
    )
VALUES
    (
        '3d1f3b00-87f3-433b-a713-178c4050411b',
        '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
        'Test',
        '{
          "test-key": "test-value"
        }',
                                '{}',
        false,
        '7dc1c4e8-4255-4874-80a0-0c12b958744b'
    );

INSERT INTO
    course_phase (
        id,
        course_id,
        name,
        restricted_data,
        student_readable_data,
        is_initial_phase,
        course_phase_type_id
    )
VALUES
    (
        '92bb0532-39e5-453d-bc50-fa61ea0128b2',
        '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
        'Template Phase',
        '{}',
        '{}',
        false,
        '7dc1c4e8-4255-4874-80a0-0c12b958744c'
    );

INSERT INTO
    course_phase (
        id,
        course_id,
        name,
        restricted_data,
        student_readable_data,
        is_initial_phase,
        course_phase_type_id
    )
VALUES
    (
        '500db7ed-2eb2-42d0-82b3-8750e12afa8a',
        '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
        'Application Phase',
        '{}',
        '{}',
        true,
        '7dc1c4e8-4255-4874-80a0-0c12b958744b'
    );

ALTER TABLE ONLY course_phase ADD CONSTRAINT course_phase_pkey PRIMARY KEY (id);

CREATE UNIQUE INDEX unique_initial_phase_per_course ON course_phase USING btree (course_id)
WHERE
    (is_initial_phase = true);

ALTER TABLE ONLY course_phase_type ADD CONSTRAINT course_phase_type_name_key UNIQUE (name);

ALTER TABLE ONLY course_phase_type ADD CONSTRAINT course_phase_type_pkey PRIMARY KEY (id);

ALTER TABLE ONLY course_phase ADD CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type (id);

CREATE TABLE
    course_phase_graph (
        from_course_phase_id uuid NOT NULL,
        to_course_phase_id uuid NOT NULL
    );

--
-- Data for Name: course_phase_graph; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--
INSERT INTO
    course_phase_graph (from_course_phase_id, to_course_phase_id)
VALUES
    (
        '500db7ed-2eb2-42d0-82b3-8750e12afa8a',
        '92bb0532-39e5-453d-bc50-fa61ea0128b2'
    );

--
-- Name: course_phase_graph unique_from_course_phase; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--
ALTER TABLE ONLY course_phase_graph ADD CONSTRAINT unique_from_course_phase UNIQUE (from_course_phase_id);

--
-- Name: course_phase_graph unique_to_course_phase; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--
ALTER TABLE ONLY course_phase_graph ADD CONSTRAINT unique_to_course_phase UNIQUE (to_course_phase_id);

--
-- Name: course_phase_graph fk_from_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--
ALTER TABLE ONLY course_phase_graph ADD CONSTRAINT fk_from_course_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase (id) ON DELETE CASCADE;

--
-- Name: course_phase_graph fk_to_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--
ALTER TABLE ONLY course_phase_graph ADD CONSTRAINT fk_to_course_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase (id) ON DELETE CASCADE;

-- Apply migration adjustments for tests
ALTER TABLE course
ADD COLUMN short_description VARCHAR(255);

ALTER TABLE course
ADD COLUMN long_description TEXT;