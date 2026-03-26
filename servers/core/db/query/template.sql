-- name: GetTemplateCoursesAdmin :many
SELECT *
FROM course
WHERE template = TRUE
ORDER BY semester_tag, name DESC;

-- name: GetTemplateCoursesRestricted :many
-- struct: Course
WITH parsed_roles AS (
  SELECT
    split_part(role, '-', 1) AS semester_tag,
    split_part(role, '-', 2) AS course_name,
    split_part(role, '-', 3) AS user_role
  FROM
    unnest($1::text[]) AS role
),
user_course_roles AS (
  SELECT
    c.id,
    c.name,
    c.semester_tag,
    c.start_date,
    c.end_date,
    c.course_type,
    c.student_readable_data,
    c.restricted_data,
    c.ects,
    c.template,
    c.short_description,
    c.long_description,
    c.archived,
    c.archived_on,
    pr.user_role
  FROM
    course c
  INNER JOIN
    parsed_roles pr
    ON c.name = pr.course_name
    AND c.semester_tag = pr.semester_tag
  WHERE
    c.template = TRUE
    AND c.archived = FALSE
)
SELECT
  ucr.id,
  ucr.name,
  ucr.start_date,
  ucr.end_date,
  ucr.semester_tag,
  ucr.course_type,
  ucr.ects,
  CASE 
    WHEN COUNT(ucr.user_role) = 1 AND MAX(ucr.user_role) = 'Student' THEN '{}'::jsonb
    ELSE ucr.restricted_data::jsonb
  END AS restricted_data,
  ucr.student_readable_data,
  ucr.template,
  ucr.short_description,
  ucr.long_description,
  ucr.archived,
  ucr.archived_on
FROM
  user_course_roles ucr
GROUP BY
  ucr.id,
  ucr.name,
  ucr.semester_tag,
  ucr.start_date,
  ucr.end_date,
  ucr.course_type,
  ucr.student_readable_data,
  ucr.ects,
  ucr.restricted_data,
  ucr.template,
  ucr.short_description,
  ucr.long_description,
  ucr.archived,
  ucr.archived_on
ORDER BY
  ucr.semester_tag, ucr.name DESC;

-- name: CheckCourseNameExists :one
SELECT EXISTS(
    SELECT 1 FROM course WHERE name = $1 AND semester_tag = $2
) AS "exists";

-- name: CheckCourseTemplateStatus :one
SELECT template
FROM course
WHERE id = $1;

-- name: GetTemplateCourseByID :one
SELECT *
FROM course
WHERE id = $1
  AND template = TRUE;

-- name: MarkCourseAsTemplate :exec
UPDATE course
SET template = TRUE
WHERE id = $1;

-- name: UnmarkCourseAsTemplate :exec
UPDATE course
SET template = FALSE
WHERE id = $1;
