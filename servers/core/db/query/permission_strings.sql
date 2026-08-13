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
-- The university login is unique (partial unique index), so it identifies the student on its own.
-- Matriculation is optional for imported students: match it when the row carries one, otherwise fall
-- back to the login. A token without a matriculation claim ($1 = '') therefore matches only a student
-- whose stored matriculation is also empty, never one that carries a different number.
SELECT CONCAT(c.semester_tag, '-', c.name, '-Student')::text AS student_role
FROM course c
JOIN course_participation cp ON c.id = cp.course_id
JOIN student s ON cp.student_id = s.id
WHERE s.university_login = $2
AND (s.matriculation_number = $1 OR COALESCE(s.matriculation_number, '') = '');

-- name: GetCoursePhaseAuthRoleMapping :one
SELECT CONCAT(c.semester_tag, '-', c.name, '-Lecturer')::text AS lecturer_role, CONCAT(c.semester_tag, '-', c.name, '-Editor')::text AS editor_role, CONCAT(c.semester_tag, '-', c.name, '-cg-')::text AS custom_role_prefix
FROM course c
JOIN course_phase cp ON c.id = cp.course_id
WHERE cp.id = $1;