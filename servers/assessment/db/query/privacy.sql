-- Queries for gdpr export and deletion.

-- name: GetAllAssessmentsByCourseParticipationIDs :many
SELECT id, course_participation_id, course_phase_id, competency_id, assessed_at, score_level
FROM assessment
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllAssessmentCompletionsByCourseParticipationIDs :many
SELECT course_participation_id, course_phase_id, completed_at, comment, grade_suggestion, completed
FROM assessment_completion
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllCategoryAssessmentsByCourseParticipationIDs :many
SELECT id,
       category_id,
       course_phase_id,
       course_participation_id,
       -- This regex simply removes "(by <author>)" in the "## ..." heading; It was added in an older migration and we don't want to leak the author's name.
       COALESCE(regexp_replace(comment, '^(##[^\n]*?)\s*\(by[^)\n]*\)\s*$', '\1', 'ng'), '')::text AS comment,
       created_at,
       updated_at
FROM category_assessment
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllEvaluationsByCourseParticipationIDs :many
SELECT id, course_participation_id, course_phase_id, competency_id, score_level, evaluated_at, type
FROM evaluation
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllEvaluationCompletionsByCourseParticipationIDs :many
SELECT id, course_participation_id, course_phase_id, completed_at, completed, type
FROM evaluation_completion
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllActionItemsByCourseParticipationIDs :many
SELECT id, course_phase_id, course_participation_id, action, created_at
FROM action_item
WHERE course_participation_id = ANY($1::uuid[]);

-- name: GetAllFeedbackItemsByCourseParticipationIDs :many
SELECT id, feedback_type, feedback_text, course_participation_id, course_phase_id, created_at, type
FROM feedback_items
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteAssessmentsByCourseParticipationIDs :exec
DELETE FROM assessment
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteAssessmentCompletionsByCourseParticipationIDs :exec
DELETE FROM assessment_completion
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteCategoryAssessmentsByCourseParticipationIDs :exec
DELETE FROM category_assessment
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteEvaluationsByRecipientOrAuthorIDs :exec
DELETE FROM evaluation
WHERE course_participation_id = ANY($1::uuid[])
   OR author_course_participation_id = ANY($1::uuid[]);

-- name: DeleteEvaluationCompletionsByRecipientOrAuthorIDs :exec
DELETE FROM evaluation_completion
WHERE course_participation_id = ANY($1::uuid[])
   OR author_course_participation_id = ANY($1::uuid[]);

-- name: DeleteActionItemsByCourseParticipationIDs :exec
DELETE FROM action_item
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteFeedbackItemsByRecipientOrAuthorIDs :exec
DELETE FROM feedback_items
WHERE course_participation_id = ANY($1::uuid[])
   OR author_course_participation_id = ANY($1::uuid[]);
