-- Insert test data for assessment schemas
INSERT INTO public.assessment_schema (id, name, description) VALUES
('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests');

-- Insert test course phase configurations with all evaluation types open
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, start,
                                        self_evaluation_enabled, self_evaluation_start, self_evaluation_deadline,
                                        peer_evaluation_enabled, peer_evaluation_start, peer_evaluation_deadline,
                                        tutor_evaluation_enabled, tutor_evaluation_start, tutor_evaluation_deadline) VALUES
('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2099-12-31 23:59:59+00', '2024-01-01 00:00:00+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00'),
('550e8400-e29b-41d4-a716-446655440000', '34561b6b-3c3a-4bc6-ba42-69eeb1514da9', '2099-12-31 23:59:59+00', '2024-01-01 00:00:00+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2024-01-01 00:00:00+00', '2099-12-31 23:59:59+00'),
-- Phase 3 has self evaluation enabled but not yet started
('550e8400-e29b-41d4-a716-446655440000', '44561b6b-3c3a-4bc6-ba42-69eeb1514da9', '2099-12-31 23:59:59+00', '2099-01-01 00:00:00+00',
 true, '2099-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2099-01-01 00:00:00+00', '2099-12-31 23:59:59+00',
 true, '2099-01-01 00:00:00+00', '2099-12-31 23:59:59+00');

-- A completed self evaluation locks further writes for that author in phase 2
INSERT INTO public.evaluation_completion (course_participation_id, course_phase_id, author_course_participation_id, completed_at, completed, type) VALUES
('ca42e447-60f9-4fe0-b297-2dae3f924fd7', '34561b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7', '2024-02-01 00:00:00+00', true, 'self');

-- Insert test data for feedback items
INSERT INTO public.feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id) VALUES
('11111111-1111-1111-1111-111111111111', 'positive', 'Great teamwork and communication skills!', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'da42e447-60f9-4fe0-b297-2dae3f924fd7'),
('22222222-2222-2222-2222-222222222222', 'negative', 'Need to improve time management', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ea42e447-60f9-4fe0-b297-2dae3f924fd7'),
('33333333-3333-3333-3333-333333333333', 'positive', 'Excellent problem-solving abilities', 'da42e447-60f9-4fe0-b297-2dae3f924fd7', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ca42e447-60f9-4fe0-b297-2dae3f924fd7'),
('44444444-4444-4444-4444-444444444444', 'negative', 'Could be more active in discussions', 'da42e447-60f9-4fe0-b297-2dae3f924fd7', '34561b6b-3c3a-4bc6-ba42-69eeb1514da9', 'ea42e447-60f9-4fe0-b297-2dae3f924fd7');
