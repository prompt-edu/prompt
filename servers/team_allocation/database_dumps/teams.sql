-- Teams table test data
BEGIN;

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
                   team_id,
                   university_login)
VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02',
        '99999999-9999-9999-9999-999999999993',
        'Alice',
        'Johnson',
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'ab12cde'),
       ('4179d58a-d00d-4fa7-94a5-397bc69fab02',
        '99999999-9999-9999-9999-999999999994',
        'Bob',
        'Williams',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        NULL);

COMMIT;