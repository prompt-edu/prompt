INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects) VALUES
  ('3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10);

INSERT INTO course_phase_type (id, name, base_url, description) VALUES
  ('7dc1c4e8-4255-4874-80a0-0c12b958744b', 'application', 'http://example.com', 'Application phase handling'),
  ('7dc1c4e8-4255-4874-80a0-0c12b958744c', 'example_component', 'core', 'Example component phase');

INSERT INTO course_phase_type_phase_provided_output_dto (id, course_phase_type_id, dto_name, version_number, endpoint_path, specification) VALUES
  ('cccccccc-cccc-cccc-cccc-cccccccccccc', '7dc1c4e8-4255-4874-80a0-0c12b958744b', 'TestDTO', 1, 'core', '{}'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', '7dc1c4e8-4255-4874-80a0-0c12b958744b', 'ResolutionDTO', 1, 'non-core', '{}');

INSERT INTO course_phase_type_phase_required_input_dto (id, course_phase_type_id, dto_name, specification) VALUES
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '7dc1c4e8-4255-4874-80a0-0c12b958744b', 'TestDTO', '{}'),
  ('ffffffff-ffff-ffff-ffff-ffffffffffff', '7dc1c4e8-4255-4874-80a0-0c12b958744b', 'ResolutionDTO', '{}');

INSERT INTO course_phase (id, course_id, name, restricted_data, student_readable_data, is_initial_phase, course_phase_type_id) VALUES
  ('3d1f3b00-87f3-433b-a713-178c4050411b', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'Test', '{"test-key":"test-value"}', '{}', false, '7dc1c4e8-4255-4874-80a0-0c12b958744b'),
  ('92bb0532-39e5-453d-bc50-fa61ea0128b2', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'Example Phase', '{}', '{}', false, '7dc1c4e8-4255-4874-80a0-0c12b958744c'),
  ('500db7ed-2eb2-42d0-82b3-8750e12afa8a', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'Application Phase', '{}', '{}', true, '7dc1c4e8-4255-4874-80a0-0c12b958744b'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'Predecessor Phase', '{"TestDTO": "restricted-test-value"}', '{}', false, '7dc1c4e8-4255-4874-80a0-0c12b958744b'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '3f42d322-e5bf-4faa-b576-51f2cab14c2e', 'Target Phase', '{}', '{}', false, '7dc1c4e8-4255-4874-80a0-0c12b958744b');

INSERT INTO phase_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_dto_id, to_course_phase_dto_id) VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'ffffffff-ffff-ffff-ffff-ffffffffffff');
