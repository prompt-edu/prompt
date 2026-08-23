--
-- PostgreSQL database dump
--

-- Dumped from database version 15.2
-- Dumped by pg_dump version 15.8 (Homebrew)

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

--
-- Name: course_type; Type: TYPE; Schema: public; Owner: prompt-postgres
--

CREATE TYPE course_type AS ENUM (
    'lecture',
    'seminar',
    'practical course'
);


--
-- Name: gender; Type: TYPE; Schema: public; Owner: prompt-postgres
--

CREATE TYPE gender AS ENUM (
    'male',
    'female',
    'diverse',
    'prefer_not_to_say'
);



--
-- Name: pass_status; Type: TYPE; Schema: public; Owner: prompt-postgres
--

CREATE TYPE pass_status AS ENUM (
    'passed',
    'failed',
    'not_assessed'
);


--
-- Name: study_degree; Type: TYPE; Schema: public; Owner: prompt-postgres
--

CREATE TYPE study_degree AS ENUM (
    'bachelor',
    'master'
);



--
-- Name: update_last_modified_column(); Type: FUNCTION; Schema: public; Owner: prompt-postgres
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
-- Name: application_answer_multi_select; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE application_answer_multi_select (
    id uuid NOT NULL,
    application_question_id uuid NOT NULL,
    course_phase_participation_id uuid NOT NULL,
    answer text[]
);


--
-- Name: application_answer_text; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE application_answer_text (
    id uuid NOT NULL,
    application_question_id uuid NOT NULL,
    course_phase_participation_id uuid NOT NULL,
    answer text
);


--
-- Name: application_assessment; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE application_assessment (
    id uuid NOT NULL,
    course_phase_participation_id uuid NOT NULL,
    score integer
);


--
-- Name: application_question_multi_select; Type: TABLE; Schema: public; Owner: prompt-postgres
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
    access_key character varying(50)
);


--
-- Name: application_question_text; Type: TABLE; Schema: public; Owner: prompt-postgres
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
    access_key character varying(50)
);


--
-- Name: course; Type: TABLE; Schema: public; Owner: prompt-postgres
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
    CONSTRAINT check_end_date_after_start_date CHECK ((end_date > start_date))
);


--
-- Name: course_participation; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_participation (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL
);


--
-- Name: course_phase; Type: TABLE; Schema: public; Owner: prompt-postgres
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
-- Name: course_phase_graph; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_phase_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL
);


--
-- Name: course_phase_participation; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_phase_participation (
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    restricted_data jsonb,
    pass_status pass_status DEFAULT 'not_assessed'::pass_status,
    last_modified timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    student_readable_data jsonb DEFAULT '{}'::jsonb
);


--
-- Name: course_phase_type; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_phase_type (
    id uuid NOT NULL,
    name text NOT NULL,
    initial_phase boolean DEFAULT false NOT NULL,
    base_url text DEFAULT 'core'::text NOT NULL,
    description text
);


--
-- Name: course_phase_type_provided_output_dto; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_phase_type_provided_output_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    version_number integer NOT NULL,
    endpoint_path text NOT NULL,
    specification jsonb NOT NULL
);

--
-- Name: course_phase_type_required_input_dto; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE course_phase_type_required_input_dto (
    id uuid NOT NULL,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    specification jsonb NOT NULL,
    optional boolean DEFAULT false NOT NULL
);


--
-- Name: meta_data_dependency_graph; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE meta_data_dependency_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL,
    from_course_phase_dto_id uuid NOT NULL,
    to_course_phase_dto_id uuid NOT NULL
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE TABLE schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);



--
-- Name: student; Type: TABLE; Schema: public; Owner: prompt-postgres
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
    last_modified timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    current_semester integer
);


--
-- Data for Name: application_answer_multi_select; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('20293fb0-e09d-4e27-9ade-803aac04de12', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', '{iPhone,MacBook}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('ca957381-e94e-408d-a07b-d2d18b0fe4d4', '65e25b73-ce47-4536-b651-a1632347d733', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', '{Patterns,"Interactive Learning"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('efd16262-9749-4831-a510-ad559d7befa2', '326d1c5c-992f-427d-8d99-e119149b46c0', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('dd7b728d-c544-407b-ad73-ebe9a3f984a9', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('940c28e7-ea99-40df-885e-146f4c7b4772', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', '{"Really nice option"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('e76f90fc-8564-440e-82ce-319e6f7fe4c3', '383a9590-fba2-4e6b-a32b-88895d55fb9b', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', '{iPad,MacBook}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('881b5573-44b1-4a43-95f1-ffa98214721c', '65e25b73-ce47-4536-b651-a1632347d733', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', '{Ferienakademie,Patterns}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('1be68c8c-20cd-4bd9-8654-f76f6d059832', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '{iPhone,iPad,Vision}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('eec0f1ad-32bd-4e17-ac25-6f811d5ebdbe', '65e25b73-ce47-4536-b651-a1632347d733', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '{Ferienakademie,Patterns,"Interactive Learning"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('47c3f72d-bbbb-4bc1-b2b6-a4f8ac05e3eb', '326d1c5c-992f-427d-8d99-e119149b46c0', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('981d4ee5-eb8b-4267-ae24-4b8079e9e8c7', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('252ebd3f-f469-4da8-b73c-54a5c558265d', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '{"Test Option"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('78c30f02-8973-4d2f-8a50-7f43a4d0d9ca', '326d1c5c-992f-427d-8d99-e119149b46c0', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('4e2eeaa3-b906-4d0c-9d2b-c34fd67d1db4', 'e77b6517-3c73-4760-90dd-e9b6283789c4', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('bc2642d5-868b-48c2-981d-a0e3b499485f', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', '{"Some more options"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('74e35105-596b-4579-b754-a7a65849687e', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', '{MacBook,iPad}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('3b2c25dd-a7f5-4486-be52-c9ad7c212d46', '65e25b73-ce47-4536-b651-a1632347d733', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', '{Ferienakademie,Patterns}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('acc0ea77-46be-48db-9413-961a3ac47e37', '326d1c5c-992f-427d-8d99-e119149b46c0', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('9a81eeeb-b18b-4de4-ae14-b145253f3dd9', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('949afd47-e752-4a6c-8d54-e0771b6adbcb', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', '{"Really nice option"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('3f38097f-b009-41a9-aa1f-c34d035c8dca', '383a9590-fba2-4e6b-a32b-88895d55fb9b', '6dbfd180-55a6-4397-8733-b2116c91d8f1', '{iPad,iPhone}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('ba62db03-a9e2-4c52-ba5d-960c5ff3eec1', '65e25b73-ce47-4536-b651-a1632347d733', '6dbfd180-55a6-4397-8733-b2116c91d8f1', '{Ferienakademie,Patterns}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('c85416bc-b82a-4b78-92ce-b858da0b5ed3', '326d1c5c-992f-427d-8d99-e119149b46c0', '6dbfd180-55a6-4397-8733-b2116c91d8f1', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('772a453f-7362-4885-934d-f3cad96b5e36', 'e77b6517-3c73-4760-90dd-e9b6283789c4', '6dbfd180-55a6-4397-8733-b2116c91d8f1', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('2bb5f5af-c258-4c47-842b-9934acca7d92', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '6dbfd180-55a6-4397-8733-b2116c91d8f1', '{"Some more options"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('e4259ec1-cc4f-40ff-84b9-b0cb019434be', '383a9590-fba2-4e6b-a32b-88895d55fb9b', '03ec52ff-7363-4306-b96a-017c40a4d1be', '{iPhone,MacBook}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('55a20be9-d971-40c5-bb59-3b2da5586320', '65e25b73-ce47-4536-b651-a1632347d733', '03ec52ff-7363-4306-b96a-017c40a4d1be', '{Ferienakademie}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('b9ec54cc-35d7-4b53-90ed-c225cb25a160', '326d1c5c-992f-427d-8d99-e119149b46c0', '03ec52ff-7363-4306-b96a-017c40a4d1be', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('b1f921e3-891c-4505-86ce-3e75193c6a31', 'e77b6517-3c73-4760-90dd-e9b6283789c4', '03ec52ff-7363-4306-b96a-017c40a4d1be', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('2fb8cfeb-58b0-4bb4-8e41-7d9f370cf7f9', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '03ec52ff-7363-4306-b96a-017c40a4d1be', '{"Some more options"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('f070a634-3f91-44f7-993d-ea1f9a0636a9', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', '{iPad,iPhone}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('e5938125-f18d-40c6-bb2a-af0ca17600eb', '65e25b73-ce47-4536-b651-a1632347d733', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', '{Patterns,Ferienakademie}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('aad3f8a3-7f0a-48c3-ada0-f114e52aaed4', '326d1c5c-992f-427d-8d99-e119149b46c0', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('d27410dc-6eac-48fe-86c3-9f95fed1cf83', '383a9590-fba2-4e6b-a32b-88895d55fb9b', '2f57b012-af8e-450a-b859-30371b353b0c', '{Vision,iPad}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('2aaa86f7-9a6c-4432-b596-471aa6780d7b', '65e25b73-ce47-4536-b651-a1632347d733', '2f57b012-af8e-450a-b859-30371b353b0c', '{Patterns,"Interactive Learning"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('71ba409b-7113-451e-8be2-82485b5ad457', '326d1c5c-992f-427d-8d99-e119149b46c0', '2f57b012-af8e-450a-b859-30371b353b0c', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('73751332-2dd9-4787-b2cd-b133aa0650d6', 'e77b6517-3c73-4760-90dd-e9b6283789c4', '2f57b012-af8e-450a-b859-30371b353b0c', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('7f0d5125-d427-4d65-b7ce-784310cb8b8a', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '2f57b012-af8e-450a-b859-30371b353b0c', '{"Some more options"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('8f803b2a-0eca-4904-adb4-5addac44ce57', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('207977a6-604f-4aff-86cc-3d9c63e50f3e', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', '{"Test Option"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('f8352710-846d-4a84-85df-3744c4f931db', '383a9590-fba2-4e6b-a32b-88895d55fb9b', '497847d9-17bb-42d1-adfe-2d26fab18f25', '{iPad,Vision}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('a74b86aa-a8c4-4f04-88a3-6b40efc77e55', '65e25b73-ce47-4536-b651-a1632347d733', '497847d9-17bb-42d1-adfe-2d26fab18f25', '{Ferienakademie,Patterns}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('2f5d887d-a81c-42ad-ba18-e848429336d6', '326d1c5c-992f-427d-8d99-e119149b46c0', '497847d9-17bb-42d1-adfe-2d26fab18f25', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('495dc3cd-037e-4582-ac25-81cf77a00499', 'e77b6517-3c73-4760-90dd-e9b6283789c4', '497847d9-17bb-42d1-adfe-2d26fab18f25', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('ac731694-a2e6-405f-8b5a-702ccb508482', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '497847d9-17bb-42d1-adfe-2d26fab18f25', '{"Test Option"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('0ae70dbb-5dd0-445b-b156-9723da12be3a', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'f8693999-4a10-4f8d-be01-416fc8cafd15', '{iPad,iPhone}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('165a238c-49fe-44d2-af82-2578ea0c9f60', '65e25b73-ce47-4536-b651-a1632347d733', 'f8693999-4a10-4f8d-be01-416fc8cafd15', '{Ferienakademie,Patterns}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('2abd391a-17c1-48bc-be9e-8c2fdcd94941', '326d1c5c-992f-427d-8d99-e119149b46c0', 'f8693999-4a10-4f8d-be01-416fc8cafd15', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('589f2098-0478-49f2-a5f7-bbd3f2e5b25e', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'f8693999-4a10-4f8d-be01-416fc8cafd15', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('e6eeb57a-20dd-41b1-9340-7c9a7aa4c2ea', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'f8693999-4a10-4f8d-be01-416fc8cafd15', '{"Some more options"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('6b367c30-16e5-400f-b822-1d88411650e6', '383a9590-fba2-4e6b-a32b-88895d55fb9b', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', '{iPhone,iPad}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('c177e53f-8f17-4d0b-a040-385cc7c49305', '65e25b73-ce47-4536-b651-a1632347d733', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', '{Ferienakademie,Patterns,"Interactive Learning"}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('b904d74f-e83a-4f04-97bf-0155ddfb0750', '326d1c5c-992f-427d-8d99-e119149b46c0', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('73dab542-bc4f-44cb-835a-0805e0367f17', 'e77b6517-3c73-4760-90dd-e9b6283789c4', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', '{Yes}');
INSERT INTO application_answer_multi_select (id, application_question_id, course_phase_participation_id, answer) VALUES ('96398e21-227a-49c3-a86d-4505a0695eb8', '8c5558c7-a1e5-42bc-89c3-d5b8b427396b', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', '{"Some more options"}');


--
-- Data for Name: application_answer_text; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('bc70225e-8c91-47b4-b5af-ec179f657f10', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', 'sdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('c077d50d-1ca8-423d-9879-60563bf200a7', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', 'sdfsdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('517d7d79-b40d-468b-9a81-63839f353558', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '497847d9-17bb-42d1-adfe-2d26fab18f25', 'sdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('8e75d564-6c42-4ec1-b736-f1c1474882b5', 'a6a04042-95d1-4765-8592-caf9560c8c3c', '497847d9-17bb-42d1-adfe-2d26fab18f25', 'dsf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('172a28d5-1788-4e1a-a840-261ab4626bf5', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', 'sdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('93a9f680-ef1c-4b8d-bc93-58c42bb9a910', 'a6a04042-95d1-4765-8592-caf9560c8c3c', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', 'sdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('da93da4f-8f09-40ee-8f1e-0f5c6105e781', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'f8693999-4a10-4f8d-be01-416fc8cafd15', 'werwer');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('a9413247-b411-4ef8-950a-3a0009beb4c3', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'f8693999-4a10-4f8d-be01-416fc8cafd15', 'werwer');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('60bf8eba-1b8b-48c8-84ce-50af80ab41a8', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', 'none');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('a9f622a5-1056-4e2b-9871-fb11241530b3', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', 'sdfdsf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('90c67ea7-7f9c-4e9b-970f-9eaab6a5ee4c', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', 'sdfdsf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('af3aabe8-a77f-4570-8e5f-5838d62ebcd4', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', 'sdfsdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('5b86c0fa-d86e-4c2e-b3f3-28f372f94888', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '6dbfd180-55a6-4397-8733-b2116c91d8f1', 'sdfsdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('204a7e61-ee86-4813-8708-4feba1ceb54f', 'a6a04042-95d1-4765-8592-caf9560c8c3c', '6dbfd180-55a6-4397-8733-b2116c91d8f1', 'sdfsdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('49a647e5-71dc-44c1-8d92-88976bb21afc', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '2f57b012-af8e-450a-b859-30371b353b0c', 'none');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('c4adb64f-71d5-445f-9378-8da85cabdc6f', 'a6a04042-95d1-4765-8592-caf9560c8c3c', '2f57b012-af8e-450a-b859-30371b353b0c', 'Test');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('5479253a-2ead-4169-b540-6b032c613366', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '03ec52ff-7363-4306-b96a-017c40a4d1be', 'sdfsdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('a9a4f01f-1eab-4c5e-81f5-2053b4927285', 'a6a04042-95d1-4765-8592-caf9560c8c3c', '03ec52ff-7363-4306-b96a-017c40a4d1be', 'sdf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('f538b5d0-ebb8-4f29-9538-ac941c054c9a', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', 'test');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('dbb2b70a-fa55-44f5-9973-5161177c14bf', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', 'sdfdsf');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('03e99d43-64f7-4262-a849-6c97a96853a8', 'fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', 'THISSSSSSS');
INSERT INTO application_answer_text (id, application_question_id, course_phase_participation_id, answer) VALUES ('7fda3d95-b745-4765-8cbe-a6a9db083e6c', 'a6a04042-95d1-4765-8592-caf9560c8c3c', 'b0df675c-5fe7-47fd-95b1-eda41012e6b2', 'THISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSSTHISSSSSSS');


--
-- Data for Name: application_assessment; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('965cc4c5-50d5-47db-8cd3-c87fe8e3fb7b', '03ec52ff-7363-4306-b96a-017c40a4d1be', 1);
INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('28737e9f-be3c-4bed-8f0a-b8526c038ab0', 'f8693999-4a10-4f8d-be01-416fc8cafd15', 10);
INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('cdc3107b-0115-4b07-99c2-ca67c9a938f5', 'eb16edff-6854-4c76-917c-fcf8cb746cfb', 1);
INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('a5f4a560-9ee7-44fa-8a9e-c6141c4d6b5c', 'a6567808-d20b-45fa-a24f-1a1f52e194c1', 2);
INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('df02df1c-811b-42e8-9b3a-f66f5cd38f7f', 'e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', 4);
INSERT INTO application_assessment (id, course_phase_participation_id, score) VALUES ('df02df1c-811b-42e8-9b3a-f66f5cd38f7a', '70415cdb-fc1e-495e-bbfe-ddf35987cebe', 4);


--
-- Data for Name: application_question_multi_select; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key) VALUES ('383a9590-fba2-4e6b-a32b-88895d55fb9b', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Available Devices', '', '', '', true, 2, 3, '{iPhone,iPad,MacBook,Vision}', 2, false, '');
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key) VALUES ('65e25b73-ce47-4536-b651-a1632347d733', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Taken Courses', '<p class="text-node"><strong>Which courses have you already taken ad the chair</strong></p><p class="text-node"></p><p class="text-node">This is a link to the <a class="link" href="https://ase.cit.tum.de" target="">DataSettings</a></p><p class="text-node"></p>', '', '', false, 0, 3, '{Ferienakademie,Patterns,"Interactive Learning"}', 4, true, 'availableDevices');
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key) VALUES ('326d1c5c-992f-427d-8d99-e119149b46c0', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Data Rights Acceptance', '<p class="text-node">I herby declare that I accept the data being <a class="link" href="https://test.de" target="">processed</a></p><p class="text-node"></p>', 'CheckBoxQuestion', 'You have to accept th', true, 0, 1, '{Yes}', 5, true, 'test');
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key) VALUES ('8c5558c7-a1e5-42bc-89c3-d5b8b427396b', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'New MultiSelect Test', 'This should also have an asterix', '', '', false, 1, 1, '{"Test Option","Some more options","Really nice option"}', 7, true, 'test');
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key) VALUES ('e77b6517-3c73-4760-90dd-e9b6283789c4', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'This is an interesting Checkbox', '<p class="text-node">The data collected here will only be used for the developer application process of the iPraktikum practical course.</p><p class="text-node">Only the lecturers have access to the personal data.</p><p class="text-node">Consequences in the absence of <a class="link" href="https://ase.cit.tum.de/" target="_blank">consent</a>:</p><p class="text-node">You have the right not to agree to this declaration of consent - however, since the data marked as mandatory fields are required to apply as a developer, we cannot complete your registration without your consent.<br><br>Your rights:</p><ul class="list-node"><li><p class="text-node">a right to information</p></li><li><p class="text-node">a right to correction or deletion or restriction of processing or a right to object to processing</p></li><li><p class="text-node">a right to data portability</p></li><li><p class="text-node">there is also a right of appeal to the Bavarian State Commissioner for Data Protection</p></li></ul><p class="text-node">Consequences of withdrawing consent:</p><p class="text-node">As soon as the cancellation notice is received, your data may not be processed further and will be deleted immediately. Withdrawing your consent does not affect the legality of the processing carried out up to that point. Please send your cancellation to the address below.<br><br>Krusche, Stephan, Technische Universität München, Institut für Informatik I1, Boltzmannstraße 3, 85748 Garching b. München, +49 (89) 289-18212, <a class="link" href="mailto:ios@in.tum.de">ios@in.tum.de</a><br>For further questions concerning data protection, please contact: <a class="link" href="http://www.datenschutz.tum.de">www.datenschutz.tum.de</a></p>', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 6, false, '');


--
-- Data for Name: application_question_text; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key) VALUES ('fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Expierenced7', '<p class="text-node"><strong><em>This again is something interesting</em></strong></p>', 'Please enter your motivation', '', '', true, 10, 1, true, 'experience');
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key) VALUES ('a6a04042-95d1-4765-8592-caf9560c8c3c', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Motivation', 'You should fill out the motivation why you want to take this absolutely amazing course.', 'Enter your motivation.', '', 'You are not allowed to enter more than 500 chars. ', true, 500, 3, false, '');
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key) VALUES ('3f935b9f-a55d-49fa-8cd0-31523f91442e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'This is an amazing question', 'Amazing', 'test', '', 'test', true, 500, 8, false, '');
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key) VALUES ('91c1c7e9-91d3-437f-8062-823b0ea4f4d2', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'test', '', '', '', '', false, 500, 9, false, '');
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key) VALUES ('35687df9-d180-4b37-b71d-69ca419fb8e9', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'New Test', '', '', '', '', false, 500, 10, true, 'test');



INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description) VALUES
    (
        'd7307be2-d3dc-496e-86f0-643bff6cc1c8',
        'iPraktikum',
        '2024-10-13',
        '2024-10-14',
        'ios2425',
        'practical course',
        10,
        '{"icon": "smartphone", "bg-color": "bg-red-100"}',
        '{}',
        FALSE,
        'Two-day iPraktikum intro',
        'Compact iteration used in integration tests.'
    );

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description) VALUES
    (
        'e12ffe63-448d-4469-a840-1699e9b328d1',
        'iPraktikum-Test',
        '2024-12-15',
        '2025-03-14',
        'ios2425',
        'practical course',
        10,
        '{"icon": "smartphone", "bg-color": "bg-blue-100"}',
        '{}',
        FALSE,
        'Extended practice run',
        'Covers template-based flows for integration tests.'
    );

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description) VALUES
    (
        'be780b32-a678-4b79-ae1c-80071771d254',
        'TestCourse',
        '2024-12-19',
        '2025-01-24',
        'ios2425',
        'practical course',
        10,
        '{"icon": "smartphone", "bg-color": "bg-blue-100"}',
        '{}',
        FALSE,
        'Developer bootcamp',
        'Covers onboarding and application data for integration tests.'
    );


--
-- Data for Name: course_participation; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_participation (id, course_id, student_id) VALUES ('1378db54-4ab4-4225-ac67-68e4345f21e2', 'be780b32-a678-4b79-ae1c-80071771d254', '1e41e383-2c3f-4149-bd67-54e41cdaebec');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('2939a7f8-fc0c-4d0a-927a-4251785f97ce', 'be780b32-a678-4b79-ae1c-80071771d254', 'bb9736b4-076f-4592-8197-9a839ac115fb');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('6515f343-6922-45d5-b0f1-3c8188039759', 'be780b32-a678-4b79-ae1c-80071771d254', '2d8c24b4-b91a-4219-9bd8-3f2502774ebc');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('ca41772a-e06d-40eb-9c4b-ab44e06a890c', 'be780b32-a678-4b79-ae1c-80071771d254', '5939210d-5c47-446e-ba55-3da992fd7aa6');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('ef63ff96-baa3-42de-9152-acdaa773d3ee', 'be780b32-a678-4b79-ae1c-80071771d254', '5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('b276ba5f-4522-4af1-800e-e9323978c971', 'be780b32-a678-4b79-ae1c-80071771d254', '2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('6a49b717-a8ca-4d16-bcd0-0bb059525269', 'be780b32-a678-4b79-ae1c-80071771d254', '33f49be1-0106-4642-8a8c-d492c841118a');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('4ac1b89e-36b2-433a-94bf-be7764fac45a', 'be780b32-a678-4b79-ae1c-80071771d254', '9c157166-dd37-42f6-98ab-f5fda439ced1');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('17f4518a-833b-49d5-a512-8f701703852e', 'be780b32-a678-4b79-ae1c-80071771d254', '28880f4d-6f8a-4826-a7c6-e2f295f6ff72');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('f6744410-cfe2-456d-96fa-e857cf989569', 'be780b32-a678-4b79-ae1c-80071771d254', '3869f209-9a21-4595-ae0e-bc6d6a3e2d63');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('395d3e34-67c0-4c0b-9921-60bd85d1b75e', 'be780b32-a678-4b79-ae1c-80071771d254', '402e1535-fb9c-494f-82f2-ab4d39d71155');


--
-- Data for Name: course_phase; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('4e736d05-c125-48f0-8fa0-848b03ca6908', 'be780b32-a678-4b79-ae1c-80071771d254', 'New Intro Course', '{}', false, '48d22f19-6cc0-417b-ac25-415fb40f2030', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('7062236a-e290-487c-be41-29b24e0afc64', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'New Team Phase', '{}', false, '627b6fb9-2106-4fce-ba6d-b68eeb546382', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d3', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'Intro Course', '{}', false, '48d22f19-6cc0-417b-ac25-415fb40f2030', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d2', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'Test LOL 5', '{}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('3a879348-6cac-4d44-b0b9-2bea94198005', 'be780b32-a678-4b79-ae1c-80071771d254', 'New Team Phase', '{}', false, '627b6fb9-2106-4fce-ba6d-b68eeb546382', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1', 'be780b32-a678-4b79-ae1c-80071771d254', 'New Interview', '{}', false, '627b6fb9-2106-4fce-ba6d-b68eeb546383', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('7ffffd38-2454-4c67-821d-5692d8086e6c', 'be780b32-a678-4b79-ae1c-80071771d254', 'New Matching', '{}', false, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abb', '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'be780b32-a678-4b79-ae1c-80071771d254', 'Dev Application', '{"exportAnswers": {"answersText": [{"key": "experience", "questionID": "fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75"}], "additionalScores": [{"key": "tech", "name": "Tech-Challenge"}], "applicationScore": true, "answersMultiSelect": [{"key": "takenCourses", "questionID": "65e25b73-ce47-4536-b651-a1632347d733"}]}, "mailingSettings": {"replyToName": "Program Management Test", "replyToEmail": "niclas@heun.net", "failedMailContent": "<p class=\"text-node\">We are sorry to inform you that you were not selected for the iPraktikum!</p>", "failedMailSubject": "iPraktikum: Not Accepted", "passedMailContent": "<p class=\"text-node\">We are delighted to inform you that we accpeted you to the ipraktikum</p>", "passedMailSubject": "iPraktikum: Accepted", "sendRejectionMail": false, "sendAcceptanceMail": false, "sendConfirmationMail": true, "confirmationMailContent": "<p class=\"text-node\">Dear {{firstName}} {{lastName}},</p><p class=\"text-node\">Thank you for applying to the iPraktikum! We are pleased to confirm that we have received your application.</p><p class=\"text-node\">We’re excited to see your interest in participating as a developer in the iPraktikum program. Below are the details of your submission:</p><ul class=\"list-node\"><li><p class=\"text-node\"><strong>Name:</strong> {{firstName}} {{lastName}}</p></li><li><p class=\"text-node\"><strong>Email:</strong> {{email}}</p></li><li><p class=\"text-node\"><strong>Matriculation Number:</strong> {{matriculationNumber}}</p></li><li><p class=\"text-node\"><strong>TUM-ID:</strong> {{universityLogin}}</p></li><li><p class=\"text-node\"><strong>Study Degree:</strong> {{studyDegree}}</p></li><li><p class=\"text-node\"><strong>Current Semester:</strong> {{currentSemester}}</p></li><li><p class=\"text-node\"><strong>Study Program:</strong> {{studyProgram}}</p></li></ul><h3 class=\"heading-node\">Important Information</h3><ol class=\"list-node\"><li><p class=\"text-node\"><strong>Matching Process:</strong><br>Please remember to prioritize the iPraktikum in your TUM matching preferences to ensure your application is considered.</p></li><li><p class=\"text-node\"><strong>Availability:</strong><br>The iPraktikum will start on <strong>{{courseStartDate}}</strong>. Attendance on-campus is mandatory, so please ensure you are available from the start date.</p></li><li><p class=\"text-node\"><strong>Resubmissions:</strong></p><ul class=\"list-node\"><li><p class=\"text-node\">If you are a TUM student with a TUM-ID ({{universityLogin}}), you can update and resubmit your application until <strong>{{applicationEndDate}}</strong> via the following link:<br><a class=\"link\" href=\"https://{{applicationURL}}\" target=\"_blank\">Application Form.</a></p></li><li><p class=\"text-node\"><strong>Only your last submission will be evaluated.</strong></p></li><li><p class=\"text-node\">Unfortunately, resubmissions are not possible for external students without a TUM-ID.</p></li></ul></li><li><p class=\"text-node\"><strong>Application Feedback:</strong><br>Please note that we do not provide direct feedback on applications. The TUM matching process will determine whether you are assigned to the course.</p></li></ol><p class=\"text-node\">If you have any further questions or need assistance, feel free to reach out to us.</p><p class=\"text-node\">Best regards,<br><strong>The iPraktikum Management Team</strong></p>", "confirmationMailSubject": "Confirmation of Your iPraktikum Application"}, "additionalScores": [{"key": "Tech-Challenge", "name": "Tech-Challenge"}, {"key": "New-Test", "name": "New-Test"}, {"key": "testwithspaces", "name": "Test with spaces"}], "sendRejectionMail": true, "applicationEndDate": "2025-02-14T10:59:00+01:00", "applicationStartDate": "2024-12-09T05:06:00+01:00", "externalStudentsAllowed": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', '{}');


--
-- Data for Name: course_phase_graph; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d2', '7062236a-e290-487c-be41-29b24e0afc64');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id) VALUES ('4e736d05-c125-48f0-8fa0-848b03ca6908', '3a879348-6cac-4d44-b0b9-2bea94198005');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', '2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id) VALUES ('2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1', '7ffffd38-2454-4c67-821d-5692d8086e6c');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id) VALUES ('7ffffd38-2454-4c67-821d-5692d8086e6c', '4e736d05-c125-48f0-8fa0-848b03ca6908');


--
-- Data for Name: course_phase_participation; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('497847d9-17bb-42d1-adfe-2d26fab18f25', 'b276ba5f-4522-4af1-800e-e9323978c971', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'not_assessed', '2025-01-08 15:23:10.318395', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('fc1c6eda-7581-4ec4-bb4a-136eda7e8dd4', 'f6744410-cfe2-456d-96fa-e857cf989569', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"New-Test": 30.00, "comments": [{"text": "test", "author": "Niclas Heun", "timestamp": "2025-01-14T09:20:37.633Z"}], "testwithspaces": 30.00, "student_last_modified": "2025-01-09T18:20:28.256593+01:00", "assessment_last_modified": "2025-01-14T10:20:37.6513+01:00"}', 'failed', '2025-01-14 14:56:51.967771', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('70415cdb-fc1e-495e-bbfe-ddf35987cebe', '6a49b717-a8ca-4d16-bcd0-0bb059525269', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"hasOwnMac": "true", "Tech-Challenge": 20, "testwithspaces": 90.90}', 'passed', '2025-01-14 14:56:51.967771', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('03ec52ff-7363-4306-b96a-017c40a4d1be', 'ca41772a-e06d-40eb-9c4b-ab44e06a890c', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"Tech-Challenge": 20, "assessment_last_modified": "2025-01-12T13:19:30.0747+01:00"}', 'passed', '2025-01-12 13:20:03.937483', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('83d88b1f-1435-4c36-b8ca-6741094f35e4', '6a49b717-a8ca-4d16-bcd0-0bb059525269', '4e736d05-c125-48f0-8fa0-848b03ca6908', '{"proficiency level": "test"}', 'passed', '2025-01-12 13:20:12.912591', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('6dbfd180-55a6-4397-8733-b2116c91d8f1', '2939a7f8-fc0c-4d0a-927a-4251785f97ce', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'not_assessed', '2025-01-07 18:21:19.97367', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('2f57b012-af8e-450a-b859-30371b353b0c', '6515f343-6922-45d5-b0f1-3c8188039759', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'not_assessed', '2025-01-07 18:21:19.97367', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('b0df675c-5fe7-47fd-95b1-eda41012e6b2', '395d3e34-67c0-4c0b-9921-60bd85d1b75e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"student_last_modified": "2025-01-14T09:48:07.641231+01:00"}', 'not_assessed', '2025-01-14 09:48:07.641231', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('f8693999-4a10-4f8d-be01-416fc8cafd15', '4ac1b89e-36b2-433a-94bf-be7764fac45a', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"comments": [{"text": "Test", "author": "Niclas Heun", "timestamp": "2025-01-14T09:30:54.475Z"}], "assessment_last_modified": "2025-01-14T10:30:54.48701+01:00"}', 'passed', '2025-01-14 10:30:54.48701', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('eb16edff-6854-4c76-917c-fcf8cb746cfb', 'ef63ff96-baa3-42de-9152-acdaa773d3ee', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"assessment_last_modified": "2025-01-14T10:51:19.995136+01:00"}', 'not_assessed', '2025-01-14 10:51:19.995136', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('a6567808-d20b-45fa-a24f-1a1f52e194c1', '17f4518a-833b-49d5-a512-8f701703852e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"assessment_last_modified": "2025-01-14T10:51:28.292706+01:00"}', 'not_assessed', '2025-01-14 10:51:28.292706', '{}');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data) VALUES ('e64843b5-ab0b-4deb-82fc-6cf763f8aa0e', '1378db54-4ab4-4225-ac67-68e4345f21e2', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{"comments": [{"text": "test", "author": "Niclas Heun", "timestamp": "2025-01-14T09:21:29.786Z"}], "assessment_last_modified": "2025-01-14T11:00:06.150053+01:00"}', 'not_assessed', '2025-01-14 11:00:06.150053', '{}');


--
-- Data for Name: course_phase_type; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_type (id, name, initial_phase, base_url, description) VALUES ('48d22f19-6cc0-417b-ac25-415fb40f2030', 'Intro Course', false, 'core', 'Introduces participants to course expectations');
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description) VALUES ('627b6fb9-2106-4fce-ba6d-b68eeb546382', 'Team Phase', false, 'core', 'Facilitates team collaboration');
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description) VALUES ('96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'Application', true, 'core', 'Collects applications and materials');
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description) VALUES ('96fb1001-b21c-4527-8b6f-2fd5f4ba3abb', 'Matching', false, 'core', 'Matches students to projects');
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description) VALUES ('627b6fb9-2106-4fce-ba6d-b68eeb546383', 'Interview', false, 'core', 'Runs interview style assessments');


--
-- Data for Name: course_phase_type_provided_output_dto; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_type_provided_output_dto (id, course_phase_type_id, dto_name, version_number, endpoint_path, specification) VALUES ('19b81ea0-a9ad-4040-ba2e-9f21081b2b30', '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'score', 1, 'core', '{"type": "integer"}');
INSERT INTO course_phase_type_provided_output_dto (id, course_phase_type_id, dto_name, version_number, endpoint_path, specification) VALUES ('c36c1d63-8cf5-4f32-a5c6-8ed9794f0896', '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'applicationAnswers', 1, 'core', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');
INSERT INTO course_phase_type_provided_output_dto (id, course_phase_type_id, dto_name, version_number, endpoint_path, specification) VALUES ('b3dbdb12-0a59-4b57-a15a-28d63cf701e8', '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'additionalScores', 1, 'core', '{"type": "array", "items": {"type": "object", "required": ["score", "key"], "properties": {"key": {"type": "string"}, "score": {"type": "number"}}}}');
INSERT INTO course_phase_type_provided_output_dto (id, course_phase_type_id, dto_name, version_number, endpoint_path, specification) VALUES ('837794af-673a-4ae7-97b8-b91270461500', '627b6fb9-2106-4fce-ba6d-b68eeb546383', 'score', 1, 'core', '{"type": "integer"}');


--
-- Data for Name: course_phase_type_required_input_dto; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_type_required_input_dto (id, course_phase_type_id, dto_name, specification) VALUES ('cfb7b9ff-8a1a-4c45-b0b4-716073cf4463', '96fb1001-b21c-4527-8b6f-2fd5f4ba3abb', 'score', '{"type": "integer"}');
INSERT INTO course_phase_type_required_input_dto (id, course_phase_type_id, dto_name, specification) VALUES ('d0fe539a-f958-4516-a19a-402d30038be6', '627b6fb9-2106-4fce-ba6d-b68eeb546383', 'score', '{"type": "integer"}');
INSERT INTO course_phase_type_required_input_dto (id, course_phase_type_id, dto_name, specification) VALUES ('9178c9dd-4356-4f8e-984d-02f76540d55d', '627b6fb9-2106-4fce-ba6d-b68eeb546383', 'applicationAnswers', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');


--
-- Data for Name: meta_data_dependency_graph; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO meta_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_dto_id, to_course_phase_dto_id) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', '2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1', 'c36c1d63-8cf5-4f32-a5c6-8ed9794f0896', '9178c9dd-4356-4f8e-984d-02f76540d55d');
INSERT INTO meta_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_dto_id, to_course_phase_dto_id) VALUES ('2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1', '7ffffd38-2454-4c67-821d-5692d8086e6c', '837794af-673a-4ae7-97b8-b91270461500', 'cfb7b9ff-8a1a-4c45-b0b4-716073cf4463');
INSERT INTO meta_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_dto_id, to_course_phase_dto_id) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', '2b1a55ad-8b1d-453f-b2b4-2373ecb35bc1', '19b81ea0-a9ad-4040-ba2e-9f21081b2b30', 'd0fe539a-f958-4516-a19a-402d30038be6');


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO schema_migrations (version, dirty) VALUES (9, false);


--
-- Data for Name: student; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('3869f209-9a21-4595-ae0e-bc6d6a3e2d63', 'Niclas', 'Heun', 'niclas.heun@tum.de', '03711126', 'ge25hok', true, 'male', 'DE', 'Computer Science', 'master', '2025-01-09 18:20:28.256593', 19);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('402e1535-fb9c-494f-82f2-ab4d39d71155', 'super long text', 'Text test', 'test@supertest.de', '', '', false, 'prefer_not_to_say', 'AI', 'test', 'bachelor', '2025-01-14 09:48:07.641231', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07', 'Niclas', 'Heun', 'niclas@heun.ent', '08888888', 'uu66uuu', true, 'male', 'DE', 'Computer Science', 'master', '2025-01-08 14:30:29.553373', 12);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('bb9736b4-076f-4592-8197-9a839ac115fb', 'Test User', 'User-Last_Nam', 'niclas10@test.de', '09111999', 'ge77hok', true, 'diverse', NULL, NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('afd0cc74-7218-4f3a-8c2b-ef88aa5a9b1e', 'Test User', 'TEst', 'test@test.de', '09987652', 'ab00lll', true, 'diverse', NULL, NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('ac3d8139-723f-4d8f-89c8-b7171b41b0d3', 'Test', 'test', 'test3@test3.de', '05555555', 'hh77hhh', true, 'female', NULL, NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('9c157166-dd37-42f6-98ab-f5fda439ced1', 'Niclas', 'Heun', 'niclas@heun.net', '', '', false, 'diverse', 'AD', 'Information Systems', 'bachelor', '2025-01-08 18:29:02.512604', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee', 'Test', 'Test', 'test2@test.de', '04511126', 'ge88hok', true, 'diverse', 'DZ', 'Information Systems', 'bachelor', '2025-01-08 18:33:36.571565', 10);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('33f49be1-0106-4642-8a8c-d492c841118a', 'Test', 'Supertest', 'supertest@test.de', '08888889', 'ab00hhh', true, 'diverse', 'AD', 'Information Systems', 'bachelor', '2025-01-08 18:36:09.56128', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('777286f4-a3e7-4bcd-bf57-13d178bf582d', 'New Test', 'user', 'niclas@heun.io', '09999222', 'oo55ooo', true, 'female', 'DZ', 'Computer Science', 'bachelor', '2025-01-08 18:39:34.368859', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('1e41e383-2c3f-4149-bd67-54e41cdaebec', 'test-lol', 'test', 'niclas@test.de', '09999999', 'as45fgh', true, 'female', 'AL', 'Information Systems', 'bachelor', '2025-01-08 18:42:55.407307', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('28880f4d-6f8a-4826-a7c6-e2f295f6ff72', 'Stefan', 'Heun', 'stefan@heun.io', '', '', false, 'male', 'DE', 'Computer Science', 'master', '2025-01-08 20:55:17.282497', 12);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('23bf3123-4f0d-473c-9ef5-d0333e29fe9a', 'Max', 'Mustermann', 'max.mustermann@tum.de', '09822222', 'ji79klj', true, 'male', NULL, NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('6381660e-6f30-4632-bfd0-5b7dd92c6fcf', 'test', 'student', 'looool@tum.de', '', '', false, 'diverse', 'AL', NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('5f2c6b09-b170-48a1-a6cf-30c7688df1f4', 'Nationality Test', 'Registered User', 'nation@test.de', '08877663', 'ab69lol', true, 'male', 'DE', NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('65828d3e-11bc-4168-8edc-3968b53f4f83', 'External ', 'TEst', 'amazing@test.de', '', '', false, 'female', 'DK', NULL, 'bachelor', '2025-01-07 18:21:19.97367', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('2d8c24b4-b91a-4219-9bd8-3f2502774ebc', 'Test-100', 'User-Update', 'user@test.de', '', '', false, 'diverse', 'AL', NULL, 'bachelor', '2025-01-07 18:22:05.814353', NULL);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('1c62c564-491b-43e3-9929-7be39509e32e', 'Niclas', 'Heun', 'test@leeeeeel.de', '00000000', 'hh88hhh', true, 'female', 'DE', 'Computer Science', 'master', '2025-01-07 22:50:17.814704', 5);
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, last_modified, current_semester) VALUES ('5939210d-5c47-446e-ba55-3da992fd7aa6', 'Niclas', 'Heuni', 'heuni@heuni.de', '', '', false, 'prefer_not_to_say', 'DE', 'Information Systems', 'bachelor', '2025-01-07 23:05:43.120086', 5);


--
-- Name: application_answer_multi_select application_answer_multi_select_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT application_answer_multi_select_pkey PRIMARY KEY (id);


--
-- Name: application_answer_text application_answer_text_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT application_answer_text_pkey PRIMARY KEY (id);


--
-- Name: application_assessment application_assessment_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT application_assessment_pkey PRIMARY KEY (id);


--
-- Name: application_question_multi_select application_question_multi_select_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT application_question_multi_select_pkey PRIMARY KEY (id);


--
-- Name: application_question_text application_question_text_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT application_question_text_pkey PRIMARY KEY (id);


--
-- Name: course_participation course_participation_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT course_participation_pkey PRIMARY KEY (id);


--
-- Name: course_phase_participation course_phase_participation_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT course_phase_participation_pkey PRIMARY KEY (id);


--
-- Name: course_phase course_phase_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT course_phase_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type course_phase_type_name_key; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type
    ADD CONSTRAINT course_phase_type_name_key UNIQUE (name);


--
-- Name: course_phase_type course_phase_type_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type
    ADD CONSTRAINT course_phase_type_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type_provided_output_dto course_phase_type_provided_output_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type_provided_output_dto
    ADD CONSTRAINT course_phase_type_provided_output_dto_pkey PRIMARY KEY (id);


--
-- Name: course_phase_type_required_input_dto course_phase_type_required_input_dto_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type_required_input_dto
    ADD CONSTRAINT course_phase_type_required_input_dto_pkey PRIMARY KEY (id);


--
-- Name: course course_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course
    ADD CONSTRAINT course_pkey PRIMARY KEY (id);


--
-- Name: meta_data_dependency_graph meta_data_dependency_graph_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY meta_data_dependency_graph
    ADD CONSTRAINT meta_data_dependency_graph_pkey PRIMARY KEY (to_course_phase_id, to_course_phase_dto_id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: student student_email_key; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY student
    ADD CONSTRAINT student_email_key UNIQUE (email);


--
-- Name: student student_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY student
    ADD CONSTRAINT student_pkey PRIMARY KEY (id);


--
-- Name: application_answer_multi_select unique_application_answer_multi_select; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT unique_application_answer_multi_select UNIQUE (course_phase_participation_id, application_question_id);


--
-- Name: application_answer_text unique_application_answer_text; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT unique_application_answer_text UNIQUE (course_phase_participation_id, application_question_id);


--
-- Name: course unique_course_identifier; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course
    ADD CONSTRAINT unique_course_identifier UNIQUE (name, semester_tag);


--
-- Name: course_participation unique_course_participation; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT unique_course_participation UNIQUE (course_id, student_id);


--
-- Name: course_phase_participation unique_course_phase_participation; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT unique_course_phase_participation UNIQUE (course_participation_id, course_phase_id);


--
-- Name: application_assessment unique_course_phase_participation_assessment; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT unique_course_phase_participation_assessment UNIQUE (course_phase_participation_id);


--
-- Name: course_phase_graph unique_from_course_phase; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT unique_from_course_phase UNIQUE (from_course_phase_id);


--
-- Name: course_phase_graph unique_to_course_phase; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT unique_to_course_phase UNIQUE (to_course_phase_id);


--
-- Name: student_matriculation_number_unique; Type: INDEX; Schema: public; Owner: prompt-postgres
--

CREATE UNIQUE INDEX student_matriculation_number_unique ON student USING btree (matriculation_number) WHERE ((matriculation_number IS NOT NULL) AND ((matriculation_number)::text <> ''::text));


--
-- Name: student_university_login_unique; Type: INDEX; Schema: public; Owner: prompt-postgres
--

CREATE UNIQUE INDEX student_university_login_unique ON student USING btree (university_login) WHERE ((university_login IS NOT NULL) AND ((university_login)::text <> ''::text));


--
-- Name: unique_initial_phase_per_course; Type: INDEX; Schema: public; Owner: prompt-postgres
--

CREATE UNIQUE INDEX unique_initial_phase_per_course ON course_phase USING btree (course_id) WHERE (is_initial_phase = true);


--
-- Name: course_phase_participation set_last_modified_course_phase_participation; Type: TRIGGER; Schema: public; Owner: prompt-postgres
--

CREATE TRIGGER set_last_modified_course_phase_participation BEFORE UPDATE ON course_phase_participation FOR EACH ROW EXECUTE FUNCTION update_last_modified_column();


--
-- Name: student set_last_modified_student; Type: TRIGGER; Schema: public; Owner: prompt-postgres
--

CREATE TRIGGER set_last_modified_student BEFORE UPDATE ON student FOR EACH ROW EXECUTE FUNCTION update_last_modified_column();


--
-- Name: application_answer_text fk_application_question; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT fk_application_question FOREIGN KEY (application_question_id) REFERENCES application_question_text(id) ON DELETE CASCADE;


--
-- Name: application_answer_multi_select fk_application_question; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT fk_application_question FOREIGN KEY (application_question_id) REFERENCES application_question_multi_select(id) ON DELETE CASCADE;


--
-- Name: course_phase fk_course; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course(id) ON DELETE CASCADE;


--
-- Name: course_participation fk_course; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course(id) ON DELETE CASCADE;


--
-- Name: course_phase_participation fk_course_participation; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT fk_course_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE;


--
-- Name: course_phase_participation fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_participation
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: application_question_text fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: application_question_multi_select fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: application_answer_text fk_course_phase_participation; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT fk_course_phase_participation FOREIGN KEY (course_phase_participation_id) REFERENCES course_phase_participation(id) ON DELETE CASCADE;


--
-- Name: application_answer_multi_select fk_course_phase_participation; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT fk_course_phase_participation FOREIGN KEY (course_phase_participation_id) REFERENCES course_phase_participation(id) ON DELETE CASCADE;


--
-- Name: application_assessment fk_course_phase_participation; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT fk_course_phase_participation FOREIGN KEY (course_phase_participation_id) REFERENCES course_phase_participation(id) ON DELETE CASCADE;


--
-- Name: course_phase_type_provided_output_dto fk_course_phase_type_provided; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type_provided_output_dto
    ADD CONSTRAINT fk_course_phase_type_provided FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_type_required_input_dto fk_course_phase_type_required; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_type_required_input_dto
    ADD CONSTRAINT fk_course_phase_type_required FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id) ON DELETE CASCADE;


--
-- Name: course_phase_graph fk_from_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT fk_from_course_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: meta_data_dependency_graph fk_from_dto; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY meta_data_dependency_graph
    ADD CONSTRAINT fk_from_dto FOREIGN KEY (from_course_phase_dto_id) REFERENCES course_phase_type_provided_output_dto(id) ON DELETE CASCADE;


--
-- Name: meta_data_dependency_graph fk_from_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY meta_data_dependency_graph
    ADD CONSTRAINT fk_from_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: course_phase fk_phase_type; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id);


--
-- Name: course_participation fk_student; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT fk_student FOREIGN KEY (student_id) REFERENCES student(id) ON DELETE CASCADE;


--
-- Name: course_phase_graph fk_to_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase_graph
    ADD CONSTRAINT fk_to_course_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- Name: meta_data_dependency_graph fk_to_dto; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY meta_data_dependency_graph
    ADD CONSTRAINT fk_to_dto FOREIGN KEY (to_course_phase_dto_id) REFERENCES course_phase_type_required_input_dto(id) ON DELETE CASCADE;


--
-- Name: meta_data_dependency_graph fk_to_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY meta_data_dependency_graph
    ADD CONSTRAINT fk_to_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

-------------------------------
-- 1. Adjust course_phase_participation
-------------------------------
-- Rename the surrogate primary key column so we can still reference its values.
ALTER TABLE course_phase_participation
  RENAME COLUMN id TO old_id;

-------------------------------
-- 2. Adjust application_answer_text
-------------------------------
-- (a) Add new columns for the composite foreign key.
ALTER TABLE application_answer_text
  ADD COLUMN new_course_participation_id uuid;

-- (b) Populate the new columns using the mapping from course_phase_participation.
UPDATE application_answer_text a
SET new_course_participation_id = cp.course_participation_id
FROM course_phase_participation cp
WHERE a.course_phase_participation_id = cp.old_id;

ALTER TABLE application_answer_text
  ALTER COLUMN new_course_participation_id SET NOT NULL;

-- (c) Drop the old foreign key and unique constraints.
ALTER TABLE application_answer_text
  DROP CONSTRAINT fk_course_phase_participation,
  DROP CONSTRAINT unique_application_answer_text;

-- (d) Remove the old surrogate column.
ALTER TABLE application_answer_text
  DROP COLUMN course_phase_participation_id;

-- (e) Rename the new columns to the desired names.
ALTER TABLE application_answer_text
  RENAME COLUMN new_course_participation_id TO course_participation_id;

-- (f) Add a new foreign key constraint on the composite columns.
ALTER TABLE application_answer_text
  ADD CONSTRAINT fk_course_participation
    FOREIGN KEY (course_participation_id)
    REFERENCES course_participation(id) ON DELETE CASCADE;

-- (g) Recreate a unique constraint that now uses the two foreign key columns.
ALTER TABLE application_answer_text
  ADD CONSTRAINT unique_application_answer_text
    UNIQUE (course_participation_id, application_question_id);

-------------------------------
-- 3. Adjust application_answer_multi_select
-------------------------------
-- (a) Add new columns.
ALTER TABLE application_answer_multi_select
  ADD COLUMN new_course_participation_id uuid;

-- (b) Populate the new columns.
UPDATE application_answer_multi_select a
SET new_course_participation_id = cp.course_participation_id
FROM course_phase_participation cp
WHERE a.course_phase_participation_id = cp.old_id;

ALTER TABLE application_answer_multi_select
  ALTER COLUMN new_course_participation_id SET NOT NULL;

-- (c) Drop the old constraints.
ALTER TABLE application_answer_multi_select
  DROP CONSTRAINT fk_course_phase_participation,
  DROP CONSTRAINT unique_application_answer_multi_select;

-- (d) Drop the old column.
ALTER TABLE application_answer_multi_select
  DROP COLUMN course_phase_participation_id;

-- (e) Rename new columns.
ALTER TABLE application_answer_multi_select
  RENAME COLUMN new_course_participation_id TO course_participation_id;

-- (f) Add the new foreign key.
ALTER TABLE application_answer_multi_select
  ADD CONSTRAINT fk_course_participation
    FOREIGN KEY (course_participation_id)
    REFERENCES course_participation(id) ON DELETE CASCADE;

-- (g) Recreate the unique constraint.
ALTER TABLE application_answer_multi_select
  ADD CONSTRAINT unique_application_answer_multi_select
    UNIQUE (course_participation_id, application_question_id);

-------------------------------
-- 4. Adjust application_assessment
-------------------------------
-- (a) Add new columns.
ALTER TABLE application_assessment
  ADD COLUMN new_course_phase_id uuid,
  ADD COLUMN new_course_participation_id uuid;

-- (b) Populate the new columns.
UPDATE application_assessment a
SET new_course_phase_id = cp.course_phase_id,
    new_course_participation_id = cp.course_participation_id
FROM course_phase_participation cp
WHERE a.course_phase_participation_id = cp.old_id;

ALTER TABLE application_assessment
  ALTER COLUMN new_course_phase_id SET NOT NULL,
  ALTER COLUMN new_course_participation_id SET NOT NULL;

-- (c) Drop the old foreign key constraint.
ALTER TABLE application_assessment
  DROP CONSTRAINT fk_course_phase_participation;

-- (d) Drop the old surrogate column.
ALTER TABLE application_assessment
  DROP COLUMN course_phase_participation_id;

-- (e) Rename the new columns.
ALTER TABLE application_assessment
  RENAME COLUMN new_course_phase_id TO course_phase_id;

ALTER TABLE application_assessment
  RENAME COLUMN new_course_participation_id TO course_participation_id;

-------------------------------
-- 5. Final Cleanup
-------------------------------
-- If you no longer need the old surrogate mapping in course_phase_participation,
-- you can drop the old_id column. (Make sure all referencing tables have been updated.)

-- Drop the old primary key constraint (assumed name).
ALTER TABLE course_phase_participation
  DROP CONSTRAINT course_phase_participation_pkey;

-- Add the new composite primary key.
ALTER TABLE course_phase_participation
  ADD PRIMARY KEY (course_participation_id, course_phase_id);

-- (f) Add the new foreign key constraint.
ALTER TABLE application_assessment
  ADD CONSTRAINT fk_course_phase_participation
    FOREIGN KEY (course_participation_id, course_phase_id)
    REFERENCES course_phase_participation (course_participation_id, course_phase_id)
    ON DELETE CASCADE;


ALTER TABLE course_phase_participation
  DROP COLUMN old_id;

-- Rename existing DTO tables to use the "participation" prefix
ALTER TABLE course_phase_type_provided_output_dto 
    RENAME TO course_phase_type_participation_provided_output_dto;

ALTER TABLE course_phase_type_required_input_dto 
    RENAME TO course_phase_type_participation_required_input_dto;

-- Rename the dependency graph table to "participation_data_dependency_graph"
ALTER TABLE meta_data_dependency_graph 
    RENAME TO participation_data_dependency_graph;
