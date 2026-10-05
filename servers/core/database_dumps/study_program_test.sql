-- Migration 0031 seeds the default study programs with random ids; pin them so tests can reference them.
UPDATE study_program SET id = 'a1000000-0000-0000-0000-000000000001' WHERE name = 'Computer Science';
UPDATE study_program SET id = 'a1000000-0000-0000-0000-000000000002' WHERE name = 'Information Systems';
UPDATE study_program SET id = 'a1000000-0000-0000-0000-000000000003' WHERE name = 'Games Engineering';
UPDATE study_program SET id = 'a1000000-0000-0000-0000-000000000004' WHERE name = 'Management and Technology';

INSERT INTO student (id, first_name, last_name, gender, study_program) VALUES
    ('b2000000-0000-0000-0000-000000000001', 'Ada', 'Lovelace', 'female', 'Computer Science'),
    ('b2000000-0000-0000-0000-000000000002', 'Alan', 'Turing', 'male', ' Computer Science '),
    ('b2000000-0000-0000-0000-000000000003', 'Grace', 'Hopper', 'female', 'Information Systems'),
    ('b2000000-0000-0000-0000-000000000004', 'Edsger', 'Dijkstra', 'male', 'Games Engineering'),
    ('b2000000-0000-0000-0000-000000000005', 'Barbara', 'Liskov', 'female', 'Robotics'),
    ('b2000000-0000-0000-0000-000000000006', 'Donald', 'Knuth', 'male', NULL),
    ('b2000000-0000-0000-0000-000000000007', 'Frances', 'Allen', 'female', '  Information Systems');
