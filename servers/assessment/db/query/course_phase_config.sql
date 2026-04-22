-- name: GetCoursePhaseConfig :one
SELECT *
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: GetCoursePhasesByAssessmentSchema :many
SELECT course_phase_id
FROM course_phase_config
WHERE assessment_schema_id = $1;

-- name: ListAssessmentSchemaCoursePhaseMappings :many
SELECT *
FROM course_phase_config
ORDER BY assessment_schema_id, course_phase_id;

-- name: IsAssessmentOpen :one
SELECT CASE
           WHEN start <= NOW() THEN true
           ELSE false
           END AS has_started
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsAssessmentDeadlinePassed :one
SELECT CASE
           WHEN deadline < NOW() THEN true
           ELSE false
           END AS is_deadline_passed
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsPeerEvaluationOpen :one
SELECT CASE
           WHEN peer_evaluation_enabled AND peer_evaluation_start <= NOW() THEN true
           ELSE false
           END AS has_peer_evaluation_started
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsPeerEvaluationDeadlinePassed :one
SELECT CASE
           WHEN peer_evaluation_deadline < NOW() THEN true
           ELSE false
           END AS is_peer_evaluation_deadline_passed
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsTutorEvaluationOpen :one
SELECT CASE
           WHEN tutor_evaluation_enabled AND tutor_evaluation_start <= NOW() THEN true
           ELSE false
           END AS has_tutor_evaluation_started
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsTutorEvaluationDeadlinePassed :one
SELECT CASE
           WHEN tutor_evaluation_deadline < NOW() THEN true
           ELSE false
           END AS is_tutor_evaluation_deadline_passed
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsSelfEvaluationOpen :one
SELECT CASE
           WHEN self_evaluation_enabled AND self_evaluation_start <= NOW() THEN true
           ELSE false
           END AS has_self_evaluation_started
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: IsSelfEvaluationDeadlinePassed :one
SELECT CASE
           WHEN self_evaluation_enabled AND self_evaluation_deadline < NOW() THEN true
           ELSE false
           END AS is_self_evaluation_deadline_passed
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: GetSelfEvaluationTimeframe :one
SELECT self_evaluation_start, self_evaluation_deadline
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: GetPeerEvaluationTimeframe :one
SELECT peer_evaluation_start, peer_evaluation_deadline
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: GetTutorEvaluationTimeframe :one
SELECT tutor_evaluation_start, tutor_evaluation_deadline
FROM course_phase_config
WHERE course_phase_id = $1;

-- name: CreateDefaultCoursePhaseConfig :exec
INSERT INTO course_phase_config (course_phase_id)
VALUES ($1);

-- name: CheckAssessmentSchemaUsageInOtherPhases :one
SELECT EXISTS(
    SELECT 1
    FROM course_phase_config cpc
    WHERE (
        cpc.assessment_schema_id = $1
        OR cpc.self_evaluation_schema = $1
        OR cpc.peer_evaluation_schema = $1
        OR cpc.tutor_evaluation_schema = $1
    )
    AND cpc.course_phase_id != $2
    AND EXISTS(
        SELECT 1
        FROM assessment a
        WHERE a.course_phase_id = cpc.course_phase_id
        AND a.competency_id IN (
            SELECT co.id
            FROM competency co
            INNER JOIN category cat ON co.category_id = cat.id
            WHERE cat.assessment_schema_id = $1
        )
        UNION
        SELECT 1
        FROM evaluation e
        WHERE e.course_phase_id = cpc.course_phase_id
        AND e.competency_id IN (
            SELECT co.id
            FROM competency co
            INNER JOIN category cat ON co.category_id = cat.id
            WHERE cat.assessment_schema_id = $1
        )
    )
) AS schema_used_in_other_phases;

-- name: CreateOrUpdateCoursePhaseConfig :exec
INSERT INTO course_phase_config (assessment_schema_id,
                                 course_phase_id,
                                 start,
                                 deadline,
                                 self_evaluation_enabled,
                                 self_evaluation_schema,
                                 self_evaluation_start,
                                 self_evaluation_deadline,
                                 peer_evaluation_enabled,
                                 peer_evaluation_schema,
                                 peer_evaluation_start,
                                 peer_evaluation_deadline,
                                 tutor_evaluation_enabled,
                                 tutor_evaluation_schema,
                                 tutor_evaluation_start,
                                 tutor_evaluation_deadline,
                                 evaluation_results_visible,
                                 grade_suggestion_visible,
                                 action_items_visible,
                                 results_released,
                                 grading_sheet_visible)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, 
        COALESCE(sqlc.narg('grade_suggestion_visible')::boolean, TRUE), 
        COALESCE(sqlc.narg('action_items_visible')::boolean, TRUE),
        COALESCE(sqlc.narg('results_released')::boolean, FALSE),
        COALESCE(sqlc.narg('grading_sheet_visible')::boolean, FALSE))
ON CONFLICT (course_phase_id)
    DO UPDATE SET assessment_schema_id      = EXCLUDED.assessment_schema_id,
                  start                     = EXCLUDED.start,
                  deadline                  = EXCLUDED.deadline,
                  self_evaluation_enabled   = EXCLUDED.self_evaluation_enabled,
                  self_evaluation_schema    = EXCLUDED.self_evaluation_schema,
                  self_evaluation_start     = EXCLUDED.self_evaluation_start,
                  self_evaluation_deadline  = EXCLUDED.self_evaluation_deadline,
                  peer_evaluation_enabled   = EXCLUDED.peer_evaluation_enabled,
                  peer_evaluation_schema    = EXCLUDED.peer_evaluation_schema,
                  peer_evaluation_start     = EXCLUDED.peer_evaluation_start,
                  peer_evaluation_deadline  = EXCLUDED.peer_evaluation_deadline,
                  tutor_evaluation_enabled  = EXCLUDED.tutor_evaluation_enabled,
                  tutor_evaluation_schema   = EXCLUDED.tutor_evaluation_schema,
                  tutor_evaluation_start    = EXCLUDED.tutor_evaluation_start,
                  tutor_evaluation_deadline = EXCLUDED.tutor_evaluation_deadline,
                  evaluation_results_visible = EXCLUDED.evaluation_results_visible,
                  grade_suggestion_visible  = COALESCE(EXCLUDED.grade_suggestion_visible, TRUE),
                  action_items_visible      = COALESCE(EXCLUDED.action_items_visible, TRUE),
                  results_released          = COALESCE(EXCLUDED.results_released, FALSE),
                  grading_sheet_visible     = COALESCE(EXCLUDED.grading_sheet_visible, FALSE);

-- name: UpdateCoursePhaseConfigAssessmentSchema :exec
UPDATE course_phase_config
SET assessment_schema_id = $2
WHERE course_phase_id = $1;

-- name: UpdateCoursePhaseConfigResultsReleased :exec
UPDATE course_phase_config
SET results_released = $2
WHERE course_phase_id = $1;

-- name: UpdateCoursePhaseConfigSelfEvaluationSchema :exec
UPDATE course_phase_config
SET self_evaluation_schema = $2
WHERE course_phase_id = $1;

-- name: UpdateCoursePhaseConfigPeerEvaluationSchema :exec
UPDATE course_phase_config
SET peer_evaluation_schema = $2
WHERE course_phase_id = $1;

-- name: UpdateCoursePhaseConfigTutorEvaluationSchema :exec
UPDATE course_phase_config
SET tutor_evaluation_schema = $2
WHERE course_phase_id = $1;
