-- name: GetApplicationQuestionsTextForCoursePhase :many
SELECT * FROM application_question_text
WHERE course_phase_id = $1;

-- name: GetApplicationQuestionsMultiSelectForCoursePhase :many
SELECT * FROM application_question_multi_select
WHERE course_phase_id = $1;

-- name: CreateApplicationAnswerText :exec
INSERT INTO application_answer_text (id, application_question_id, course_participation_id, answer)
VALUES ($1, $2, $3, $4);

-- name: CreateApplicationAnswerMultiSelect :exec
INSERT INTO application_answer_multi_select (id, application_question_id, course_participation_id, answer)
VALUES ($1, $2, $3, $4);

-- name: CreateApplicationAnswerFileUpload :exec
INSERT INTO application_answer_file_upload (id, application_question_id, course_participation_id, file_id)
VALUES ($1, $2, $3, $4);

-- name: GetApplicationExists :one
SELECT EXISTS (
    SELECT 1
    FROM course_phase_participation cpp
    WHERE cpp.course_phase_id = $1
    AND cpp.course_participation_id = $2
);


-- name: CreateApplicationQuestionText :exec
INSERT INTO application_question_text (id, course_phase_id, title, description, placeholder, validation_regex, error_message, is_required, allowed_length, order_num, accessible_for_other_phases, access_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);

-- name: CreateApplicationQuestionMultiSelect :exec
INSERT INTO application_question_multi_select (id, course_phase_id, title, description, placeholder, error_message, is_required, min_select, max_select, options, order_num, accessible_for_other_phases, access_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13);

-- name: UpdateApplicationQuestionMultiSelect :exec
UPDATE application_question_multi_select
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    placeholder = COALESCE($4, placeholder),
    error_message = COALESCE($5, error_message),
    is_required = COALESCE($6, is_required),
    min_select = COALESCE($7, min_select),
    max_select = COALESCE($8, max_select),
    options = COALESCE($9, options),
    order_num = COALESCE($10, order_num), 
    accessible_for_other_phases = COALESCE($11, accessible_for_other_phases),
    access_key = COALESCE($12, access_key)
WHERE id = $1;

-- name: UpdateApplicationQuestionText :exec
UPDATE application_question_text
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    placeholder = COALESCE($4, placeholder),
    validation_regex = COALESCE($5, validation_regex),
    error_message = COALESCE($6, error_message),
    is_required = COALESCE($7, is_required),
    allowed_length = COALESCE($8, allowed_length),
    order_num = COALESCE($9, order_num),
    accessible_for_other_phases = COALESCE($10, accessible_for_other_phases),
    access_key = COALESCE($11, access_key)
WHERE id = $1;



-- name: CheckIfCoursePhaseIsApplicationPhase :one
SELECT 
    cpt.name = 'Application' AS is_application
FROM 
    course_phase cp
JOIN 
    course_phase_type cpt
ON 
    cp.course_phase_type_id = cpt.id
WHERE 
    cp.id = $1;

-- name: DeleteApplicationQuestionText :exec
DELETE FROM application_question_text
WHERE id = $1;

-- name: DeleteApplicationQuestionMultiSelect :exec
DELETE FROM application_question_multi_select
WHERE id = $1;

-- name: GetAllOpenApplicationPhases :many
SELECT 
    cp.id AS course_phase_id,
    c.name AS course_name,
    c.start_date, 
    c.end_date,
    c.course_type, 
    c.ects,
    c.short_description,
    c.long_description,
    (cp.restricted_data->>'applicationEndDate')::text AS application_end_date,
    (cp.restricted_data->>'externalStudentsAllowed')::boolean AS external_students_allowed,
    (cp.restricted_data->>'universityLoginAvailable')::boolean AS university_login_available
FROM 
    course_phase cp
JOIN 
    course_phase_type cpt
    ON cp.course_phase_type_id = cpt.id
JOIN 
    course c
    ON cp.course_id = c.id
WHERE 
    cp.is_initial_phase = true
    AND cpt.name = 'Application'
    AND (cp.restricted_data->>'applicationEndDate')::timestamp > NOW()
    AND (cp.restricted_data->>'applicationStartDate')::timestamp < NOW();

-- name: GetOpenApplicationPhase :one
SELECT 
    cp.id AS course_phase_id,
    c.name AS course_name,
    c.start_date, 
    c.end_date,
    c.course_type, 
    c.ects,
    c.short_description,
    c.long_description,
    (cp.restricted_data->>'applicationEndDate')::text AS application_end_date,
    (cp.restricted_data->>'externalStudentsAllowed')::boolean AS external_students_allowed,
    (cp.restricted_data->>'universityLoginAvailable')::boolean AS university_login_available
FROM 
    course_phase cp
JOIN 
    course_phase_type cpt
    ON cp.course_phase_type_id = cpt.id
JOIN 
    course c
    ON cp.course_id = c.id
WHERE 
    cp.id = $1
    AND cp.is_initial_phase = true
    AND cpt.name = 'Application'
    AND (cp.restricted_data->>'applicationEndDate')::timestamp > NOW()
    AND (cp.restricted_data->>'applicationStartDate')::timestamp < NOW();

-- name: GetApplicationExistsForStudent :one
SELECT EXISTS (
    SELECT 1
    FROM course_participation cp
    INNER JOIN course_phase ph ON cp.course_id = ph.course_id
    WHERE cp.student_id = $1 AND ph.id = $2
);

-- name: CheckIfCoursePhaseIsOpenApplicationPhase :one
SELECT 
    cpt.name = 'Application' AS is_application,
    (cp.restricted_data->>'universityLoginAvailable')::boolean AS university_login_available 
FROM 
    course_phase cp
JOIN 
    course_phase_type cpt
ON 
    cp.course_phase_type_id = cpt.id
WHERE 
    cp.id = $1
    AND (cp.restricted_data->>'applicationEndDate')::timestamp > NOW();

-- name: GetApplicationAnswersTextForCourseParticipationID :many
SELECT aat.*
FROM application_answer_text aat
JOIN application_question_text aqt ON aat.application_question_id = aqt.id
WHERE aqt.course_phase_id = $1 AND aat.course_participation_id = $2;

-- name: GetApplicationAnswersMultiSelectForCourseParticipationID :many
SELECT aams.*
FROM application_answer_multi_select aams
JOIN application_question_multi_select aqms ON aams.application_question_id = aqms.id
WHERE aqms.course_phase_id = $1 AND aams.course_participation_id = $2;

-- name: CreateOrOverwriteApplicationAnswerText :exec 
INSERT INTO application_answer_text (id, application_question_id, course_participation_id, answer)
VALUES ($1, $2, $3, $4)
ON CONFLICT (course_participation_id, application_question_id)
DO UPDATE
SET answer = EXCLUDED.answer;

-- name: CreateOrOverwriteApplicationAnswerMultiSelect :exec 
INSERT INTO application_answer_multi_select (id, application_question_id, course_participation_id, answer)
VALUES ($1, $2, $3, $4)
ON CONFLICT (course_participation_id, application_question_id)
DO UPDATE
SET answer = EXCLUDED.answer;

-- name: GetAllApplicationParticipations :many
SELECT
    cpp.course_phase_id,
    cpp.course_participation_id,
    cpp.pass_status,
    cpp.restricted_data,
    s.id AS student_id,
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.has_university_account,
    s.gender, 
    s.nationality,
    s.study_degree,
    s.study_program,
    s.current_semester,
    a.score
FROM
    course_phase_participation cpp
JOIN
    course_participation cp ON cpp.course_participation_id = cp.id
JOIN
    student s ON cp.student_id = s.id
LEFT JOIN
    application_assessment a ON cpp.course_participation_id = a.course_participation_id AND cpp.course_phase_id = a.course_phase_id
WHERE
    cpp.course_phase_id = $1;

-- name: UpdateApplicationAssessment :exec
INSERT INTO application_assessment (id, course_phase_id, course_participation_id, score)
VALUES (
    gen_random_uuid(),    
    $1,                   
    $2, 
    $3             
)
ON CONFLICT (course_phase_id, course_participation_id) 
DO UPDATE 
SET score = EXCLUDED.score; 

-- name: BatchUpdateAdditionalScores :exec
WITH updates AS (
  SELECT 
    UNNEST(sqlc.arg(course_participation_ids)::uuid[]) AS course_participation_id,
    UNNEST(sqlc.arg(scores)::numeric[]) AS score,
    sqlc.arg(score_name)::text[] AS path -- Use $3 as a JSON path array
)
UPDATE course_phase_participation
SET    
    restricted_data = jsonb_set(
        COALESCE(restricted_data, '{}'),
        updates.path, -- Use dynamic path
        to_jsonb(ROUND(updates.score, 2)) -- Convert the float score to JSONB
    )
FROM updates
WHERE 
    course_phase_participation.course_participation_id = updates.course_participation_id
    AND course_phase_participation.course_phase_id = sqlc.arg(course_phase_id)::uuid;


-- name: GetExistingAdditionalScores :one
SELECT 
    restricted_data->>'additional_scores' AS additional_scores
FROM
    course_phase
WHERE
    id = $1;

-- name: UpdateExistingAdditionalScores :exec
UPDATE course_phase
SET restricted_data = restricted_data || $2
WHERE id = $1;

-- name: DeleteApplications :exec
DELETE FROM course_participation
WHERE id IN (
      SELECT cpp.course_participation_id
      FROM course_phase_participation cpp
      WHERE cpp.course_participation_id = ANY(sqlc.arg(course_participation_ids)::uuid[])
        AND cpp.course_phase_id = sqlc.arg(course_phase_id)::uuid -- ensures that only applications for the given course phase are deleted
  );

-- name: StoreApplicationAnswerUpdateTimestamp :exec
UPDATE course_phase_participation
SET restricted_data = jsonb_set(
    COALESCE(restricted_data, '{}'), -- Ensure meta_data is not NULL
    '{student_last_modified}', -- Path to the key
    to_jsonb(NOW())::jsonb     -- Value to set
)
WHERE 
 course_phase_id = $1
 AND course_participation_id = $2;

-- name: StoreApplicationAssessmentUpdateTimestamp :exec
UPDATE course_phase_participation
SET restricted_data = jsonb_set(
    COALESCE(restricted_data, '{}'), -- Ensure meta_data is not NULL
    '{assessment_last_modified}', -- Path to the key
    to_jsonb(NOW())::jsonb     -- Value to set
)
WHERE course_phase_id = $1
 AND course_participation_id = $2;

-- name: AcceptApplicationIfAutoAccept :exec
UPDATE course_phase_participation
SET pass_status = 'passed'
WHERE course_phase_id = $1
  AND course_participation_id = $2
  AND (
    SELECT (restricted_data->>'autoAccept')::boolean
    FROM course_phase
    WHERE id = $1
  ) = true;

-- name: GetApplicationPhaseIDForCourse :one
SELECT cp.id
FROM course_phase cp
JOIN course_phase_type cpt ON cp.course_phase_type_id = cpt.id
WHERE cp.course_id = $1
  AND cpt.name = 'Application'
LIMIT 1;

-- name: GetApplicationQuestionsFileUploadForCoursePhase :many
SELECT * FROM application_question_file_upload
WHERE course_phase_id = $1;

-- name: CreateApplicationQuestionFileUpload :exec
INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: UpdateApplicationQuestionFileUpload :exec
UPDATE application_question_file_upload
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    is_required = COALESCE($4, is_required),
    allowed_file_types = COALESCE($5, allowed_file_types),
    max_file_size_mb = COALESCE($6, max_file_size_mb),
    order_num = COALESCE($7, order_num),
    accessible_for_other_phases = COALESCE($8, accessible_for_other_phases),
    access_key = COALESCE($9, access_key)
WHERE id = $1;

-- name: DeleteApplicationQuestionFileUpload :exec
DELETE FROM application_question_file_upload
WHERE id = $1;

-- name: GetApplicationAnswersFileUploadForCourseParticipationID :many
SELECT aafu.*
FROM application_answer_file_upload aafu
JOIN application_question_file_upload aqfu ON aafu.application_question_id = aqfu.id
WHERE aqfu.course_phase_id = $1 AND aafu.course_participation_id = $2;

-- name: GetApplicationAnswerFileUploadByQuestionAndParticipation :one
SELECT * FROM application_answer_file_upload
WHERE application_question_id = $1 AND course_participation_id = $2;

-- name: CreateOrOverwriteApplicationAnswerFileUpload :exec 
INSERT INTO application_answer_file_upload (id, application_question_id, course_participation_id, file_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (course_participation_id, application_question_id)
DO UPDATE
SET file_id = EXCLUDED.file_id;

