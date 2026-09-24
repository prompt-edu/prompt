-- name: DeleteCertificateDownloadsByCoursePhase :exec
DELETE FROM certificate_download
WHERE
    course_phase_id = $1;

-- name: DeleteCoursePhaseConfigByCoursePhase :exec
-- Removes the template, release date, and student page text of the course phase.
DELETE FROM course_phase_config
WHERE
    course_phase_id = $1;
