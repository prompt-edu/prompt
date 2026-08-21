-- Fixture data for the presentation service tests. The schema itself comes from
-- db/migration/0001_schema.up.sql, which testutils applies first, so this file only ever
-- contains rows.
--
-- Stable IDs:
--   10000000-…-0001  course phase (individual targets, independent feedback)
--   10000000-…-0002  course phase (team targets, shared feedback)
--   20000000-…-000N  feedback categories
--   30000000-…-000N  presentation slots
--   40000000-…-000N  presentations
--   50000000-…-000N  course participations / teams used as presentation targets

INSERT INTO course_phase_config (course_phase_id, target_mode, feedback_mode) VALUES
    ('10000000-0000-0000-0000-000000000001', 'individual', 'independent'),
    ('10000000-0000-0000-0000-000000000002', 'team', 'shared'),
    -- Kept free of presentations and feedback, so category mutations are not locked and
    -- constraint conflicts can be exercised directly.
    ('10000000-0000-0000-0000-000000000003', 'individual', 'independent');

INSERT INTO feedback_category (id, course_phase_id, name, description, position) VALUES
    ('20000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', 'Delivery', 'How was it presented?', 0),
    ('20000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', 'Content', 'Was it correct?', 1),
    ('20000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', 'Teamwork', 'How did the team split the work?', 0),
    ('20000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000003', 'Delivery', 'How was it presented?', 0);

INSERT INTO presentation_slot (id, course_phase_id, start_time, end_time, location) VALUES
    ('30000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '2999-01-01 10:00:00+00', '2999-01-01 10:15:00+00', 'Room 1'),
    ('30000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '2999-01-01 10:15:00+00', '2999-01-01 10:30:00+00', 'Room 1'),
    -- Unassigned, so slot-deletion tests have a slot they are allowed to remove.
    ('30000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000001', '2999-01-01 10:30:00+00', '2999-01-01 10:45:00+00', NULL),
    ('30000000-0000-0000-0000-000000000004', '10000000-0000-0000-0000-000000000002', '2999-01-01 11:00:00+00', '2999-01-01 11:30:00+00', 'Room 2');

INSERT INTO presentation (id, course_phase_id, slot_id, target_type, target_id, target_name) VALUES
    ('40000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000001', 'individual', '50000000-0000-0000-0000-000000000001', 'Ada Lovelace'),
    ('40000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000001', '30000000-0000-0000-0000-000000000002', 'individual', '50000000-0000-0000-0000-000000000002', 'Alan Turing'),
    ('40000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000002', '30000000-0000-0000-0000-000000000004', 'team', '50000000-0000-0000-0000-000000000003', 'Team Rocket');
