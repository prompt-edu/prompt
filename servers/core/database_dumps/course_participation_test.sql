INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects) VALUES
    ('3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10),
    ('918977e1-2d27-4b55-9064-8504ff027a1a', 'New fancy course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10);

INSERT INTO student (id, first_name, last_name, email, gender) VALUES
    ('3d1f3b00-87f3-433b-a713-178c4050411a', 'Ada', 'Lovelace', 'ada.lovelace@example.com', 'female'),
    ('7dc1c4e8-4255-4874-80a0-0c12b958744b', 'Alan', 'Turing', 'alan.turing@example.com', 'male'),
    ('500db7ed-2eb2-42d0-82b3-8750e12afa8b', 'Grace', 'Hopper', 'grace.hopper@example.com', 'female'),
    ('7dc1c4e8-4255-4874-80a0-0c12b958744a', 'Edsger', 'Dijkstra', 'edsger.dijkstra@example.com', 'male');

INSERT INTO course_participation (id, course_id, student_id) VALUES ('6e19bab2-53d0-4b6a-ac02-33b23988401a', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', '3d1f3b00-87f3-433b-a713-178c4050411a');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('8713d7bc-1542-4366-88a9-1fa50945b052', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', '7dc1c4e8-4255-4874-80a0-0c12b958744b');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('0e762fdd-c4fa-49f4-9c38-c90160cc6caa', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', '500db7ed-2eb2-42d0-82b3-8750e12afa8b');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('65dcc535-a9ab-4421-a2bc-0f09780ca59e', '918977e1-2d27-4b55-9064-8504ff027a1a', '500db7ed-2eb2-42d0-82b3-8750e12afa8b');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('ec679792-f9e1-423c-80fa-a9f3324fefa8', '918977e1-2d27-4b55-9064-8504ff027a1a', '7dc1c4e8-4255-4874-80a0-0c12b958744b');
INSERT INTO course_participation (id, course_id, student_id) VALUES ('9f061396-e208-4b00-bc8c-f3a04bc0a212', '918977e1-2d27-4b55-9064-8504ff027a1a', '7dc1c4e8-4255-4874-80a0-0c12b958744a');
