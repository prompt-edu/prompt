-- name: CreateCompetency :exec
INSERT INTO competency (id,
                        category_id,
                        name,
                        short_name,
                        description,
                        description_very_bad,
                        description_bad,
                        description_ok,
                        description_good,
                        description_very_good,
                        weight)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);

-- name: CheckCompetencyNameExists :one
-- Check if a competency name already exists within a given category
SELECT EXISTS(
    SELECT 1 FROM competency
    WHERE category_id = $1 AND name = $2
);

-- name: GetCompetency :one
SELECT *
FROM competency
WHERE id = $1;

-- name: GetAssessmentSchemaIDByCompetency :one
SELECT cat.assessment_schema_id
FROM competency comp
INNER JOIN category cat ON comp.category_id = cat.id
WHERE comp.id = $1;

-- name: ListCompetencies :many
SELECT *
FROM competency;

-- name: ListCompetenciesForCoursePhase :many
WITH phase_config AS (
    SELECT assessment_schema_id, self_evaluation_schema, peer_evaluation_schema, tutor_evaluation_schema
    FROM course_phase_config
    WHERE course_phase_id = sqlc.arg(course_phase_id)
)
SELECT comp.*
FROM competency comp
INNER JOIN category cat ON comp.category_id = cat.id
INNER JOIN assessment_schema s ON cat.assessment_schema_id = s.id
LEFT JOIN phase_config pc ON TRUE
WHERE s.source_phase_id IS NULL
   OR s.source_phase_id = sqlc.arg(course_phase_id)
   OR s.id = pc.assessment_schema_id
   OR s.id = pc.self_evaluation_schema
   OR s.id = pc.peer_evaluation_schema
   OR s.id = pc.tutor_evaluation_schema
ORDER BY comp.name;

-- name: ListCompetenciesByCategory :many
SELECT *
FROM competency
WHERE category_id = $1;

-- name: UpdateCompetency :exec
UPDATE competency
SET category_id           = $2,
    name                  = $3,
    short_name            = $4,
    description           = $5,
    description_very_bad  = $6,
    description_bad       = $7,
    description_ok        = $8,
    description_good      = $9,
    description_very_good = $10,
    weight                = $11
WHERE id = $1;

-- name: DeleteCompetency :exec
DELETE
FROM competency
WHERE id = $1;