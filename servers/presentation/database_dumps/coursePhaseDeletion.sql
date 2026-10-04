-- Fixture data for the course phase deletion tests. The schema comes from db/migration/, which
-- testutils applies first, so this file only contains rows. Both phases hold a row in every table,
-- so the tests can tell a phase-scoped deletion from one that also reaches the other phase.
--
-- Stable IDs:
--   11000000-…-0001  course phase under deletion
--   11000000-…-0002  course phase that must be retained
--   21000000-…-000N  feedback categories
--   31000000-…-000N  presentation slots
--   41000000-…-000N  presentations
--   51000000-…-000N  course participations used as presentation targets
--   61000000-…-000N  presentation materials
--   71000000-…-000N  feedback forms

INSERT INTO course_phase_config (course_phase_id, target_mode, feedback_mode) VALUES
    ('11000000-0000-0000-0000-000000000001', 'individual', 'independent'),
    ('11000000-0000-0000-0000-000000000002', 'individual', 'independent');

INSERT INTO feedback_category (id, course_phase_id, name, description, position) VALUES
    ('21000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', 'Delivery', 'How was it presented?', 0),
    ('21000000-0000-0000-0000-000000000002', '11000000-0000-0000-0000-000000000002', 'Delivery', 'How was it presented?', 0);

INSERT INTO presentation_slot (id, course_phase_id, start_time, end_time, location) VALUES
    ('31000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', '2999-01-01 10:00:00+00', '2999-01-01 10:15:00+00', 'Room 1'),
    -- Unassigned, so the deletion has to remove slots that no presentation points at as well.
    ('31000000-0000-0000-0000-000000000002', '11000000-0000-0000-0000-000000000001', '2999-01-01 10:15:00+00', '2999-01-01 10:30:00+00', NULL),
    ('31000000-0000-0000-0000-000000000003', '11000000-0000-0000-0000-000000000002', '2999-01-01 11:00:00+00', '2999-01-01 11:15:00+00', 'Room 2');

INSERT INTO presentation (id, course_phase_id, slot_id, target_type, target_id, target_name) VALUES
    ('41000000-0000-0000-0000-000000000001', '11000000-0000-0000-0000-000000000001', '31000000-0000-0000-0000-000000000001', 'individual', '51000000-0000-0000-0000-000000000001', 'Ada Lovelace'),
    ('41000000-0000-0000-0000-000000000002', '11000000-0000-0000-0000-000000000002', '31000000-0000-0000-0000-000000000003', 'individual', '51000000-0000-0000-0000-000000000002', 'Alan Turing');

INSERT INTO presentation_material (id, presentation_id, original_filename, content_type, size_bytes, storage_key, state, uploader_user_id, uploader_name) VALUES
    ('61000000-0000-0000-0000-000000000001', '41000000-0000-0000-0000-000000000001', 'slides.pdf', 'application/pdf', 1024,
     'presentations/11000000-0000-0000-0000-000000000001/41000000-0000-0000-0000-000000000001/61000000-0000-0000-0000-000000000001/slides.pdf', 'ready', 'ada', 'Ada Lovelace'),
    ('61000000-0000-0000-0000-000000000002', '41000000-0000-0000-0000-000000000001', 'draft.pdf', 'application/pdf', 0,
     'presentations/11000000-0000-0000-0000-000000000001/41000000-0000-0000-0000-000000000001/61000000-0000-0000-0000-000000000002/draft.pdf', 'pending', 'ada', 'Ada Lovelace'),
    ('61000000-0000-0000-0000-000000000003', '41000000-0000-0000-0000-000000000002', 'slides.pdf', 'application/pdf', 1024,
     'presentations/11000000-0000-0000-0000-000000000002/41000000-0000-0000-0000-000000000002/61000000-0000-0000-0000-000000000003/slides.pdf', 'ready', 'alan', 'Alan Turing');

INSERT INTO feedback_form (id, presentation_id, scope_key, evaluator_user_id, evaluator_name, status, submitted_at) VALUES
    ('71000000-0000-0000-0000-000000000001', '41000000-0000-0000-0000-000000000001', 'user:grace', 'grace', 'Grace Hopper', 'submitted', '2999-01-01 10:20:00+00'),
    ('71000000-0000-0000-0000-000000000002', '41000000-0000-0000-0000-000000000002', 'user:grace', 'grace', 'Grace Hopper', 'submitted', '2999-01-01 11:20:00+00');

INSERT INTO feedback_answer (feedback_form_id, category_id, value, updated_by_user_id, updated_by_name) VALUES
    ('71000000-0000-0000-0000-000000000001', '21000000-0000-0000-0000-000000000001', 'Clear and well paced.', 'grace', 'Grace Hopper'),
    ('71000000-0000-0000-0000-000000000002', '21000000-0000-0000-0000-000000000002', 'Strong demo.', 'grace', 'Grace Hopper');

INSERT INTO feedback_contributor (feedback_form_id, user_id, name) VALUES
    ('71000000-0000-0000-0000-000000000001', 'grace', 'Grace Hopper'),
    ('71000000-0000-0000-0000-000000000002', 'grace', 'Grace Hopper');
