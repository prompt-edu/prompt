-- name: CreateOrUpdateAssessment :exec
INSERT INTO assessment (id,
                        course_participation_id,
                        course_phase_id,
                        competency_id,
                        score_level,
                        assessed_at,
                        author,
                        author_id)
VALUES (gen_random_uuid(), $1, $2, $3, $4, CURRENT_TIMESTAMP, $5, $6)
ON CONFLICT (course_participation_id, course_phase_id, competency_id)
    DO UPDATE
    SET score_level = EXCLUDED.score_level,
        assessed_at = CURRENT_TIMESTAMP,
        author      = EXCLUDED.author,
        author_id   = EXCLUDED.author_id;

-- name: DeleteAssessment :exec
DELETE
FROM assessment
WHERE id = $1;

-- name: GetAssessment :one
SELECT *
FROM assessment
WHERE id = $1;

-- name: ListAssessmentsByCoursePhase :many
SELECT *
FROM assessment
WHERE course_phase_id = $1;

-- name: ListAssessmentsByStudentInPhase :many
SELECT *
FROM assessment
WHERE course_participation_id = $1
  AND course_phase_id = $2;

-- name: ListAssessmentsByCompetencyInPhase :many
SELECT *
FROM assessment
WHERE competency_id = $1
  AND course_phase_id = $2;

-- name: ListAssessmentsByCategoryInPhase :many
SELECT a.*
FROM assessment a
         JOIN competency c ON a.competency_id = c.id
WHERE c.category_id = $1
  AND a.course_phase_id = $2;

-- name: CountRemainingAssessmentsForStudent :one
WITH total_competencies AS (SELECT COUNT(*) AS total
                            FROM competency c
                                     INNER JOIN category_course_phase ccp ON c.category_id = ccp.category_id
                            WHERE ccp.course_phase_id = $2),
     assessed_competencies AS (SELECT COUNT(*) AS assessed
                               FROM assessment a
                                        INNER JOIN competency c ON a.competency_id = c.id
                                        INNER JOIN category_course_phase ccp ON c.category_id = ccp.category_id
                               WHERE a.course_participation_id = $1
                                 AND a.course_phase_id = $2
                                 AND ccp.course_phase_id = $2),
     remaining_per_category AS (SELECT c.category_id,
                                       COUNT(*) - COUNT(ass.id) AS remaining_assessments
                                FROM competency c
                                         INNER JOIN category_course_phase ccp ON c.category_id = ccp.category_id
                                         LEFT JOIN assessment ass ON ass.competency_id = c.id
                                    AND ass.course_participation_id = $1
                                    AND ass.course_phase_id = $2
                                WHERE ccp.course_phase_id = $2
                                GROUP BY c.category_id)
SELECT (SELECT total FROM total_competencies) - (SELECT assessed FROM assessed_competencies) AS remaining_assessments,
       json_agg(
               json_build_object(
                       'categoryID', rpc.category_id,
                       'remainingAssessments', rpc.remaining_assessments
               )
       )                                                                                     AS categories
FROM remaining_per_category rpc;

-- name: PhaseHasAssessmentData :one
-- Blank comments/actions are excluded: the UI creates empty action items on click and category
-- comments have no delete route, so a bare EXISTS would permanently lock the phase. The regex
-- covers tabs and newlines too, which btrim's default (spaces only) would leave in place.
SELECT EXISTS(SELECT 1 FROM assessment a WHERE a.course_phase_id = $1)
           OR EXISTS(SELECT 1 FROM assessment_completion ac WHERE ac.course_phase_id = $1)
           OR EXISTS(SELECT 1
                     FROM category_assessment ca
                     WHERE ca.course_phase_id = $1
                       AND ca.comment !~ '^[[:space:]]*$')
           OR EXISTS(SELECT 1
                     FROM action_item ai
                     WHERE ai.course_phase_id = $1
                       AND ai.action !~ '^[[:space:]]*$') AS has_assessment_data;
