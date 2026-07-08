-- Allocations table test data
BEGIN;

-- Schema for allocations
DO $$ BEGIN
    CREATE TYPE skill_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS team (
    id uuid NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    course_phase_id uuid NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name),
    CONSTRAINT team_id_course_phase_uk UNIQUE (id, course_phase_id)
);

CREATE TABLE IF NOT EXISTS allocations (
    id UUID NOT NULL PRIMARY KEY,
    course_participation_id UUID NOT NULL,
    team_id UUID NOT NULL,
    course_phase_id UUID NOT NULL,
    student_full_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    -- still enforce that team_id itself exists...
    FOREIGN KEY (team_id) REFERENCES team (id) ON DELETE CASCADE,
    -- …and also enforce that the phase matches the one on the team:
    FOREIGN KEY (team_id, course_phase_id) REFERENCES team (id, course_phase_id) ON DELETE CASCADE,
    CONSTRAINT allocations_participation_phase_uk UNIQUE (course_participation_id, course_phase_id)
);

ALTER TABLE allocations
    DROP COLUMN student_full_name,
    ADD COLUMN student_first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN student_last_name  TEXT NOT NULL DEFAULT '';

CREATE TABLE IF NOT EXISTS tutor (
    course_phase_id         uuid NOT NULL,
    course_participation_id uuid NOT NULL,
    first_name              text NOT NULL,
    last_name               text NOT NULL,
    team_id                 uuid NOT NULL,
    university_login        text,
    PRIMARY KEY (course_phase_id, course_participation_id),
    FOREIGN KEY (team_id, course_phase_id) REFERENCES team (id, course_phase_id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_tutor_phase_login
    ON tutor (course_phase_id, university_login)
    WHERE university_login IS NOT NULL;

-- Test data
-- Teams for allocations
INSERT INTO team (id, name, course_phase_id) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Team Alpha', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Team Beta', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Team Gamma', '4179d58a-d00d-4fa7-94a5-397bc69fab02');

-- Allocations for testing
INSERT INTO allocations (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name) VALUES
('e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', '99999999-9999-9999-9999-999999999991', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'John', 'Doe'),
('e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', '99999999-9999-9999-9999-999999999992', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Jane', 'Smith'),
('e3e3e3e3-e3e3-e3e3-e3e3-e3e3e3e3e3e3', '99999999-9999-9999-9999-999999999993', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Bob', 'Johnson');

-- Tutor scoped to Team Alpha for access-control tests
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login) VALUES
('4179d58a-d00d-4fa7-94a5-397bc69fab02', '99999999-9999-9999-9999-999999999994', 'Alice', 'Johnson', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'ab12cde');

COMMIT;
