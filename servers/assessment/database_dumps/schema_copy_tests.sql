-- SCENARIO 1: Schema with no assessments in any phase (for TestUpdateCategory_NoAssessments)
-- This schema is only used in Phase 1 and has no assessments anywhere
INSERT INTO public.assessment_schema (id, name, description)
VALUES ('00000000-0000-0000-0000-000000000001', 'Test Schema 1 - No Assessments',
        'Schema with no assessments for testing no-copy scenario');

-- SCENARIO 2: Schema used in multiple phases, with assessments in "other" phase
-- (for TestUpdateCategory_WithAssessmentsInOtherPhase)
INSERT INTO public.assessment_schema (id, name, description, source_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', 'Test Schema 2 - With Other Phase Assessments',
        'Schema used in multiple phases for testing copy scenario',
        '10000000-0000-0000-0000-000000000002');

-- SCENARIO 3: Schema with assessments only in the same phase (for TestUpdateCategory_WithAssessmentsInSamePhase)
INSERT INTO public.assessment_schema (id, name, description, source_phase_id)
VALUES ('00000000-0000-0000-0000-000000000003', 'Test Schema 3 - Same Phase Assessments',
        'Schema with assessments only in same phase for testing no-copy scenario',
        '10000000-0000-0000-0000-000000000003');

INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000001', '10000000-0000-0000-0000-000000000001');

-- Phase 2 is the owner
-- Phase 2a: Current phase (where we'll make the update)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000002');

-- Phase 2b: Other phase (has assessments, should trigger copy)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000002', '10000000-0000-0000-0000-000000000004');

INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id)
VALUES ('00000000-0000-0000-0000-000000000003', '10000000-0000-0000-0000-000000000003');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES
    ('20000000-0000-0000-0000-000000000001', 'Test Category 1', 'First test category', 1, 'TC1',
     '00000000-0000-0000-0000-000000000001');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES
    ('20000000-0000-0000-0000-000000000002', 'Test Category 2', 'Second test category', 1, 'TC2',
     '00000000-0000-0000-0000-000000000002'),
    ('20000000-0000-0000-0000-000000000003', 'Test Category 2-B', 'Another category in schema 2', 2, 'TC2-B',
     '00000000-0000-0000-0000-000000000002');

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES
    ('20000000-0000-0000-0000-000000000004', 'Test Category 3', 'Third test category', 2, 'TC3',
     '00000000-0000-0000-0000-000000000003');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name,
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES
    ('30000000-0000-0000-0000-000000000001', '20000000-0000-0000-0000-000000000001', 'Test Competency 1-1',
     'First competency in category 1', 1, 'TC1-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name,
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES
    ('30000000-0000-0000-0000-000000000002', '20000000-0000-0000-0000-000000000002', 'Test Competency 2-1',
     'First competency in category 2', 1, 'TC2-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good'),
    ('30000000-0000-0000-0000-000000000003', '20000000-0000-0000-0000-000000000003', 'Test Competency 2-B-1',
     'Competency in category 2-B', 1, 'TC2-B-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

INSERT INTO public.competency (id, category_id, name, description, weight, short_name,
    description_very_bad, description_bad, description_ok, description_good, description_very_good)
VALUES
    ('30000000-0000-0000-0000-000000000004', '20000000-0000-0000-0000-000000000004', 'Test Competency 3-1',
     'Competency in category 3', 1, 'TC3-1',
     'Very bad', 'Bad', 'OK', 'Good', 'Very good');

-- Assessments in the "other" phase (Phase 2b/Phase 4) - this will trigger the copy
INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, author)
VALUES
    ('40000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000001',
     '10000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000002',
     'good', 'Test Author');

-- Evaluation in the "other" phase (Phase 2b/Phase 4)
INSERT INTO public.evaluation (id, course_participation_id, course_phase_id, competency_id, score_level,
    author_course_participation_id, type)
VALUES
    ('45000000-0000-0000-0000-000000000001', '50000000-0000-0000-0000-000000000002',
     '10000000-0000-0000-0000-000000000004', '30000000-0000-0000-0000-000000000003',
     'very_good', '50000000-0000-0000-0000-000000000001', 'peer');

-- Assessments in the "same" phase (Phase 3) - should NOT trigger copy when updating from Phase 3
INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, author)
VALUES
    ('40000000-0000-0000-0000-000000000002', '50000000-0000-0000-0000-000000000003',
     '10000000-0000-0000-0000-000000000003', '30000000-0000-0000-0000-000000000004',
     'ok', 'Test Author Same Phase');
