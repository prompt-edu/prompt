-- name: GetPermissionStringByCourseID :one
SELECT CONCAT(semester_tag, '-', name) AS course_identifier
FROM course
WHERE id = $1;

-- name: GetPermissionStringByCourseParticipationID :one
SELECT CONCAT(c.semester_tag, '-', c.name) AS course_identifier
FROM course c
JOIN course_participation cp ON c.id = cp.course_id
WHERE cp.id = $1;

-- name: GetPermissionStringByCoursePhaseID :one
SELECT CONCAT(c.semester_tag, '-', c.name) AS course_identifier
FROM course c
JOIN course_phase cp ON c.id = cp.course_id
WHERE cp.id = $1;

-- name: GetStudentRoleStrings :many
SELECT CONCAT(c.semester_tag, '-', c.name, '-Student')::text AS student_role
FROM course c
JOIN course_participation cp ON c.id = cp.course_id
JOIN student s ON cp.student_id = s.id
WHERE s.university_login = $2
AND (s.matriculation_number = $1 OR s.matriculation_number IS NULL OR $1 = '');

-- name: GetCoursePhaseAuthRoleMapping :one
SELECT CONCAT(c.semester_tag, '-', c.name, '-Lecturer')::text AS lecturer_role, CONCAT(c.semester_tag, '-', c.name, '-Editor')::text AS editor_role, CONCAT(c.semester_tag, '-', c.name, '-cg-')::text AS custom_role_prefix
FROM course c
JOIN course_phase cp ON c.id = cp.course_id
WHERE cp.id = $1;