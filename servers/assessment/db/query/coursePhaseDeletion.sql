-- Queries for deleting everything this service stores for a course phase. Every table except the
-- schema tables (assessment_schema, category, competency) is keyed by course_phase_id. Schemas are
-- owned by a phase through assessment_schema.source_phase_id, but other phases can use them too,
-- see TransferOwnedAssessmentSchemas.

-- name: DeleteAssessmentsByCoursePhase :exec
DELETE
FROM assessment
WHERE course_phase_id = $1;

-- name: DeleteIndependentAssessmentsByCoursePhase :exec
DELETE
FROM independent_assessment
WHERE course_phase_id = $1;

-- name: DeleteCategoryAssessmentsByCoursePhase :exec
DELETE
FROM category_assessment
WHERE course_phase_id = $1;

-- name: DeleteAssessmentCompletionsByCoursePhase :exec
DELETE
FROM assessment_completion
WHERE course_phase_id = $1;

-- name: DeleteEvaluationsByCoursePhase :exec
DELETE
FROM evaluation
WHERE course_phase_id = $1;

-- name: DeleteEvaluationCompletionsByCoursePhase :exec
DELETE
FROM evaluation_completion
WHERE course_phase_id = $1;

-- name: DeleteFeedbackItemsByCoursePhase :exec
DELETE
FROM feedback_items
WHERE course_phase_id = $1;

-- name: DeleteActionItemsByCoursePhase :exec
DELETE
FROM action_item
WHERE course_phase_id = $1;

-- name: DeleteEvaluationRemindersByCoursePhase :exec
DELETE
FROM evaluation_reminder
WHERE course_phase_id = $1;

-- name: DeleteResultsReleasedMailsByCoursePhase :exec
DELETE
FROM results_released_mail
WHERE course_phase_id = $1;

-- name: DeleteCoursePhaseConfigByCoursePhase :exec
DELETE
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: TransferOwnedAssessmentSchemas :exec
-- Hands every schema the course phase owns to another phase that still uses it, so deleting the
-- phase neither breaks that phase nor leaves a schema without an owner behind. A copied course
-- phase reuses the source phase's schemas, so this is the usual case when last semester's course is
-- deleted. A phase uses a schema when its config selects it or when its assessment data points
-- into it. The new owner is the first such phase by config start, the order migration
-- 0020_add_source_phase_id used to assign the original owners. Run it after the course phase's own
-- rows are deleted, so the phase itself no longer counts as a user.
WITH schema_usage AS (
    SELECT assessment_schema_id AS schema_id, course_phase_id
    FROM course_phase_config
    UNION
    SELECT self_evaluation_schema, course_phase_id
    FROM course_phase_config
    UNION
    SELECT peer_evaluation_schema, course_phase_id
    FROM course_phase_config
    UNION
    SELECT tutor_evaluation_schema, course_phase_id
    FROM course_phase_config
    UNION
    SELECT cat.assessment_schema_id, a.course_phase_id
    FROM assessment a
             JOIN competency co ON co.id = a.competency_id
             JOIN category cat ON cat.id = co.category_id
    UNION
    SELECT cat.assessment_schema_id, ia.course_phase_id
    FROM independent_assessment ia
             JOIN competency co ON co.id = ia.competency_id
             JOIN category cat ON cat.id = co.category_id
    UNION
    SELECT cat.assessment_schema_id, e.course_phase_id
    FROM evaluation e
             JOIN competency co ON co.id = e.competency_id
             JOIN category cat ON cat.id = co.category_id
    UNION
    SELECT cat.assessment_schema_id, ca.course_phase_id
    FROM category_assessment ca
             JOIN category cat ON cat.id = ca.category_id
),
     new_owner AS (
         SELECT DISTINCT ON (su.schema_id) su.schema_id, su.course_phase_id
         FROM schema_usage su
                  JOIN assessment_schema s ON s.id = su.schema_id
                  LEFT JOIN course_phase_config cpc ON cpc.course_phase_id = su.course_phase_id
         WHERE s.source_phase_id = sqlc.arg(course_phase_id)::uuid
           AND su.course_phase_id != sqlc.arg(course_phase_id)::uuid
         ORDER BY su.schema_id, cpc.start NULLS LAST, su.course_phase_id
     )
UPDATE assessment_schema s
SET source_phase_id = new_owner.course_phase_id,
    updated_at      = CURRENT_TIMESTAMP
FROM new_owner
WHERE s.id = new_owner.schema_id;

-- name: DeleteAssessmentSchemasOwnedByCoursePhase :exec
-- Deletes the schemas the course phase still owns, cascading to their categories and competencies.
-- Run it after TransferOwnedAssessmentSchemas, which hands every schema another phase uses to that
-- phase, so no other phase's config or assessment data can point into what is deleted here.
DELETE
FROM assessment_schema
WHERE source_phase_id = sqlc.arg(course_phase_id)::uuid;
