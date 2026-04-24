-- Survey table test data
BEGIN;

-- Schema for survey
CREATE TABLE IF NOT EXISTS survey_timeframe (
    course_phase_id uuid NOT NULL PRIMARY KEY,
    survey_start TIMESTAMP NOT NULL,
    survey_deadline TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS survey_timeframe_profile (
    course_phase_id uuid NOT NULL,
    profile VARCHAR(64) NOT NULL,
    survey_start TIMESTAMP NOT NULL,
    survey_deadline TIMESTAMP NOT NULL,
    PRIMARY KEY (course_phase_id, profile)
);

CREATE TABLE IF NOT EXISTS team (
    id uuid NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    course_phase_id uuid NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name)
);

ALTER TABLE team
ADD COLUMN team_type VARCHAR(32) NOT NULL DEFAULT 'standard';

CREATE TABLE IF NOT EXISTS allocation_profile (
    course_phase_id UUID NOT NULL PRIMARY KEY,
    profile VARCHAR(64) NOT NULL DEFAULT 'standard'
);

CREATE TABLE IF NOT EXISTS student_team_preference_response (
    course_participation_id uuid NOT NULL,
    team_id uuid NOT NULL,
    preference INT NOT NULL,
    PRIMARY KEY (course_participation_id, team_id),
    FOREIGN KEY (team_id) REFERENCES team(id) ON DELETE CASCADE
);

-- Test data
INSERT INTO survey_timeframe (course_phase_id, survey_start, survey_deadline) VALUES
('4179d58a-d00d-4fa7-94a5-397bc69fab02', '2024-01-01 10:00:00', '2024-01-31 23:59:59'),
('5179d58a-d00d-4fa7-94a5-397bc69fab03', '2024-02-01 10:00:00', '2024-02-28 23:59:59');

INSERT INTO survey_timeframe_profile (course_phase_id, profile, survey_start, survey_deadline) VALUES
('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'standard', '2024-01-01 10:00:00', '2024-01-31 23:59:59'),
('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'standard', '2024-02-01 10:00:00', '2024-02-28 23:59:59');

INSERT INTO allocation_profile (course_phase_id, profile) VALUES
('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'standard'),
('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'standard');

-- Teams for preferences
INSERT INTO team (id, name, course_phase_id) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Team Alpha', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Team Beta', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Team Gamma', '4179d58a-d00d-4fa7-94a5-397bc69fab02');

-- Student team preferences for testing
INSERT INTO student_team_preference_response (course_participation_id, team_id, preference) VALUES
('99999999-9999-9999-9999-999999999991', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1),
('99999999-9999-9999-9999-999999999991', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2),
('99999999-9999-9999-9999-999999999991', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 3),
('99999999-9999-9999-9999-999999999992', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 1),
('99999999-9999-9999-9999-999999999992', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 2),
('99999999-9999-9999-9999-999999999992', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 3);

COMMIT;
