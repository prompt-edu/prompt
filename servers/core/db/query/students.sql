-- name: GetStudent :one
SELECT * FROM student
WHERE id = $1 LIMIT 1;

-- name: GetStudentByCourseParticipationID :one
SELECT s.*
FROM student s
INNER JOIN course_participation cp ON s.id = cp.student_id
WHERE cp.id = $1;

-- name: GetAllStudents :many
SELECT * FROM student;

-- name: CreateStudent :one
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender, nationality, study_program, study_degree, current_semester)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: UpdateStudent :one
UPDATE student
SET first_name = $2,
    last_name = $3,
    email = $4,
    matriculation_number = $5,
    university_login = $6,
    has_university_account = $7,
    gender = $8,
    nationality = $9,
    study_program = $10,
    study_degree = $11,
    current_semester = $12
WHERE id = $1
RETURNING *;

-- name: GetStudentByEmail :one
SELECT * FROM student
WHERE email = $1 LIMIT 1;

-- name: GetStudentByMatriculationNumberAndUniversityLogin :one
SELECT * FROM student
WHERE matriculation_number = $1
  AND university_login = $2
LIMIT 1;

-- name: SearchStudents :many
SELECT *
FROM student
WHERE (first_name || ' ' || last_name) ILIKE '%' || $1 || '%'
   OR first_name ILIKE '%' || $1 || '%'
   OR last_name ILIKE '%' || $1 || '%'
   OR email ILIKE '%' || $1 || '%'
   OR matriculation_number ILIKE '%' || $1 || '%'
   OR university_login ILIKE '%' || $1 || '%';

-- name: GetStudentEmails :many
SELECT id, email
FROM student
WHERE id = ANY($1::uuid[]);

-- name: GetStudentsByEmail :many
SELECT * FROM student
WHERE email = ANY($1::text[]);

-- name: GetStudentUniversityLogins :many
SELECT id, university_login
FROM student
WHERE id = ANY($1::uuid[]);

-- name: GetAllStudentsWithCourseParticipations :many
SELECT
  s.id AS student_id,
  s.first_name AS student_first_name,
  s.last_name AS student_last_name,
  s.email AS student_email,
  s.has_university_account AS student_has_university_account,
  s.current_semester,
  s.study_program,
  COALESCE(
    jsonb_agg(
      jsonb_build_object(
        'courseId', c.id,
        'courseName', c.name,
        'studentReadableData', c.student_readable_data
      )
    ) FILTER (WHERE c.id IS NOT NULL),
    '[]'::jsonb
  )::jsonb AS courses,
  COALESCE(
    (
      SELECT jsonb_agg(jsonb_build_object('id', nt.id, 'name', nt.name, 'color', nt.color) ORDER BY nt.name)
      FROM (
        SELECT DISTINCT nt.id, nt.name, nt.color
        FROM note n
        JOIN note_tag_relation ntr ON ntr.note_id = n.id
        JOIN note_tag nt ON nt.id = ntr.tag_id
        WHERE n.for_student = s.id
          AND n.date_deleted IS NULL
      ) nt
    ),
    '[]'::jsonb
  )::jsonb AS note_tags
FROM student s
LEFT JOIN course_participation cp
  ON cp.student_id = s.id
LEFT JOIN course c
  ON c.id = cp.course_id
GROUP BY
  s.id,
  s.first_name,
  s.last_name,
  s.email,
  s.has_university_account;

-- name: GetStudentEnrollments :one
WITH RECURSIVE
course_participations AS (
    SELECT
        cp.id AS course_participation_id,
        cp.student_id,
        cp.course_id
    FROM course_participation cp
    WHERE cp.student_id = $1
),
phase_sequence AS (
    SELECT
        cph.id,
        cph.course_id,
        cph.name,
        cph.is_initial_phase,
        cph.course_phase_type_id,
        1 AS sequence_order
    FROM course_phase cph
    WHERE cph.is_initial_phase = true
    UNION ALL
    SELECT
        cph.id,
        cph.course_id,
        cph.name,
        cph.is_initial_phase,
        cph.course_phase_type_id,
        ps.sequence_order + 1
    FROM course_phase cph
    INNER JOIN course_phase_graph g
        ON g.to_course_phase_id = cph.id
    INNER JOIN phase_sequence ps
        ON g.from_course_phase_id = ps.id
),
course_phases AS (
    SELECT
        cp.student_id,
        cp.course_id,
        cp.course_participation_id,
        COALESCE(
            jsonb_agg(
                jsonb_build_object(
                    'coursePhaseId', ps.id,
                    'name', ps.name,
                    'isInitialPhase', ps.is_initial_phase,
                    'coursePhaseType', jsonb_build_object(
                        'id', cpt.id,
                        'name', cpt.name
                    ),
                    'passStatus', COALESCE(cpp.pass_status, 'not_assessed'),
                    'lastModified', cpp.last_modified::text
                )
                ORDER BY ps.sequence_order
            ),
            '[]'::jsonb
        ) AS course_phases
    FROM course_participations cp
    INNER JOIN phase_sequence ps
        ON ps.course_id = cp.course_id
    INNER JOIN course_phase_type cpt
        ON ps.course_phase_type_id = cpt.id
    LEFT JOIN course_phase_participation cpp
        ON cpp.course_participation_id = cp.course_participation_id
       AND cpp.course_phase_id = ps.id
    GROUP BY cp.student_id, cp.course_id, cp.course_participation_id
)
SELECT
    COALESCE(
        jsonb_agg(
            jsonb_build_object(
                'courseId', c.id,
                'courseParticipationId', cp.course_participation_id,
                'name', c.name,
                'semesterTag', c.semester_tag,
                'courseType', c.course_type,
                'ects', c.ects,
                'startDate', to_char(
                    c.start_date::timestamptz,
                    'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'
                ),
                'studentReadableData', c.student_readable_data,
                'endDate', to_char(
                    c.end_date::timestamptz,
                    'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'
                ),
                'longDescription', c.long_description,
                'coursePhases', cp.course_phases
            )
            ORDER BY c.start_date DESC, c.name
        ) FILTER (WHERE c.id IS NOT NULL),
        '[]'::jsonb
    )::jsonb AS courses
FROM student s
LEFT JOIN course_phases cp
    ON s.id = cp.student_id
LEFT JOIN course c
    ON cp.course_id = c.id
WHERE s.id = $1
GROUP BY s.id;
