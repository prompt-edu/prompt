-- Course phase deletion test data
-- Fills every table this service stores course phase scoped data in for two course phases, so the
-- tests can assert that only the deleted phase is affected.
BEGIN;

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
