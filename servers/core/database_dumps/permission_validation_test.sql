INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e11', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10);

INSERT INTO course_phase_type (id, name, base_url, description) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e12', 'application', 'http://example.com', 'Application phase handling');

INSERT INTO course_phase (id, course_id, name, restricted_data, student_readable_data, is_initial_phase, course_phase_type_id) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e13', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e11', 'Application Phase', '{}', '{}', true, '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e12');

INSERT INTO student (id, first_name, last_name, email, gender) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e14', 'Ada', 'Lovelace', 'ada.lovelace@example.com', 'female');

INSERT INTO course_participation (id, course_id, student_id) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e15', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e11', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e14');
