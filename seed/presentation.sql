-- PROMPT demo seed: presentation database.
--
-- The Presentation phase of iPraktikumDemo
-- (f0000006-0000-0000-0000-000000000006), configured for team presentations
-- with shared feedback.
--
-- target_id references a team in the TEAM_ALLOCATION database - a cross-service
-- link by UUID value that nothing enforces. The ids must match seed/team_allocation.sql.
--
-- feedback_answer.category_id is ON DELETE RESTRICT, so presentations (and their
-- forms and answers) go before the categories and the config they hang off.

DELETE FROM presentation WHERE course_phase_id = 'f0000006-0000-0000-0000-000000000006';
DELETE FROM presentation_slot WHERE course_phase_id = 'f0000006-0000-0000-0000-000000000006';
DELETE FROM feedback_category WHERE course_phase_id = 'f0000006-0000-0000-0000-000000000006';
DELETE FROM course_phase_config WHERE course_phase_id = 'f0000006-0000-0000-0000-000000000006';

INSERT INTO course_phase_config (course_phase_id, target_mode, feedback_mode, required_material_types)
    VALUES ('f0000006-0000-0000-0000-000000000006', 'team', 'shared', '{slides}');

INSERT INTO feedback_category (id, course_phase_id, name, description, position)
    VALUES ('4a000001-0000-0000-0000-000000000001', 'f0000006-0000-0000-0000-000000000006', 'Product', 'Does the demo tell a convincing product story?', 0),
           ('4a000002-0000-0000-0000-000000000002', 'f0000006-0000-0000-0000-000000000006', 'Engineering', 'Is the technical solution sound and well explained?', 1),
           ('4a000003-0000-0000-0000-000000000003', 'f0000006-0000-0000-0000-000000000006', 'Delivery', 'Was the presentation itself clear and well paced?', 2);

INSERT INTO presentation_slot (id, course_phase_id, start_time, end_time, location)
    VALUES ('4b000001-0000-0000-0000-000000000001', 'f0000006-0000-0000-0000-000000000006', '2026-09-24 10:00:00+02', '2026-09-24 10:30:00+02', 'MI HS1'),
           ('4b000002-0000-0000-0000-000000000002', 'f0000006-0000-0000-0000-000000000006', '2026-09-24 10:30:00+02', '2026-09-24 11:00:00+02', 'MI HS1');

INSERT INTO presentation (id, course_phase_id, slot_id, target_type, target_id, target_name)
    VALUES ('4c000001-0000-0000-0000-000000000001', 'f0000006-0000-0000-0000-000000000006', '4b000001-0000-0000-0000-000000000001', 'team', '3a000001-0000-0000-0000-000000000001', 'Team Kepler'),
           ('4c000002-0000-0000-0000-000000000002', 'f0000006-0000-0000-0000-000000000006', '4b000002-0000-0000-0000-000000000002', 'team', '3a000002-0000-0000-0000-000000000002', 'Team Hubble');
