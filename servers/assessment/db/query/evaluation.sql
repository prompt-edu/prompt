-- name: CreateOrUpdateEvaluation :exec
INSERT INTO evaluation (course_participation_id,
                        course_phase_id,
                        competency_id,
                        score_level,
                        author_course_participation_id,
                        type)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (course_participation_id, course_phase_id, competency_id, author_course_participation_id)
    DO UPDATE SET score_level  = EXCLUDED.score_level,
                  type         = EXCLUDED.type,
                  evaluated_at = CURRENT_TIMESTAMP;

-- name: DeleteEvaluation :exec
DELETE
FROM evaluation
WHERE id = $1;

-- name: GetEvaluationsByPhase :many
SELECT *
FROM evaluation
WHERE course_phase_id = $1;

-- name: GetEvaluationsForParticipantInPhase :many
SELECT *
FROM evaluation
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND type != 'tutor';

-- name: GetEvaluationsForTutorInPhase :many
SELECT *
FROM evaluation
WHERE course_participation_id = $1
  AND course_phase_id = $2
  AND type = 'tutor';

-- name: GetEvaluationsForAuthorInPhase :many
SELECT *
FROM evaluation
WHERE author_course_participation_id = $1
  AND course_phase_id = $2;

-- name: GetEvaluationByID :one
SELECT *
FROM evaluation
WHERE id = $1;

-- name: CountRemainingEvaluationsForStudent :one
WITH total_competencies AS (SELECT COUNT(*) AS total
                            FROM competency c
                                     INNER JOIN category cat ON c.category_id = cat.id
                                     INNER JOIN course_phase_config cpc ON
                                CASE
                                    WHEN $4::assessment_type = 'self' THEN
                                        cat.assessment_schema_id = cpc.self_evaluation_schema
                                    WHEN $4::assessment_type = 'peer' THEN
                                        cat.assessment_schema_id = cpc.peer_evaluation_schema
                                    WHEN $4::assessment_type = 'tutor' THEN
                                        cat.assessment_schema_id = cpc.tutor_evaluation_schema
                                    ELSE
                                        cat.assessment_schema_id = cpc.assessment_schema_id
                                    END
                            WHERE cpc.course_phase_id = $3),
     evaluated_competencies AS (SELECT COUNT(*) AS evaluated
                                FROM evaluation e
                                         INNER JOIN competency c ON e.competency_id = c.id
                                         INNER JOIN category cat ON c.category_id = cat.id
                                         INNER JOIN course_phase_config cpc ON
                                    CASE
                                        WHEN $4::assessment_type = 'self' THEN
                                            cat.assessment_schema_id = cpc.self_evaluation_schema
                                        WHEN $4::assessment_type = 'peer' THEN
                                            cat.assessment_schema_id = cpc.peer_evaluation_schema
                                        WHEN $4::assessment_type = 'tutor' THEN
                                            cat.assessment_schema_id = cpc.tutor_evaluation_schema
                                        ELSE
                                            cat.assessment_schema_id = cpc.assessment_schema_id
                                        END
                                WHERE e.course_participation_id = $1
                                  AND e.course_phase_id = $3
                                  AND e.author_course_participation_id = $2
                                  AND e.type = $4::assessment_type
                                  AND cpc.course_phase_id = $3)
SELECT (SELECT total FROM total_competencies) - (SELECT evaluated FROM evaluated_competencies) AS remaining_evaluations;
