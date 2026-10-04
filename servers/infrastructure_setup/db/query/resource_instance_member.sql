-- name: DeleteInstanceMembers :exec
DELETE FROM resource_instance_member
WHERE resource_instance_id = $1;

-- name: InsertInstanceMembers :exec
-- One row per person the instance's latest run was for. The arrays are parallel: the
-- n-th granted flag belongs to the n-th participation, and two unnests in one select
-- list advance together.
INSERT INTO resource_instance_member (resource_instance_id, course_participation_id, granted)
SELECT sqlc.arg(resource_instance_id)::uuid,
       unnest(sqlc.arg(course_participation_ids)::uuid[]),
       unnest(sqlc.arg(granted)::boolean[]);

-- name: ListInstanceMembersByCoursePhase :many
SELECT member.resource_instance_id,
       member.course_participation_id,
       member.granted
FROM resource_instance_member AS member
    JOIN resource_instance AS instance ON instance.id = member.resource_instance_id
WHERE instance.course_phase_id = $1;

-- name: GetInstanceMembershipsByCourseParticipationIDs :many
-- The resources a subject was provisioned into as a member, team resources included,
-- for the privacy export. The config is joined in so the export names what the
-- resource is rather than an opaque instance id.
SELECT instance.course_phase_id,
       config.provider_type,
       config.resource_type,
       config.scope,
       instance.resolved_name,
       instance.external_url,
       member.granted
FROM resource_instance_member AS member
    JOIN resource_instance AS instance ON instance.id = member.resource_instance_id
    JOIN resource_config AS config ON config.id = instance.resource_config_id
WHERE member.course_participation_id = ANY(sqlc.arg(course_participation_ids)::uuid[])
ORDER BY instance.created_at;

-- name: DeleteInstanceMembersByCourseParticipationIDs :exec
DELETE FROM resource_instance_member
WHERE course_participation_id = ANY(sqlc.arg(course_participation_ids)::uuid[]);

-- name: ListInstancesForParticipant :many
-- The instances a student sees: the ones provisioned for them personally and the ones
-- whose latest run was for them as a member, their team's included. granted is null
-- for an instance whose run predates member tracking.
SELECT instance.id,
       instance.resource_config_id,
       instance.status,
       instance.external_url,
       instance.target_name,
       instance.resolved_name,
       member.granted
FROM resource_instance AS instance
    LEFT JOIN resource_instance_member AS member
        ON member.resource_instance_id = instance.id
       AND member.course_participation_id = sqlc.arg(course_participation_id)
WHERE instance.course_phase_id = sqlc.arg(course_phase_id)
  AND (instance.course_participation_id = sqlc.arg(course_participation_id)
       OR member.course_participation_id IS NOT NULL)
ORDER BY instance.created_at;
