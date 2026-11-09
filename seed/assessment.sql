-- PROMPT demo seed: assessment database.
--
-- The Assessment phase of iPraktikumDemo (f0000005-0000-0000-0000-000000000005):
-- a rubric, graded participants, and completed self-, peer- and tutor-evaluation rounds.
--
-- The seed brings its own assessment_schema rows rather than reusing the four
-- defaults the migrations create ('Assessment Template', 'Self/Peer/Tutor
-- Evaluation Template'): those are created with gen_random_uuid(), carry no
-- categories, and are the fallback every other phase relies on.
--
-- course_phase_config references the schemas with ON DELETE RESTRICT, so it is
-- deleted before them.

DELETE FROM results_released_mail WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM evaluation_reminder WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM action_item WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM feedback_items WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM evaluation_completion WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM evaluation WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment_completion WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM category_assessment WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM course_phase_config WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment_schema WHERE id IN ('5a000001-0000-0000-0000-000000000001', '5a000002-0000-0000-0000-000000000002', '5a000003-0000-0000-0000-000000000003',
                                          '5a000004-0000-0000-0000-000000000004');

INSERT INTO assessment_schema (id, name, description, source_phase_id)
    VALUES ('5a000001-0000-0000-0000-000000000001', 'iPraktikumDemo Rubric', 'The rubric the demo course grades against.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000002-0000-0000-0000-000000000002', 'iPraktikumDemo Self Evaluation', 'What students assess about themselves at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000003-0000-0000-0000-000000000003', 'iPraktikumDemo Tutor Evaluation', 'What students assess about their team''s tutors at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000004-0000-0000-0000-000000000004', 'iPraktikumDemo Peer Evaluation', 'What students assess about their teammates at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005');

INSERT INTO category (id, name, description, weight, short_name, assessment_schema_id)
    VALUES ('5b000001-0000-0000-0000-000000000001', 'Engineering', 'Technical contribution to the team project.', 2, 'ENG', '5a000001-0000-0000-0000-000000000001'),
           ('5b000002-0000-0000-0000-000000000002', 'Collaboration', 'Working with the team and the industry partner.', 1, 'COL', '5a000001-0000-0000-0000-000000000001'),
           ('5b000003-0000-0000-0000-000000000003', 'Self Reflection', 'How the student judges their own contribution.', 1, 'SELF', '5a000002-0000-0000-0000-000000000002'),
           ('5b000004-0000-0000-0000-000000000004', 'Tutoring', 'How the tutor supported the team.', 1, 'TUT', '5a000003-0000-0000-0000-000000000003'),
           ('5b000005-0000-0000-0000-000000000005', 'Learning', 'What the student learned over the semester.', 1, 'LRN', '5a000002-0000-0000-0000-000000000002'),
           ('5b000006-0000-0000-0000-000000000006', 'Teamwork', 'How the teammate worked with the team.', 1, 'TW', '5a000004-0000-0000-0000-000000000004'),
           ('5b000007-0000-0000-0000-000000000007', 'Contribution', 'What the teammate contributed to the project.', 1, 'CON', '5a000004-0000-0000-0000-000000000004');

INSERT INTO competency (id, category_id, name, description, weight, short_name, description_very_bad, description_bad, description_ok, description_good, description_very_good)
    VALUES ('5c000001-0000-0000-0000-000000000001', '5b000001-0000-0000-0000-000000000001', 'Code Quality', 'Readable, tested, maintainable code.', 2, 'CQ',
            'Code rarely compiles or is unreviewable.', 'Frequent review findings, little testing.', 'Reasonable code with some review findings.', 'Clean code, reviews pass with minor comments.', 'Consistently exemplary, others learn from it.'),
           ('5c000002-0000-0000-0000-000000000002', '5b000001-0000-0000-0000-000000000001', 'Architecture', 'Designing features that fit the system.', 1, 'AR',
            'No grasp of the system structure.', 'Designs conflict with the architecture.', 'Follows the given architecture.', 'Extends the architecture soundly.', 'Shapes the architecture for the whole team.'),
           ('5c000003-0000-0000-0000-000000000003', '5b000002-0000-0000-0000-000000000002', 'Teamwork', 'Reliability and support within the team.', 1, 'TW',
            'Absent and unreliable.', 'Needs constant follow-up.', 'Delivers what was agreed.', 'Actively unblocks others.', 'Holds the team together.'),
           ('5c000004-0000-0000-0000-000000000004', '5b000002-0000-0000-0000-000000000002', 'Communication', 'Clarity towards team and stakeholders.', 1, 'CM',
            'Communicates nothing.', 'Unclear and infrequent.', 'Understandable when asked.', 'Clear and proactive.', 'Sets the standard for the course.'),
           ('5c000005-0000-0000-0000-000000000005', '5b000003-0000-0000-0000-000000000003', 'Own Contribution', 'How the student rates their own impact.', 1, 'OC',
            'I contributed nothing.', 'I contributed less than I wanted.', 'I contributed my share.', 'I contributed more than my share.', 'I drove the project.'),
           ('5c000006-0000-0000-0000-000000000006', '5b000004-0000-0000-0000-000000000004', 'Availability', 'Reachable when the team needed help.', 1, 'AV',
            'Never reachable.', 'Hard to reach.', 'Reachable in the meetings.', 'Reachable within a day.', 'Always there when it mattered.'),
           ('5c000007-0000-0000-0000-000000000007', '5b000004-0000-0000-0000-000000000004', 'Feedback', 'Useful feedback on the team''s work.', 1, 'FB',
            'No feedback at all.', 'Vague feedback.', 'Feedback when asked.', 'Concrete, timely feedback.', 'Feedback that moved the project forward.'),
           ('5c000008-0000-0000-0000-000000000008', '5b000005-0000-0000-0000-000000000005', 'Handling Feedback', 'How the student used feedback.', 1, 'HF',
            'I ignored feedback.', 'I rarely acted on feedback.', 'I acted on most feedback.', 'I asked for feedback actively.', 'I turned feedback into lasting habits.'),
           ('5c000009-0000-0000-0000-000000000009', '5b000005-0000-0000-0000-000000000005', 'Problem Solving', 'How the student got unstuck.', 1, 'PS',
            'I stayed stuck.', 'I needed a lot of help.', 'I solved most problems with some help.', 'I solved problems on my own.', 'I helped others solve theirs.'),
           ('5c00000a-0000-0000-0000-00000000000a', '5b000005-0000-0000-0000-000000000005', 'Technical Growth', 'How much the student grew technically.', 1, 'TG',
            'I learned nothing new.', 'I learned a little.', 'I learned the basics of our stack.', 'I became confident with our stack.', 'I can now teach our stack to others.'),
           ('5c00000b-0000-0000-0000-00000000000b', '5b000003-0000-0000-0000-000000000003', 'Ownership', 'How much responsibility the student took.', 1, 'OW',
            'I avoided responsibility.', 'I needed reminders.', 'I owned my tasks.', 'I owned features end to end.', 'I owned the product with the team.'),
           ('5c00000c-0000-0000-0000-00000000000c', '5b000003-0000-0000-0000-000000000003', 'Time Management', 'How the student planned their work.', 1, 'TM',
            'I missed most deadlines.', 'I was often late.', 'I met most deadlines.', 'I planned ahead and met all deadlines.', 'I helped the team plan as well.'),
           ('5c00000d-0000-0000-0000-00000000000d', '5b000006-0000-0000-0000-000000000006', 'Reliability', 'Delivers what was agreed.', 1, 'RE',
            'Never delivered.', 'Often missed commitments.', 'Delivered most of what was agreed.', 'Always delivered.', 'Delivered and covered for others.'),
           ('5c00000e-0000-0000-0000-00000000000e', '5b000006-0000-0000-0000-000000000006', 'Communication', 'Keeps the team informed.', 1, 'CM',
            'Never shared progress.', 'Shared little.', 'Answered when asked.', 'Kept everyone informed.', 'Made the team''s communication better.'),
           ('5c00000f-0000-0000-0000-00000000000f', '5b000007-0000-0000-0000-000000000007', 'Technical Contribution', 'Quality and amount of work.', 1, 'TC',
            'Contributed nothing usable.', 'Contributed little.', 'Contributed a fair share.', 'Contributed more than a fair share.', 'Carried key parts of the project.');

-- Results released, self-, peer- and tutor-evaluation enabled and bound to the
-- seed's own schemas. The windows below sit on the ios2526
-- semester like every other date in the seed (see seed/README.md), so they close
-- for good once that semester is past.
INSERT INTO course_phase_config (course_phase_id, assessment_schema_id, self_evaluation_schema, start, deadline,
                                 self_evaluation_enabled, self_evaluation_start, self_evaluation_deadline,
                                 peer_evaluation_schema, peer_evaluation_enabled, peer_evaluation_start, peer_evaluation_deadline,
                                 tutor_evaluation_schema, tutor_evaluation_enabled, tutor_evaluation_start, tutor_evaluation_deadline,
                                 evaluation_results_visible, grade_suggestion_visible, action_items_visible,
                                 results_released, grading_sheet_visible, assessment_enabled)
    VALUES ('f0000005-0000-0000-0000-000000000005', '5a000001-0000-0000-0000-000000000001', '5a000002-0000-0000-0000-000000000002',
            '2026-07-01 00:00:00+02', '2026-09-20 23:59:00+02',
            true, '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
            '5a000004-0000-0000-0000-000000000004', true, '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
            '5a000003-0000-0000-0000-000000000003', true, '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
            true, true, true, true, true, true);

-- Tutor assessments: one row per competency per participant.
INSERT INTO assessment (id, course_participation_id, course_phase_id, competency_id, author, author_id, score_level)
SELECT ('5d0000' || lpad(row_number() OVER (ORDER BY graded.participation, graded.competency)::text, 2, '0') || '-0000-0000-0000-000000000001')::uuid,
       graded.participation, 'f0000005-0000-0000-0000-000000000005', graded.competency, 'Seeded Tutor', 'seed-tutor', graded.level
FROM (VALUES
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'very_good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'very_good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'ok'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'very_good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'ok'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'ok'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'ok'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'bad'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'ok'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'ok'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'very_good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'very_good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'good'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5c000001-0000-0000-0000-000000000001'::uuid, 'ok'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5c000002-0000-0000-0000-000000000002'::uuid, 'bad'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5c000003-0000-0000-0000-000000000003'::uuid, 'ok'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5c000004-0000-0000-0000-000000000004'::uuid, 'ok'::score_level)
     ) AS graded(participation, competency, level);

INSERT INTO category_assessment (id, category_id, course_phase_id, course_participation_id, comment, author, author_id)
SELECT ('5e0000' || lpad(row_number() OVER (ORDER BY commented.participation, commented.category)::text, 2, '0') || '-0000-0000-0000-000000000001')::uuid,
       commented.category, 'f0000005-0000-0000-0000-000000000005', commented.participation, commented.comment, 'Seeded Tutor', 'seed-tutor'
FROM (VALUES
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Carried the networking layer and reviewed most pull requests.'),
        ('cd000001-0000-0000-0000-000000000001'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'The person the team asked when they were stuck.'),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Solid feature work, occasionally over-engineered.'),
        ('cd000002-0000-0000-0000-000000000002'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'Ran the partner meetings and kept notes for everyone.'),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Grew a lot over the semester after a slow start.'),
        ('cd000003-0000-0000-0000-000000000003'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'Dependable in stand-ups.'),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Struggled with Swift; strongest on the backend tasks.'),
        ('cd000004-0000-0000-0000-000000000004'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'Communicated blockers early, which helped the team.'),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Best technical contributor of the cohort.'),
        ('cd000005-0000-0000-0000-000000000005'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'Could involve quieter team members more.'),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5b000001-0000-0000-0000-000000000001'::uuid, 'Delivered the agreed scope, little beyond it.'),
        ('cd000006-0000-0000-0000-000000000006'::uuid, '5b000002-0000-0000-0000-000000000002'::uuid, 'Quiet, but present and reliable.')
     ) AS commented(participation, category, comment);

INSERT INTO assessment_completion (course_participation_id, course_phase_id, author, comment, grade_suggestion, completed, completed_at)
    VALUES ('cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Outstanding semester.', 1.0, true, '2026-09-21 10:00:00+02'),
           ('cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Strong all round.', 1.3, true, '2026-09-21 10:05:00+02'),
           ('cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Good development over the semester.', 2.3, true, '2026-09-21 10:10:00+02'),
           ('cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Passed, with room to grow on iOS.', 3.0, true, '2026-09-21 10:15:00+02'),
           ('cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Exceptional engineering work.', 1.0, true, '2026-09-21 10:20:00+02'),
           ('cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'Seeded Tutor', 'Solid, dependable contribution.', 2.7, true, '2026-09-21 10:25:00+02');

-- Self evaluations: the author is the participant themselves. Levels per
-- competency: Own Contribution, Ownership, Time Management (Self Reflection),
-- Handling Feedback, Problem Solving, Technical Growth (Learning).
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.participation, competency.id)::text, 2, '0') || '-0000-0000-0000-000000000001')::uuid,
       rated.participation, 'f0000005-0000-0000-0000-000000000005', competency.id, competency.level::score_level, rated.participation, 'self'
FROM (VALUES
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'very_good', 'good', 'good', 'very_good', 'good', 'very_good'),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'good', 'good', 'ok', 'good', 'very_good', 'good'),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'ok', 'good', 'ok', 'good', 'ok', 'good'),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'ok', 'good', 'ok', 'good', 'bad', 'very_good'),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'good', 'very_good', 'good', 'ok', 'very_good', 'good'),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'ok', 'ok', 'bad', 'good', 'ok', 'ok')
     ) AS rated(participation, own_contribution, ownership, time_management, handling_feedback, problem_solving, technical_growth)
CROSS JOIN LATERAL (VALUES ('5c000005-0000-0000-0000-000000000005'::uuid, rated.own_contribution),
                           ('5c00000b-0000-0000-0000-00000000000b'::uuid, rated.ownership),
                           ('5c00000c-0000-0000-0000-00000000000c'::uuid, rated.time_management),
                           ('5c000008-0000-0000-0000-000000000008'::uuid, rated.handling_feedback),
                           ('5c000009-0000-0000-0000-000000000009'::uuid, rated.problem_solving),
                           ('5c00000a-0000-0000-0000-00000000000a'::uuid, rated.technical_growth)) AS competency(id, level);

INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
    VALUES ('60000001-0000-0000-0000-000000000001', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', true, '2026-06-29 18:00:00+02', 'self'),
           ('60000002-0000-0000-0000-000000000002', 'cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', true, '2026-06-29 18:10:00+02', 'self'),
           ('60000003-0000-0000-0000-000000000003', 'cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', true, '2026-06-29 18:20:00+02', 'self'),
           ('60000004-0000-0000-0000-000000000004', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', true, '2026-06-29 18:30:00+02', 'self'),
           ('60000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', true, '2026-06-29 18:40:00+02', 'self'),
           ('60000006-0000-0000-0000-000000000006', 'cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', true, '2026-06-29 18:50:00+02', 'self');

-- Peer evaluations: every student rates both teammates (Team Kepler: cd1, cd3,
-- cd5; Team Hubble: cd2, cd4, cd6, see team_allocation.sql). Levels per
-- competency: Reliability, Communication, Technical Contribution.
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.peer, rated.author, competency.id)::text, 2, '0') || '-0000-0000-0000-000000000003')::uuid,
       rated.peer, 'f0000005-0000-0000-0000-000000000005', competency.id, competency.level::score_level, rated.author, 'peer'
FROM (VALUES
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, 'very_good', 'good', 'very_good'),
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, 'good', 'very_good', 'good'),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, 'good', 'ok', 'good'),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, 'ok', 'ok', 'good'),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, 'very_good', 'good', 'very_good'),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, 'very_good', 'very_good', 'good'),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, 'good', 'very_good', 'good'),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, 'good', 'good', 'ok'),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, 'ok', 'bad', 'ok'),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, 'bad', 'ok', 'ok'),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, 'good', 'ok', 'good'),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, 'ok', 'good', 'ok')
     ) AS rated(peer, author, reliability, communication, technical_contribution)
CROSS JOIN LATERAL (VALUES ('5c00000d-0000-0000-0000-00000000000d'::uuid, rated.reliability),
                           ('5c00000e-0000-0000-0000-00000000000e'::uuid, rated.communication),
                           ('5c00000f-0000-0000-0000-00000000000f'::uuid, rated.technical_contribution)) AS competency(id, level);

-- One completion per student and teammate: course_participation_id is the teammate.
INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
SELECT ('600001' || lpad(row_number() OVER (ORDER BY peer, author)::text, 2, '0') || '-0000-0000-0000-000000000003')::uuid,
       peer, 'f0000005-0000-0000-0000-000000000005', author, true, '2026-06-29 18:30:00+02', 'peer'
FROM (SELECT DISTINCT course_participation_id AS peer, author_course_participation_id AS author
      FROM evaluation
      WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005' AND type = 'peer') AS pairs;

-- Tutor evaluations: every student rates each tutor of their team (see
-- team_allocation.sql) on both tutoring competencies.
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.tutor, rated.author, competency.id)::text, 2, '0') || '-0000-0000-0000-000000000002')::uuid,
       rated.tutor, 'f0000005-0000-0000-0000-000000000005', competency.id, competency.level, rated.author, 'tutor'
FROM (VALUES
        ('cd000009-0000-0000-0000-000000000009'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, 'very_good'::score_level, 'good'::score_level),
        ('cd000009-0000-0000-0000-000000000009'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, 'good'::score_level, 'good'::score_level),
        ('cd000009-0000-0000-0000-000000000009'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, 'very_good'::score_level, 'very_good'::score_level),
        ('cd000010-0000-0000-0000-000000000010'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, 'good'::score_level, 'very_good'::score_level),
        ('cd000010-0000-0000-0000-000000000010'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, 'bad'::score_level, 'ok'::score_level),
        ('cd000010-0000-0000-0000-000000000010'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level, 'good'::score_level),
        ('cd000011-0000-0000-0000-000000000011'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, 'ok'::score_level, 'good'::score_level),
        ('cd000011-0000-0000-0000-000000000011'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, 'good'::score_level, 'ok'::score_level),
        ('cd000011-0000-0000-0000-000000000011'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, 'ok'::score_level, 'ok'::score_level)
     ) AS rated(tutor, author, availability, feedback)
CROSS JOIN LATERAL (VALUES ('5c000006-0000-0000-0000-000000000006'::uuid, rated.availability),
                           ('5c000007-0000-0000-0000-000000000007'::uuid, rated.feedback)) AS competency(id, level);

-- One completion per student and tutor: course_participation_id is the tutor.
INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
SELECT ('600000' || lpad((6 + row_number() OVER (ORDER BY tutor, author))::text, 2, '0') || '-0000-0000-0000-000000000002')::uuid,
       tutor, 'f0000005-0000-0000-0000-000000000005', author, true, '2026-06-29 19:00:00+02', 'tutor'
FROM (SELECT DISTINCT course_participation_id AS tutor, author_course_participation_id AS author
      FROM evaluation
      WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005' AND type = 'tutor') AS pairs;

INSERT INTO feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id, type)
    VALUES ('6a000001-0000-0000-0000-000000000001', 'positive', 'I was always available when the team needed a reviewer.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', 'self'),
           ('6a000002-0000-0000-0000-000000000002', 'negative', 'I sometimes take on too much instead of delegating.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', 'self'),
           ('6a000003-0000-0000-0000-000000000003', 'positive', 'Her code reviews taught us a lot about Swift.', 'cd000009-0000-0000-0000-000000000009', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', 'tutor'),
           ('6a000004-0000-0000-0000-000000000004', 'positive', 'Always had time for a quick call before the sprint review.', 'cd000010-0000-0000-0000-000000000010', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', 'tutor'),
           ('6a000005-0000-0000-0000-000000000005', 'negative', 'Feedback on pull requests sometimes took several days.', 'cd000010-0000-0000-0000-000000000010', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'tutor'),
           ('6a000006-0000-0000-0000-000000000006', 'positive', 'Bruno always made time to pair on the backend.', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', 'peer'),
           ('6a000007-0000-0000-0000-000000000007', 'negative', 'Bruno could share his progress more often in stand-ups.', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', 'peer'),
           ('6a000008-0000-0000-0000-000000000008', 'positive', 'Alice kept our sprint board up to date for everyone.', 'cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', 'peer');

INSERT INTO action_item (id, course_phase_id, course_participation_id, action, author)
    VALUES ('6b000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'Pair with a Swift-experienced team member on the next feature.', 'Seeded Tutor'),
           ('6b000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', 'Take ownership of one feature end to end.', 'Seeded Tutor');
