INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests');

-- Config for phase where action items should be visible (results_released = true)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, true);

-- Config for phase where action items should not be visible (results_released = false)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, false);

-- Config for service tests (assessment is open, allows creating/editing action items)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, start, deadline, action_items_visible, results_released) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '2020-01-01 00:00:00+00', '2030-12-31 23:59:59+00', true, false);

-- Completed assessment for the visible action items test
INSERT INTO public.assessment_completion (course_participation_id, course_phase_id, completed_at, author, comment, grade_suggestion, completed) VALUES
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2025-01-15 10:00:00+00', 'test_author', 'Test completion', 4.5, true);

-- Test data for visibility tests
INSERT INTO public.action_item (id, course_phase_id, course_participation_id, action, author) VALUES
    ('a1111111-1111-1111-1111-111111111111', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for visible scenario', 'tester'),
    ('a2222222-2222-2222-2222-222222222222', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Test action item for not visible scenario', 'tester');
