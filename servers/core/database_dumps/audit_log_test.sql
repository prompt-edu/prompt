-- Minimal course_phase for course_id resolution (GetCourseIDByCoursePhaseID).

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects) VALUES
  ('22222222-2222-2222-2222-222222222222', 'Audit Course', '2024-10-01', '2025-03-31', 'ws2425', 'practical course', 10);

INSERT INTO course_phase_type (id, name) VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Application');

INSERT INTO course_phase (id, course_id, name, is_initial_phase, course_phase_type_id) VALUES
  ('11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'Application', true, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');
