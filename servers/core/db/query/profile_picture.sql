-- name: GetProfilePictureByUserID :one
SELECT pp.user_id, pp.university_login, pp.file_id, pp.updated_at, f.storage_key
FROM profile_picture pp
JOIN files f ON f.id = pp.file_id
WHERE pp.user_id = $1;

-- name: UpsertProfilePicture :one
INSERT INTO profile_picture (user_id, university_login, file_id)
VALUES ($1, $2, $3)
ON CONFLICT (user_id) DO UPDATE
SET university_login = EXCLUDED.university_login,
    file_id          = EXCLUDED.file_id,
    updated_at       = CURRENT_TIMESTAMP
RETURNING *;

-- name: DeleteProfilePictureByUserID :exec
DELETE FROM profile_picture
WHERE user_id = $1;

-- name: GetProfilePictureStorageKeysByUserIDs :many
SELECT pp.user_id, f.storage_key
FROM profile_picture pp
JOIN files f ON f.id = pp.file_id
WHERE pp.user_id = ANY(sqlc.arg(user_ids)::uuid[]);

-- If several accounts share a university login, the most recent upload wins.
-- name: GetProfilePictureStorageKeysByStudentIDs :many
SELECT DISTINCT ON (s.id) s.id AS student_id, f.storage_key
FROM student s
JOIN profile_picture pp ON pp.university_login = s.university_login
JOIN files f ON f.id = pp.file_id
WHERE s.id = ANY(sqlc.arg(student_ids)::uuid[])
ORDER BY s.id, pp.updated_at DESC;

-- name: GetProfilePictureStorageKeysByCourseParticipationIDs :many
SELECT DISTINCT ON (cp.id) cp.id AS course_participation_id, f.storage_key
FROM course_participation cp
JOIN student s ON s.id = cp.student_id
JOIN profile_picture pp ON pp.university_login = s.university_login
JOIN files f ON f.id = pp.file_id
WHERE cp.id = ANY(sqlc.arg(course_participation_ids)::uuid[])
ORDER BY cp.id, pp.updated_at DESC;

-- name: GetProfilePictureFileIDsForStudent :many
SELECT pp.file_id
FROM profile_picture pp
JOIN student s ON s.university_login = pp.university_login
WHERE s.id = $1;

-- Serializes picture changes of one user until the transaction ends, so replacing a picture always
-- sees the file it replaces, even for the first upload when no row exists to lock yet.
-- name: LockProfilePictureOfUser :exec
SELECT pg_advisory_xact_lock(hashtextextended(sqlc.arg(user_id)::uuid::text, 0));
