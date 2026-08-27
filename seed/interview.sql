-- PROMPT demo seed: interview database.
--
-- Slots, assignments and reviews for the iPraktikumDemo Interview phase
-- (f0000002-0000-0000-0000-000000000002). The answers reference the
-- interviewQuestions configured on that phase in seed/core.sql, and the Matching
-- phase resolves the scores below over REST.
--
-- Only the demo course is seeded. The e2e suite's interview phase stays empty:
-- its specs create and clean up their own slots, keyed by location.

DELETE FROM interview_slot WHERE course_phase_id = 'f0000002-0000-0000-0000-000000000002';
DELETE FROM interview_review WHERE course_phase_id = 'f0000002-0000-0000-0000-000000000002';

INSERT INTO interview_slot (id, course_phase_id, start_time, end_time, location, capacity)
    VALUES ('1a000001-0000-0000-0000-000000000001', 'f0000002-0000-0000-0000-000000000002', '2026-03-20 09:00:00+01', '2026-03-20 09:30:00+01', 'MI 00.08.038', 2),
           ('1a000002-0000-0000-0000-000000000002', 'f0000002-0000-0000-0000-000000000002', '2026-03-20 09:30:00+01', '2026-03-20 10:00:00+01', 'MI 00.08.038', 2),
           ('1a000003-0000-0000-0000-000000000003', 'f0000002-0000-0000-0000-000000000002', '2026-03-20 10:00:00+01', '2026-03-20 10:30:00+01', 'MI 00.08.038', 2),
           ('1a000004-0000-0000-0000-000000000004', 'f0000002-0000-0000-0000-000000000002', '2026-03-21 14:00:00+01', '2026-03-21 14:30:00+01', 'MI 01.09.014', 2);

INSERT INTO interview_assignment (id, interview_slot_id, course_participation_id)
    VALUES ('1b000001-0000-0000-0000-000000000001', '1a000001-0000-0000-0000-000000000001', 'cd000001-0000-0000-0000-000000000001'),
           ('1b000002-0000-0000-0000-000000000002', '1a000001-0000-0000-0000-000000000001', 'cd000002-0000-0000-0000-000000000002'),
           ('1b000003-0000-0000-0000-000000000003', '1a000002-0000-0000-0000-000000000002', 'cd000003-0000-0000-0000-000000000003'),
           ('1b000004-0000-0000-0000-000000000004', '1a000002-0000-0000-0000-000000000002', 'cd000004-0000-0000-0000-000000000004'),
           ('1b000005-0000-0000-0000-000000000005', '1a000003-0000-0000-0000-000000000003', 'cd000005-0000-0000-0000-000000000005'),
           ('1b000006-0000-0000-0000-000000000006', '1a000003-0000-0000-0000-000000000003', 'cd000006-0000-0000-0000-000000000006');

INSERT INTO interview_review (course_phase_id, course_participation_id, score, interviewer, interview_answers)
    VALUES ('f0000002-0000-0000-0000-000000000002', 'cd000001-0000-0000-0000-000000000001', 5, 'Lector Lector', '[{"questionID": 1, "answer": "<p>Two SwiftUI apps published to TestFlight.</p>"}, {"questionID": 2, "answer": "<p>Strong iOS background, took the lead in every group project.</p>"}]'),
           ('f0000002-0000-0000-0000-000000000002', 'cd000002-0000-0000-0000-000000000002', 4, 'Lector Lector', '[{"questionID": 1, "answer": "<p>A Kotlin side project and a university web app.</p>"}, {"questionID": 2, "answer": "<p>Solid all round, comfortable pairing.</p>"}]'),
           ('f0000002-0000-0000-0000-000000000002', 'cd000003-0000-0000-0000-000000000003', 4, 'Eduard Eduard', '[{"questionID": 1, "answer": "<p>A student-run scheduling tool.</p>"}, {"questionID": 2, "answer": "<p>Good communicator, asks for feedback early.</p>"}]'),
           ('f0000002-0000-0000-0000-000000000002', 'cd000004-0000-0000-0000-000000000004', 3, 'Eduard Eduard', '[{"questionID": 1, "answer": "<p>Mostly coursework, no shipped app yet.</p>"}, {"questionID": 2, "answer": "<p>Limited Swift experience, but eager to learn.</p>"}]'),
           ('f0000002-0000-0000-0000-000000000002', 'cd000005-0000-0000-0000-000000000005', 5, 'Eduard Eduard', '[{"questionID": 1, "answer": "<p>A games portfolio with two Unity titles.</p>"}, {"questionID": 2, "answer": "<p>Excellent portfolio and a clear sense of ownership.</p>"}]'),
           ('f0000002-0000-0000-0000-000000000002', 'cd000006-0000-0000-0000-000000000006', 3, 'Lector Lector', '[{"questionID": 1, "answer": "<p>A first Android prototype from a seminar.</p>"}, {"questionID": 2, "answer": "<p>Motivated, needs mentoring in the first weeks.</p>"}]');
