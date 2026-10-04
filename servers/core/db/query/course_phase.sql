-- name: GetCoursePhase :one
SELECT cp.*, cpt.name AS course_phase_type_name
FROM course_phase cp
INNER JOIN course_phase_type cpt ON cp.course_phase_type_id = cpt.id
WHERE cp.id = $1
LIMIT 1;

-- name: GetAllCoursePhaseForCourse :many
SELECT cp.*, cpt.name AS course_phase_type_name
FROM course_phase cp
INNER JOIN course_phase_type cpt ON cp.course_phase_type_id = cpt.id
WHERE cp.course_id = $1;

-- name: CreateCoursePhase :one
INSERT INTO course_phase (id, course_id, name, is_initial_phase, restricted_data, student_readable_data, course_phase_type_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateCoursePhase :exec
UPDATE course_phase
SET
    name = COALESCE($2, name),
    restricted_data = restricted_data || $3,
    student_readable_data = student_readable_data || $4
WHERE id = $1;

-- name: DeleteCoursePhase :exec
DELETE FROM course_phase
WHERE id = $1;

-- name: GetCourseIDByCoursePhaseID :one
SELECT course_id
FROM course_phase
WHERE id = $1
LIMIT 1;

-- name: GetResolutionsForCoursePhase :many
SELECT po.dto_name, cpt.base_url, po.endpoint_path, mdg.from_course_phase_id
FROM participation_data_dependency_graph mdg
JOIN course_phase_type_participation_provided_output_dto po
  ON po.id = mdg.from_course_phase_DTO_id
JOIN course_phase_type cpt
  ON cpt.id = po.course_phase_type_id
WHERE mdg.to_course_phase_id = $1
    AND po.endpoint_path <> 'core';



-- name: GetPrevCoursePhaseDataFromCore :one
SELECT jsonb_object_agg(sub.dto_name, sub.dto_value) AS incoming_dto_objects
FROM (
  SELECT pod.dto_name,
         COALESCE(cp.student_readable_data -> pod.dto_name,
                  cp.restricted_data -> pod.dto_name) AS dto_value
  FROM phase_data_dependency_graph pdg
  JOIN course_phase cp
    ON cp.id = pdg.from_course_phase_id
  JOIN course_phase_type_phase_provided_output_dto pod
    ON pod.id = pdg.from_course_phase_DTO_id
  WHERE pdg.to_course_phase_id = $1
    AND pod.endpoint_path = 'core'
) sub;

-- name: GetPrevCoursePhaseDataResolution :many
SELECT po.dto_name, cpt.base_url, po.endpoint_path, dg.from_course_phase_id
FROM phase_data_dependency_graph dg
JOIN course_phase_type_phase_provided_output_dto po
  ON po.id = dg.from_course_phase_DTO_id
JOIN course_phase_type cpt
  ON cpt.id = po.course_phase_type_id
WHERE dg.to_course_phase_id = $1
    AND po.endpoint_path <> 'core';

-- name: GetCoursePhaseDeletionTargets :many
SELECT cp.id, cpt.name AS course_phase_type_name, cpt.base_url
FROM course_phase cp
JOIN course_phase_type cpt ON cpt.id = cp.course_phase_type_id
WHERE cp.id = ANY(sqlc.arg(course_phase_ids)::uuid[]);
