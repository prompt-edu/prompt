-- name: CreateOrUpdateEvaluationCompletion :exec
INSERT INTO evaluation_completion (course_participation_id,
                                   course_phase_id,
                                   author_course_participation_id,
                                   completed_at,
                                   completed,
                                   type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_participation_id, course_phase_id, author_course_participation_id)
    DO UPDATE
    SET completed_at = EXCLUDED.completed_at,
        completed    = EXCLUDED.completed,
        type         = EXCLUDED.type;

-- name: MarkEvaluationAsFinished :exec
INSERT INTO evaluation_completion (course_participation_id,
                   course_phase_id,
                   author_course_participation_id,
                   completed_at,
                   completed,
                   type)
VALUES ($1, $2, $3, $4, true, $5)
ON CONFLICT (course_participation_id, course_phase_id, author_course_participation_id)
  DO UPDATE
  SET completed_at = EXCLUDED.completed_at,
    completed    = true,
    type         = EXCLUDED.type;

-- name: UnmarkEvaluationAsFinished :exec
UPDATE evaluation_completion
SET completed = false
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND author_course_participation_id = $3;

-- name: DeleteEvaluationCompletion :exec
DELETE
FROM evaluation_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND author_course_participation_id = $3;

-- name: GetEvaluationCompletionsByCoursePhase :many
SELECT *
FROM evaluation_completion
WHERE course_phase_id = $1;

-- name: GetEvaluationCompletion :one
SELECT *
FROM evaluation_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND author_course_participation_id = $3;

-- name: CheckEvaluationCompletionExists :one
SELECT EXISTS (SELECT 1
               FROM evaluation_completion
               WHERE course_participation_id = $1
                 AND course_phase_id = $2
                 AND author_course_participation_id = $3);

-- name: GetEvaluationCompletionsForParticipantInPhase :many
SELECT *
FROM evaluation_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: GetEvaluationCompletionsForAuthorInPhase :many
SELECT *
FROM evaluation_completion
WHERE author_course_participation_id = $1
  AND course_phase_id = $2;

-- name: GetEvaluationCompletionByType :one
SELECT *
FROM evaluation_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND author_course_participation_id = $3
  AND type = $4;
