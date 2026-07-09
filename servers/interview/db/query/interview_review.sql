-- name: UpsertInterviewReview :one
INSERT INTO interview_review (
    course_phase_id,
    course_participation_id,
    score,
    interviewer,
    interview_answers,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, now()
)
ON CONFLICT (course_phase_id, course_participation_id)
DO UPDATE SET
    score = EXCLUDED.score,
    interviewer = EXCLUDED.interviewer,
    interview_answers = EXCLUDED.interview_answers,
    updated_at = now()
RETURNING *;

-- name: GetInterviewReview :one
SELECT * FROM interview_review
WHERE course_phase_id = $1
  AND course_participation_id = $2;

-- name: GetInterviewReviewsByCoursePhase :many
SELECT * FROM interview_review
WHERE course_phase_id = $1
ORDER BY course_participation_id ASC;

-- name: GetScoredInterviewReviewsByCoursePhase :many
SELECT * FROM interview_review
WHERE course_phase_id = $1
  AND score IS NOT NULL
ORDER BY course_participation_id ASC;

-- name: DeleteInterviewReviewByParticipation :exec
DELETE FROM interview_review
WHERE course_participation_id = ANY(@course_participation_ids::uuid[]);
