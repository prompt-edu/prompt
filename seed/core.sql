-- PROMPT demo seed: core database.
--
-- Data only. The core server owns the schema (startup `migrate up`) and creates
-- every course_phase_type with a random UUID, so this file runs AFTER the server
-- has started and resolves phase types by name. scripts/seed.sh enforces that
-- ordering. Full documentation: docs/contributor/guide/seeding.md
--
-- The seed is authoritative for the rows it owns: it deletes them and inserts
-- them again, so re-running reconciles drift instead of accumulating. Everything
-- inside the seeded courses is therefore disposable. scripts/seed.sh runs each
-- file with --single-transaction, so the file itself carries no BEGIN/COMMIT.


-- Deleting a seeded student is the one destructive step: student rows are global
-- and course_participation.student_id is ON DELETE CASCADE, so a participation
-- created outside the seeded courses would be silently destroyed. Notes cascade
-- too, and privacy_deletion_request.student_id is ON DELETE SET NULL. Refuse
-- rather than damage anything the seed does not own.
DO $$
DECLARE
    seeded_students uuid[] := ARRAY[
        '3869f209-9a21-4595-ae0e-bc6d6a3e2d63',
        '402e1535-fb9c-494f-82f2-ab4d39d71155',
        '5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07',
        'bb9736b4-076f-4592-8197-9a839ac115fb',
        'afd0cc74-7218-4f3a-8c2b-ef88aa5a9b1e',
        'ac3d8139-723f-4d8f-89c8-b7171b41b0d3',
        '9c157166-dd37-42f6-98ab-f5fda439ced1',
        '2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee',
        '33f49be1-0106-4642-8a8c-d492c841118a',
        '777286f4-a3e7-4bcd-bf57-13d178bf582d',
        '1e41e383-2c3f-4149-bd67-54e41cdaebec',
        '28880f4d-6f8a-4826-a7c6-e2f295f6ff72',
        '23bf3123-4f0d-473c-9ef5-d0333e29fe9a',
        '6381660e-6f30-4632-bfd0-5b7dd92c6fcf',
        '5f2c6b09-b170-48a1-a6cf-30c7688df1f4',
        '65828d3e-11bc-4168-8edc-3968b53f4f83',
        '2d8c24b4-b91a-4219-9bd8-3f2502774ebc',
        '1c62c564-491b-43e3-9929-7be39509e32e',
        '5939210d-5c47-446e-ba55-3da992fd7aa6',
        'e0000005-0000-0000-0000-000000000005',
        'a5000007-0000-4000-8000-000000000007',
        'e0000009-0000-0000-0000-000000000009',
        'e0000010-0000-0000-0000-000000000010',
        'e1000001-0000-0000-0000-000000000001',
        'e1000002-0000-0000-0000-000000000002',
        'e1000003-0000-0000-0000-000000000003',
        'e1000004-0000-0000-0000-000000000004',
        'e1000005-0000-0000-0000-000000000005',
        'e1000006-0000-0000-0000-000000000006'
    ];
    seeded_courses uuid[] := ARRAY[
        'd7307be2-d3dc-496e-86f0-643bff6cc1c8',
        'e12ffe63-448d-4469-a840-1699e9b328d1',
        'be780b32-a678-4b79-ae1c-80071771d254',
        'c0000001-0000-0000-0000-000000000001',
        'c0000002-0000-0000-0000-000000000002',
        'c0000003-0000-0000-0000-000000000003'
    ];
    blocker text;
BEGIN
    SELECT string_agg(reason, ', ') INTO blocker FROM (
        SELECT 'participation outside the seeded courses' AS reason
            FROM course_participation
            WHERE student_id = ANY (seeded_students) AND NOT (course_id = ANY (seeded_courses))
        UNION
        SELECT 'privacy export' FROM privacy_export WHERE student_id = ANY (seeded_students)
        UNION
        SELECT 'privacy deletion request' FROM privacy_deletion_request WHERE student_id = ANY (seeded_students)
        UNION
        SELECT 'note' FROM note WHERE for_student = ANY (seeded_students)
    ) AS blockers;

    IF blocker IS NOT NULL THEN
        RAISE EXCEPTION 'refusing to reseed: seeded students are referenced outside the seed (%). See seed/README.md.', blocker;
    END IF;
END $$;

-- Courses cascade to their phases, participations, graph edges, application data
-- and mail campaigns. Phase types are owned by the server and never deleted.
DELETE FROM course WHERE id IN (
    'd7307be2-d3dc-496e-86f0-643bff6cc1c8',
    'e12ffe63-448d-4469-a840-1699e9b328d1',
    'be780b32-a678-4b79-ae1c-80071771d254',
    'c0000001-0000-0000-0000-000000000001',
    'c0000002-0000-0000-0000-000000000002',
    'c0000003-0000-0000-0000-000000000003'
);

DELETE FROM student WHERE id IN (
    '3869f209-9a21-4595-ae0e-bc6d6a3e2d63',
    '402e1535-fb9c-494f-82f2-ab4d39d71155',
    '5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07',
    'bb9736b4-076f-4592-8197-9a839ac115fb',
    'afd0cc74-7218-4f3a-8c2b-ef88aa5a9b1e',
    'ac3d8139-723f-4d8f-89c8-b7171b41b0d3',
    '9c157166-dd37-42f6-98ab-f5fda439ced1',
    '2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee',
    '33f49be1-0106-4642-8a8c-d492c841118a',
    '777286f4-a3e7-4bcd-bf57-13d178bf582d',
    '1e41e383-2c3f-4149-bd67-54e41cdaebec',
    '28880f4d-6f8a-4826-a7c6-e2f295f6ff72',
    '23bf3123-4f0d-473c-9ef5-d0333e29fe9a',
    '6381660e-6f30-4632-bfd0-5b7dd92c6fcf',
    '5f2c6b09-b170-48a1-a6cf-30c7688df1f4',
    '65828d3e-11bc-4168-8edc-3968b53f4f83',
    '2d8c24b4-b91a-4219-9bd8-3f2502774ebc',
    '1c62c564-491b-43e3-9929-7be39509e32e',
    '5939210d-5c47-446e-ba55-3da992fd7aa6',
    'e0000005-0000-0000-0000-000000000005',
    'a5000007-0000-4000-8000-000000000007',
    'e0000009-0000-0000-0000-000000000009',
    'e0000010-0000-0000-0000-000000000010',
    'e1000001-0000-0000-0000-000000000001',
    'e1000002-0000-0000-0000-000000000002',
    'e1000003-0000-0000-0000-000000000003',
    'e1000004-0000-0000-0000-000000000004',
    'e1000005-0000-0000-0000-000000000005',
    'e1000006-0000-0000-0000-000000000006'
);

-- The one phase type nothing owns: core creates the other eight at startup and
-- the example server never registers itself, so the seed has to create it for
-- the phases below to resolve it by name. Nothing reconciles the row
-- afterwards, which makes `base_url` seed-owned configuration: it is right
-- wherever core fronts the phase services (Docker, production), and wrong for a
-- host-run example server, which core would otherwise address as
-- http://localhost:8086/example-service/api. The id stays random like every
-- core-created type - no seed file references it.
INSERT INTO course_phase_type (id, name, initial_phase, base_url, description)
    VALUES (gen_random_uuid(), 'example_component', false,
            '{CORE_HOST}/example-service/api', 'Example phase')
    ON CONFLICT (name) DO NOTHING;

-- ─── E2E fixtures ───────────────────────────────────────────────────────────
--
-- Courses, students and phases the Playwright suite asserts against by pinned
-- UUID (e2e/src/data/constants.ts). Ported from the former e2e/seed/e2e_seed.sql.

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('3869f209-9a21-4595-ae0e-bc6d6a3e2d63', 'Niclas', 'Heun', 'niclas.heun@tum.de', '03711126', 'ge25hok', true, 'male', 'DE', 'Computer Science', 'master', 19, '2025-01-09 18:20:28.256593');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('402e1535-fb9c-494f-82f2-ab4d39d71155', 'super long text', 'Text test', 'test@supertest.de', '', '', false, 'prefer_not_to_say', 'AI', 'test', 'bachelor', 5, '2025-01-14 09:48:07.641231');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07', 'Niclas', 'Heun', 'niclas@heun.ent', '08888888', 'uu66uuu', true, 'male', 'DE', 'Computer Science', 'master', 12, '2025-01-08 14:30:29.553373');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('bb9736b4-076f-4592-8197-9a839ac115fb', 'Test User', 'User-Last_Nam', 'niclas10@test.de', '09111999', 'ge77hok', true, 'diverse', NULL, NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('afd0cc74-7218-4f3a-8c2b-ef88aa5a9b1e', 'Test User', 'TEst', 'test@test.de', '09987652', 'ab00lll', true, 'diverse', NULL, NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('ac3d8139-723f-4d8f-89c8-b7171b41b0d3', 'Test', 'test', 'test3@test3.de', '05555555', 'hh77hhh', true, 'female', NULL, NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('9c157166-dd37-42f6-98ab-f5fda439ced1', 'Niclas', 'Heun', 'niclas@heun.net', '', '', false, 'diverse', 'AD', 'Information Systems', 'bachelor', 5, '2025-01-08 18:29:02.512604');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee', 'Test', 'Test', 'test2@test.de', '04511126', 'ge88hok', true, 'diverse', 'DZ', 'Information Systems', 'bachelor', 10, '2025-01-08 18:33:36.571565');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('33f49be1-0106-4642-8a8c-d492c841118a', 'Test', 'Supertest', 'supertest@test.de', '08888889', 'ab00hhh', true, 'diverse', 'AD', 'Information Systems', 'bachelor', 5, '2025-01-08 18:36:09.56128');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('777286f4-a3e7-4bcd-bf57-13d178bf582d', 'New Test', 'user', 'niclas@heun.io', '09999222', 'oo55ooo', true, 'female', 'DZ', 'Computer Science', 'bachelor', 5, '2025-01-08 18:39:34.368859');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('1e41e383-2c3f-4149-bd67-54e41cdaebec', 'test-lol', 'test', 'niclas@test.de', '09999999', 'as45fgh', true, 'female', 'AL', 'Information Systems', 'bachelor', 5, '2025-01-08 18:42:55.407307');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('28880f4d-6f8a-4826-a7c6-e2f295f6ff72', 'Stefan', 'Heun', 'stefan@heun.io', '', '', false, 'male', 'DE', 'Computer Science', 'master', 12, '2025-01-08 20:55:17.282497');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('23bf3123-4f0d-473c-9ef5-d0333e29fe9a', 'Max', 'Mustermann', 'max.mustermann@tum.de', '09822222', 'ji79klj', true, 'male', NULL, NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('6381660e-6f30-4632-bfd0-5b7dd92c6fcf', 'test', 'student', 'looool@tum.de', '', '', false, 'diverse', 'AL', NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('5f2c6b09-b170-48a1-a6cf-30c7688df1f4', 'Nationality Test', 'Registered User', 'nation@test.de', '08877663', 'ab69lol', true, 'male', 'DE', NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('65828d3e-11bc-4168-8edc-3968b53f4f83', 'External ', 'TEst', 'amazing@test.de', '', '', false, 'female', 'DK', NULL, 'bachelor', NULL, '2025-01-07 18:21:19.97367');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('2d8c24b4-b91a-4219-9bd8-3f2502774ebc', 'Test-100', 'User-Update', 'user@test.de', '', '', false, 'diverse', 'AL', NULL, 'bachelor', NULL, '2025-01-07 18:22:05.814353');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('1c62c564-491b-43e3-9929-7be39509e32e', 'Niclas', 'Heun', 'test@leeeeeel.de', '00000000', 'hh88hhh', true, 'female', 'DE', 'Computer Science', 'master', 5, '2025-01-07 22:50:17.814704');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('5939210d-5c47-446e-ba55-3da992fd7aa6', 'Niclas', 'Heuni', 'heuni@heuni.de', '', '', false, 'prefer_not_to_say', 'DE', 'Information Systems', 'bachelor', 5, '2025-01-07 23:05:43.120086');
-- The two Keycloak e2e student users (matriculation_number/university_login
-- must match the realm user attributes in e2e/keycloak/realm.json). `student`
-- (Stan, 00000005/no42tum) is also the iPraktikumFull participant a0000001, so
-- its full-course participation resolves to a DB-derived Student role.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('e0000005-0000-0000-0000-000000000005', 'Stan', 'Stan', 'pgdp_enjoyer@example.com', '00000005', 'no42tum', true, 'male', 'DE', 'Computer Science', 'bachelor', 3, '2025-01-09 12:00:00.000000');
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('a5000007-0000-4000-8000-000000000007', 'Selma', 'Second', 'second_student@example.com', '00000007', 'st70two', true, 'female', 'DE', 'Computer Science', 'bachelor', 3, '2025-01-09 12:00:00.000000');
-- Subject of the privacy deletion-approval spec, which deletes this student. Nothing
-- else may reference it; it maps to the Keycloak user `privacy-student`.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('e0000009-0000-0000-0000-000000000009', 'Priya', 'Vacy', 'privacy_subject@example.com', '00000009', 'pv99tum', true, 'female', 'DE', 'Computer Science', 'bachelor', 3, '2025-01-09 12:00:00.000000');
-- Subject of the admin-initiated deletion spec. last_modified is far enough in the
-- past to match the student table's "not modified in N years" filter, which a student
-- created during the run never can. No Keycloak account: admins delete this one.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester, last_modified)
    VALUES ('e0000010-0000-0000-0000-000000000010', 'Ida', 'Inactive', 'inactive_subject@example.com', '00000010', 'ii10tum', true, 'female', 'DE', 'Computer Science', 'bachelor', 3, '2015-01-01 12:00:00.000000');

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('d7307be2-d3dc-496e-86f0-643bff6cc1c8', 'iPraktikum', '2024-10-13', '2025-02-14', 'ios2425', 'practical course', 10, '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', false, 'iOS practical course', 'The iPraktikum is a hands-on iOS development course.', false, NULL);
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('e12ffe63-448d-4469-a840-1699e9b328d1', 'iPraktikum-Test', '2024-12-15', '2025-03-15', 'ios2425', 'practical course', 10, '{"icon": "graduation-cap", "bg-color": "bg-green-100"}', '{"icon": "graduation-cap", "bg-color": "bg-green-100"}', false, 'Test variant', 'A test course.', false, NULL);
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('be780b32-a678-4b79-ae1c-80071771d254', 'TestCourse', '2024-12-19', '2025-04-19', 'ios2425', 'seminar', 5, '{"icon": "book", "bg-color": "bg-purple-100"}', '{"icon": "book", "bg-color": "bg-purple-100"}', false, 'Seminar course', 'A seminar.', false, NULL);
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('c0000001-0000-0000-0000-000000000001', 'iPraktikumFull', '2025-04-01', '2025-09-30', 'ios2425', 'practical course', 10, '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', false, 'Full-cycle practical course', 'A practical course spanning application, interview, matching, team allocation, and assessment. Seeded with participations and course-scoped roles for e2e.', false, NULL);
-- Owned by the application welcome-text spec, which mutates the phase's welcomeText.
-- Its own course, because only one initial phase is allowed per course.
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('c0000003-0000-0000-0000-000000000003', 'iPraktikumWelcome', '2025-04-01', '2025-09-30', 'ios2425', 'practical course', 10, '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', '{"icon": "graduation-cap", "bg-color": "bg-blue-100"}', false, 'Welcome text fixture course', 'Owns the Application phase used by the welcome-text spec.', false, NULL);

-- An open Application phase on iPraktikum, so the file-upload endpoints accept
-- uploads (applicationEndDate in the future → CheckIfCoursePhaseIsOpenApplicationPhase passes).
-- applicationStartDate + explicit externalStudentsAllowed are required by the stricter
-- GetOpenApplicationPhase query that backs the public /apply form (start<NOW; non-null bool casts).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa1111-0000-0000-0000-0000000000a1', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2099-12-31T23:59:59", "externalStudentsAllowed": false, "universityLoginAvailable": false}', true, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
-- A Self Team Allocation phase on iPraktikum (follows Application in the
-- graph), plus one on TestCourse with no participants and no graph edge: the
-- negative-auth fixture (the e2e students are not enrolled in TestCourse).
-- Plus the full phase graph for iPraktikumFull: Application -> Interview ->
-- Matching -> Team Allocation -> Assessment. Its Application phase is open
-- (start in the past, end in the far future) so it also qualifies as an open
-- application phase.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa2222-0000-0000-0000-0000000000a2', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'Self Team Allocation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Self Team Allocation'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa3333-0000-0000-0000-0000000000a3', 'be780b32-a678-4b79-ae1c-80071771d254', 'Self Team Allocation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Self Team Allocation'), '{}');
-- Standalone (no graph edge) phase owned by the lecturer-overview spec: teams
-- formed there never collide with the student journey's phase when Playwright
-- runs the two spec files in parallel workers.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa4444-0000-0000-0000-0000000000a4', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'Self Team Allocation Overview', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Self Team Allocation'), '{}');
-- The open Application phase on fullCourse. universityLoginAvailable must stay true: it is what
-- sends an already logged-in student from /apply/:phaseId straight to /apply/:phaseId/authenticated
-- instead of the login card, which the auth session-persistence spec asserts.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000001-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2099-12-31T23:59:59", "externalStudentsAllowed": true, "universityLoginAvailable": true}', true, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
-- Standalone import-mode Application phase on fullCourse (no graph edge, reach by API). Non-initial
-- because fullCourse already has an initial Application phase (unique_initial_phase_per_course), and
-- the import endpoint keys off the phase type, not is_initial_phase. applicationMode=import closes
-- the public apply flow and enables the CSV import endpoint. Owned by the application-import API spec.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000011-0000-0000-0000-000000000011', 'c0000001-0000-0000-0000-000000000001', 'CSV Import Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2099-12-31T23:59:59", "externalStudentsAllowed": false, "universityLoginAvailable": true, "applicationMode": "import"}', false, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000015-0000-0000-0000-000000000015', 'c0000001-0000-0000-0000-000000000001', 'Custom Scores Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2099-12-31T23:59:59", "externalStudentsAllowed": false, "universityLoginAvailable": true, "useCustomScores": true, "additionalScores": [{"key": "exercisescore", "name": "Exercise Score"}]}', false, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000002-0000-0000-0000-000000000002', 'c0000001-0000-0000-0000-000000000001', 'Interview', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Interview'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000003-0000-0000-0000-000000000003', 'c0000001-0000-0000-0000-000000000001', 'Matching', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Matching'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000004-0000-0000-0000-000000000004', 'c0000001-0000-0000-0000-000000000001', 'Team Allocation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Team Allocation'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000005-0000-0000-0000-000000000005', 'c0000001-0000-0000-0000-000000000001', 'Assessment', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
-- Team Allocation fixtures (no graph edge, route by URL): b3000001 is the
-- standalone lecturer-journey phase on iPraktikumFull (Stan + Selma participate)
-- so its team creation + published allocation never collide with the graph
-- Team Allocation phase (d0000004) used by the smoke / API specs. b3000003 is
-- the student-journey phase (Stan participates) so its own published allocation
-- stays isolated too. b3000002 is the TestCourse negative-auth fixture (no
-- participants; the e2e students are not enrolled in TestCourse).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('b3000001-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'Team Allocation Journey', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Team Allocation'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('b3000003-0000-0000-0000-000000000003', 'c0000001-0000-0000-0000-000000000001', 'Team Allocation Student', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Team Allocation'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('b3000002-0000-0000-0000-000000000002', 'be780b32-a678-4b79-ae1c-80071771d254', 'Team Allocation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Team Allocation'), '{}');
-- Standalone assessment fixture phases (no graph edges — the graph's UNIQUE
-- from/to constraints force a chain, and non-graph phases still route by URL,
-- they are just filtered from the course sidebar). One phase per spec file so
-- release state and schema locking never cross parallel Playwright files:
-- d0000006 = student visibility spec (Stan + Selma participate),
-- d0000007 = self-evaluation spec (Stan participates),
-- d0000008 = TestCourse negative-auth fixture (no participants; the e2e
-- students are not enrolled in TestCourse),
-- d0000009 = print spec (Stan participates),
-- d0000012 = CampusOnline grade export spec (Stan + Selma participate),
-- d0000013 = evaluation-only spec (Stan participates).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000006-0000-0000-0000-000000000006', 'c0000001-0000-0000-0000-000000000001', 'Assessment Visibility', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000007-0000-0000-0000-000000000007', 'c0000001-0000-0000-0000-000000000001', 'Assessment Self Evaluation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000008-0000-0000-0000-000000000008', 'be780b32-a678-4b79-ae1c-80071771d254', 'Assessment', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000009-0000-0000-0000-000000000009', 'c0000001-0000-0000-0000-000000000001', 'Assessment Print', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000012-0000-0000-0000-000000000012', 'c0000001-0000-0000-0000-000000000001', 'Assessment Grade Export', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000013-0000-0000-0000-000000000013', 'c0000001-0000-0000-0000-000000000001', 'Assessment Evaluation Only', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}');
-- Standalone Example phases (no graph edges, route by URL). The example phase is
-- a minimal placeholder module, so it needs no participants or config:
-- d000000f = MF smoke + lecturer-info API auth on iPraktikumFull,
-- d0000010 = TestCourse negative-auth fixture (course-lecturer of iPraktikumFull
-- is not a lecturer of TestCourse, so its info endpoint must reject them).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000f-0000-0000-0000-00000000000f', 'c0000001-0000-0000-0000-000000000001', 'Example', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'example_component'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000010-0000-0000-0000-000000000010', 'be780b32-a678-4b79-ae1c-80071771d254', 'Example', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'example_component'), '{}');
-- Standalone Matching phase (no graph edge) owned by the matching lecturer
-- re-import spec, so its pass_status mutation never collides with the graph
-- Matching phase (d0000003) used by the smoke / student / API specs.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000e-0000-0000-0000-00000000000e', 'c0000001-0000-0000-0000-000000000001', 'Matching Re-Import', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Matching'), '{}');
-- A CLOSED Application phase on TestCourse (applicationEndDate in the past):
-- the negative fixture for the public apply endpoints (GET 404, POST 400).
-- TestCourse has no other initial phase, so unique_initial_phase_per_course holds.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa5555-0000-0000-0000-0000000000a5', 'be780b32-a678-4b79-ae1c-80071771d254', 'Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2020-06-30T23:59:59", "externalStudentsAllowed": true, "universityLoginAvailable": false}', true, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
-- The welcomeText deliberately carries a script tag and an inline event handler:
-- the spec asserts DOMPurify strips both before the applicant sees the page.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000030-0000-0000-0000-000000000030', 'c0000003-0000-0000-0000-000000000003', 'Welcome Application', '{"applicationStartDate": "2020-01-01T00:00:00", "applicationEndDate": "2099-12-31T23:59:59", "externalStudentsAllowed": true, "universityLoginAvailable": true, "welcomeText": "<p>Welcome to the PROMPT e2e welcome course.</p><p><a href=\"https://example.com/handbook\">Course handbook</a></p><script>window.__welcomeTextXss=1</script><img src=\"x\" onerror=\"window.__welcomeTextXss=1\">"}', true, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}');
-- Certificate phases. d000000d is the graph-tail phase on iPraktikumFull
-- (smoke + API-auth reads, left unconfigured). d000000a / d000000b / d0000016 are
-- standalone fixtures (no graph edge, route by URL) so the lecturer journey's
-- template/release state, the student journey's downloads and the download-page
-- instructor text never cross when Playwright runs the spec files in parallel.
-- d000000c is the TestCourse negative-auth fixture (no participants; the e2e
-- students are not enrolled).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000d-0000-0000-0000-00000000000d', 'c0000001-0000-0000-0000-000000000001', 'Certificate', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000a-0000-0000-0000-00000000000a', 'c0000001-0000-0000-0000-000000000001', 'Certificate Lecturer', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000b-0000-0000-0000-00000000000b', 'c0000001-0000-0000-0000-000000000001', 'Certificate Student', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000016-0000-0000-0000-000000000016', 'c0000001-0000-0000-0000-000000000001', 'Certificate Student Page Text', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}');
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d000000c-0000-0000-0000-00000000000c', 'be780b32-a678-4b79-ae1c-80071771d254', 'Certificate', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}');
-- Standalone Presentation phase on fullCourse for the Module Federation and
-- API-proxy smoke test. It is intentionally not part of the graph so the
-- existing course lifecycle fixtures remain unchanged.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('d0000014-0000-0000-0000-000000000014', 'c0000001-0000-0000-0000-000000000001', 'Presentation', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Presentation'), '{}');
-- Standalone interview fixture phase (no graph edge, see above):
-- aaaa6666 = TestCourse negative-auth fixture for the interview API (no
-- participants; the e2e users hold no TestCourse roles).
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('aaaa6666-0000-0000-0000-0000000000a6', 'be780b32-a678-4b79-ae1c-80071771d254', 'Interview', '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Interview'), '{}');

INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('aaaa1111-0000-0000-0000-0000000000a1', 'aaaa2222-0000-0000-0000-0000000000a2');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('d0000001-0000-0000-0000-000000000001', 'd0000002-0000-0000-0000-000000000002');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('d0000002-0000-0000-0000-000000000002', 'd0000003-0000-0000-0000-000000000003');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('d0000003-0000-0000-0000-000000000003', 'd0000004-0000-0000-0000-000000000004');
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('d0000004-0000-0000-0000-000000000004', 'd0000005-0000-0000-0000-000000000005');
-- Certificate appended to the tail of the iPraktikumFull chain (after Assessment).
INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('d0000005-0000-0000-0000-000000000005', 'd000000d-0000-0000-0000-00000000000d');

-- The Keycloak e2e users `student` / `student2` enrolled in iPraktikum
-- (matched to the student rows below via matriculation_number/university_login
-- token claims). Neither is enrolled in TestCourse — its self team allocation
-- phase is the negative-auth fixture.
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('ca000005-0000-4000-8000-000000000005', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'e0000005-0000-0000-0000-000000000005');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('ca000007-0000-4000-8000-000000000007', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'a5000007-0000-4000-8000-000000000007');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('ca000009-0000-4000-8000-000000000009', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'e0000009-0000-0000-0000-000000000009');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('ca000010-0000-4000-8000-000000000010', 'd7307be2-d3dc-496e-86f0-643bff6cc1c8', 'e0000010-0000-0000-0000-000000000010');
-- iPraktikumFull participations. a0000001 is the same Keycloak `student` user (Stan,
-- 00000005/no42tum) that the iPraktikum self team allocation fixtures above use.
-- ca000008 enrolls `student2` (Selma) too, so assessment visibility tests have a
-- second in-course student.
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'e0000005-0000-0000-0000-000000000005');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'c0000001-0000-0000-0000-000000000001', 'a5000007-0000-4000-8000-000000000007');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'c0000001-0000-0000-0000-000000000001', '3869f209-9a21-4595-ae0e-bc6d6a3e2d63');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'c0000001-0000-0000-0000-000000000001', '5eb545c2-c2eb-4c77-9c0f-46ccf7c45d07');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'c0000001-0000-0000-0000-000000000001', '2428d311-4ad4-4d91-a46e-e5e2a5a4a3ee');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000005-0000-0000-0000-000000000005', 'c0000001-0000-0000-0000-000000000001', '23bf3123-4f0d-473c-9ef5-d0333e29fe9a');
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('a0000006-0000-0000-0000-000000000006', 'c0000001-0000-0000-0000-000000000001', '777286f4-a3e7-4bcd-bf57-13d178bf582d');

-- Both e2e students participate in the iPraktikum self team allocation phase
-- (needed by the student UI's own-participation lookup and the lecturer
-- participants table; backend auth checks course-level enrollment).
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000005-0000-4000-8000-000000000005', 'aaaa2222-0000-0000-0000-0000000000a2', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000007-0000-4000-8000-000000000007', 'aaaa2222-0000-0000-0000-0000000000a2', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000005-0000-4000-8000-000000000005', 'aaaa4444-0000-0000-0000-0000000000a4', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000007-0000-4000-8000-000000000007', 'aaaa4444-0000-0000-0000-0000000000a4', '{}', 'passed', '{}');
-- Funnel across the iPraktikumFull graph. course_participation a0000001 (the seeded `student` user)
-- appears in every phase; the roster narrows toward Assessment.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000005-0000-0000-0000-000000000005', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000006-0000-0000-0000-000000000006', 'd0000001-0000-0000-0000-000000000001', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000002-0000-0000-0000-000000000002', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'd0000002-0000-0000-0000-000000000002', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'd0000002-0000-0000-0000-000000000002', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'd0000002-0000-0000-0000-000000000002', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000005-0000-0000-0000-000000000005', 'd0000002-0000-0000-0000-000000000002', '{}', 'passed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000003-0000-0000-0000-000000000003', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'd0000003-0000-0000-0000-000000000003', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'd0000003-0000-0000-0000-000000000003', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'd0000003-0000-0000-0000-000000000003', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000004-0000-0000-0000-000000000004', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'd0000004-0000-0000-0000-000000000004', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'd0000004-0000-0000-0000-000000000004', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'd0000004-0000-0000-0000-000000000004', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000005-0000-0000-0000-000000000005', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000002-0000-0000-0000-000000000002', 'd0000005-0000-0000-0000-000000000005', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000003-0000-0000-0000-000000000003', 'd0000005-0000-0000-0000-000000000005', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, last_modified, student_readable_data)
    VALUES ('a0000004-0000-0000-0000-000000000004', 'd0000005-0000-0000-0000-000000000005', '{}', 'not_assessed', '2025-01-09 18:20:28.256593', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000015-0000-0000-0000-000000000015', '{"exercisescore": 87.5}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'd0000015-0000-0000-0000-000000000015', '{}', 'not_assessed', '{}');
-- Standalone assessment fixture phases (see the course_phase inserts below):
-- Stan + Selma in the visibility phase, Stan in the self-evaluation phase,
-- Stan in the print phase, Stan + Selma in the grade export phase, Stan in the
-- evaluation-only phase.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000006-0000-0000-0000-000000000006', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'd0000006-0000-0000-0000-000000000006', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000007-0000-0000-0000-000000000007', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000009-0000-0000-0000-000000000009', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000012-0000-0000-0000-000000000012', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'd0000012-0000-0000-0000-000000000012', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000013-0000-0000-0000-000000000013', '{}', 'not_assessed', '{}');
-- Standalone matching re-import phase (see the course_phase insert below):
-- Stan + Selma participate, each carrying a matching score in restricted_data.
-- pass_status starts 'not_assessed' so the lecturer re-import journey can flip
-- it to 'passed' without depending on other specs.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd000000e-0000-0000-0000-00000000000e', '{"score": 90}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'd000000e-0000-0000-0000-00000000000e', '{"score": 85}', 'not_assessed', '{}');
-- Certificate phases (see the course_phase inserts below): Stan participates in
-- the graph-tail phase (smoke + API reads) and all three standalone journey phases
-- (lecturer participants table + staff download, student self-download, and the
-- instructor text on the student download page).
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd000000d-0000-0000-0000-00000000000d', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd000000a-0000-0000-0000-00000000000a', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd000000b-0000-0000-0000-00000000000b', '{}', 'passed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'd0000016-0000-0000-0000-000000000016', '{}', 'passed', '{}');
-- Standalone Team Allocation journey phase (see the course_phase inserts below):
-- Stan + Selma participate so the lecturer participants table lists them and the
-- published allocation can target Stan's participation.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'b3000001-0000-0000-0000-000000000001', '{}', 'not_assessed', '{}');
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('ca000008-0000-4000-8000-000000000008', 'b3000001-0000-0000-0000-000000000001', '{}', 'not_assessed', '{}');
-- Standalone Team Allocation student-journey phase (see the course_phase inserts
-- below): Stan participates so he can open the survey remote and read his own
-- published allocation via the API.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, pass_status, student_readable_data)
    VALUES ('a0000001-0000-0000-0000-000000000001', 'b3000003-0000-0000-0000-000000000003', '{}', 'not_assessed', '{}');

-- Required text question on the iPraktikumFull Application phase, so the
-- application journey exercises the configurable form (answersText round-trip).
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key)
    VALUES ('ab000001-0000-0000-0000-000000000001', 'd0000001-0000-0000-0000-000000000001', 'Motivation', 'Why do you want to join this course?', 'Your motivation', '', '', true, 500, 1, false, '');

INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
    VALUES ('bbbb0001-0000-0000-0000-0000000000b1', 'aaaa1111-0000-0000-0000-0000000000a1', 'Upload your CV', 'Attach your CV.', false, '.txt,.pdf', 50, 0, false, NULL);


-- ─── Demo course: iPraktikumDemo ────────────────────────────────────────────
--
-- The full-course example from issue #1564: every phase type, populated, with
-- settings. It is deliberately a course of its own rather than an extension of
-- iPraktikumFull, because the e2e journeys own the phases of that course (the
-- assessment lecturer journey, for instance, asserts "1/4 final" on its graph
-- tail) and seeded phase data there would change what they see.
--
-- The graph is a strict chain: course_phase_graph is UNIQUE on both endpoints.
-- Team Allocation and Self Team Allocation are mutually exclusive alternatives
-- in a real course, so Self Team Allocation and the developer-only Example
-- phase sit off the graph. They carry full data but are not part of the
-- sequential student flow.

INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester)
    VALUES ('e1000001-0000-0000-0000-000000000001', 'Alice', 'Anderson', 'alice.anderson@example.com', '01000001', 'al01tum', true, 'female', 'DE', 'Informatics', 'master', 2),
           ('e1000002-0000-0000-0000-000000000002', 'Bruno', 'Baumann', 'bruno.baumann@example.com', '01000002', 'br02tum', true, 'male', 'DE', 'Informatics', 'master', 3),
           ('e1000003-0000-0000-0000-000000000003', 'Carla', 'Chen', 'carla.chen@example.com', '01000003', 'ca03tum', true, 'female', 'CN', 'Information Systems', 'master', 1),
           ('e1000004-0000-0000-0000-000000000004', 'David', 'Doerr', 'david.doerr@example.com', '01000004', 'da04tum', true, 'male', 'AT', 'Informatics', 'bachelor', 5),
           ('e1000005-0000-0000-0000-000000000005', 'Elif', 'Erdem', 'elif.erdem@example.com', '01000005', 'el05tum', true, 'female', 'TR', 'Games Engineering', 'master', 2),
           ('e1000006-0000-0000-0000-000000000006', 'Felix', 'Faber', 'felix.faber@example.com', '01000006', 'fe06tum', true, 'male', 'DE', 'Informatics', 'bachelor', 6);

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
    VALUES ('c0000002-0000-0000-0000-000000000002', 'iPraktikumDemo', '2026-04-01', '2026-09-30', 'ios2526', 'practical course', 10,
            '{"icon": "graduation-cap", "bg-color": "bg-amber-100"}',
            '{"icon": "graduation-cap", "bg-color": "bg-amber-100"}',
            false, 'The seeded demo course',
            'A complete course with data in every phase: application, interview, matching, team allocation, assessment, presentation and certificate, plus an off-graph self team allocation and example phase.',
            false, NULL);

-- The chain. Each phase carries the settings its module reads out of
-- restricted_data, so the demo opens on configured screens rather than on
-- empty-state forms.
INSERT INTO course_phase (id, course_id, name, restricted_data, is_initial_phase, course_phase_type_id, student_readable_data)
    VALUES ('f0000001-0000-0000-0000-000000000001', 'c0000002-0000-0000-0000-000000000002', 'Application',
            '{"applicationStartDate": "2026-01-01T00:00:00", "applicationEndDate": "2026-03-15T23:59:59", "externalStudentsAllowed": true, "universityLoginAvailable": true, "autoAccept": false, "additionalScores": [{"key": "devSkill", "name": "Development Skill", "threshold": 50}]}',
            true, (SELECT id FROM course_phase_type WHERE name = 'Application'), '{}'),
           ('f0000002-0000-0000-0000-000000000002', 'c0000002-0000-0000-0000-000000000002', 'Interview',
            '{"interviewQuestions": [{"id": 1, "question": "Which projects has the candidate built so far?", "orderNum": 0}, {"id": 2, "question": "How well would the candidate work in a team?", "orderNum": 1}]}',
            false, (SELECT id FROM course_phase_type WHERE name = 'Interview'), '{}'),
           ('f0000003-0000-0000-0000-000000000003', 'c0000002-0000-0000-0000-000000000002', 'Matching',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Matching'), '{}'),
           ('f0000004-0000-0000-0000-000000000004', 'c0000002-0000-0000-0000-000000000002', 'Team Allocation',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Team Allocation'), '{}'),
           ('f0000005-0000-0000-0000-000000000005', 'c0000002-0000-0000-0000-000000000002', 'Assessment',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Assessment'), '{}'),
           ('f0000006-0000-0000-0000-000000000006', 'c0000002-0000-0000-0000-000000000002', 'Presentation',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Presentation'), '{}'),
           ('f0000007-0000-0000-0000-000000000007', 'c0000002-0000-0000-0000-000000000002', 'Certificate',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Certificate'), '{}'),
           ('f0000008-0000-0000-0000-000000000008', 'c0000002-0000-0000-0000-000000000002', 'Self Team Allocation',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'Self Team Allocation'), '{}'),
           ('f0000009-0000-0000-0000-000000000009', 'c0000002-0000-0000-0000-000000000002', 'Example',
            '{}', false, (SELECT id FROM course_phase_type WHERE name = 'example_component'), '{}');

INSERT INTO course_phase_graph (from_course_phase_id, to_course_phase_id)
    VALUES ('f0000001-0000-0000-0000-000000000001', 'f0000002-0000-0000-0000-000000000002'),
           ('f0000002-0000-0000-0000-000000000002', 'f0000003-0000-0000-0000-000000000003'),
           ('f0000003-0000-0000-0000-000000000003', 'f0000004-0000-0000-0000-000000000004'),
           ('f0000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005'),
           ('f0000005-0000-0000-0000-000000000005', 'f0000006-0000-0000-0000-000000000006'),
           ('f0000006-0000-0000-0000-000000000006', 'f0000007-0000-0000-0000-000000000007');

-- Stan and Selma are the Keycloak `student` / `student2` users, so a local login
-- lands on a fully populated course. cd000007 and cd000008 are the two applicants
-- who do not make it past the application phase.
INSERT INTO course_participation (id, course_id, student_id)
    VALUES ('cd000001-0000-0000-0000-000000000001', 'c0000002-0000-0000-0000-000000000002', 'e0000005-0000-0000-0000-000000000005'),
           ('cd000002-0000-0000-0000-000000000002', 'c0000002-0000-0000-0000-000000000002', 'a5000007-0000-4000-8000-000000000007'),
           ('cd000003-0000-0000-0000-000000000003', 'c0000002-0000-0000-0000-000000000002', 'e1000001-0000-0000-0000-000000000001'),
           ('cd000004-0000-0000-0000-000000000004', 'c0000002-0000-0000-0000-000000000002', 'e1000002-0000-0000-0000-000000000002'),
           ('cd000005-0000-0000-0000-000000000005', 'c0000002-0000-0000-0000-000000000002', 'e1000003-0000-0000-0000-000000000003'),
           ('cd000006-0000-0000-0000-000000000006', 'c0000002-0000-0000-0000-000000000002', 'e1000004-0000-0000-0000-000000000004'),
           ('cd000007-0000-0000-0000-000000000007', 'c0000002-0000-0000-0000-000000000002', 'e1000005-0000-0000-0000-000000000005'),
           ('cd000008-0000-0000-0000-000000000008', 'c0000002-0000-0000-0000-000000000002', 'e1000006-0000-0000-0000-000000000006');

-- The funnel: eight applicants, six of whom pass into the rest of the course.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, student_readable_data, pass_status)
    VALUES ('cd000001-0000-0000-0000-000000000001', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000002-0000-0000-0000-000000000002', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000003-0000-0000-0000-000000000003', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000004-0000-0000-0000-000000000004', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000005-0000-0000-0000-000000000005', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000006-0000-0000-0000-000000000006', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'passed'),
           ('cd000007-0000-0000-0000-000000000007', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'failed'),
           ('cd000008-0000-0000-0000-000000000008', 'f0000001-0000-0000-0000-000000000001', '{}', '{}', 'failed');

-- The interview scores live in the interview service's own interview_review
-- table (seed/interview.sql); the Matching phase resolves them over REST
-- through the participation data dependency graph wired below.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, student_readable_data, pass_status)
    VALUES ('cd000001-0000-0000-0000-000000000001', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed'),
           ('cd000002-0000-0000-0000-000000000002', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed'),
           ('cd000003-0000-0000-0000-000000000003', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed'),
           ('cd000004-0000-0000-0000-000000000004', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed'),
           ('cd000005-0000-0000-0000-000000000005', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed'),
           ('cd000006-0000-0000-0000-000000000006', 'f0000002-0000-0000-0000-000000000002', '{}', '{}', 'passed');

-- Every later phase carries the same six participants.
INSERT INTO course_phase_participation (course_participation_id, course_phase_id, restricted_data, student_readable_data, pass_status)
SELECT cp.id, phase.id, '{}', '{}', 'passed'::pass_status
FROM course_participation cp
CROSS JOIN (VALUES ('f0000003-0000-0000-0000-000000000003'::uuid),
                   ('f0000004-0000-0000-0000-000000000004'::uuid),
                   ('f0000005-0000-0000-0000-000000000005'::uuid),
                   ('f0000006-0000-0000-0000-000000000006'::uuid),
                   ('f0000007-0000-0000-0000-000000000007'::uuid),
                   ('f0000008-0000-0000-0000-000000000008'::uuid),
                   ('f0000009-0000-0000-0000-000000000009'::uuid)) AS phase(id)
WHERE cp.course_id = 'c0000002-0000-0000-0000-000000000002'
  AND cp.id NOT IN ('cd000007-0000-0000-0000-000000000007', 'cd000008-0000-0000-0000-000000000008');

-- ─── Application form and answers ───────────────────────────────────────────

INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key)
    VALUES ('fa000001-0000-0000-0000-000000000001', 'f0000001-0000-0000-0000-000000000001', 'Motivation', 'Why do you want to join the iPraktikum?', 'Your motivation', '', '', true, 500, 1, true, 'motivation');

INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key)
    VALUES ('fa000002-0000-0000-0000-000000000002', 'f0000001-0000-0000-0000-000000000001', 'Prior experience', 'Which technologies have you already worked with?', 'Select all that apply', 'Select at least one', true, 1, 4, '{Swift,SwiftUI,Kotlin,Go}', 2, true, 'experience');

INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
    VALUES ('fa000003-0000-0000-0000-000000000003', 'f0000001-0000-0000-0000-000000000001', 'Curriculum vitae', 'Attach your CV as a PDF.', false, '.pdf', 10, 3, false, NULL);

-- Answers to the upload question are deliberately absent: they need a `files`
-- row backed by a real object in SeaweedFS, which SQL alone cannot create.
INSERT INTO application_answer_text (id, application_question_id, answer, course_participation_id)
    VALUES ('fb000001-0000-0000-0000-000000000001', 'fa000001-0000-0000-0000-000000000001', 'I have wanted to build iOS apps since my first Swift tutorial.', 'cd000001-0000-0000-0000-000000000001'),
           ('fb000002-0000-0000-0000-000000000002', 'fa000001-0000-0000-0000-000000000001', 'I am looking for a project where design and engineering meet.', 'cd000002-0000-0000-0000-000000000002'),
           ('fb000003-0000-0000-0000-000000000003', 'fa000001-0000-0000-0000-000000000001', 'The industry partner projects are what draw me to this course.', 'cd000003-0000-0000-0000-000000000003'),
           ('fb000004-0000-0000-0000-000000000004', 'fa000001-0000-0000-0000-000000000001', 'I want to learn how a real team ships software.', 'cd000004-0000-0000-0000-000000000004'),
           ('fb000005-0000-0000-0000-000000000005', 'fa000001-0000-0000-0000-000000000001', 'I have built two apps on my own and want feedback from professionals.', 'cd000005-0000-0000-0000-000000000005'),
           ('fb000006-0000-0000-0000-000000000006', 'fa000001-0000-0000-0000-000000000001', 'Mobile development is the direction I want my studies to take.', 'cd000006-0000-0000-0000-000000000006'),
           ('fb000007-0000-0000-0000-000000000007', 'fa000001-0000-0000-0000-000000000001', 'I am curious about the course but have little experience yet.', 'cd000007-0000-0000-0000-000000000007'),
           ('fb000008-0000-0000-0000-000000000008', 'fa000001-0000-0000-0000-000000000001', 'A friend recommended the course to me.', 'cd000008-0000-0000-0000-000000000008');

INSERT INTO application_answer_multi_select (id, application_question_id, answer, course_participation_id)
    VALUES ('fc000001-0000-0000-0000-000000000001', 'fa000002-0000-0000-0000-000000000002', '{Swift,SwiftUI}', 'cd000001-0000-0000-0000-000000000001'),
           ('fc000002-0000-0000-0000-000000000002', 'fa000002-0000-0000-0000-000000000002', '{Swift,Go}', 'cd000002-0000-0000-0000-000000000002'),
           ('fc000003-0000-0000-0000-000000000003', 'fa000002-0000-0000-0000-000000000002', '{Kotlin}', 'cd000003-0000-0000-0000-000000000003'),
           ('fc000004-0000-0000-0000-000000000004', 'fa000002-0000-0000-0000-000000000002', '{Swift,SwiftUI,Go}', 'cd000004-0000-0000-0000-000000000004'),
           ('fc000005-0000-0000-0000-000000000005', 'fa000002-0000-0000-0000-000000000002', '{SwiftUI}', 'cd000005-0000-0000-0000-000000000005'),
           ('fc000006-0000-0000-0000-000000000006', 'fa000002-0000-0000-0000-000000000002', '{Kotlin,Go}', 'cd000006-0000-0000-0000-000000000006'),
           ('fc000007-0000-0000-0000-000000000007', 'fa000002-0000-0000-0000-000000000002', '{Go}', 'cd000007-0000-0000-0000-000000000007'),
           ('fc000008-0000-0000-0000-000000000008', 'fa000002-0000-0000-0000-000000000002', '{Kotlin}', 'cd000008-0000-0000-0000-000000000008');

INSERT INTO application_assessment (id, score, course_phase_id, course_participation_id)
    VALUES ('fd000001-0000-0000-0000-000000000001', 92, 'f0000001-0000-0000-0000-000000000001', 'cd000001-0000-0000-0000-000000000001'),
           ('fd000002-0000-0000-0000-000000000002', 88, 'f0000001-0000-0000-0000-000000000001', 'cd000002-0000-0000-0000-000000000002'),
           ('fd000003-0000-0000-0000-000000000003', 75, 'f0000001-0000-0000-0000-000000000001', 'cd000003-0000-0000-0000-000000000003'),
           ('fd000004-0000-0000-0000-000000000004', 70, 'f0000001-0000-0000-0000-000000000001', 'cd000004-0000-0000-0000-000000000004'),
           ('fd000005-0000-0000-0000-000000000005', 95, 'f0000001-0000-0000-0000-000000000001', 'cd000005-0000-0000-0000-000000000005'),
           ('fd000006-0000-0000-0000-000000000006', 64, 'f0000001-0000-0000-0000-000000000001', 'cd000006-0000-0000-0000-000000000006'),
           ('fd000007-0000-0000-0000-000000000007', 41, 'f0000001-0000-0000-0000-000000000001', 'cd000007-0000-0000-0000-000000000007'),
           ('fd000008-0000-0000-0000-000000000008', 38, 'f0000001-0000-0000-0000-000000000001', 'cd000008-0000-0000-0000-000000000008');

-- ─── Inter-phase data dependencies ──────────────────────────────────────────
--
-- The DTO descriptors belong to the phase TYPES and are created by the core
-- server on startup, so they are resolved by (type name, dto name) rather than
-- pinned. This is the wiring that makes a downstream phase see an upstream
-- phase's data: without it a phase's participants table shows no resolved
-- columns at all.

INSERT INTO participation_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_dto_id, to_course_phase_dto_id)
SELECT edge.from_phase, edge.to_phase, provided.id, required.id
FROM (VALUES
        ('f0000001-0000-0000-0000-000000000001'::uuid, 'Application', 'score',           'f0000002-0000-0000-0000-000000000002'::uuid, 'Interview',       'score'),
        ('f0000001-0000-0000-0000-000000000001'::uuid, 'Application', 'applicationAnswers', 'f0000002-0000-0000-0000-000000000002'::uuid, 'Interview',    'applicationAnswers'),
        ('f0000002-0000-0000-0000-000000000002'::uuid, 'Interview',   'score',           'f0000003-0000-0000-0000-000000000003'::uuid, 'Matching',        'score'),
        ('f0000001-0000-0000-0000-000000000001'::uuid, 'Application', 'scoreLevel',      'f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'scoreLevel'),
        ('f0000001-0000-0000-0000-000000000001'::uuid, 'Application', 'applicationAnswers', 'f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'applicationAnswers'),
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teamAllocation', 'f0000005-0000-0000-0000-000000000005'::uuid, 'Assessment',   'teamAllocation'),
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teamAllocation', 'f0000006-0000-0000-0000-000000000006'::uuid, 'Presentation', 'teamAllocation'),
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teamAllocation', 'f0000007-0000-0000-0000-000000000007'::uuid, 'Certificate',  'teamAllocation')
     ) AS edge(from_phase, from_type, from_dto, to_phase, to_type, to_dto)
JOIN course_phase_type from_type ON from_type.name = edge.from_type
JOIN course_phase_type to_type   ON to_type.name = edge.to_type
JOIN course_phase_type_participation_provided_output_dto provided
     ON provided.course_phase_type_id = from_type.id AND provided.dto_name = edge.from_dto
JOIN course_phase_type_participation_required_input_dto required
     ON required.course_phase_type_id = to_type.id AND required.dto_name = edge.to_dto;

INSERT INTO phase_data_dependency_graph (from_course_phase_id, to_course_phase_id, from_course_phase_DTO_id, to_course_phase_DTO_id)
SELECT edge.from_phase, edge.to_phase, provided.id, required.id
FROM (VALUES
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teams', 'f0000005-0000-0000-0000-000000000005'::uuid, 'Assessment',   'teams'),
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teams', 'f0000006-0000-0000-0000-000000000006'::uuid, 'Presentation', 'teams'),
        ('f0000004-0000-0000-0000-000000000004'::uuid, 'Team Allocation', 'teams', 'f0000007-0000-0000-0000-000000000007'::uuid, 'Certificate',  'teams')
     ) AS edge(from_phase, from_type, from_dto, to_phase, to_type, to_dto)
JOIN course_phase_type from_type ON from_type.name = edge.from_type
JOIN course_phase_type to_type   ON to_type.name = edge.to_type
JOIN course_phase_type_phase_provided_output_dto provided
     ON provided.course_phase_type_id = from_type.id AND provided.dto_name = edge.from_dto
JOIN course_phase_type_phase_required_input_dto required
     ON required.course_phase_type_id = to_type.id AND required.dto_name = edge.to_dto;

-- ─── Mailing ────────────────────────────────────────────────────────────────

INSERT INTO mail_campaign (id, course_id, name, subject, body, target_course_phase_id, target_pass_statuses, status, created_by_id, created_by_email, created_by_name)
    VALUES ('fe000001-0000-0000-0000-000000000001', 'c0000002-0000-0000-0000-000000000002',
            'Kickoff invitation', 'Welcome to the iPraktikum',
            E'Dear {{firstName}},\n\nyou have been accepted into the iPraktikum. The kickoff meeting takes place in the first week of the semester.\n\nBest regards,\nthe teaching team',
            'f0000001-0000-0000-0000-000000000001', '{passed}', 'draft',
            'seed-lecturer', 'lecturer@example.com', 'Seeded Lecturer');

INSERT INTO mail_campaign_recipient (id, campaign_id, course_id, course_participation_id, email, status)
SELECT ('ff0000' || lpad(row_number() OVER (ORDER BY s.last_name)::text, 2, '0') || '-0000-0000-0000-000000000001')::uuid,
       'fe000001-0000-0000-0000-000000000001', 'c0000002-0000-0000-0000-000000000002', cp.id, s.email, 'pending'
FROM course_participation cp
JOIN student s ON s.id = cp.student_id
JOIN course_phase_participation cpp
     ON cpp.course_participation_id = cp.id
    AND cpp.course_phase_id = 'f0000001-0000-0000-0000-000000000001'
WHERE cp.course_id = 'c0000002-0000-0000-0000-000000000002'
  AND cpp.pass_status = 'passed';
