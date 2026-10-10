-- Test data for storage module tests: one course phase that files can be attached to.

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type) VALUES
('33333333-3333-3333-3333-333333333333', 'Test Course', '2024-10-01', '2025-03-31', 'WS2024', 'practical course');

INSERT INTO course_phase_type (id, name) VALUES
('66666666-6666-6666-6666-666666666666', 'Application');

INSERT INTO course_phase (id, course_id, name, is_initial_phase, course_phase_type_id) VALUES
('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', 'Test Phase', true, '66666666-6666-6666-6666-666666666666');
