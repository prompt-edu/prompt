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
-- Name: course_phase; Type: TABLE; Schema: public; Owner: prompt-postgres
--
CREATE TABLE
  course_phase_type (
    id uuid NOT NULL,
    name text NOT NULL,
    description text
  );

CREATE TABLE
  course_phase (
    id uuid NOT NULL,
    course_id uuid NOT NULL,
    name text,
    meta_data jsonb,
    is_initial_phase boolean NOT NULL,
    course_phase_type_id uuid NOT NULL
  );

INSERT INTO
  course_phase_type (id, name, description)
VALUES
  (
    '7dc1c4e8-4255-4874-80a0-0c12b958744b',
    'application',
    'Test Description'
  );

INSERT INTO
  course_phase_type (id, name, description)
VALUES
  (
    '7dc1c4e8-4255-4874-80a0-0c12b958744c',
    'example_component',
    'Test Description'
  );

--
-- Data for Name: course_phase; Type: TABLE DATA; Schema: public; Owner: prompt-postgres
--
INSERT INTO
  course_phase (
    id,
    course_id,
    name,
    meta_data,
    is_initial_phase,
    course_phase_type_id
  )
VALUES
  (
    '3d1f3b00-87f3-433b-a713-178c4050411b',
    '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
    'Test',
    '{"test-key":"test-value"}',
    false,
    '7dc1c4e8-4255-4874-80a0-0c12b958744b'
  );

INSERT INTO
  course_phase (
    id,
    course_id,
    name,
    meta_data,
    is_initial_phase,
    course_phase_type_id
  )
VALUES
  (
    '92bb0532-39e5-453d-bc50-fa61ea0128b2',
    '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
    'Example Phase',
    '{}',
    false,
    '7dc1c4e8-4255-4874-80a0-0c12b958744c'
  );

INSERT INTO
  course_phase (
    id,
    course_id,
    name,
    meta_data,
    is_initial_phase,
    course_phase_type_id
  )
VALUES
  (
    '500db7ed-2eb2-42d0-82b3-8750e12afa8a',
    '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
    'Application Phase',
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

ALTER TABLE ONLY course_phase ADD CONSTRAINT fk_phase_type FOREIGN KEY (course_phase_type_id) REFERENCES public.course_phase_type (id);

ALTER TABLE course_phase
RENAME COLUMN meta_data TO restricted_data;

ALTER TABLE course_phase
ADD COLUMN student_readable_data jsonb DEFAULT '{}';

CREATE TABLE
  course_phase_type_phase_provided_output_dto (
    id uuid PRIMARY KEY,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    version_number integer NOT NULL,
    endpoint_path text NOT NULL,
    specification jsonb NOT NULL,
    CONSTRAINT fk_course_phase_type_phase_provided FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type (id) -- adjust if needed
    ON DELETE CASCADE
  );

-- Required Input DTO table for phases
CREATE TABLE
  course_phase_type_phase_required_input_dto (
    id uuid PRIMARY KEY,
    course_phase_type_id uuid NOT NULL,
    dto_name text NOT NULL,
    specification jsonb NOT NULL,
    optional boolean DEFAULT false NOT NULL,
    CONSTRAINT fk_course_phase_type_phase_required FOREIGN KEY (course_phase_type_id) REFERENCES course_phase_type (id) -- adjust if needed
    ON DELETE CASCADE
  );

-- Dependency graph table for phases
CREATE TABLE
  phase_data_dependency_graph (
    from_course_phase_id uuid NOT NULL,
    to_course_phase_id uuid NOT NULL,
    from_course_phase_DTO_id uuid NOT NULL,
    to_course_phase_DTO_id uuid NOT NULL,
    PRIMARY KEY (to_course_phase_id, to_course_phase_DTO_id),
    CONSTRAINT fk_from_phase_phase FOREIGN KEY (from_course_phase_id) REFERENCES course_phase (id) ON DELETE CASCADE,
    CONSTRAINT fk_to_phase_phase FOREIGN KEY (to_course_phase_id) REFERENCES course_phase (id) ON DELETE CASCADE,
    CONSTRAINT fk_from_dto_phase FOREIGN KEY (from_course_phase_DTO_id) REFERENCES course_phase_type_phase_provided_output_dto (id) ON DELETE CASCADE,
    CONSTRAINT fk_to_dto_phase FOREIGN KEY (to_course_phase_DTO_id) REFERENCES course_phase_type_phase_required_input_dto (id) ON DELETE CASCADE
  );

-- Add base_url column to course_phase_type (if not already present)
ALTER TABLE course_phase_type
ADD COLUMN IF NOT EXISTS base_url text;

UPDATE course_phase_type
SET
  base_url = 'http://example.com'
WHERE
  id = '7dc1c4e8-4255-4874-80a0-0c12b958744b';

-- Ensure description column exists for newer schema expectations
ALTER TABLE course_phase_type
ADD COLUMN IF NOT EXISTS description text;

UPDATE course_phase_type
SET
  description = 'Application phase handling'
WHERE
  name = 'application';

UPDATE course_phase_type
SET
  description = 'Example component phase'
WHERE
  name = 'example_component';

-- Insert a predecessor phase and a target phase for testing
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
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
    'Predecessor Phase',
    '{"TestDTO": "restricted-test-value"}',
    '{}',
    false,
    '7dc1c4e8-4255-4874-80a0-0c12b958744b'
  ),
  (
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    '3f42d322-e5bf-4faa-b576-51f2cab14c2e',
    'Target Phase',
    '{}',
    '{}',
    false,
    '7dc1c4e8-4255-4874-80a0-0c12b958744b'
  );

-- Insert provided output DTO rows for the predecessor phase
-- Core DTO (for GetPrevCoursePhaseDataFromCore)
INSERT INTO
  course_phase_type_phase_provided_output_dto (
    id,
    course_phase_type_id,
    dto_name,
    version_number,
    endpoint_path,
    specification
  )
VALUES
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '7dc1c4e8-4255-4874-80a0-0c12b958744b',
    'TestDTO',
    1,
    'core',
    '{}'
  );

-- Non-core DTO (for resolution)
INSERT INTO
  course_phase_type_phase_provided_output_dto (
    id,
    course_phase_type_id,
    dto_name,
    version_number,
    endpoint_path,
    specification
  )
VALUES
  (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    '7dc1c4e8-4255-4874-80a0-0c12b958744b',
    'ResolutionDTO',
    1,
    'non-core',
    '{}'
  );

-- Insert required input DTO rows for dependency graph links
INSERT INTO
  course_phase_type_phase_required_input_dto (id, course_phase_type_id, dto_name, specification)
VALUES
  (
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    '7dc1c4e8-4255-4874-80a0-0c12b958744b',
    'TestDTO',
    '{}'
  ),
  (
    'ffffffff-ffff-ffff-ffff-ffffffffffff',
    '7dc1c4e8-4255-4874-80a0-0c12b958744b',
    'ResolutionDTO',
    '{}'
  );

-- Insert dependency graph rows linking the predecessor phase to the target phase.
-- This creates two connections: one for the core DTO and one for the resolution DTO.
INSERT INTO
  phase_data_dependency_graph (
    from_course_phase_id,
    to_course_phase_id,
    from_course_phase_DTO_id,
    to_course_phase_DTO_id
  )
VALUES
  (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'
  ),
  (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'ffffffff-ffff-ffff-ffff-ffffffffffff'
  );
