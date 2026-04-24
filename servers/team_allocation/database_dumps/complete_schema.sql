-- Comprehensive database dump for team allocation tests
-- Contains schema from all migrations and test data
BEGIN;

-- Schema from 0001_schema.up.sql
CREATE TYPE skill_level AS ENUM ('novice', 'intermediate', 'advanced', 'expert');

CREATE TABLE
    team (
        id uuid NOT NULL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        course_phase_id uuid NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name),
        CONSTRAINT team_id_course_phase_uk UNIQUE (id, course_phase_id)
    );

CREATE TABLE
    student_team_preference_response (
        course_participation_id uuid NOT NULL,
        team_id uuid NOT NULL,
        preference INT NOT NULL,
        PRIMARY KEY (course_participation_id, team_id),
        FOREIGN KEY (team_id) REFERENCES team (id) ON DELETE CASCADE
    );

CREATE TABLE
    survey_timeframe (
        course_phase_id uuid NOT NULL PRIMARY KEY,
        survey_start TIMESTAMP NOT NULL,
        survey_deadline TIMESTAMP NOT NULL
    );

CREATE TABLE
    survey_timeframe_profile (
        course_phase_id uuid NOT NULL,
        profile VARCHAR(64) NOT NULL,
        survey_start TIMESTAMP NOT NULL,
        survey_deadline TIMESTAMP NOT NULL,
        PRIMARY KEY (course_phase_id, profile)
    );

CREATE TABLE
    skill (
        id uuid NOT NULL PRIMARY KEY,
        course_phase_id uuid NOT NULL,
        name VARCHAR(255) NOT NULL
    );

CREATE TABLE
    student_skill_response (
        course_participation_id uuid NOT NULL,
        skill_id uuid NOT NULL,
        skill_level skill_level NOT NULL,
        PRIMARY KEY (course_participation_id, skill_id),
        FOREIGN KEY (skill_id) REFERENCES skill (id) ON DELETE CASCADE
    );

-- Schema from 0003_allocations.up.sql
CREATE TABLE
    allocations (
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
ADD COLUMN student_last_name TEXT NOT NULL DEFAULT '';

ALTER TABLE team
ADD COLUMN team_type VARCHAR(32) NOT NULL DEFAULT 'standard';

ALTER TABLE student_skill_response
ADD COLUMN preference_mode VARCHAR(16) NOT NULL DEFAULT 'teams';

ALTER TABLE student_skill_response
DROP CONSTRAINT student_skill_response_pkey;

ALTER TABLE student_skill_response
ADD PRIMARY KEY (course_participation_id, skill_id, preference_mode);

CREATE TABLE allocation_profile (
    course_phase_id UUID NOT NULL PRIMARY KEY,
    profile VARCHAR(64) NOT NULL DEFAULT 'standard'
);

-- Test data
-- Course phase for testing
INSERT INTO
    survey_timeframe (course_phase_id, survey_start, survey_deadline)
VALUES
    (
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        '2024-01-01 00:00:00',
        '2030-12-31 23:59:59'
    );

INSERT INTO
    survey_timeframe_profile (course_phase_id, profile, survey_start, survey_deadline)
VALUES
    (
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'standard',
        '2024-01-01 00:00:00',
        '2030-12-31 23:59:59'
    );

INSERT INTO allocation_profile (course_phase_id, profile)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'standard');

-- Skills for testing
INSERT INTO
    skill (id, course_phase_id, name)
VALUES
    (
        '11111111-1111-1111-1111-111111111111',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'Java'
    ),
    (
        '22222222-2222-2222-2222-222222222222',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'Python'
    ),
    (
        '33333333-3333-3333-3333-333333333333',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'JavaScript'
    ),
    (
        '44444444-4444-4444-4444-444444444444',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'Database Design'
    );

-- Teams for testing
INSERT INTO
    team (id, name, course_phase_id)
VALUES
    (
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Team Alpha',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'
    ),
    (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'Team Beta',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        'Team Gamma',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'
    );

-- Student skill responses for testing
INSERT INTO
    student_skill_response (course_participation_id, skill_id, skill_level)
VALUES
    (
        '99999999-9999-9999-9999-999999999991',
        '11111111-1111-1111-1111-111111111111',
        'advanced'
    ),
    (
        '99999999-9999-9999-9999-999999999991',
        '22222222-2222-2222-2222-222222222222',
        'intermediate'
    ),
    (
        '99999999-9999-9999-9999-999999999992',
        '11111111-1111-1111-1111-111111111111',
        'novice'
    ),
    (
        '99999999-9999-9999-9999-999999999992',
        '33333333-3333-3333-3333-333333333333',
        'expert'
    );

-- Student team preferences for testing
INSERT INTO
    student_team_preference_response (course_participation_id, team_id, preference)
VALUES
    (
        '99999999-9999-9999-9999-999999999991',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        1
    ),
    (
        '99999999-9999-9999-9999-999999999991',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        2
    ),
    (
        '99999999-9999-9999-9999-999999999992',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        1
    ),
    (
        '99999999-9999-9999-9999-999999999992',
        'cccccccc-cccc-cccc-cccc-cccccccccccc',
        2
    );

-- Allocations for testing
INSERT INTO
    allocations (
        id,
        course_participation_id,
        team_id,
        course_phase_id,
        student_first_name,
        student_last_name
    )
VALUES
    (
        'e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1',
        '99999999-9999-9999-9999-999999999991',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'John',
        'Doe'
    ),
    (
        'e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2',
        '99999999-9999-9999-9999-999999999992',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'Jane',
        'Smith'
    );

COMMIT;
