-- PROMPT demo seed: assessment database.
--
-- The Assessment phase of iPraktikumDemo (f0000005-0000-0000-0000-000000000005):
-- a rubric, graded participants, and completed self, peer and tutor evaluation
-- rounds.
--
-- The seed brings its own assessment_schema rows rather than reusing the four
-- defaults the migrations create ('Assessment Template', 'Self/Peer/Tutor
-- Evaluation Template'): those are created with gen_random_uuid(), carry no
-- categories, and are the fallback every other phase relies on, so the demo
-- config binds every schema column to one of the seed's own schemas.
--
-- Peer and tutor evaluations follow the teams the phase reads from Team
-- Allocation (seed/team_allocation.sql): Team Kepler is cd000001, cd000003 and
-- cd000005 with tutor cd000007, Team Hubble is cd000002, cd000004 and cd000006
-- with tutor cd000008.
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
DELETE FROM assessment_schema WHERE id IN ('5a000001-0000-0000-0000-000000000001', '5a000002-0000-0000-0000-000000000002',
                                           '5a000003-0000-0000-0000-000000000003', '5a000004-0000-0000-0000-000000000004');

INSERT INTO assessment_schema (id, name, description, source_phase_id)
    VALUES ('5a000001-0000-0000-0000-000000000001', 'iPraktikumDemo Rubric', 'The rubric the demo course grades against.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000002-0000-0000-0000-000000000002', 'iPraktikumDemo Self Evaluation', 'What students assess about themselves at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000003-0000-0000-0000-000000000003', 'iPraktikumDemo Peer Evaluation', 'What students assess about their teammates at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000004-0000-0000-0000-000000000004', 'iPraktikumDemo Tutor Evaluation', 'What students assess about their team tutor at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005');

INSERT INTO category (id, name, description, weight, short_name, assessment_schema_id)
    VALUES ('5b000001-0000-0000-0000-000000000001', 'Engineering', 'Technical contribution to the team project.', 2, 'ENG', '5a000001-0000-0000-0000-000000000001'),
           ('5b000002-0000-0000-0000-000000000002', 'Collaboration', 'Working with the team and the industry partner.', 1, 'COL', '5a000001-0000-0000-0000-000000000001'),
           ('5b000003-0000-0000-0000-000000000003', 'Self Reflection', 'How the student judges their own contribution.', 1, 'SELF', '5a000002-0000-0000-0000-000000000002'),
           ('5b000004-0000-0000-0000-000000000004', 'Team Contribution', 'How a teammate contributed to the team.', 1, 'PEER', '5a000003-0000-0000-0000-000000000003'),
           ('5b000005-0000-0000-0000-000000000005', 'Tutoring', 'How the tutor supported the team.', 1, 'TUT', '5a000004-0000-0000-0000-000000000004');

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
           ('5c000006-0000-0000-0000-000000000006', '5b000004-0000-0000-0000-000000000004', 'Reliability', 'Delivers what was promised, on time.', 1, 'REL',
            'Never delivers what was agreed.', 'Often late or incomplete.', 'Usually delivers what was agreed.', 'Delivers reliably and on time.', 'The teammate everyone counts on.'),
           ('5c000007-0000-0000-0000-000000000007', '5b000004-0000-0000-0000-000000000004', 'Team Support', 'Helps teammates and shares knowledge.', 1, 'SUP',
            'Works against the team.', 'Rarely helps others.', 'Helps when asked.', 'Offers help unprompted.', 'Makes the whole team better.'),
           ('5c000008-0000-0000-0000-000000000008', '5b000005-0000-0000-0000-000000000005', 'Availability', 'Reachable and responsive when the team needs help.', 1, 'AV',
            'Never reachable.', 'Hard to reach.', 'Reachable at the agreed times.', 'Responds quickly.', 'There whenever the team is stuck.'),
           ('5c000009-0000-0000-0000-000000000009', '5b000005-0000-0000-0000-000000000005', 'Feedback Quality', 'Gives actionable, constructive feedback.', 1, 'FQ',
            'Gives no feedback.', 'Feedback is vague or late.', 'Feedback is useful when given.', 'Feedback is concrete and timely.', 'Feedback that changed how we work.');

-- Results released, and the self, peer and tutor evaluation rounds enabled,
-- each bound to the seed's own schema. The three rounds share one window. The
-- windows below sit on the ios2526 semester like every other date in the seed
-- (see seed/README.md), so they close for good once that semester is past.
INSERT INTO course_phase_config (course_phase_id, assessment_schema_id, start, deadline,
                                 self_evaluation_enabled, self_evaluation_schema, self_evaluation_start, self_evaluation_deadline,
                                 peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_start, peer_evaluation_deadline,
                                 tutor_evaluation_enabled, tutor_evaluation_schema, tutor_evaluation_start, tutor_evaluation_deadline,
                                 evaluation_results_visible, grade_suggestion_visible, action_items_visible,
                                 results_released, grading_sheet_visible, assessment_enabled)
    VALUES ('f0000005-0000-0000-0000-000000000005', '5a000001-0000-0000-0000-000000000001',
            '2026-07-01 00:00:00+02', '2026-09-20 23:59:00+02',
            true, '5a000002-0000-0000-0000-000000000002', '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
            true, '5a000003-0000-0000-0000-000000000003', '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
            true, '5a000004-0000-0000-0000-000000000004', '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
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

-- Self evaluations: the author is the participant themselves.
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.participation)::text, 2, '0') || '-0000-0000-0000-000000000001')::uuid,
       rated.participation, 'f0000005-0000-0000-0000-000000000005', '5c000005-0000-0000-0000-000000000005', rated.level, rated.participation, 'self'
FROM (VALUES
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'very_good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'ok'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'ok'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'good'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level)
     ) AS rated(participation, level);

INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
    VALUES ('60000001-0000-0000-0000-000000000001', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', true, '2026-06-29 18:00:00+02', 'self'),
           ('60000002-0000-0000-0000-000000000002', 'cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', true, '2026-06-29 18:10:00+02', 'self'),
           ('60000003-0000-0000-0000-000000000003', 'cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', true, '2026-06-29 18:20:00+02', 'self'),
           ('60000004-0000-0000-0000-000000000004', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', true, '2026-06-29 18:30:00+02', 'self'),
           ('60000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', true, '2026-06-29 18:40:00+02', 'self'),
           ('60000006-0000-0000-0000-000000000006', 'cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', true, '2026-06-29 18:50:00+02', 'self');

INSERT INTO feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id, type)
    VALUES ('6a000001-0000-0000-0000-000000000001', 'positive', 'I was always available when the team needed a reviewer.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', 'self'),
           ('6a000002-0000-0000-0000-000000000002', 'negative', 'I sometimes take on too much instead of delegating.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', 'self');

-- Peer evaluations: every student rates both teammates on both peer
-- competencies, so each student has the two distinct raters a student-facing
-- peer average needs. course_participation_id is the teammate being rated,
-- author_course_participation_id the teammate who rated them.
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.participation, rated.author, rated.competency)::text, 2, '0') || '-0000-0000-0000-000000000002')::uuid,
       rated.participation, 'f0000005-0000-0000-0000-000000000005', rated.competency, rated.level, rated.author, 'peer'
FROM (VALUES
        -- Team Kepler
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'very_good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'very_good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'good'::score_level),
        ('cd000001-0000-0000-0000-000000000001'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'very_good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level),
        ('cd000003-0000-0000-0000-000000000003'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'very_good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'very_good'::score_level),
        ('cd000005-0000-0000-0000-000000000005'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'ok'::score_level),
        -- Team Hubble
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'very_good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'very_good'::score_level),
        ('cd000002-0000-0000-0000-000000000002'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'good'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level),
        ('cd000004-0000-0000-0000-000000000004'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'good'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'ok'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000006-0000-0000-0000-000000000006'::uuid, 'ok'::score_level),
        ('cd000006-0000-0000-0000-000000000006'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000007-0000-0000-0000-000000000007'::uuid, 'bad'::score_level)
     ) AS rated(participation, author, competency, level);

INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
    VALUES ('60000007-0000-0000-0000-000000000007', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', true, '2026-06-29 19:00:00+02', 'peer'),
           ('60000008-0000-0000-0000-000000000008', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', true, '2026-06-29 19:05:00+02', 'peer'),
           ('60000009-0000-0000-0000-000000000009', 'cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', true, '2026-06-29 19:10:00+02', 'peer'),
           ('60000010-0000-0000-0000-000000000010', 'cd000003-0000-0000-0000-000000000003', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', true, '2026-06-29 19:15:00+02', 'peer'),
           ('60000011-0000-0000-0000-000000000011', 'cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', true, '2026-06-29 19:20:00+02', 'peer'),
           ('60000012-0000-0000-0000-000000000012', 'cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', true, '2026-06-29 19:25:00+02', 'peer'),
           ('60000013-0000-0000-0000-000000000013', 'cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', true, '2026-06-29 19:30:00+02', 'peer'),
           ('60000014-0000-0000-0000-000000000014', 'cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', true, '2026-06-29 19:35:00+02', 'peer'),
           ('60000015-0000-0000-0000-000000000015', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', true, '2026-06-29 19:40:00+02', 'peer'),
           ('60000016-0000-0000-0000-000000000016', 'cd000004-0000-0000-0000-000000000004', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', true, '2026-06-29 19:45:00+02', 'peer'),
           ('60000017-0000-0000-0000-000000000017', 'cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', true, '2026-06-29 19:50:00+02', 'peer'),
           ('60000018-0000-0000-0000-000000000018', 'cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', true, '2026-06-29 19:55:00+02', 'peer');

INSERT INTO feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id, type)
    VALUES ('6a000003-0000-0000-0000-000000000003', 'positive', 'Stan reviewed my pull requests within hours and explained every comment.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', 'peer'),
           ('6a000004-0000-0000-0000-000000000004', 'negative', 'Stan could hand over more of the networking work instead of doing it alone.', 'cd000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', 'peer'),
           ('6a000005-0000-0000-0000-000000000005', 'positive', 'Carla''s prototypes made our partner meetings much easier.', 'cd000005-0000-0000-0000-000000000005', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', 'peer'),
           ('6a000006-0000-0000-0000-000000000006', 'positive', 'Selma kept the whole team aligned with her meeting notes.', 'cd000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', 'peer'),
           ('6a000007-0000-0000-0000-000000000007', 'negative', 'David was hard to reach during the last sprint.', 'cd000006-0000-0000-0000-000000000006', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'peer');

-- Tutor evaluations: every student rates their team's tutor. Here
-- course_participation_id is the tutor being rated (the tutor row's
-- course_participation_id in seed/team_allocation.sql, which is what the
-- results page at /tutors/:tutorId looks up) and
-- author_course_participation_id the student who rated them.
INSERT INTO evaluation (id, course_participation_id, course_phase_id, competency_id, score_level, author_course_participation_id, type)
SELECT ('5f0000' || lpad(row_number() OVER (ORDER BY rated.tutor, rated.author, rated.competency)::text, 2, '0') || '-0000-0000-0000-000000000003')::uuid,
       rated.tutor, 'f0000005-0000-0000-0000-000000000005', rated.competency, rated.level, rated.author, 'tutor'
FROM (VALUES
        -- Eduard, tutor of Team Kepler
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'very_good'::score_level),
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000001-0000-0000-0000-000000000001'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'good'::score_level),
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'good'::score_level),
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000003-0000-0000-0000-000000000003'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'ok'::score_level),
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'very_good'::score_level),
        ('cd000007-0000-0000-0000-000000000007'::uuid, 'cd000005-0000-0000-0000-000000000005'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'very_good'::score_level),
        -- Lector, tutor of Team Hubble
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'good'::score_level),
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000002-0000-0000-0000-000000000002'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'ok'::score_level),
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'ok'::score_level),
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000004-0000-0000-0000-000000000004'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'good'::score_level),
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000008-0000-0000-0000-000000000008'::uuid, 'good'::score_level),
        ('cd000008-0000-0000-0000-000000000008'::uuid, 'cd000006-0000-0000-0000-000000000006'::uuid, '5c000009-0000-0000-0000-000000000009'::uuid, 'ok'::score_level)
     ) AS rated(tutor, author, competency, level);

INSERT INTO evaluation_completion (id, course_participation_id, course_phase_id, author_course_participation_id, completed, completed_at, type)
    VALUES ('60000019-0000-0000-0000-000000000019', 'cd000007-0000-0000-0000-000000000007', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', true, '2026-06-30 10:00:00+02', 'tutor'),
           ('60000020-0000-0000-0000-000000000020', 'cd000007-0000-0000-0000-000000000007', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', true, '2026-06-30 10:05:00+02', 'tutor'),
           ('60000021-0000-0000-0000-000000000021', 'cd000007-0000-0000-0000-000000000007', 'f0000005-0000-0000-0000-000000000005', 'cd000005-0000-0000-0000-000000000005', true, '2026-06-30 10:10:00+02', 'tutor'),
           ('60000022-0000-0000-0000-000000000022', 'cd000008-0000-0000-0000-000000000008', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', true, '2026-06-30 10:15:00+02', 'tutor'),
           ('60000023-0000-0000-0000-000000000023', 'cd000008-0000-0000-0000-000000000008', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', true, '2026-06-30 10:20:00+02', 'tutor'),
           ('60000024-0000-0000-0000-000000000024', 'cd000008-0000-0000-0000-000000000008', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', true, '2026-06-30 10:25:00+02', 'tutor');

INSERT INTO feedback_items (id, feedback_type, feedback_text, course_participation_id, course_phase_id, author_course_participation_id, type)
    VALUES ('6a000008-0000-0000-0000-000000000008', 'positive', 'Eduard always had time for a quick call before a deadline.', 'cd000007-0000-0000-0000-000000000007', 'f0000005-0000-0000-0000-000000000005', 'cd000001-0000-0000-0000-000000000001', 'tutor'),
           ('6a000009-0000-0000-0000-000000000009', 'negative', 'Code review feedback sometimes came only after the sprint had ended.', 'cd000007-0000-0000-0000-000000000007', 'f0000005-0000-0000-0000-000000000005', 'cd000003-0000-0000-0000-000000000003', 'tutor'),
           ('6a000010-0000-0000-0000-000000000010', 'positive', 'The weekly check-ins kept our team on track.', 'cd000008-0000-0000-0000-000000000008', 'f0000005-0000-0000-0000-000000000005', 'cd000002-0000-0000-0000-000000000002', 'tutor'),
           ('6a000011-0000-0000-0000-000000000011', 'negative', 'I would have liked more concrete feedback on our architecture.', 'cd000008-0000-0000-0000-000000000008', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'tutor');

INSERT INTO action_item (id, course_phase_id, course_participation_id, action, author)
    VALUES ('6b000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'Pair with a Swift-experienced team member on the next feature.', 'Seeded Tutor'),
           ('6b000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', 'Take ownership of one feature end to end.', 'Seeded Tutor');
