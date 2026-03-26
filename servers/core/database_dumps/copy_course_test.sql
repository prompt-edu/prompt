--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.13 (Homebrew)

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

--
-- Name: course_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE course_type AS ENUM (
    'lecture',
    'seminar',
    'practical course'
);


--
-- Name: gender; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE gender AS ENUM (
    'male',
    'female',
    'diverse',
    'prefer_not_to_say'
);


--
-- Name: pass_status; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE pass_status AS ENUM (
    'passed',
    'failed',
    'not_assessed'
);


--
-- Name: study_degree; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE study_degree AS ENUM (
    'bachelor',
    'master'
);


--
-- Name: update_last_modified_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION update_last_modified_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  NEW.last_modified = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: application_answer_multi_select; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_answer_multi_select (
    id uuid NOT NULL,
    application_question_id uuid NOT NULL,
    answer text[],
    course_participation_id uuid NOT NULL
);


--
-- Name: application_answer_text; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_answer_text (
    id uuid NOT NULL,
    application_question_id uuid NOT NULL,
    answer text,
    course_participation_id uuid NOT NULL
);


--
-- Name: application_assessment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_assessment (
    id uuid NOT NULL,
    score integer,
    course_phase_id uuid NOT NULL,
    course_participation_id uuid NOT NULL
);


--
-- Name: application_question_multi_select; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_question_multi_select (
    id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    title text,
    description text,
    placeholder text,
    error_message text,
    is_required boolean,
    min_select integer,
    max_select integer,
    options text[],
    order_num integer,
    accessible_for_other_phases boolean DEFAULT false,
    access_key character varying(50) DEFAULT ''::character varying
);


--
-- Name: application_question_text; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_question_text (
    id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    title text,
    description text,
    placeholder text,
    validation_regex text,
    error_message text,
    is_required boolean,
    allowed_length integer,
    order_num integer,
    accessible_for_other_phases boolean DEFAULT false,
    access_key character varying(50) DEFAULT ''::character varying
);

--
-- Name: application_question_file_upload; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_question_file_upload (
    id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    title text NOT NULL,
    description text,
    is_required boolean NOT NULL DEFAULT false,
    allowed_file_types text,
    max_file_size_mb integer,
    order_num integer NOT NULL,
    accessible_for_other_phases boolean NOT NULL DEFAULT false,
    access_key text
);


--
-- Name: files; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE files (
    id uuid NOT NULL,
    filename character varying(500) NOT NULL,
    original_filename character varying(500) NOT NULL,
    content_type character varying(200) NOT NULL,
    size_bytes bigint NOT NULL,
    storage_key character varying(500) NOT NULL,
    storage_provider character varying(50) NOT NULL DEFAULT 'seaweedfs'::character varying,
    uploaded_by_user_id character varying(200) NOT NULL,
    uploaded_by_email character varying(200),
    course_phase_id uuid,
    description text,
    tags character varying(100)[],
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at timestamp without time zone
);


--
-- Name: application_answer_file_upload; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE application_answer_file_upload (
    id uuid NOT NULL,
    application_question_id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    file_id uuid NOT NULL
);


--
-- Name: course; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course (
    id uuid NOT NULL,
    name text NOT NULL,
    start_date date,
    end_date date,
    semester_tag text,
    course_type course_type NOT NULL,
    ects integer,
    restricted_data jsonb,
    student_readable_data jsonb DEFAULT '{}'::jsonb,
    template boolean NOT NULL DEFAULT FALSE,
    short_description character varying(255),
    long_description text,
    archived boolean NOT NULL DEFAULT FALSE,
    archived_on timestamp with time zone,
    CONSTRAINT check_end_date_after_start_date CHECK (
        template = true
            OR end_date > start_date
        )
);


--
-- Name: course_participation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_participation (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL
);


--
-- Name: course_phase; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    name text,
    restricted_data jsonb,
    is_initial_phase boolean NOT NULL,
    course_phase_type_id uuid NOT NULL,
    student_readable_data jsonb DEFAULT '{}'::jsonb
);


--
-- Name: course_phase_graph; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL
);


--
-- Name: course_phase_participation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_participation (
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    restricted_data jsonb,
    pass_status pass_status DEFAULT 'not_assessed'::pass_status,
    last_modified timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    student_readable_data jsonb DEFAULT '{}'::jsonb
);


--
-- Name: course_phase_type; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_type (
    id uuid NOT NULL,
    name text NOT NULL,
    initial_phase boolean DEFAULT false NOT NULL,
    base_url text DEFAULT 'core'::text NOT NULL,
    description text
);


--
-- Name: course_phase_type_participation_provided_output_dto; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_type_participation_provided_output_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    version_number integer NOT NULL,
    endpoint_path text NOT NULL,
    specification jsonb NOT NULL
);


--
-- Name: course_phase_type_participation_required_input_dto; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_type_participation_required_input_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    specification jsonb NOT NULL
);


--
-- Name: course_phase_type_phase_provided_output_dto; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_type_phase_provided_output_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    version_number integer NOT NULL,
    endpoint_path text NOT NULL,
    specification jsonb NOT NULL
);


--
-- Name: course_phase_type_phase_required_input_dto; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE course_phase_type_phase_required_input_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    specification jsonb NOT NULL
);


--
-- Name: participation_data_dependency_graph; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE participation_data_dependency_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL,
    from_course_phase_dto_id uuid NOT NULL,
    to_course_phase_dto_id uuid NOT NULL
);


--
-- Name: phase_data_dependency_graph; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE phase_data_dependency_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL,
    from_course_phase_dto_id uuid NOT NULL,
    to_course_phase_dto_id uuid NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: student; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE student (
    id uuid NOT NULL,
    first_name character varying(50),
    last_name character varying(50),
    email character varying(255),
    matriculation_number character varying(30),
    university_login character varying(20),
    has_university_account boolean,
    gender gender NOT NULL,
    nationality character varying(2),
    study_program character varying(100),
    study_degree study_degree DEFAULT 'bachelor'::study_degree NOT NULL,
    current_semester integer,
    last_modified timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Data for Name: application_answer_multi_select; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: application_answer_text; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: application_assessment; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: application_question_multi_select; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO application_question_multi_select VALUES ('9df42ab6-1607-4c78-a1bd-28c58399ed55', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Where are you saved?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 2, false, '');
INSERT INTO application_question_multi_select VALUES ('a540c8d0-b5ee-4510-a671-2368a5997b66', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Are you copied with the course?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 1, false, '');
INSERT INTO application_question_multi_select VALUES ('e9abbd7f-f7ae-4096-911c-c62fe33105fa', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Where are you saved?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 2, false, '');
INSERT INTO application_question_multi_select VALUES ('c20829f9-d1d2-4952-95ee-1ebbff01a432', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Are you copied with the course?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 1, false, '');


--
-- Data for Name: application_question_text; Type: TABLE DATA; Schema: public; Owner: -
--


--
-- Data for Name: course; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
VALUES
    (
        'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd',
        'Master Test',
        '2025-05-19',
        '2025-06-30',
        'ss25',
        'practical course',
        10,
        '{}',
        '{"icon": "school", "bg-color": "bg-teal-100"}',
        FALSE,
        'Hands-on master course',
        'Detailed description for the Master Test course used in copy tests.',
        FALSE,
        NULL
    );

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
VALUES
    (
        'c1f8060d-7381-4b64-a6ea-5ba8e8ac88ee',
        'Template Test',
        '2025-05-19',
        '2025-08-30',
        'ss25',
        'practical course',
        10,
        '{}',
        '{"icon": "school", "bg-color": "bg-teal-100"}',
        TRUE,
        'Template for future courses',
        'Long-form description for the template course copy flow.',
        FALSE,
        NULL
    );


--
-- Data for Name: course_participation; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: course_phase; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Application', '{"autoAccept": true, "applicationEndDate": "2025-06-01T23:59:00+02:00", "applicationStartDate": "2025-05-19T00:00:00+02:00", "externalStudentsAllowed": false, "universityLoginAvailable": true}', true, '3258275d-a76b-40f8-bb1a-95618299b8ac', '{}');
INSERT INTO course_phase VALUES ('93693f81-9c49-4183-ae70-c0ee3742560d', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Matching', '{}', false, 'c313bd84-bc7b-4a5a-aca3-77a526e02f57', '{}');
INSERT INTO course_phase VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Interview', '{"interviewQuestions": [{"id": 1747661413157, "orderNum": 0, "question": "Question 1"}, {"id": 1747661417361, "orderNum": 1, "question": "Question 2"}]}', false, '0542034b-87eb-4f91-ac90-b2e1536450de', '{}');


--
-- Data for Name: course_phase_graph; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase_graph VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '93693f81-9c49-4183-ae70-c0ee3742560d');
INSERT INTO course_phase_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123');


--
-- Data for Name: course_phase_participation; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: course_phase_type; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase_type VALUES ('0542034b-87eb-4f91-ac90-b2e1536450de', 'Interview', false, 'core', 'Test Description');
INSERT INTO course_phase_type VALUES ('c313bd84-bc7b-4a5a-aca3-77a526e02f57', 'Matching', false, 'core', 'Test Description');
INSERT INTO course_phase_type VALUES ('8c97fd14-e1fe-4b12-bf7d-c542e620e8d8', 'Intro Course Developer', false, 'http://localhost:8082/intro-course/api', 'Test Description');
INSERT INTO course_phase_type VALUES ('e0fe2692-4e06-47db-b80f-92817e9a7566', 'DevOps Challenge', false, 'core', 'Test Description');
INSERT INTO course_phase_type VALUES ('613f8b0a-2200-4650-b0bb-6a26a0a140e8', 'Assessment', false, 'http://localhost:8085/assessment/api', 'Test Description');
INSERT INTO course_phase_type VALUES ('88ca3586-7748-4152-8c89-5fbb6e113587', 'Team Allocation', false, 'http://localhost:8083/team-allocation/api', 'Test Description');
INSERT INTO course_phase_type VALUES ('cc63d311-6992-4b08-9719-a83b92f45e8f', 'Self Team Allocation', false, 'http://localhost:8084/self-team-allocation/api', 'Test Description');
INSERT INTO course_phase_type VALUES ('3258275d-a76b-40f8-bb1a-95618299b8ac', 'Application', true, 'core', 'Test Description');


--
-- Data for Name: course_phase_type_participation_provided_output_dto; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('57f45359-3fd6-47b7-822e-74e57b5f530c', '0542034b-87eb-4f91-ac90-b2e1536450de', 'score', 1, 'core', '{"type": "integer"}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('a67b6b12-bd5e-4be4-8353-fcc98fd7013d', '8c97fd14-e1fe-4b12-bf7d-c542e620e8d8', 'devices', 1, '/devices', '{"type": "array", "items": {"enum": ["IPhone", "IPad", "MacBook", "AppleWatch"], "type": "string"}}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('9a73af69-0ee1-46f1-bb3e-0d6775c7f02e', '613f8b0a-2200-4650-b0bb-6a26a0a140e8', 'scoreLevel', 1, '/student-assessment/scoreLevel', '{"enum": ["novice", "intermediate", "advanced", "expert"], "type": "string"}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('ebd9ac97-909e-4865-818a-faf2c5c22480', '88ca3586-7748-4152-8c89-5fbb6e113587', 'teamAllocation', 1, '/allocation', '{"type": "string"}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('df569553-b5d7-4ee5-8df2-5eb10fe083c5', 'cc63d311-6992-4b08-9719-a83b92f45e8f', 'teamAllocation', 1, '/allocation', '{"type": "string"}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('d7c3d526-c95d-4e23-8c6b-e0a561464ef3', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'score', 1, 'core', '{"type": "integer"}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('3c657de6-3fe1-41f3-a3c7-ceee83442815', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'applicationAnswers', 1, 'core', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');
INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('2ce41974-47e2-46fe-ad92-09159f08e030', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'additionalScores', 1, 'core', '{"type": "array", "items": {"type": "object", "required": ["score", "key"], "properties": {"key": {"type": "string"}, "score": {"type": "number"}}}}');


--
-- Data for Name: course_phase_type_participation_required_input_dto; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('67a66cf6-5505-41f2-a72e-2f247b386b83', '0542034b-87eb-4f91-ac90-b2e1536450de', 'score', '{"type": "integer"}');
INSERT INTO course_phase_type_participation_required_input_dto VALUES ('7ee14e46-4dd0-4281-a3da-26e00859a62e', '0542034b-87eb-4f91-ac90-b2e1536450de', 'applicationAnswers', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');
INSERT INTO course_phase_type_participation_required_input_dto VALUES ('0dfdb01d-fffd-4ebc-b191-3201aa2d0fc3', 'c313bd84-bc7b-4a5a-aca3-77a526e02f57', 'score', '{"type": "integer"}');
INSERT INTO course_phase_type_participation_required_input_dto VALUES ('faad5951-3af7-4189-adf3-298a545082af', '88ca3586-7748-4152-8c89-5fbb6e113587', 'applicationAnswers', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');
INSERT INTO course_phase_type_participation_required_input_dto VALUES ('afae0c00-a7ae-40f9-b707-af19393a58b7', '88ca3586-7748-4152-8c89-5fbb6e113587', 'devices', '{"type": "array", "items": {"enum": ["IPhone", "IPad", "MacBook", "AppleWatch"], "type": "string"}}');
INSERT INTO course_phase_type_participation_required_input_dto VALUES ('2f17451e-d905-4484-aebc-a8c5191a4557', '88ca3586-7748-4152-8c89-5fbb6e113587', 'scoreLevel', '{"enum": ["novice", "intermediate", "advanced", "expert"], "type": "string"}');


--
-- Data for Name: course_phase_type_phase_provided_output_dto; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO course_phase_type_phase_provided_output_dto VALUES ('c539e8fa-e8fb-4ed3-9cab-feacf3d7d8d1', '88ca3586-7748-4152-8c89-5fbb6e113587', 'teams', 1, '/team', '{"type": "array", "items": {"type": "object", "required": ["id", "name"], "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}}');
INSERT INTO course_phase_type_phase_provided_output_dto VALUES ('eded55c7-1077-4945-9ae6-dbdac205f9cf', 'cc63d311-6992-4b08-9719-a83b92f45e8f', 'teams', 1, '/team', '{"type": "array", "items": {"type": "object", "required": ["id", "name"], "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}}');


--
-- Data for Name: course_phase_type_phase_required_input_dto; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: participation_data_dependency_graph; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO participation_data_dependency_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', 'd7c3d526-c95d-4e23-8c6b-e0a561464ef3', '67a66cf6-5505-41f2-a72e-2f247b386b83');
INSERT INTO participation_data_dependency_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '3c657de6-3fe1-41f3-a3c7-ceee83442815', '7ee14e46-4dd0-4281-a3da-26e00859a62e');
INSERT INTO participation_data_dependency_graph VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '93693f81-9c49-4183-ae70-c0ee3742560d', '57f45359-3fd6-47b7-822e-74e57b5f530c', '0dfdb01d-fffd-4ebc-b191-3201aa2d0fc3');


--
-- Data for Name: phase_data_dependency_graph; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO schema_migrations VALUES (16, false);


--
-- Data for Name: student; Type: TABLE DATA; Schema: public; Owner: -
--



--
-- Name: application_answer_multi_select application_answer_multi_select_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT application_answer_multi_select_pkey PRIMARY KEY (id);


--
-- Name: application_answer_text application_answer_text_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT application_answer_text_pkey PRIMARY KEY (id);


--
-- Name: application_assessment application_assessment_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT application_assessment_pkey PRIMARY KEY (id);


--
-- Name: application_question_multi_select application_question_multi_select_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT application_question_multi_select_pkey PRIMARY KEY (id);


--
-- Name: application_question_text application_question_text_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT application_question_text_pkey PRIMARY KEY (id);

--
-- Name: application_question_file_upload application_question_file_upload_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_file_upload
    ADD CONSTRAINT application_question_file_upload_pkey PRIMARY KEY (id);

--
-- Name: application_answer_file_upload application_answer_file_upload_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_file_upload
    ADD CONSTRAINT application_answer_file_upload_pkey PRIMARY KEY (id);

--
-- Name: application_answer_file_upload unique_file_upload_answer; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_file_upload
    ADD CONSTRAINT unique_file_upload_answer UNIQUE (course_participation_id, application_question_id);

--
-- Name: files files_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY files
    ADD CONSTRAINT files_pkey PRIMARY KEY (id);

--
-- Name: files files_storage_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY files
    ADD CONSTRAINT files_storage_key_key UNIQUE (storage_key);


--
-- Name: course_participation course_participation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT course_participation_pkey PRIMARY KEY (id);


--
-- Name: course_phase_participation course_phase_participation_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT course_phase_participation_pkey PRIMARY KEY (course_participation_id, course_phase_id);


--
-- Name: course_phase course_phase_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT course_phase_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type course_phase_type_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type
    ADD CONSTRAINT course_phase_type_name_key UNIQUE (name);


--
-- Name: course_phase_type_phase_provided_output_dto course_phase_type_phase_provided_output_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_phase_provided_output_dto
    ADD CONSTRAINT course_phase_type_phase_provided_output_dto_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type_phase_required_input_dto course_phase_type_phase_required_input_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_phase_required_input_dto
    ADD CONSTRAINT course_phase_type_phase_required_input_dto_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type course_phase_type_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type
    ADD CONSTRAINT course_phase_type_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type_participation_provided_output_dto course_phase_type_provided_output_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_participation_provided_output_dto
    ADD CONSTRAINT course_phase_type_provided_output_dto_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type_participation_required_input_dto course_phase_type_required_input_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_participation_required_input_dto
    ADD CONSTRAINT course_phase_type_required_input_dto_pkey PRIMARY KEY (id);


--
-- Name: course course_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course
    ADD CONSTRAINT course_pkey PRIMARY KEY (id);


--
-- Name: participation_data_dependency_graph meta_data_dependency_graph_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY participation_data_dependency_graph
    ADD CONSTRAINT meta_data_dependency_graph_pkey PRIMARY KEY (to_course_phase_id, to_course_phase_dto_id);


--
-- Name: phase_data_dependency_graph phase_data_dependency_graph_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY phase_data_dependency_graph
    ADD CONSTRAINT phase_data_dependency_graph_pkey PRIMARY KEY (to_course_phase_id, to_course_phase_dto_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: student student_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY student
    ADD CONSTRAINT student_email_key UNIQUE (email);


--
-- Name: student student_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY student
    ADD CONSTRAINT student_pkey PRIMARY KEY (id);


--
-- Name: application_answer_multi_select unique_application_answer_multi_select; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT unique_application_answer_multi_select UNIQUE (course_participation_id, application_question_id);


--
-- Name: application_answer_text unique_application_answer_text; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT unique_application_answer_text UNIQUE (course_participation_id, application_question_id);


--
-- Name: course unique_course_identifier; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course
    ADD CONSTRAINT unique_course_identifier UNIQUE (name, semester_tag);


--
-- Name: course_participation unique_course_participation; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT unique_course_participation UNIQUE (course_id, student_id);


--
-- Name: course_phase_participation unique_course_phase_participation; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT unique_course_phase_participation UNIQUE (course_participation_id, course_phase_id);


--
-- Name: application_assessment unique_course_phase_participation_assessment; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT unique_course_phase_participation_assessment UNIQUE (course_phase_id, course_participation_id);


--
-- Name: course_phase_graph unique_from_course_phase; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT unique_from_course_phase UNIQUE (from_course_phase_id);


--
-- Name: course_phase_graph unique_to_course_phase; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT unique_to_course_phase UNIQUE (to_course_phase_id);


--
-- Name: student_matriculation_number_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX student_matriculation_number_unique ON student USING btree (matriculation_number) WHERE ((matriculation_number IS NOT NULL) AND ((matriculation_number)::text <> ''::text));


--
-- Name: student_university_login_unique; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX student_university_login_unique ON student USING btree (university_login) WHERE ((university_login IS NOT NULL) AND ((university_login)::text <> ''::text));


--
-- Name: unique_initial_phase_per_course; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX unique_initial_phase_per_course ON course_phase USING btree (course_id) WHERE (is_initial_phase = true);


--
-- Name: course_phase_participation set_last_modified_course_phase_participation; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_last_modified_course_phase_participation BEFORE UPDATE ON course_phase_participation FOR EACH ROW EXECUTE FUNCTION update_last_modified_column();


--
-- Name: student set_last_modified_student; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER set_last_modified_student BEFORE UPDATE ON student FOR EACH ROW EXECUTE FUNCTION update_last_modified_column();


--
-- Name: application_answer_text fk_application_question; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT fk_application_question FOREIGN KEY (application_question_id) REFERENCES application_question_text(id) ON DELETE CASCADE;


--
-- Name: application_answer_multi_select fk_application_question; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT fk_application_question FOREIGN KEY (application_question_id) REFERENCES application_question_multi_select(id) ON DELETE CASCADE;


--
-- Name: course_phase fk_course; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course(id) ON DELETE CASCADE;


--
-- Name: course_participation fk_course; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course(id) ON DELETE CASCADE;


--
-- Name: course_phase_participation fk_course_participation; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT fk_course_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE;


--
-- Name: application_answer_text fk_course_participation; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT fk_course_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE;


--
-- Name: application_answer_multi_select fk_course_participation; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT fk_course_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE;


--
-- Name: course_phase_participation fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: application_question_text fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: application_question_multi_select fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;

--
-- Name: application_question_file_upload fk_application_question_file_upload_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_question_file_upload
    ADD CONSTRAINT fk_application_question_file_upload_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;

--
-- Name: files fk_files_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY files
    ADD CONSTRAINT fk_files_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE SET NULL;

--
-- Name: application_answer_file_upload fk_application_answer_file_upload_question; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_file_upload
    ADD CONSTRAINT fk_application_answer_file_upload_question FOREIGN KEY (application_question_id) REFERENCES application_question_file_upload(id) ON DELETE CASCADE;

--
-- Name: application_answer_file_upload fk_application_answer_file_upload_participation; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_file_upload
    ADD CONSTRAINT fk_application_answer_file_upload_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE;

--
-- Name: application_answer_file_upload fk_application_answer_file_upload_file; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_answer_file_upload
    ADD CONSTRAINT fk_application_answer_file_upload_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE;


--
-- Name: application_assessment fk_course_phase_participation; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT fk_course_phase_participation FOREIGN KEY (course_participation_id, course_phase_id) REFERENCES course_phase_participation(course_participation_id, course_phase_id) ON DELETE CASCADE;


--
-- Name: course_phase_type_phase_provided_output_dto fk_course_phase_type_phase_provided; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_phase_provided_output_dto
    ADD CONSTRAINT fk_course_phase_type_phase_provided FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_type_phase_required_input_dto fk_course_phase_type_phase_required; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_phase_required_input_dto
    ADD CONSTRAINT fk_course_phase_type_phase_required FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_type_participation_provided_output_dto fk_course_phase_type_provided; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_participation_provided_output_dto
    ADD CONSTRAINT fk_course_phase_type_provided FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_type_participation_required_input_dto fk_course_phase_type_required; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_type_participation_required_input_dto
    ADD CONSTRAINT fk_course_phase_type_required FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_graph fk_from_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT fk_from_course_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: participation_data_dependency_graph fk_from_dto; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY participation_data_dependency_graph
    ADD CONSTRAINT fk_from_dto FOREIGN KEY (from_course_phase_dto_id) REFERENCES course_phase_type_participation_provided_output_dto(id) ON DELETE CASCADE;


--
-- Name: phase_data_dependency_graph fk_from_dto_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY phase_data_dependency_graph
    ADD CONSTRAINT fk_from_dto_phase FOREIGN KEY (from_course_phase_dto_id) REFERENCES course_phase_type_phase_provided_output_dto(id) ON DELETE CASCADE;


--
-- Name: participation_data_dependency_graph fk_from_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY participation_data_dependency_graph
    ADD CONSTRAINT fk_from_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: phase_data_dependency_graph fk_from_phase_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY phase_data_dependency_graph
    ADD CONSTRAINT fk_from_phase_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: course_phase fk_phase_type; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id);


--
-- Name: course_participation fk_student; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT fk_student FOREIGN KEY (student_id) REFERENCES student(id) ON DELETE CASCADE;


--
-- Name: course_phase_graph fk_to_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT fk_to_course_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: participation_data_dependency_graph fk_to_dto; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY participation_data_dependency_graph
    ADD CONSTRAINT fk_to_dto FOREIGN KEY (to_course_phase_dto_id) REFERENCES course_phase_type_participation_required_input_dto(id) ON DELETE CASCADE;


--
-- Name: phase_data_dependency_graph fk_to_dto_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY phase_data_dependency_graph
    ADD CONSTRAINT fk_to_dto_phase FOREIGN KEY (to_course_phase_dto_id) REFERENCES course_phase_type_phase_required_input_dto(id) ON DELETE CASCADE;


--
-- Name: participation_data_dependency_graph fk_to_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY participation_data_dependency_graph
    ADD CONSTRAINT fk_to_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: phase_data_dependency_graph fk_to_phase_phase; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY phase_data_dependency_graph
    ADD CONSTRAINT fk_to_phase_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--
