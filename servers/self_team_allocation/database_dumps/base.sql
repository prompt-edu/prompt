--
-- Minimal dataset for self-team allocation module tests
--

INSERT INTO public.team (id, name, course_phase_id, created_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Alpha Team', '11111111-1111-1111-1111-111111111111', '2024-01-01 10:00:00'),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Beta Team', '11111111-1111-1111-1111-111111111111', '2024-01-02 10:00:00'),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Gamma Team', '22222222-2222-2222-2222-222222222222', '2024-01-03 10:00:00');

INSERT INTO public.assignments (
    id,
    course_participation_id,
    team_id,
    course_phase_id,
    created_at,
    updated_at,
    student_first_name,
    student_last_name
) VALUES
    ('11111111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'aaaa1111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '2024-01-05 10:00:00', '2024-01-05 10:00:00', 'Alice', 'Anderson'),
    ('22222222-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'bbbb1111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111', '2024-01-06 11:00:00', '2024-01-06 11:00:00', 'Bob', 'Brown'),
    ('33333333-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'cccc1111-1111-1111-1111-111111111111', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '11111111-1111-1111-1111-111111111111', '2024-01-07 12:00:00', '2024-01-07 12:00:00', 'Charlie', 'Clark'),
    ('44444444-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'dddd2222-2222-2222-2222-222222222222', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '22222222-2222-2222-2222-222222222222', '2024-02-01 09:00:00', '2024-02-01 09:00:00', 'Dana', 'Davis');

INSERT INTO public.timeframe (course_phase_id, starttime, endtime) VALUES
    ('11111111-1111-1111-1111-111111111111', '2020-01-01 00:00:00+00', '2035-01-01 00:00:00+00'),
    ('22222222-2222-2222-2222-222222222222', '2099-01-01 00:00:00+00', '2100-01-01 00:00:00+00');

INSERT INTO public.tutor (course_phase_id, course_participation_id, first_name, last_name, team_id) VALUES
    ('11111111-1111-1111-1111-111111111111', 'eeee1111-1111-1111-1111-111111111111', 'Tara', 'Tutor', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa');

-- Tutor scope phase: a project lead (pl34ead) leads Lovelace and Turing, Hamilton has no tutor.
INSERT INTO public.team (id, name, course_phase_id, created_at) VALUES
    ('dddddddd-0000-0000-0000-000000000001', 'Lovelace', '33333333-3333-3333-3333-333333333333', '2024-03-01 10:00:00'),
    ('dddddddd-0000-0000-0000-000000000002', 'Turing', '33333333-3333-3333-3333-333333333333', '2024-03-01 10:00:00'),
    ('dddddddd-0000-0000-0000-000000000003', 'Hamilton', '33333333-3333-3333-3333-333333333333', '2024-03-01 10:00:00');

INSERT INTO public.assignments (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name) VALUES
    ('dddddddd-aaaa-0000-0000-000000000001', 'dddd0001-0000-0000-0000-000000000001', 'dddddddd-0000-0000-0000-000000000001', '33333333-3333-3333-3333-333333333333', 'Lia', 'Lovelace'),
    ('dddddddd-aaaa-0000-0000-000000000002', 'dddd0002-0000-0000-0000-000000000002', 'dddddddd-0000-0000-0000-000000000002', '33333333-3333-3333-3333-333333333333', 'Tom', 'Turing'),
    ('dddddddd-aaaa-0000-0000-000000000003', 'dddd0003-0000-0000-0000-000000000003', 'dddddddd-0000-0000-0000-000000000003', '33333333-3333-3333-3333-333333333333', 'Hal', 'Hamilton');

INSERT INTO public.tutor (course_phase_id, course_participation_id, first_name, last_name, team_id, university_login) VALUES
    ('33333333-3333-3333-3333-333333333333', '99999999-9999-9999-9999-999999999997', 'Petra', 'Lead', 'dddddddd-0000-0000-0000-000000000001', 'pl34ead'),
    ('33333333-3333-3333-3333-333333333333', '99999999-9999-9999-9999-999999999997', 'Petra', 'Lead', 'dddddddd-0000-0000-0000-000000000002', 'pl34ead');
