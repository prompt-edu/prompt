SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);
SET client_min_messages = warning;

CREATE TABLE student (
    id uuid PRIMARY KEY,
    first_name character varying(50),
    last_name character varying(50),
    study_program character varying(100)
);

CREATE TABLE study_program (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name character varying(100) NOT NULL,
    short_name character varying(20)
);

CREATE UNIQUE INDEX study_program_name_unique ON study_program (lower(trim(name)));
CREATE UNIQUE INDEX study_program_label_unique ON study_program (lower(trim(coalesce(short_name, name))));

INSERT INTO study_program (id, name, short_name) VALUES
    ('a1000000-0000-0000-0000-000000000001', 'Computer Science', 'CS'),
    ('a1000000-0000-0000-0000-000000000002', 'Information Systems', 'IS'),
    ('a1000000-0000-0000-0000-000000000003', 'Games Engineering', 'GE'),
    ('a1000000-0000-0000-0000-000000000004', 'Management and Technology', 'M&T');

INSERT INTO student (id, first_name, last_name, study_program) VALUES
    ('b2000000-0000-0000-0000-000000000001', 'Ada', 'Lovelace', 'Computer Science'),
    ('b2000000-0000-0000-0000-000000000002', 'Alan', 'Turing', ' Computer Science '),
    ('b2000000-0000-0000-0000-000000000003', 'Grace', 'Hopper', 'Information Systems'),
    ('b2000000-0000-0000-0000-000000000004', 'Edsger', 'Dijkstra', 'Games Engineering'),
    ('b2000000-0000-0000-0000-000000000005', 'Barbara', 'Liskov', 'Robotics'),
    ('b2000000-0000-0000-0000-000000000006', 'Donald', 'Knuth', NULL),
    ('b2000000-0000-0000-0000-000000000007', 'Frances', 'Allen', '  Information Systems');
