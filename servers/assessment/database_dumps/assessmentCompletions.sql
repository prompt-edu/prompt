-- Insert the default assessment schema
INSERT INTO public.assessment_schema (id, name, description)
VALUES ('550e8400-e29b-41d4-a716-446655440000', 'Intro Course Assessment Schema', 'This is the default assessment schema.');

-- Insert some sample course_phase_config records
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '24461b6b-3c3a-4bc6-ba42-69eeb1514da9', '2025-01-01 00:00:00+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', true, true, true);

-- Course phase config for not visible scenario (both grade suggestions and action items hidden)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '3517a3e3-fe60-40e0-8a5e-8f39049c12c3', '2025-01-01 00:00:00+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', false, false, false);

-- Course phase config for before deadline scenario (deadline in the future)
INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, grade_suggestion_visible, action_items_visible, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '4179d58a-d00d-4fa7-94a5-397bc69fab02', '2099-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', true, true, false);

INSERT INTO public.course_phase_config (assessment_schema_id, course_phase_id, deadline, self_evaluation_enabled, self_evaluation_schema, self_evaluation_deadline, peer_evaluation_enabled, peer_evaluation_schema, peer_evaluation_deadline, start, self_evaluation_start, peer_evaluation_start, results_released)
VALUES ('550e8400-e29b-41d4-a716-446655440000', '319f28d4-8877-400e-9450-d49077aae7fe', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', true, '550e8400-e29b-41d4-a716-446655440000', '2025-12-31 23:59:59+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', '2024-01-01 00:00:00+00', false);

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES (
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'Version Control',
        '',
        1,
        'Git',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES (
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'User Interface',
        '',
        1,
        'UI',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.category (id, name, description, weight, short_name, assessment_schema_id)
VALUES (
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Fundamentals in Software Engineering',
        '',
        1,
        'SE',
        '550e8400-e29b-41d4-a716-446655440000'
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'GitLab Project Management',
        'Understand GitLab’s collaboration features, including issue tracking, merge request workflows, and navigation of the issue board.',
        'Can create and manage repositories and basic issues.',
        'Can manage issues, labels, and milestones; creates and reviews MRs. ',
        'Can manage issue workflows, works together with Tutor on Reviews, and enforces best practices.',
        'Defines and optimizes issue tracking and MR processes for efficient collaboration. ',
        'Describes the GitLab project management features and how they can be used to manage the intro course app project.',
        1,
        'GitLab PM'
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '0431b736-7fab-4333-b83e-fe3927f32475',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Requirements & Backlog Management',
        'Understand and apply structured documentation techniques such as product backlogs and requirement artifacts. Use common principles as Abbots Technique or FURPS+',
        'Writes simple user stories but lacks clarity and adherence to best practices.',
        'Creates well-structured user stories using INVEST criteria and documents them systematically.',
        'Manages a product backlog effectively, refining requirements iteratively.',
        'Ensures requirement traceability, prioritization, and alignment with long-term goals.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Architecture and System Design',
        'Understand and apply architectural principles, including top-level architecture, subsystem decomposition, and deployment diagrams.',
        'Recognizes key architectural elements and struggles with formal documentation.',
        'Creates high-level system architecture with some guidance.',
        'Develops and evaluates scalable architectures tailored to project requirements.',
        'Designs architecture with clear subsystem decomposition, using SDD and API specifications.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        '9107c0aa-15b7-4967-bf62-6fa131f08bee',
        'Software Engineering & Modeling',
        'Recognize why modeling is important in software engineering and how it contributes to structured development.',
        'Has a basic understanding of software engineering and modeling but struggles to phrase their significance.',
        'Understands the role of modeling in software engineering and can explain why it is useful.',
        'Applies model-based approaches to structure software development and can justify their importance.',
        'Critically evaluates modeling techniques and adapts them to project-specific needs.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'Low-Fidelity Mockups / Prototyping',
        'Learn UI wireframing and prototyping using pen & paper (not tool-bound. Create low-fidelity mockups for the intro course app and the first SwiftUI non-functional requirement (NFR). ',
        'Sketches basic UI ideas but lacks structure and clarity.',
        'Creates structured wireframes with a focus on layout and usability.',
        'Designs clear and functional low-fidelity prototypes, considering user flow and NFRs.',
        'Rapidly iterates on mockups, ensuring usability and alignment with design principles.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        '31aea83e-407b-4428-a5da-b25dd562832b',
        '815b159b-cab3-49b4-8060-c4722d59241d',
        'Apple''s Human Interface Guidelines',
        'Understand the Human Interface Guidelines (HIG) and how SwiftUI supports cross-platform compliance, especially for non-functional requirements of the intro course app.',
        'Recognizes that HIG exists and affects app design.',
        'Understands core principles of HIG and applies basic guidelines in SwiftUI.',
        'Integrates HIG principles effectively, ensuring usability and consistency.',
        'Applies HIG to enhance usability and consistency across platforms, aligning with Apple’s best practices.',
        'Very Good',
        1,
        NULL
    );

INSERT INTO public.competency (id, category_id, name, description, description_very_bad, description_bad, description_ok, description_good, description_very_good, weight, short_name)
VALUES (
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        '25f1c984-ba31-4cf2-aa8e-5662721bf44e',
        'Git Basics',
        'Learn and apply Git workflows, key commands, and best practices, including branching models, commit conventions, and the differences between CLI and GUI tools.',
        'Can initialize a repository and commit changes.',
        'Can create branches, merge changes, and resolve basic conflicts.',
        'Follows a structured Git workflow with a clear branching model, adopts meaningful commit messages.',
        'Optimizes Git usage, enforces best practices, mentors others in efficient Git workflows.',
        'Very Good',
        2,
        NULL
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        '1950fdb7-d736-4fe6-81f9-b8b1cf7c85df',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 20:16:50.331851+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        '4422e68c-d042-43f0-a725-a2b2b7e387d8',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '20725c05-bfd7-45a7-a981-d092e14f98d3',
        'ok',
        '2025-05-11 21:41:57.148125+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        'ba5dbcbf-3496-4de0-bff7-d0b668201123',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '0431b736-7fab-4333-b83e-fe3927f32475',
        'bad',
        '2025-05-13 16:27:33.380423+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        'e8139c0d-130a-440c-8b8f-98b2c3be3d90',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '36af9432-0b0e-49e0-93d0-5044b7bed1c8',
        'bad',
        '2025-05-13 16:27:33.811249+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        '64e97c29-7221-47e4-a802-59638b7b872a',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2fc14584-d82c-47c2-9f75-22276d9809ef',
        'bad',
        '2025-05-13 16:27:34.383677+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        '7db6d645-69f7-4bbf-85d8-03f486149b37',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '54dbdc81-8566-4353-ace4-e2a8252a8c59',
        'bad',
        '2025-05-13 16:27:35.807434+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        '9de5b596-431f-4145-9ba9-31b4565cc73d',
        'ca42e447-60f9-4fe0-b297-2dae3f924fd7',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '31aea83e-407b-4428-a5da-b25dd562832b',
        'bad',
        '2025-05-13 16:27:36.73803+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment (id, course_participation_id, course_phase_id, competency_id, score_level, assessed_at, author, author_id)
VALUES (
        'bd3c570b-db28-456d-ae30-882e01b243cc',
        '319f28d4-8877-400e-9450-d49077aae7fe',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        'eb36bf49-87c2-429b-a87e-a930630a3fe3',
        'good',
        '2025-05-11 21:46:32.204333+02',
        'Maximilian Rapp',
        ''
    );

INSERT INTO public.assessment_completion (course_participation_id, course_phase_id, completed_at, author, comment, grade_suggestion, completed)
VALUES (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '24461b6b-3c3a-4bc6-ba42-69eeb1514da9',
        '2025-05-13 16:36:46.782123+02',
        'Maximilian Rapp',
        'Test comment for visible scenario',
        4.5,
        true
    );

-- Assessment completion for not visible scenario
INSERT INTO public.assessment_completion (course_participation_id, course_phase_id, completed_at, author, comment, grade_suggestion, completed)
VALUES (
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
        '3517a3e3-fe60-40e0-8a5e-8f39049c12c3',
        '2025-05-13 16:36:46.782123+02',
        'Maximilian Rapp',
        'Test comment for not visible scenario',
        3.5,
        true
    );
