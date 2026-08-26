-- PROMPT demo seed: assessment database.
--
-- The Assessment phase of iPraktikumDemo (f0000005-0000-0000-0000-000000000005):
-- a rubric, graded participants, and a completed self evaluation round.
--
-- The seed brings its own assessment_schema rows rather than reusing the four
-- defaults the migrations create ('Assessment Template', 'Self/Peer/Tutor
-- Evaluation Template'): those are created with gen_random_uuid(), carry no
-- categories, and are the fallback every other phase relies on. The peer and
-- tutor schema columns are therefore left out of the config insert so their
-- migration-set DEFAULTs still apply.
--
-- course_phase_config references the schemas with ON DELETE RESTRICT, so it is
-- deleted before them.

DELETE FROM action_item WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM feedback_items WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM evaluation_completion WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM evaluation WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment_completion WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM category_assessment WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM course_phase_config WHERE course_phase_id = 'f0000005-0000-0000-0000-000000000005';
DELETE FROM assessment_schema WHERE id IN ('5a000001-0000-0000-0000-000000000001', '5a000002-0000-0000-0000-000000000002');

INSERT INTO assessment_schema (id, name, description, source_phase_id)
    VALUES ('5a000001-0000-0000-0000-000000000001', 'iPraktikumDemo Rubric', 'The rubric the demo course grades against.', 'f0000005-0000-0000-0000-000000000005'),
           ('5a000002-0000-0000-0000-000000000002', 'iPraktikumDemo Self Evaluation', 'What students assess about themselves at the end of the demo course.', 'f0000005-0000-0000-0000-000000000005');

INSERT INTO category (id, name, description, weight, short_name, assessment_schema_id)
    VALUES ('5b000001-0000-0000-0000-000000000001', 'Engineering', 'Technical contribution to the team project.', 2, 'ENG', '5a000001-0000-0000-0000-000000000001'),
           ('5b000002-0000-0000-0000-000000000002', 'Collaboration', 'Working with the team and the industry partner.', 1, 'COL', '5a000001-0000-0000-0000-000000000001'),
           ('5b000003-0000-0000-0000-000000000003', 'Self Reflection', 'How the student judges their own contribution.', 1, 'SELF', '5a000002-0000-0000-0000-000000000002');

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
            'I contributed nothing.', 'I contributed less than I wanted.', 'I contributed my share.', 'I contributed more than my share.', 'I drove the project.');

-- Assessment window closed, results released, self evaluation enabled and bound
-- to the seed's own self evaluation schema. peer_evaluation_schema and
-- tutor_evaluation_schema are omitted so their column DEFAULTs apply.
INSERT INTO course_phase_config (course_phase_id, assessment_schema_id, self_evaluation_schema, start, deadline,
                                 self_evaluation_enabled, self_evaluation_start, self_evaluation_deadline,
                                 evaluation_results_visible, grade_suggestion_visible, action_items_visible,
                                 results_released, grading_sheet_visible, assessment_enabled)
    VALUES ('f0000005-0000-0000-0000-000000000005', '5a000001-0000-0000-0000-000000000001', '5a000002-0000-0000-0000-000000000002',
            '2026-07-01 00:00:00+02', '2026-09-20 23:59:00+02',
            true, '2026-06-15 00:00:00+02', '2026-06-30 23:59:00+02',
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

INSERT INTO action_item (id, course_phase_id, course_participation_id, action, author)
    VALUES ('6b000001-0000-0000-0000-000000000001', 'f0000005-0000-0000-0000-000000000005', 'cd000004-0000-0000-0000-000000000004', 'Pair with a Swift-experienced team member on the next feature.', 'Seeded Tutor'),
           ('6b000002-0000-0000-0000-000000000002', 'f0000005-0000-0000-0000-000000000005', 'cd000006-0000-0000-0000-000000000006', 'Take ownership of one feature end to end.', 'Seeded Tutor');
