INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
VALUES
    (
        'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd',
        'Master Test',
        '2025-05-19',
        '2025-06-30',
        'ss25',
        'practical course',
        10,
        '{}',
        '{"icon": "school", "bg-color": "bg-teal-100"}',
        FALSE,
        'Hands-on master course',
        'Detailed description for the Master Test course used in copy tests.',
        FALSE,
        NULL
    );

INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data, student_readable_data, template, short_description, long_description, archived, archived_on)
VALUES
    (
        'c1f8060d-7381-4b64-a6ea-5ba8e8ac88ee',
        'Template Test',
        '2025-05-19',
        '2025-08-30',
        'ss25',
        'practical course',
        10,
        '{}',
        '{"icon": "school", "bg-color": "bg-teal-100"}',
        TRUE,
        'Template for future courses',
        'Long-form description for the template course copy flow.',
        FALSE,
        NULL
    );

INSERT INTO course_phase_type VALUES ('0542034b-87eb-4f91-ac90-b2e1536450de', 'Interview', false, 'core', 'Test Description');

INSERT INTO course_phase_type VALUES ('c313bd84-bc7b-4a5a-aca3-77a526e02f57', 'Matching', false, 'core', 'Test Description');

INSERT INTO course_phase_type VALUES ('8c97fd14-e1fe-4b12-bf7d-c542e620e8d8', 'Intro Course Developer', false, 'http://localhost:8082/intro-course/api', 'Test Description');

INSERT INTO course_phase_type VALUES ('e0fe2692-4e06-47db-b80f-92817e9a7566', 'DevOps Challenge', false, 'core', 'Test Description');

INSERT INTO course_phase_type VALUES ('613f8b0a-2200-4650-b0bb-6a26a0a140e8', 'Assessment', false, 'http://localhost:8085/assessment/api', 'Test Description');

INSERT INTO course_phase_type VALUES ('88ca3586-7748-4152-8c89-5fbb6e113587', 'Team Allocation', false, 'http://localhost:8083/team-allocation/api', 'Test Description');

INSERT INTO course_phase_type VALUES ('cc63d311-6992-4b08-9719-a83b92f45e8f', 'Self Team Allocation', false, 'http://localhost:8084/self-team-allocation/api', 'Test Description');

INSERT INTO course_phase_type VALUES ('3258275d-a76b-40f8-bb1a-95618299b8ac', 'Application', true, 'core', 'Test Description');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('57f45359-3fd6-47b7-822e-74e57b5f530c', '0542034b-87eb-4f91-ac90-b2e1536450de', 'score', 1, 'core', '{"type": "integer"}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('a67b6b12-bd5e-4be4-8353-fcc98fd7013d', '8c97fd14-e1fe-4b12-bf7d-c542e620e8d8', 'devices', 1, '/devices', '{"type": "array", "items": {"enum": ["IPhone", "IPad", "MacBook", "AppleWatch"], "type": "string"}}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('9a73af69-0ee1-46f1-bb3e-0d6775c7f02e', '613f8b0a-2200-4650-b0bb-6a26a0a140e8', 'scoreLevel', 1, '/student-assessment/scoreLevel', '{"enum": ["novice", "intermediate", "advanced", "expert"], "type": "string"}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('ebd9ac97-909e-4865-818a-faf2c5c22480', '88ca3586-7748-4152-8c89-5fbb6e113587', 'teamAllocation', 1, '/allocation', '{"type": "string"}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('df569553-b5d7-4ee5-8df2-5eb10fe083c5', 'cc63d311-6992-4b08-9719-a83b92f45e8f', 'teamAllocation', 1, '/allocation', '{"type": "string"}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('d7c3d526-c95d-4e23-8c6b-e0a561464ef3', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'score', 1, 'core', '{"type": "integer"}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('3c657de6-3fe1-41f3-a3c7-ceee83442815', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'applicationAnswers', 1, 'core', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}');

INSERT INTO course_phase_type_participation_provided_output_dto VALUES ('2ce41974-47e2-46fe-ad92-09159f08e030', '3258275d-a76b-40f8-bb1a-95618299b8ac', 'additionalScores', 1, 'core', '{"type": "array", "items": {"type": "object", "required": ["score", "key"], "properties": {"key": {"type": "string"}, "score": {"type": "number"}}}}');

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('67a66cf6-5505-41f2-a72e-2f247b386b83', '0542034b-87eb-4f91-ac90-b2e1536450de', 'score', '{"type": "integer"}', false);

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('7ee14e46-4dd0-4281-a3da-26e00859a62e', '0542034b-87eb-4f91-ac90-b2e1536450de', 'applicationAnswers', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}', false);

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('0dfdb01d-fffd-4ebc-b191-3201aa2d0fc3', 'c313bd84-bc7b-4a5a-aca3-77a526e02f57', 'score', '{"type": "integer"}', false);

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('faad5951-3af7-4189-adf3-298a545082af', '88ca3586-7748-4152-8c89-5fbb6e113587', 'applicationAnswers', '{"type": "array", "items": {"oneOf": [{"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["text"], "type": "string"}, "answer": {"type": "string"}, "order_num": {"type": "integer"}}}, {"type": "object", "required": ["answer", "key", "order_num", "type"], "properties": {"key": {"type": "string"}, "type": {"enum": ["multiselect"], "type": "string"}, "answer": {"type": "array", "items": {"type": "string"}}, "order_num": {"type": "integer"}}}]}}', false);

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('afae0c00-a7ae-40f9-b707-af19393a58b7', '88ca3586-7748-4152-8c89-5fbb6e113587', 'devices', '{"type": "array", "items": {"enum": ["IPhone", "IPad", "MacBook", "AppleWatch"], "type": "string"}}', false);

INSERT INTO course_phase_type_participation_required_input_dto VALUES ('2f17451e-d905-4484-aebc-a8c5191a4557', '88ca3586-7748-4152-8c89-5fbb6e113587', 'scoreLevel', '{"enum": ["novice", "intermediate", "advanced", "expert"], "type": "string"}', false);

INSERT INTO course_phase_type_phase_provided_output_dto VALUES ('c539e8fa-e8fb-4ed3-9cab-feacf3d7d8d1', '88ca3586-7748-4152-8c89-5fbb6e113587', 'teams', 1, '/team', '{"type": "array", "items": {"type": "object", "required": ["id", "name"], "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}}');

INSERT INTO course_phase_type_phase_provided_output_dto VALUES ('eded55c7-1077-4945-9ae6-dbdac205f9cf', 'cc63d311-6992-4b08-9719-a83b92f45e8f', 'teams', 1, '/team', '{"type": "array", "items": {"type": "object", "required": ["id", "name"], "properties": {"id": {"type": "string"}, "name": {"type": "string"}}}}');

INSERT INTO course_phase VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Application', '{"autoAccept": true, "applicationEndDate": "2025-06-01T23:59:00+02:00", "applicationStartDate": "2025-05-19T00:00:00+02:00", "externalStudentsAllowed": false, "universityLoginAvailable": true}', true, '3258275d-a76b-40f8-bb1a-95618299b8ac', '{}');

INSERT INTO course_phase VALUES ('93693f81-9c49-4183-ae70-c0ee3742560d', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Matching', '{}', false, 'c313bd84-bc7b-4a5a-aca3-77a526e02f57', '{}');

INSERT INTO course_phase VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', 'c1f8060d-7381-4b64-a6ea-5ba8e8ac88dd', 'Interview', '{"interviewQuestions": [{"id": 1747661413157, "orderNum": 0, "question": "Question 1"}, {"id": 1747661417361, "orderNum": 1, "question": "Question 2"}]}', false, '0542034b-87eb-4f91-ac90-b2e1536450de', '{}');

INSERT INTO course_phase_graph VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '93693f81-9c49-4183-ae70-c0ee3742560d');

INSERT INTO course_phase_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123');

INSERT INTO participation_data_dependency_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', 'd7c3d526-c95d-4e23-8c6b-e0a561464ef3', '67a66cf6-5505-41f2-a72e-2f247b386b83');

INSERT INTO participation_data_dependency_graph VALUES ('bd727106-2dc0-4c44-a804-2efde26101ae', '0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '3c657de6-3fe1-41f3-a3c7-ceee83442815', '7ee14e46-4dd0-4281-a3da-26e00859a62e');

INSERT INTO participation_data_dependency_graph VALUES ('0bf6eb6c-ff6f-40f4-af63-a005e2c8d123', '93693f81-9c49-4183-ae70-c0ee3742560d', '57f45359-3fd6-47b7-822e-74e57b5f530c', '0dfdb01d-fffd-4ebc-b191-3201aa2d0fc3');

INSERT INTO application_question_multi_select VALUES ('9df42ab6-1607-4c78-a1bd-28c58399ed55', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Where are you saved?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 2, false, '');

INSERT INTO application_question_multi_select VALUES ('a540c8d0-b5ee-4510-a671-2368a5997b66', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Are you copied with the course?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 1, false, '');

INSERT INTO application_question_multi_select VALUES ('e9abbd7f-f7ae-4096-911c-c62fe33105fa', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Where are you saved?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 2, false, '');

INSERT INTO application_question_multi_select VALUES ('c20829f9-d1d2-4952-95ee-1ebbff01a432', 'bd727106-2dc0-4c44-a804-2efde26101ae', 'Are you copied with the course?', '', 'CheckBoxQuestion', '', true, 0, 1, '{Yes}', 1, false, '');
