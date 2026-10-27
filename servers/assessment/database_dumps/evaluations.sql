-- Evaluations test data
-- Insert test assessment schemas
INSERT INTO assessment_schema (id, name, description) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', 'Test Assessment Schema', 'Test schema for unit tests'),
    ('550e8400-e29b-41d4-a716-446655440001', 'Self Evaluation Schema', 'This is the default self evaluation schema.'),
    ('550e8400-e29b-41d4-a716-446655440002', 'Peer Evaluation Schema', 'This is the default peer evaluation schema.');

-- Insert test course phase configurations
INSERT INTO course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, tutor_evaluation_enabled, tutor_evaluation_start, tutor_evaluation_deadline) VALUES
    ('550e8400-e29b-41d4-a716-446655440000', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440002', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', true, '2024-01-01 00:00:00+00', '2030-12-31 23:59:59+00'),
    ('550e8400-e29b-41d4-a716-446655440000', '5179d58a-d00d-4fa7-94a5-397bc69fab03', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440001', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440002', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', false, NULL, NULL);

-- Insert test categories
INSERT INTO category (id, name, short_name, description, weight, assessment_schema_id) VALUES
    ('12345678-1234-1234-1234-123456789012', 'Test Category 1', 'TC1', 'First test category', 1, '550e8400-e29b-41d4-a716-446655440000'),
    ('22345678-1234-1234-1234-123456789012', 'Test Category 2', 'TC2', 'Second test category', 2, '550e8400-e29b-41d4-a716-446655440000');

-- Insert test competencies
INSERT INTO competency (id, name, description, category_id) VALUES
    ('c1234567-1234-1234-1234-123456789012', 'Test Competency 1', 'First test competency', '12345678-1234-1234-1234-123456789012'),
    ('c2234567-1234-1234-1234-123456789012', 'Test Competency 2', 'Second test competency', '12345678-1234-1234-1234-123456789012'),
    ('c3234567-1234-1234-1234-123456789012', 'Test Competency 3', 'Third test competency', '12345678-1234-1234-1234-123456789012');

-- Insert test evaluations
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type, evaluated_at) VALUES
    -- Self evaluations
    ('e1234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', 'self', '2024-01-15 10:00:00+00'),
    ('e2234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'very_good', '01234567-1234-1234-1234-123456789012', 'self', '2024-01-15 11:00:00+00'),

    -- Peer evaluations
    ('e3234567-1234-1234-1234-123456789012', '01234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'ok', '02234567-1234-1234-1234-123456789012', 'peer', '2024-01-16 10:00:00+00'),
    ('e4234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', 'peer', '2024-01-16 11:00:00+00'),

    -- More test data for different scenarios
    ('e5234567-1234-1234-1234-123456789012', '02234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c2234567-1234-1234-1234-123456789012', 'bad', '02234567-1234-1234-1234-123456789012', 'self', '2024-01-17 10:00:00+00'),
    ('e6234567-1234-1234-1234-123456789012', '03234567-1234-1234-1234-123456789012', '5179d58a-d00d-4fa7-94a5-397bc69fab03', 'c3234567-1234-1234-1234-123456789012', 'very_bad', '03234567-1234-1234-1234-123456789012', 'self', '2024-01-18 10:00:00+00'),

    -- Peer evaluation whose subject is no longer in the author's team, used to assert that
    -- deleting it answers 403 rather than 500
    ('e7234567-1234-1234-1234-123456789012', '0f234567-1234-1234-1234-123456789012', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'c1234567-1234-1234-1234-123456789012', 'good', '01234567-1234-1234-1234-123456789012', 'peer', '2024-01-19 10:00:00+00');
