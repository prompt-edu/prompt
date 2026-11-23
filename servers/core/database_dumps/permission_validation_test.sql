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

-- An org with a child org, for the org rules of the course checks. The ase org has a course (with a
-- phase and a participation) and a template; its child org ase-lab has a course of its own.
INSERT INTO org (id, parent_org_id, name, slug, created_at) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e21', NULL, 'Applied Software Engineering', 'ase', '2026-01-01 00:00:00+00'),
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e22', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e21', 'ASE Lab', 'ase-lab', '2026-01-01 00:00:00+00');

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, template, org_id) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e23', 'Agile', '2025-10-01', '2026-02-01', 'ws2526', 'practical course', 10, false, '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e21'),
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e26', 'AgileTemplate', NULL, NULL, 'tmpl', 'practical course', 10, true, '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e21'),
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e27', 'Lab', '2025-10-01', '2026-02-01', 'ws2526', 'practical course', 10, false, '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e22');

INSERT INTO course_phase (id, course_id, name, restricted_data, student_readable_data, is_initial_phase, course_phase_type_id) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e24', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e23', 'Application Phase', '{}', '{}', true, '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e12');

INSERT INTO course_participation (id, course_id, student_id) VALUES
  ('5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e25', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e23', '5a0e1c3e-6d1b-4c55-9e0a-2f7b8c9d0e14');
