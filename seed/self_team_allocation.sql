-- PROMPT demo seed: self_team_allocation database.
--
-- The off-graph Self Team Allocation phase of iPraktikumDemo
-- (f0000008-0000-0000-0000-000000000008). It is off the phase graph because
-- course_phase_graph is a strict chain and a real course runs either the
-- lecturer-driven Team Allocation or this student-driven alternative, never both.
--
-- Student names are denormalized copies of the core `student` rows; no foreign
-- key keeps them in sync, so they must match seed/core.sql by hand.

DELETE FROM team WHERE course_phase_id = 'f0000008-0000-0000-0000-000000000008';
DELETE FROM timeframe WHERE course_phase_id = 'f0000008-0000-0000-0000-000000000008';

INSERT INTO team (id, name, course_phase_id)
    VALUES ('2a000001-0000-0000-0000-000000000001', 'Team Aurora', 'f0000008-0000-0000-0000-000000000008'),
           ('2a000002-0000-0000-0000-000000000002', 'Team Borealis', 'f0000008-0000-0000-0000-000000000008');

INSERT INTO timeframe (course_phase_id, starttime, endtime)
    VALUES ('f0000008-0000-0000-0000-000000000008', '2026-04-07 08:00:00+00', '2026-04-14 23:59:00+00');

INSERT INTO assignments (id, course_participation_id, team_id, course_phase_id, student_first_name, student_last_name)
    VALUES ('2b000001-0000-0000-0000-000000000001', 'cd000001-0000-0000-0000-000000000001', '2a000001-0000-0000-0000-000000000001', 'f0000008-0000-0000-0000-000000000008', 'Stan', 'Stan'),
           ('2b000002-0000-0000-0000-000000000002', 'cd000003-0000-0000-0000-000000000003', '2a000001-0000-0000-0000-000000000001', 'f0000008-0000-0000-0000-000000000008', 'Alice', 'Anderson'),
           ('2b000003-0000-0000-0000-000000000003', 'cd000005-0000-0000-0000-000000000005', '2a000001-0000-0000-0000-000000000001', 'f0000008-0000-0000-0000-000000000008', 'Carla', 'Chen'),
           ('2b000004-0000-0000-0000-000000000004', 'cd000002-0000-0000-0000-000000000002', '2a000002-0000-0000-0000-000000000002', 'f0000008-0000-0000-0000-000000000008', 'Selma', 'Second'),
           ('2b000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', '2a000002-0000-0000-0000-000000000002', 'f0000008-0000-0000-0000-000000000008', 'Bruno', 'Baumann'),
           ('2b000006-0000-0000-0000-000000000006', 'cd000006-0000-0000-0000-000000000006', '2a000002-0000-0000-0000-000000000002', 'f0000008-0000-0000-0000-000000000008', 'David', 'Doerr');
