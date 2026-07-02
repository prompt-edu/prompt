-- Survey table test data
BEGIN;

-- Schema for survey
CREATE TABLE IF NOT EXISTS survey_timeframe (
    course_phase_id uuid NOT NULL PRIMARY KEY,
    survey_start TIMESTAMP NOT NULL,
    survey_deadline TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS team (
    id uuid NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    course_phase_id uuid NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_course_phase_team UNIQUE (course_phase_id, name)
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

-- Teams for preferences
INSERT INTO team (id, name, course_phase_id) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Team Alpha',   '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Team Beta',    '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Team Gamma',   '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Team Delta',   '4179d58a-d00d-4fa7-94a5-397bc69fab02'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Team Epsilon', '4179d58a-d00d-4fa7-94a5-397bc69fab02');

-- Student team preferences for testing (6 students, each ranking all 5 teams).
-- Students 1–5 form a cyclic Latin square so every team appears at every rank.
-- Student 6 repeats student 1's ordering, making Alpha/Beta slightly more popular.
INSERT INTO student_team_preference_response (course_participation_id, team_id, preference) VALUES
-- Student 1: Alpha=1, Beta=2, Gamma=3, Delta=4, Epsilon=5
('99999999-9999-9999-9999-999999999991', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1),
('99999999-9999-9999-9999-999999999991', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2),
('99999999-9999-9999-9999-999999999991', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 3),
('99999999-9999-9999-9999-999999999991', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 4),
('99999999-9999-9999-9999-999999999991', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 5),
-- Student 2: Beta=1, Gamma=2, Delta=3, Epsilon=4, Alpha=5
('99999999-9999-9999-9999-999999999992', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 1),
('99999999-9999-9999-9999-999999999992', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 2),
('99999999-9999-9999-9999-999999999992', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 3),
('99999999-9999-9999-9999-999999999992', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 4),
('99999999-9999-9999-9999-999999999992', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 5),
-- Student 3: Gamma=1, Delta=2, Epsilon=3, Alpha=4, Beta=5
('99999999-9999-9999-9999-999999999993', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 1),
('99999999-9999-9999-9999-999999999993', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 2),
('99999999-9999-9999-9999-999999999993', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 3),
('99999999-9999-9999-9999-999999999993', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 4),
('99999999-9999-9999-9999-999999999993', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 5),
-- Student 4: Delta=1, Epsilon=2, Alpha=3, Beta=4, Gamma=5
('99999999-9999-9999-9999-999999999994', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 1),
('99999999-9999-9999-9999-999999999994', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 2),
('99999999-9999-9999-9999-999999999994', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 3),
('99999999-9999-9999-9999-999999999994', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 4),
('99999999-9999-9999-9999-999999999994', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 5),
-- Student 5: Epsilon=1, Alpha=2, Beta=3, Gamma=4, Delta=5
('99999999-9999-9999-9999-999999999995', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 1),
('99999999-9999-9999-9999-999999999995', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 2),
('99999999-9999-9999-9999-999999999995', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 3),
('99999999-9999-9999-9999-999999999995', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 4),
('99999999-9999-9999-9999-999999999995', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 5),
-- Student 6: Alpha=1, Beta=2, Gamma=3, Delta=4, Epsilon=5
('99999999-9999-9999-9999-999999999996', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 1),
('99999999-9999-9999-9999-999999999996', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 2),
('99999999-9999-9999-9999-999999999996', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 3),
('99999999-9999-9999-9999-999999999996', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 4),
('99999999-9999-9999-9999-999999999996', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 5);

COMMIT;
