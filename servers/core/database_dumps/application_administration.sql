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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: course_phase_type; Type: TABLE; Schema: public; Owner: prompt-postgres
--

CREATE type gender as enum ('male', 'female', 'diverse', 'prefer_not_to_say');


CREATE TABLE student (
    id uuid NOT NULL,
    first_name character varying(50),
    last_name character varying(50),
    email character varying(255),
    matriculation_number character varying(30),
    university_login character varying(20),
    has_university_account boolean,
    gender gender NOT NULL
);


--
-- Data for Name: student; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('3a774200-39a7-4656-bafb-92b7210a93c1', 'John', 'Doe', 'existingstudent@example.com', '03711111', 'ab12cde', true, 'male');

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('b1f97ee7-fd11-4556-8c75-d0c2714e7082', 'Test', 'Student', 'test@example.com', '03788888', 'ab12cde', true, 'male');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('15ae3969-bcb7-4d5b-8245-c305d13d671b', 'Another', 'Student', 'test@example.com', '03788888', 'ab12cde', true, 'male');



create type course_type as enum ('lecture', 'seminar', 'practical course');


CREATE TABLE course (
    id uuid NOT NULL,
    name text NOT NULL,
    start_date date,
    end_date date,
    semester_tag text,
    course_type course_type NOT NULL,
    ects integer,
    meta_data jsonb
);

CREATE TABLE course_participation (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL
);

ALTER TABLE ONLY course_participation
    ADD CONSTRAINT course_participation_pkey PRIMARY KEY (id);

CREATE TYPE pass_status AS ENUM ('passed', 'failed', 'not_assessed');

CREATE TABLE course_phase_participation (
    id uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    pass_status pass_status,
    meta_data jsonb
);

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
-- Data for Name: course; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, meta_data) VALUES ('be780b32-a678-4b79-ae1c-80071771d254', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{"icon": "apple", "bg-color": "bg-orange-100"}');



CREATE TABLE course_phase_type (
    id uuid NOT NULL,
    name text NOT NULL,
    required_input_meta_data jsonb DEFAULT '{}'::jsonb NOT NULL,
    provided_output_meta_data jsonb DEFAULT '{}'::jsonb NOT NULL,
  initial_phase boolean DEFAULT false NOT NULL,
  description text
);

--
-- Data for Name: course_phase_type; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase_type (id, name, required_input_meta_data, provided_output_meta_data, initial_phase, description) VALUES ('48d22f19-6cc0-417b-ac25-415fb40f2030', 'Intro Course', '[{"name": "hasOwnMac", "type": "boolean"}]', '[{"name": "proficiency level", "type": "string"}]', false, 'Introduces course basics');
INSERT INTO course_phase_type (id, name, required_input_meta_data, provided_output_meta_data, initial_phase, description) VALUES ('96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'Application', '[]', '[{"name": "hasOwnMac", "type": "boolean"}, {"name": "devices", "type": "array"}]', true, 'Handles application intake');
INSERT INTO course_phase_type (id, name, required_input_meta_data, provided_output_meta_data, initial_phase, description) VALUES ('627b6fb9-2106-4fce-ba6d-b68eeb546382', 'Team Phase', '[{"name": "proficiency level", "type": "string"}, {"name": "devices", "type": "array"}]', '[]', false, 'Runs team collaboration phase');




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


CREATE TABLE course_phase (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    name text,
    meta_data jsonb,
    is_initial_phase boolean NOT NULL,
    course_phase_type_id uuid NOT NULL
);

--
-- Data for Name: course_phase; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO course_phase (id, course_id, name, meta_data, is_initial_phase, course_phase_type_id) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'be780b32-a678-4b79-ae1c-80071771d254', 'Dev Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": false, "universityLoginAvailable": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');
INSERT INTO course_phase (id, course_id, name, meta_data, is_initial_phase, course_phase_type_id) VALUES ('7062236a-e290-487c-be41-29b24e0afc64', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'New Team Phase', '{}', false, '627b6fb9-2106-4fce-ba6d-b68eeb546382');
INSERT INTO course_phase (id, course_id, name, meta_data, is_initial_phase, course_phase_type_id) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d3', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'Intro Course', '{}', false, '48d22f19-6cc0-417b-ac25-415fb40f2030');


--
-- Name: course_phase course_phase_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT course_phase_pkey PRIMARY KEY (id);


--
-- Name: unique_initial_phase_per_course; Type: INDEX; Schema: public; Owner: prompt-postgres
--

CREATE UNIQUE INDEX unique_initial_phase_per_course ON course_phase USING btree (course_id) WHERE (is_initial_phase = true);


--
-- Name: course_phase fk_phase_type; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY course_phase
    ADD CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type(id);


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
    order_num integer
);


--
-- Data for Name: application_question_multi_select; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num) VALUES ('65e25b73-ce47-4536-b651-a1632347d733', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Taken Courses', 'Which courses have you already taken ad the chair', '', '', false, 0, 3, '{Ferienakademie,Patterns,"Interactive Learning"}', 4);
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num) VALUES ('383a9590-fba2-4e6b-a32b-88895d55fb9b', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Available Devices', '', '', '', false, 0, 4, '{iPhone,iPad,MacBook,Vision}', 2);


--
-- Name: application_question_multi_select application_question_multi_select_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT application_question_multi_select_pkey PRIMARY KEY (id);


--
-- Name: application_question_multi_select fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_multi_select
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


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
    order_num integer
);

--
-- Data for Name: application_question_text; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--

INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num) VALUES ('a6a04042-95d1-4765-8592-caf9560c8c3c', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Motivation', 'You should fill out the motivation why you want to take this absolutely amazing course.', 'Enter your motivation.', '', 'You are not allowed to enter more than 500 chars. ', true, 500, 3);
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num) VALUES ('fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Expierence', '', '', '', '', false, 500, 1);


--
-- Name: application_question_text application_question_text_pkey; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT application_question_text_pkey PRIMARY KEY (id);


--
-- Name: application_question_text fk_course_phase; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_question_text
    ADD CONSTRAINT fk_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

ALTER TABLE ONLY application_answer_multi_select
    ADD CONSTRAINT unique_application_answer_multi_select UNIQUE (course_phase_participation_id, application_question_id);


--
-- Name: application_answer_text unique_application_answer_text; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_answer_text
    ADD CONSTRAINT unique_application_answer_text UNIQUE (course_phase_participation_id, application_question_id);



INSERT INTO course_participation (id, course_id, student_id) VALUES ('82d7efae-d545-4cc5-9b94-5d0ee1e50d25', 'be780b32-a678-4b79-ae1c-80071771d254', 'b1f97ee7-fd11-4556-8c75-d0c2714e7082');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('32aa070e-67c3-4a69-852a-ba3b5e849a4d', 'be780b32-a678-4b79-ae1c-80071771d254', '15ae3969-bcb7-4d5b-8245-c305d13d671b');


INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, meta_data, pass_status) VALUES ('0c58232d-1a67-44e6-b4dc-69e95373b976', '82d7efae-d545-4cc5-9b94-5d0ee1e50d25', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'passed');
INSERT INTO course_phase_participation (id, course_participation_id, course_phase_id, meta_data, pass_status) VALUES ('f5e61de3-6b6a-494e-a0ac-a18f1f9262e1', '32aa070e-67c3-4a69-852a-ba3b5e849a4d', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'not_assessed');

CREATE TABLE application_assessment (
    id uuid NOT NULL,
    course_phase_participation_id uuid NOT NULL,
    score integer
);


ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT application_assessment_pkey PRIMARY KEY (id);


--
-- Name: application_assessment unique_course_phase_participation_assessment; Type: CONSTRAINT; Schema: public; Owner: prompt-postgres
--

ALTER TABLE ONLY application_assessment
    ADD CONSTRAINT unique_course_phase_participation_assessment UNIQUE (course_phase_participation_id);


--
-- Name: application_assessment fk_course_phase_participation; Type: FK CONSTRAINT; Schema: public; Owner: prompt-postgres
--


ALTER TABLE student
    ADD COLUMN nationality VARCHAR(2);

CREATE TYPE study_degree AS ENUM (
  'bachelor',
  'master'
);

ALTER TABLE student
ADD COLUMN study_program varchar(100),
ADD COLUMN study_degree study_degree NOT NULL DEFAULT 'bachelor',
ADD COLUMN current_semester int,
ADD COLUMN last_modified timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE course_phase_participation
ADD COLUMN last_modified timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP;


-- Create the function to update the last_modified column
CREATE OR REPLACE FUNCTION update_last_modified_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.last_modified = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create triggers to automatically update last_modified
CREATE TRIGGER set_last_modified_student
BEFORE UPDATE ON student
FOR EACH ROW
EXECUTE FUNCTION update_last_modified_column();

CREATE TRIGGER set_last_modified_course_phase_participation
BEFORE UPDATE ON course_phase_participation
FOR EACH ROW
EXECUTE FUNCTION update_last_modified_column();

CREATE TABLE meta_data_dependency_graph (
    from_phase_id uuid NOT NULL,
    to_phase_id   uuid NOT NULL,
    PRIMARY KEY (from_phase_id, to_phase_id),
    CONSTRAINT fk_from_phase
      FOREIGN KEY (from_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE,
    CONSTRAINT fk_to_phase
      FOREIGN KEY (to_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE
);

-- Add new fields to application_question_text
ALTER TABLE application_question_text
ADD COLUMN accessible_for_other_phases boolean DEFAULT false,
ADD COLUMN access_key VARCHAR(50);

-- Add new fields to application_question_multi_select
ALTER TABLE application_question_multi_select
ADD COLUMN accessible_for_other_phases boolean DEFAULT false,
ADD COLUMN access_key VARCHAR(50);


-- Apply migration
-- 1) Adjust course
ALTER TABLE course
RENAME COLUMN meta_data TO restricted_data;

ALTER TABLE course
ADD COLUMN student_readable_data jsonb DEFAULT '{}';

ALTER TABLE course
ADD COLUMN short_description VARCHAR(255);

ALTER TABLE course
ADD COLUMN long_description TEXT;

-- 2) Adjust course_phase
ALTER TABLE course_phase
RENAME COLUMN meta_data TO restricted_data;

ALTER TABLE course_phase
ADD COLUMN student_readable_data jsonb DEFAULT '{}';

-- 3) Adjust course_phase_participation
ALTER TABLE course_phase_participation
RENAME COLUMN meta_data TO restricted_data;

ALTER TABLE course_phase_participation
ADD COLUMN student_readable_data jsonb DEFAULT '{}';


-- CoursePhaseParticipationID Migration
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
  DROP CONSTRAINT unique_application_answer_text;

-- (d) Remove the old surrogate column.
ALTER TABLE application_answer_text
  DROP COLUMN course_phase_participation_id;

-- (e) Rename the new columns to the desired names.
ALTER TABLE application_answer_text
  RENAME COLUMN new_course_participation_id TO course_participation_id;

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
  DROP CONSTRAINT unique_application_answer_multi_select;

-- (d) Drop the old column.
ALTER TABLE application_answer_multi_select
  DROP COLUMN course_phase_participation_id;

-- (e) Rename new columns.
ALTER TABLE application_answer_multi_select
  RENAME COLUMN new_course_participation_id TO course_participation_id;

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

-- Add the new composite primary key.
ALTER TABLE course_phase_participation
  ADD PRIMARY KEY (course_participation_id, course_phase_id);


ALTER TABLE application_assessment
  ADD CONSTRAINT fk_course_phase_participation
    FOREIGN KEY (course_participation_id, course_phase_id)
    REFERENCES course_phase_participation (course_participation_id, course_phase_id)
    ON DELETE CASCADE;

ALTER TABLE application_assessment
  ADD CONSTRAINT unique_course_phase_participation
  UNIQUE (course_phase_id, course_participation_id);


ALTER TABLE course_phase_participation
  DROP COLUMN old_id;

-- Rename the dependency graph table to "participation_data_dependency_graph"
ALTER TABLE meta_data_dependency_graph 
    RENAME TO participation_data_dependency_graph;

-- Add files table required by file upload answers
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    content_type VARCHAR(200) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    storage_key VARCHAR(500) NOT NULL UNIQUE,
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'seaweedfs',
    uploaded_by_user_id VARCHAR(200) NOT NULL,
    uploaded_by_email VARCHAR(200),
    course_phase_id UUID,
    description TEXT,
    tags VARCHAR(100)[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    CONSTRAINT fk_files_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE SET NULL
);

CREATE INDEX idx_files_storage_key ON files(storage_key);
CREATE INDEX idx_files_uploaded_by ON files(uploaded_by_user_id);
CREATE INDEX idx_files_course_phase_id ON files(course_phase_id) WHERE course_phase_id IS NOT NULL;
CREATE INDEX idx_files_created_at ON files(created_at DESC);
CREATE INDEX idx_files_deleted_at ON files(deleted_at) WHERE deleted_at IS NULL;

-- Add sample uploaded file for required file upload answer tests
INSERT INTO files (
    id,
    filename,
    original_filename,
    content_type,
    size_bytes,
    storage_key,
    storage_provider,
    uploaded_by_user_id,
    uploaded_by_email,
    course_phase_id,
    description,
    tags
) VALUES (
    'd3d04042-95d1-4765-8592-caf9560c8c3f',
    'resume_seeded.pdf',
    'resume.pdf',
    'application/pdf',
    1024,
    'course-phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/resume_seeded.pdf',
    'seaweedfs',
    'external',
    'seed@example.com',
    '4179d58a-d00d-4fa7-94a5-397bc69fab02',
    'Seed file for application router tests',
    '{application,resume}'
);

-- Add application_question_file_upload table for file upload questions
CREATE TABLE IF NOT EXISTS application_question_file_upload (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_required BOOLEAN NOT NULL DEFAULT false,
    allowed_file_types TEXT,
    max_file_size_mb INTEGER,
    order_num INTEGER NOT NULL,
    accessible_for_other_phases BOOLEAN NOT NULL DEFAULT false,
    access_key TEXT,
    CONSTRAINT fk_application_question_file_upload_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE
);

CREATE INDEX idx_application_question_file_upload_course_phase_id ON application_question_file_upload(course_phase_id);
CREATE INDEX idx_application_question_file_upload_order_num ON application_question_file_upload(course_phase_id, order_num);

-- Add application_answer_file_upload table for file upload answers
CREATE TABLE IF NOT EXISTS application_answer_file_upload (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_question_id UUID NOT NULL,
    course_participation_id UUID NOT NULL,
    file_id UUID NOT NULL,
    CONSTRAINT fk_application_answer_file_upload_question FOREIGN KEY (application_question_id) REFERENCES application_question_file_upload(id) ON DELETE CASCADE,
    CONSTRAINT fk_application_answer_file_upload_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation(id) ON DELETE CASCADE,
    CONSTRAINT fk_application_answer_file_upload_file FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    CONSTRAINT unique_file_upload_answer UNIQUE (course_participation_id, application_question_id)
);

CREATE INDEX idx_application_answer_file_upload_question ON application_answer_file_upload(application_question_id);
CREATE INDEX idx_application_answer_file_upload_participation ON application_answer_file_upload(course_participation_id);
CREATE INDEX idx_application_answer_file_upload_file ON application_answer_file_upload(file_id);

-- Add sample file upload question for testing
INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
VALUES 
    ('b1b04042-95d1-4765-8592-caf9560c8c3d', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Resume Upload', 'Please upload your resume', true, '.pdf,.doc,.docx', 10, 3, false, null),
    ('c2c04042-95d1-4765-8592-caf9560c8c3e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Portfolio', 'Upload your portfolio (optional)', false, '.pdf,.zip', 20, 4, false, null);
