-- Course phase deletion test data
-- Contains every table this service stores course phase scoped data in, filled for two course
-- phases so the tests can assert that only the deleted phase is affected.
BEGIN;

CREATE TYPE skill_level AS ENUM ('very_bad', 'bad', 'ok', 'good', 'very_good');

CREATE TABLE
    team
(
    id              uuid         NOT NULL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    course_phase_id uuid         NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name),
    CONSTRAINT team_id_course_phase_uk UNIQUE (id, course_phase_id)
);

CREATE TABLE
    student_team_preference_response
(
    course_participation_id uuid NOT NULL,
    team_id                 uuid NOT NULL,
    preference              INT  NOT NULL,
    PRIMARY KEY (course_participation_id, team_id),
    FOREIGN KEY (team_id) REFERENCES team (id) ON DELETE CASCADE
);

CREATE TABLE
    survey_timeframe
(
    course_phase_id uuid        NOT NULL PRIMARY KEY,
    survey_start    TIMESTAMPTZ NOT NULL,
    survey_deadline TIMESTAMPTZ NOT NULL
);

CREATE TABLE
    skill
(
    id              uuid         NOT NULL PRIMARY KEY,
    course_phase_id uuid         NOT NULL,
    name            VARCHAR(255) NOT NULL
);

CREATE TABLE
    student_skill_response
(
    course_participation_id uuid        NOT NULL,
    skill_id                uuid        NOT NULL,
    skill_level             skill_level NOT NULL,
    PRIMARY KEY (course_participation_id, skill_id),
    FOREIGN KEY (skill_id) REFERENCES skill (id) ON DELETE CASCADE
);

CREATE TABLE
    allocations
(
    id                      UUID      NOT NULL PRIMARY KEY,
    course_participation_id UUID      NOT NULL,
    team_id                 UUID      NOT NULL,
    course_phase_id         UUID      NOT NULL,
    student_first_name      TEXT      NOT NULL DEFAULT '',
    student_last_name       TEXT      NOT NULL DEFAULT '',
    created_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (team_id) REFERENCES team (id) ON DELETE CASCADE,
    FOREIGN KEY (team_id, course_phase_id) REFERENCES team (id, course_phase_id) ON DELETE CASCADE,
    CONSTRAINT allocations_participation_phase_uk UNIQUE (course_participation_id, course_phase_id)
);

CREATE TABLE
    tutor
(
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

CREATE TABLE
    tease_workspace
(
    course_phase_id   uuid PRIMARY KEY,
    constraints       jsonb       NOT NULL DEFAULT '[]'::jsonb,
    locked_students   jsonb       NOT NULL DEFAULT '[]'::jsonb,
    allocations_draft jsonb       NOT NULL DEFAULT '[]'::jsonb,
    algorithm_type    varchar(64),
    updated_by        uuid,
    last_saved_at     timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_exported_at  timestamptz
);

-- Test data
-- Phase 4179d58a-... is the phase under deletion, phase 5179d58a-... must stay untouched.
INSERT INTO team (id, name, course_phase_id)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Team Alpha', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
       ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Team Beta', '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
       ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Team Delta', '5179d58a-d00d-4fa7-94a5-397bc69fab03');

INSERT INTO skill (id, course_phase_id, name)
VALUES ('11111111-1111-1111-1111-111111111111', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Java'),
       ('22222222-2222-2222-2222-222222222222', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Python'),
       ('33333333-3333-3333-3333-333333333333', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'Kotlin');

INSERT INTO student_team_preference_response (course_participation_id, team_id, preference)
VALUES ('99999999-9999-9999-9999-999999999991', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1),
       ('99999999-9999-9999-9999-999999999991', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2),
       ('99999999-9999-9999-9999-999999999991', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 1);

INSERT INTO student_skill_response (course_participation_id, skill_id, skill_level)
VALUES ('99999999-9999-9999-9999-999999999991', '11111111-1111-1111-1111-111111111111', 'good'),
       ('99999999-9999-9999-9999-999999999992', '22222222-2222-2222-2222-222222222222', 'ok'),
       ('99999999-9999-9999-9999-999999999991', '33333333-3333-3333-3333-333333333333', 'very_good');

INSERT INTO allocations (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name)
VALUES ('e1e1e1e1-e1e1-e1e1-e1e1-e1e1e1e1e1e1', '99999999-9999-9999-9999-999999999991',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'John', 'Doe'),
       ('e2e2e2e2-e2e2-e2e2-e2e2-e2e2e2e2e2e2', '99999999-9999-9999-9999-999999999992',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Jane', 'Smith'),
       ('e3e3e3e3-e3e3-e3e3-e3e3-e3e3e3e3e3e3', '99999999-9999-9999-9999-999999999991',
        'dddddddd-dddd-dddd-dddd-dddddddddddd', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'John', 'Doe');

INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', '99999999-9999-9999-9999-999999999993', 'Alice', 'Johnson',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'ab12cde'),
       ('5179d58a-d00d-4fa7-94a5-397bc69fab03', '99999999-9999-9999-9999-999999999994', 'Bob', 'Williams',
        'dddddddd-dddd-dddd-dddd-dddddddddddd', 'cd34efg');

INSERT INTO survey_timeframe (course_phase_id, survey_start, survey_deadline)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', '2024-01-01 00:00:00+00', '2030-12-31 23:59:59+00'),
       ('5179d58a-d00d-4fa7-94a5-397bc69fab03', '2024-01-01 00:00:00+00', '2030-12-31 23:59:59+00');

INSERT INTO tease_workspace (course_phase_id, algorithm_type)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'basic'),
       ('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'basic');

COMMIT;
