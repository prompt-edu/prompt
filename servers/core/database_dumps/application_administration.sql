INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('3a774200-39a7-4656-bafb-92b7210a93c1', 'John', 'Doe', 'existingstudent@example.com', '03711111', 'ab12cde', true, 'male');

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('b1f97ee7-fd11-4556-8c75-d0c2714e7082', 'Test', 'Student', 'test@example.com', '03788888', 'cd34efg', true, 'male');

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('15ae3969-bcb7-4d5b-8245-c305d13d671b', 'Another', 'Student', 'another.student@example.com', '03788889', 'ef56ghi', true, 'male');

-- Dedicated university-account student, used only to assert the external route cannot
-- attach to a university account by name match (GHSA-5mjh-vmpg-g6xp). No other test
-- mutates this row, so the assertion is independent of test ordering.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('a0000097-0000-0000-0000-000000000097', 'Uni', 'Account', 'uni-account@example.com', '03700077', 'qq00qqq', true, 'male');

-- University student who applies again on a phase without university login, where the
-- external route is the only way in (GHSA-5mjh-vmpg-g6xp).
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender)
VALUES ('a0000096-0000-0000-0000-000000000096', 'Returning', 'Student', 'returning-student@example.com', '03700066', 'qq00qqr', true, 'male');

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('be780b32-a678-4b79-ae1c-80071771d254', 'iPraktikum', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{"icon": "apple", "bg-color": "bg-orange-100"}');

-- Own course, because only one initial phase is allowed per course.
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('c0000099-0000-0000-0000-000000000099', 'Welcome Text Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}');

-- Own course for an application phase that does NOT allow external students, used
-- to assert the external route is rejected there (GHSA-5mjh-vmpg-g6xp).
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('c0000098-0000-0000-0000-000000000098', 'External Disabled Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}');

-- Own course for an application phase without university login and without external
-- students: university students apply through the external route there.
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('c0000097-0000-0000-0000-000000000097', 'No Login Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}');

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d1', 'Team Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}');

INSERT INTO course_phase_type (id, name, initial_phase, description) VALUES ('48d22f19-6cc0-417b-ac25-415fb40f2030', 'Intro Course', false, 'Introduces course basics');

INSERT INTO course_phase_type (id, name, initial_phase, description) VALUES ('96fb1001-b21c-4527-8b6f-2fd5f4ba3abc', 'Application', true, 'Handles application intake');

INSERT INTO course_phase_type (id, name, initial_phase, description) VALUES ('627b6fb9-2106-4fce-ba6d-b68eeb546382', 'Team Phase', false, 'Runs team collaboration phase');

INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('4179d58a-d00d-4fa7-94a5-397bc69fab02', 'be780b32-a678-4b79-ae1c-80071771d254', 'Dev Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": true, "universityLoginAvailable": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');

-- Carries a welcomeText; the phase above deliberately has none, so both branches are covered.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('d0000099-0000-0000-0000-000000000099', 'c0000099-0000-0000-0000-000000000099', 'Welcome Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": true, "universityLoginAvailable": true, "welcomeText": "<p>Welcome to the course.</p>"}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');

-- Application phase with external students disabled (externalStudentsAllowed false).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('d0000098-0000-0000-0000-000000000098', 'c0000098-0000-0000-0000-000000000098', 'External Disabled Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": false, "universityLoginAvailable": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');

-- Application phase without university login and with external students disabled.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('d0000097-0000-0000-0000-000000000097', 'c0000097-0000-0000-0000-000000000097', 'No Login Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": false, "universityLoginAvailable": false}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');

INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('7062236a-e290-487c-be41-29b24e0afc64', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'New Team Phase', '{}', false, '627b6fb9-2106-4fce-ba6d-b68eeb546382');

INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('e12ffe63-448d-4469-a840-1699e9b328d3', 'e12ffe63-448d-4469-a840-1699e9b328d1', 'Intro Course', '{}', false, '48d22f19-6cc0-417b-ac25-415fb40f2030');

INSERT INTO course_participation (id, course_id, student_id) VALUES ('82d7efae-d545-4cc5-9b94-5d0ee1e50d25', 'be780b32-a678-4b79-ae1c-80071771d254', 'b1f97ee7-fd11-4556-8c75-d0c2714e7082');

INSERT INTO course_participation (id, course_id, student_id) VALUES ('32aa070e-67c3-4a69-852a-ba3b5e849a4d', 'be780b32-a678-4b79-ae1c-80071771d254', '15ae3969-bcb7-4d5b-8245-c305d13d671b');

INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status) VALUES ('82d7efae-d545-4cc5-9b94-5d0ee1e50d25', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'passed');

INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status) VALUES ('32aa070e-67c3-4a69-852a-ba3b5e849a4d', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '{}', 'not_assessed');

INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num) VALUES ('a6a04042-95d1-4765-8592-caf9560c8c3c', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Motivation', 'You should fill out the motivation why you want to take this absolutely amazing course.', 'Enter your motivation.', '', 'You are not allowed to enter more than 500 chars. ', true, 500, 3);

INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num) VALUES ('fc8bda6d-280e-4a5e-9ebd-4bd8b68aab75', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Expierence', '', '', '', '', false, 500, 1);

INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num) VALUES ('65e25b73-ce47-4536-b651-a1632347d733', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Taken Courses', 'Which courses have you already taken ad the chair', '', '', false, 0, 3, '{Ferienakademie,Patterns,"Interactive Learning"}', 4);

INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num) VALUES ('383a9590-fba2-4e6b-a32b-88895d55fb9b', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Available Devices', '', '', '', false, 0, 4, '{iPhone,iPad,MacBook,Vision}', 2);

INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
VALUES
    ('b1b04042-95d1-4765-8592-caf9560c8c3d', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Resume Upload', 'Please upload your resume', true, '.pdf,.doc,.docx', 10, 3, false, null),
    ('c2c04042-95d1-4765-8592-caf9560c8c3e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Portfolio', 'Upload your portfolio (optional)', false, '.pdf,.zip', 20, 4, false, null);

INSERT INTO files (
    id,
    filename,
    original_filename,
    content_type,
    size_bytes,
    storage_key,
    storage_provider,
    uploaded_by_user_id,
    uploaded_by_email,
    course_phase_id,
    description,
    tags
) VALUES (
    'd3d04042-95d1-4765-8592-caf9560c8c3f',
    'resume_seeded.pdf',
    'resume.pdf',
    'application/pdf',
    1024,
    'course-phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/resume_seeded.pdf',
    'seaweedfs',
    'external',
    'seed@example.com',
    '4179d58a-d00d-4fa7-94a5-397bc69fab02',
    'Seed file for application router tests',
    '{application,resume}'
),
(
    'd3d04042-95d1-4765-8592-caf9560c8c40',
    'resume_applicant.pdf',
    'resume.pdf',
    'application/pdf',
    1024,
    'course-phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/resume_applicant.pdf',
    'seaweedfs',
    'applicant-user-id',
    'existingstudent@example.com',
    '4179d58a-d00d-4fa7-94a5-397bc69fab02',
    'Seed file uploaded by the authenticated test applicant',
    '{application,resume}'
),
(
    'd3d04042-95d1-4765-8592-caf9560c8c41',
    'resume_other_phase.pdf',
    'resume.pdf',
    'application/pdf',
    1024,
    'course-phase/d0000099-0000-0000-0000-000000000099/resume_other_phase.pdf',
    'seaweedfs',
    'external',
    'seed@example.com',
    'd0000099-0000-0000-0000-000000000099',
    'Seed file uploaded for another course phase',
    '{application,resume}'
);

INSERT INTO files (
    id,
    filename,
    original_filename,
    content_type,
    size_bytes,
    storage_key,
    storage_provider,
    uploaded_by_user_id,
    uploaded_by_email,
    course_phase_id,
    description,
    tags,
    deleted_at
) VALUES (
    'd3d04042-95d1-4765-8592-caf9560c8c42',
    'resume_deleted.pdf',
    'resume.pdf',
    'application/pdf',
    1024,
    'course-phase/4179d58a-d00d-4fa7-94a5-397bc69fab02/resume_deleted.pdf',
    'seaweedfs',
    'applicant-user-id',
    'existingstudent@example.com',
    '4179d58a-d00d-4fa7-94a5-397bc69fab02',
    'Soft-deleted seed file of the authenticated test applicant',
    '{application,resume}',
    CURRENT_TIMESTAMP
);

-- Application phases outside the intended window, used to assert that submissions and
-- uploads are rejected there just like the public listing hides them (GHSA-pv93-3r7g-rhpq).
-- Each gets its own course, because only one initial phase is allowed per course.
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data) VALUES ('c0000095-0000-0000-0000-000000000095', 'Not Yet Open Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}');
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, archived, archived_on) VALUES ('c0000094-0000-0000-0000-000000000094', 'Archived Course', '2024-10-01', '2025-01-01', 'ios24245', 'practical course', 10, '{}', true, CURRENT_TIMESTAMP);

INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('d0000095-0000-0000-0000-000000000095', 'c0000095-0000-0000-0000-000000000095', 'Not Yet Open Application', '{"applicationEndDate": "2099-01-18T00:00:00.000Z", "applicationStartDate": "2098-12-24T00:00:00.000Z", "externalStudentsAllowed": true, "universityLoginAvailable": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id) VALUES ('d0000094-0000-0000-0000-000000000094', 'c0000094-0000-0000-0000-000000000094', 'Archived Application', '{"applicationEndDate": "2030-01-18T00:00:00.000Z", "applicationStartDate": "2024-12-24T00:00:00.000Z", "externalStudentsAllowed": true, "universityLoginAvailable": true}', true, '96fb1001-b21c-4527-8b6f-2fd5f4ba3abc');
