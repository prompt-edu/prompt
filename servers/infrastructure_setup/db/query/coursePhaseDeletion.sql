-- Deletes every row this service stores for one course phase, called when the phase is
-- permanently removed in core.
--
-- The tables are deleted from the leaves up even though the schema cascades
-- (resource_instance -> resource_config -> provider_config): an explicit delete per root
-- table keeps the handler correct if a migration ever weakens one of those cascades, and
-- the tests count the raw tables so a weakened cascade fails loudly.

-- name: DeleteResourceInstancesByCoursePhase :exec
DELETE FROM resource_instance
WHERE course_phase_id = $1;

-- name: DeleteResourceConfigsByCoursePhase :exec
DELETE FROM resource_config
WHERE course_phase_id = $1;

-- name: DeleteProviderConfigsByCoursePhase :exec
DELETE FROM provider_config
WHERE course_phase_id = $1;

-- name: DeleteCoursePhaseConfigByCoursePhase :exec
DELETE FROM course_phase_config
WHERE course_phase_id = $1;
