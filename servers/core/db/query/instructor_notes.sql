-- name: GetStudentNotesForStudent :many
SELECT * FROM note_with_versions WHERE for_student = $1 ORDER BY date_created ASC;

-- name: GetAllStudentNotes :many
SELECT * FROM note_with_versions ORDER BY date_created DESC;

-- name: GetSingleStudentNoteByID :one
SELECT * FROM note WHERE id = $1 LIMIT 1;

-- name: GetSingleNoteWithVersionsByID :one
SELECT * FROM note_with_versions WHERE id = $1 LIMIT 1;


-- name: GetLatestNoteVersionForNoteId :one
SELECT nv.version_number
FROM note_version nv
WHERE nv.for_note = $1
ORDER BY nv.version_number DESC
LIMIT 1;

-- name: CreateNoteVersion :one
INSERT INTO note_version ( id, content, date_created, version_number, for_note )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: CreateNote :one
INSERT INTO note (id, for_student, author, author_name, author_email, date_created, date_deleted, deleted_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: DeleteNote :one
UPDATE note
SET date_deleted = now(), deleted_by = $2
WHERE id = $1
AND date_deleted IS NULL
RETURNING *;

-- name: GetAllTags :many
SELECT id, name, color FROM note_tag ORDER BY name ASC;

-- name: GetTagByID :one
SELECT id, name, color FROM note_tag WHERE id = $1 LIMIT 1;

-- name: CreateTag :one
INSERT INTO note_tag (id, name, color) VALUES ($1, $2, $3) RETURNING id, name, color;

-- name: UpdateTag :one
UPDATE note_tag SET name = $2, color = $3 WHERE id = $1 RETURNING id, name, color;

-- name: DeleteTag :exec
DELETE FROM note_tag WHERE id = $1;

-- name: GetTagsForNote :many
SELECT nt.id, nt.name, nt.color FROM note_tag nt
JOIN note_tag_relation ntr ON ntr.tag_id = nt.id
WHERE ntr.note_id = $1
ORDER BY nt.name ASC;

-- name: AddTagToNote :exec
INSERT INTO note_tag_relation (note_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: RemoveTagFromNote :exec
DELETE FROM note_tag_relation WHERE note_id = $1 AND tag_id = $2;

-- name: RemoveAllTagsFromNote :exec
DELETE FROM note_tag_relation WHERE note_id = $1;

