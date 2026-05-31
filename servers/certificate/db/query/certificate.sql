-- name: GetCoursePhaseConfig :one
SELECT * FROM course_phase_config WHERE course_phase_id = $1 LIMIT 1;

-- name: CreateCoursePhaseConfig :one
INSERT INTO
    course_phase_config (course_phase_id)
VALUES ($1)
RETURNING
    *;

-- name: UpdateCoursePhaseConfig :one
UPDATE course_phase_config
SET
    template_content = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE
    course_phase_id = $1
RETURNING
    *;

-- name: UpsertCoursePhaseConfig :one
INSERT INTO
    course_phase_config (
        course_phase_id,
        template_content,
        updated_at,
        updated_by
    )
VALUES ($1, $2, NOW(), $3)
ON CONFLICT (course_phase_id) DO
UPDATE
SET
    template_content = EXCLUDED.template_content,
    updated_at = NOW(),
    updated_by = EXCLUDED.updated_by
RETURNING
    *;

-- name: UpdateReleaseDate :one
UPDATE course_phase_config
SET
    release_date = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE
    course_phase_id = $1
RETURNING
    *;

-- name: HasDownloads :one
SELECT EXISTS (
        SELECT 1
        FROM certificate_download
        WHERE
            course_phase_id = $1
    ) AS has_downloads;

-- name: GetCertificateDownload :one
SELECT *
FROM certificate_download
WHERE
    student_id = $1
    AND course_phase_id = $2
LIMIT 1;

-- name: ListCertificateDownloadsByCoursePhase :many
SELECT *
FROM certificate_download
WHERE
    course_phase_id = $1
ORDER BY last_download DESC;

-- name: RecordCertificateDownload :one
INSERT INTO
    certificate_download (
        student_id,
        course_phase_id,
        first_download,
        last_download,
        download_count
    )
VALUES ($1, $2, NOW(), NOW(), 1)
ON CONFLICT (student_id, course_phase_id) DO
UPDATE
SET
    last_download = NOW(),
    download_count = certificate_download.download_count + 1
RETURNING
    *;

-- name: DeleteCertificateDownload :exec
DELETE FROM certificate_download
WHERE
    student_id = $1
    AND course_phase_id = $2;

-- name: CountDownloadsByCoursePhase :one
SELECT COUNT(*) AS download_count
FROM certificate_download
WHERE
    course_phase_id = $1;