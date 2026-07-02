-- name: GetCourse :one
SELECT *
FROM course
WHERE id = $1
LIMIT 1;


-- name: GetAllActiveCoursesAdmin :many
SELECT c.*
FROM course c
ORDER BY c.template,
         c.semester_tag,
         c.name DESC;

-- name: GetAllActiveCoursesRestricted :many
-- struct: Course
 WITH parsed_roles AS (SELECT split_part(role, '-', 1) AS semester_tag,
          split_part(role, '-', 2) AS course_name,
          split_part(role, '-', 3) AS user_role
        FROM unnest($1::text[]) AS role),
  user_course_roles AS (SELECT c.id,
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
          FROM course c
            INNER JOIN
           parsed_roles pr
           ON c.name = pr.course_name
            AND c.semester_tag = pr.semester_tag)
SELECT ucr.id,
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
FROM user_course_roles ucr
GROUP BY ucr.id,
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
HAVING NOT (
           ucr.archived
           AND COUNT(ucr.user_role) = 1
           AND MAX(ucr.user_role) = 'Student'
       )
ORDER BY ucr.template,
         ucr.semester_tag,
         ucr.name DESC;

-- name: CreateCourse :one
INSERT INTO course (id, name, start_date, end_date, semester_tag, course_type, ects, restricted_data,
                    student_readable_data, template, short_description, long_description)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: CheckCoursePhasesBelongToCourse :one
WITH matched_phases AS (SELECT id
                        FROM course_phase
                        WHERE id = ANY ($1::uuid[])
                          AND course_id = $2)
SELECT CASE
           WHEN COUNT(*) = cardinality($1::uuid[]) THEN true
           ELSE false
           END AS all_exist_and_match_course
FROM matched_phases;

-- name: ArchiveCourse :one
UPDATE course
SET archived = $2,
    archived_on = $3
WHERE id = $1
RETURNING *;


-- name: UpdateCourse :exec
UPDATE course
SET restricted_data       = restricted_data || $2,
    student_readable_data = student_readable_data || $3,
    start_date            = COALESCE($4, start_date),
    end_date              = COALESCE($5, end_date),
    ects                  = COALESCE($6, ects),
    course_type           = COALESCE(sqlc.narg('course_type'), course_type),
    short_description     = COALESCE($7, short_description),
    long_description      = COALESCE($8, long_description)
WHERE id = $1;


-- name: DeleteCourse :exec
DELETE
FROM course
WHERE id = $1;


-- name: GetOwnCourses :many
SELECT c.id
FROM course c
         JOIN course_participation cp ON c.id = cp.course_id
         JOIN student s ON cp.student_id = s.id
WHERE s.matriculation_number = $1
  AND s.university_login = $2;
