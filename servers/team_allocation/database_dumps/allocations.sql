-- Allocations table test data
BEGIN;

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

-- A second phase owns the write fixtures, so allocation writes cannot disturb the
-- read assertions on the phase above regardless of test order.
INSERT INTO team (id, name, course_phase_id) VALUES
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Team Delta', '5179d58a-d00d-4fa7-94a5-397bc69fab03'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Team Epsilon', '5179d58a-d00d-4fa7-94a5-397bc69fab03');

INSERT INTO allocations (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name) VALUES
('e6e6e6e6-e6e6-e6e6-e6e6-e6e6e6e6e6e6', '99999999-9999-9999-9999-999999999996', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'Dana', 'Delta'),
('e7e7e7e7-e7e7-e7e7-e7e7-e7e7e7e7e7e7', '99999999-9999-9999-9999-999999999997', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'Erin', 'Epsilon');

-- Tutor scoped to Team Delta in the write-fixture phase
INSERT INTO tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login) VALUES
('5179d58a-d00d-4fa7-94a5-397bc69fab03', '99999999-9999-9999-9999-999999999998', 'Dave', 'Delta', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'ab12cde');

-- A phase without tutors, where editors keep unscoped reads
INSERT INTO team (id, name, course_phase_id) VALUES
('f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0', 'Team Zeta', '6179d58a-d00d-4fa7-94a5-397bc69fab04');

INSERT INTO allocations (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name) VALUES
('e8e8e8e8-e8e8-e8e8-e8e8-e8e8e8e8e8e8', '99999999-9999-9999-9999-9999999999b1', 'f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0', '6179d58a-d00d-4fa7-94a5-397bc69fab04', 'Zoe', 'Zeta'),
('e9e9e9e9-e9e9-e9e9-e9e9-e9e9e9e9e9e9', '99999999-9999-9999-9999-9999999999b2', 'f0f0f0f0-f0f0-f0f0-f0f0-f0f0f0f0f0f0', '6179d58a-d00d-4fa7-94a5-397bc69fab04', 'Zack', 'Zeta');

COMMIT;
