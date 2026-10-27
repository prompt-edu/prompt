INSERT INTO public.student (id, first_name, last_name, email, matriculation_number, university_login, study_degree, current_semester, study_program, gender)
VALUES
  ('66666666-6666-6666-6666-666666666666', 'Alice', 'Anderson', 'alice@example.com', '00100001', 'ab12cde', 'bachelor', 3, 'Informatics', 'diverse'),
  ('77777777-7777-7777-7777-777777777777', 'Bob',   'Brown',    'bob@example.com',   '00100002', 'cd34efg', 'master',   2, 'Informatics', 'diverse'),
  ('88888888-8888-8888-8888-888888888888', 'Carol', 'Clark',    'carol@example.com', '00100003', 'ef56ghi', 'bachelor', 1, 'Informatics', 'diverse'),
  ('99999999-9999-9999-9999-999999999999', 'Dan',   'Doe',      NULL,                '00100004', 'gh78ijk', 'master',   4, 'Informatics', 'diverse');

INSERT INTO public.course (id, name, start_date, end_date, semester_tag, course_type, restricted_data)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'Campaign Test Course',
    '2025-04-01',
    '2025-09-30',
    'ss25',
    'practical course',
    '{
      "mailingSettings": {
        "replyToEmail": "replyto@example.com",
        "replyToName": "Course Team",
        "ccAddresses": [],
        "bccAddresses": []
      }
    }'::jsonb
);

-- A separate course + phase used to verify a campaign cannot target a phase in
-- an unrelated course.
INSERT INTO public.course (id, name, start_date, end_date, semester_tag, course_type, restricted_data)
VALUES (
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'Other Course',
    '2025-04-01',
    '2025-09-30',
    'ss25',
    'practical course',
    '{}'::jsonb
);

INSERT INTO public.course_phase_type (id, name, initial_phase, base_url, description)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    'Assessment',
    false,
    'http://assessment.test/assessment/api',
    'Assessment phase'
);

INSERT INTO public.course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
VALUES (
    'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
    'dddddddd-dddd-dddd-dddd-dddddddddddd',
    'Other Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
);

INSERT INTO public.course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
VALUES
  (
    '33333333-3333-3333-3333-333333333333',
    '11111111-1111-1111-1111-111111111111',
    'Assessment Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
  ),
  (
    'cccccccc-cccc-cccc-cccc-cccccccccccc',
    '11111111-1111-1111-1111-111111111111',
    'Empty Phase',
    '{}'::jsonb,
    false,
    '22222222-2222-2222-2222-222222222222',
    '{}'::jsonb
  );

INSERT INTO public.course_participation (id, course_id, student_id)
VALUES
  ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', '66666666-6666-6666-6666-666666666666'),
  ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', '77777777-7777-7777-7777-777777777777'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888888'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '99999999-9999-9999-9999-999999999999');

INSERT INTO public.course_phase_participation (course_participation_id, course_phase_id, pass_status)
VALUES
  ('44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333', 'passed'),
  ('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333', 'failed'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '33333333-3333-3333-3333-333333333333', 'not_assessed'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '33333333-3333-3333-3333-333333333333', 'passed');
