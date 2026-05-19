-- name: GetConfirmationMailingInformation :one
SELECT 
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.study_degree,
    s.current_semester,
    s.study_program,
    c.name AS course_name,
    c.start_date AS course_start_date,
    c.end_date AS course_end_date,
    COALESCE((p.restricted_data->>'applicationEndDate'), '')::text AS application_end_date,
    COALESCE((p.restricted_data->'mailingSettings'->>'confirmationMailSubject'), '')::text AS confirmation_mail_subject,
    COALESCE((p.restricted_data->'mailingSettings'->>'confirmationMailContent'), '')::text AS confirmation_mail_content,
    COALESCE((p.restricted_data->'mailingSettings'->>'sendConfirmationMail')::boolean, false)::boolean AS send_confirmation_mail
FROM 
    course_phase_participation cpp
JOIN 
    course_participation cp ON cpp.course_participation_id = cp.id
JOIN 
    student s ON cp.student_id = s.id
JOIN 
    course_phase p ON cpp.course_phase_id = p.id
JOIN 
    course c ON p.course_id = c.id
WHERE 
    cpp.course_participation_id = $1
    AND cpp.course_phase_id = $2;

-- name: GetFailedMailingInformation :one
SELECT
    c.name AS course_name,
    c.start_date AS course_start_date,
    c.end_date AS course_end_date,
    COALESCE((p.restricted_data->'mailingSettings'->>'failedMailSubject'), '')::text AS mail_subject,
    COALESCE((p.restricted_data->'mailingSettings'->>'failedMailContent'), '')::text AS mail_content
FROM
    course_phase p
JOIN
    course c ON p.course_id = c.id
WHERE
    p.id = $1;

-- name: GetPassedMailingInformation :one
SELECT
    c.name AS course_name,
    c.start_date AS course_start_date,
    c.end_date AS course_end_date,
    COALESCE((p.restricted_data->'mailingSettings'->>'passedMailSubject'), '')::text AS mail_subject,
    COALESCE((p.restricted_data->'mailingSettings'->>'passedMailContent'), '')::text AS mail_content
FROM
    course_phase p
JOIN
    course c ON p.course_id = c.id
WHERE
    p.id = $1;

-- name: GetParticipantMailingInformation :many
SELECT
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.study_degree,
    s.current_semester,
    s.study_program
FROM
    course_phase p
JOIN
    course_phase_participation cpp ON p.id = cpp.course_phase_id
JOIN
    course_participation cp ON cpp.course_participation_id = cp.id
JOIN
    student s ON cp.student_id = s.id
WHERE
    p.id = $1
AND 
    cpp.pass_status = $2;

-- name: GetCourseMailingSettingsForCoursePhaseID :one
SELECT
    COALESCE((c.restricted_data->'mailingSettings'->>'replyToEmail')::text, '')::text AS reply_to_email,
    COALESCE((c.restricted_data->'mailingSettings'->>'replyToName')::text, '')::text AS reply_to_name,
    COALESCE((c.restricted_data->'mailingSettings'->>'ccAddresses')::jsonb, '[]')::jsonb AS cc_addresses,
    COALESCE((c.restricted_data->'mailingSettings'->>'bccAddresses')::jsonb, '[]')::json AS bcc_addresses
FROM 
  course c
INNER JOIN
  course_phase p ON c.id = p.course_id
WHERE
  p.id = $1;

-- name: GetParticipantMailingInformationByIDs :many
SELECT
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.study_degree,
    s.current_semester,
    s.study_program
FROM
    course_phase p
JOIN
    course_phase_participation cpp ON p.id = cpp.course_phase_id
JOIN
    course_participation cp ON cpp.course_participation_id = cp.id
JOIN
    student s ON cp.student_id = s.id
WHERE
    p.id = $1
AND
    cpp.course_participation_id = ANY($2::uuid[]);

-- name: UpdateAssessmentReminderLastSentAt :exec
UPDATE course_phase
SET restricted_data = jsonb_set(
  COALESCE(restricted_data, '{}'::jsonb),
  ARRAY['mailingSettings', 'assessmentReminder', 'lastSentAtByType', $2::text],
  to_jsonb($3::text),
  true
)
WHERE id = $1;
