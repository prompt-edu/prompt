-- name: ListStudyPrograms :many
SELECT * FROM study_program ORDER BY name ASC;

-- name: GetStudyProgramByIDForUpdate :one
SELECT * FROM study_program WHERE id = $1 FOR UPDATE;

-- name: CreateStudyProgram :one
INSERT INTO study_program (id, name, short_name) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateStudyProgram :one
UPDATE study_program SET name = $2, short_name = $3 WHERE id = $1 RETURNING *;

-- name: DeleteStudyProgram :execrows
DELETE FROM study_program WHERE id = $1;

-- name: RenameStudentStudyProgram :execrows
UPDATE student
SET study_program = sqlc.arg(new_name)::text
WHERE trim(study_program) = sqlc.arg(old_name)::text;

-- name: CountStudentsPerStudyProgram :many
SELECT sp.id AS study_program_id, COUNT(s.id) AS student_count
FROM study_program sp
LEFT JOIN student s ON trim(s.study_program) = sp.name
GROUP BY sp.id;
