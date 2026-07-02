-- Skills table test data
BEGIN;

-- Schema for skills
DO $$ BEGIN
    CREATE TYPE skill_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS skill (
    id uuid NOT NULL PRIMARY KEY,
    course_phase_id uuid NOT NULL,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS student_skill_response (
    course_participation_id uuid NOT NULL,
    skill_id uuid NOT NULL,
    skill_level skill_level NOT NULL,
    PRIMARY KEY (course_participation_id, skill_id),
    FOREIGN KEY (skill_id) REFERENCES skill(id) ON DELETE CASCADE
);

-- Test data
INSERT INTO skill (id, course_phase_id, name) VALUES
('11111111-1111-1111-1111-111111111111', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Java'),
('22222222-2222-2222-2222-222222222222', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Python'),
('33333333-3333-3333-3333-333333333333', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'JavaScript'),
('44444444-4444-4444-4444-444444444444', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Database Design'),
('55555555-5555-5555-5555-555555555555', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'C++');

-- Student skill responses for testing
INSERT INTO student_skill_response (course_participation_id, skill_id, skill_level) VALUES
('99999999-9999-9999-9999-999999999991', '11111111-1111-1111-1111-111111111111', 'good'),
('99999999-9999-9999-9999-999999999991', '22222222-2222-2222-2222-222222222222', 'ok'),
('99999999-9999-9999-9999-999999999992', '11111111-1111-1111-1111-111111111111', 'bad'),
('99999999-9999-9999-9999-999999999992', '33333333-3333-3333-3333-333333333333', 'very_good'),
('99999999-9999-9999-9999-999999999993', '44444444-4444-4444-4444-444444444444', 'ok');

COMMIT;
