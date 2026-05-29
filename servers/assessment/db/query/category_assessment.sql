-- name: CreateOrUpdateCategoryAssessment :exec
INSERT INTO category_assessment (id, category_id, course_phase_id, course_participation_id,
                                 comment, author, author_id, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, now(), now())
ON CONFLICT (category_id, course_phase_id, course_participation_id)
    DO UPDATE
    SET comment    = EXCLUDED.comment,
        author     = EXCLUDED.author,
        author_id  = EXCLUDED.author_id,
        updated_at = now();

-- name: ListCategoryAssessmentsByStudentInPhase :many
SELECT *
FROM category_assessment
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: ListCategoryAssessmentsByCoursePhase :many
SELECT *
FROM category_assessment
WHERE course_phase_id = $1;

-- name: GetCategoryAssessment :one
SELECT *
FROM category_assessment
WHERE id = $1;

-- name: DeleteCategoryAssessment :exec
DELETE
FROM category_assessment
WHERE id = $1;
