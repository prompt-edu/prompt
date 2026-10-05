-- name: CreateOrUpdateIndependentAssessment :exec
INSERT INTO independent_assessment (course_participation_id,
                                    course_phase_id,
                                    competency_id,
                                    score_level,
                                    author,
                                    author_id)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_participation_id, course_phase_id, competency_id, author_id)
    DO UPDATE
    SET score_level = EXCLUDED.score_level,
        assessed_at = CURRENT_TIMESTAMP,
        author      = EXCLUDED.author;

-- name: ListIndependentAssessmentsByStudentInPhase :many
SELECT *
FROM independent_assessment
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: DeleteOwnIndependentAssessment :one
DELETE
FROM independent_assessment
WHERE id = $1
  AND course_phase_id = $2
  AND author_id = $3
RETURNING course_participation_id;
