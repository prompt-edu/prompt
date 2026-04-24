-- Teams table test data
BEGIN;

-- Schema for teams
CREATE TABLE
    IF NOT EXISTS team
(
    id              uuid         NOT NULL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    course_phase_id uuid         NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name),
    CONSTRAINT team_id_course_phase_uk UNIQUE (id, course_phase_id)
);

ALTER TABLE team
    ADD COLUMN team_type VARCHAR(32) NOT NULL DEFAULT 'standard';

CREATE TABLE allocation_profile
(
    course_phase_id UUID        NOT NULL PRIMARY KEY,
    profile         VARCHAR(64) NOT NULL DEFAULT 'standard'
);

CREATE TABLE
    IF NOT EXISTS allocations
(
    id                      UUID      NOT NULL PRIMARY KEY,
    course_participation_id UUID      NOT NULL,
    team_id                 UUID      NOT NULL,
    course_phase_id         UUID      NOT NULL,
    student_full_name       TEXT      NOT NULL DEFAULT '',
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
    PRIMARY KEY (course_phase_id, course_participation_id),
    FOREIGN KEY (team_id, course_phase_id) REFERENCES team (id, course_phase_id) ON DELETE CASCADE
);

ALTER TABLE allocations
    DROP COLUMN student_full_name,
    ADD COLUMN student_first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN student_last_name  TEXT NOT NULL DEFAULT '';

-- Test data
INSERT INTO allocation_profile (course_phase_id, profile)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'standard'),
       ('5179d58a-d00d-4fa7-94a5-397bc69fab03', 'standard');

INSERT INTO team (id, name, course_phase_id)
VALUES ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'Team Alpha',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
       ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        'Team Beta',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
       ('cccccccc-cccc-cccc-cccc-cccccccccccc',
        'Team Gamma',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
       ('dddddddd-dddd-dddd-dddd-dddddddddddd',
        'Team Delta',
        '5179d58a-d00d-4fa7-94a5-397bc69fab03');

-- Sample allocations for team tests
INSERT INTO allocations (id,
                         course_participation_id,
                         team_id,
                         course_phase_id,
                         student_first_name,
                         student_last_name)
VALUES ('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
        '99999999-9999-9999-9999-999999999991',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'John',
        'Doe'),
       ('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
        '99999999-9999-9999-9999-999999999992',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '4179d58a-d00d-4fa7-94a5-397bc69fab02',
        'Jane',
        'Smith');

-- Sample tutors for team tests
INSERT INTO tutor (course_phase_id,
                   course_participation_id,
                   first_name,
                   last_name,
                   team_id)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02',
        '99999999-9999-9999-9999-999999999993',
        'Alice',
        'Johnson',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
       ('4179d58a-d00d-4fa7-94a5-397bc69fab02',
        '99999999-9999-9999-9999-999999999994',
        'Bob',
        'Williams',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb');

COMMIT;
