-- name: GetAllGrades :many
SELECT course_participation_id, grade_suggestion
FROM assessment_completion
WHERE course_phase_id = $1
  AND completed = true;

-- name: GetStudentGrade :one
SELECT grade_suggestion
FROM assessment_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND completed = true;

-- name: CreateOrUpdateAssessmentCompletion :exec
INSERT INTO assessment_completion (course_participation_id,
                                   course_phase_id,
                                   completed_at,
                                   author,
                                   comment,
                                   grade_suggestion,
                                   completed)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (course_participation_id, course_phase_id)
    DO UPDATE
    SET completed_at     = EXCLUDED.completed_at,
        author           = EXCLUDED.author,
        comment          = EXCLUDED.comment,
        grade_suggestion = EXCLUDED.grade_suggestion,
        completed        = EXCLUDED.completed;

-- name: MarkAssessmentAsFinished :exec
UPDATE assessment_completion
SET completed    = true,
    completed_at = $3,
    author       = $4
WHERE course_participation_id = $1
  AND course_phase_id = $2;


-- name: UnmarkAssessmentAsFinished :exec
UPDATE assessment_completion
SET completed = false
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: DeleteAssessmentCompletion :exec
DELETE
FROM assessment_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: GetAssessmentCompletionsByCoursePhase :many
SELECT *
FROM assessment_completion
WHERE course_phase_id = $1;

-- name: GetAssessmentCompletion :one
SELECT *
FROM assessment_completion
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: CheckAssessmentCompletionExists :one
SELECT EXISTS (SELECT 1
               FROM assessment_completion
               WHERE course_participation_id = $1
                 AND course_phase_id = $2);
