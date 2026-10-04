-- name: ListPrivacyCalls :many
-- No actor id: on a call the subject did not make, it identifies the lecturer.
SELECT ai_call.id,
       ai_call.course_phase_id,
       (ai_call.actor_id = @actor_id::text)::boolean AS is_actor,
       ai_call.actor_role,
       ai_call.feature,
       ai_call.template,
       ai_call.template_version,
       ai_call.requested_model,
       ai_call.served_model,
       ai_call.outcome,
       ai_call.finish_reason,
       ai_call.prompt_tokens,
       ai_call.completion_tokens,
       ai_call.streamed,
       ai_call.requested_at,
       ai_call.completed_at,
       content.request,
       content.response,
       (restriction.call_id IS NOT NULL)::boolean AS restricted,
       EXISTS (SELECT 1
               FROM ai_call_subject subject
               WHERE subject.call_id = ai_call.id
                 AND subject.course_participation_id = ANY (@course_participation_ids::uuid[])) AS is_subject,
       EXISTS (SELECT 1
               FROM ai_call_subject subject
               WHERE subject.call_id = ai_call.id
                 AND NOT subject.course_participation_id = ANY (@course_participation_ids::uuid[])) AS has_other_subjects
FROM ai_call
         LEFT JOIN ai_call_content content ON content.call_id = ai_call.id
         LEFT JOIN ai_call_content_restriction restriction ON restriction.call_id = ai_call.id
WHERE ai_call.actor_id = @actor_id::text
   OR ai_call.id IN (SELECT call_id
                     FROM ai_call_subject
                     WHERE course_participation_id = ANY (@course_participation_ids::uuid[]))
ORDER BY ai_call.requested_at, ai_call.id;

-- name: ListPrivacyEvents :many
SELECT id, call_id, type, data, created_at, (actor_id = @actor_id::text)::boolean AS is_actor
FROM ai_call_event
WHERE call_id = ANY (@call_ids::uuid[])
ORDER BY created_at, id;

-- name: ListSubjectCalls :many
SELECT DISTINCT ai_call.id, ai_call.feature
FROM ai_call
         JOIN ai_call_subject subject ON subject.call_id = ai_call.id
WHERE subject.course_participation_id = ANY (@course_participation_ids::uuid[]);

-- name: RestrictCallContents :exec
INSERT INTO ai_call_content_restriction (call_id)
SELECT unnest(@call_ids::uuid[])
ON CONFLICT (call_id) DO NOTHING;

-- name: DeleteSubjectCallContents :exec
WITH deleted_subjects AS (DELETE
                          FROM ai_call_subject
                          WHERE call_id = ANY (@call_ids::uuid[])
                            AND course_participation_id = ANY (@course_participation_ids::uuid[])),
     deleted_restrictions AS (DELETE FROM ai_call_content_restriction WHERE call_id = ANY (@call_ids::uuid[]))
DELETE
FROM ai_call_content
WHERE call_id = ANY (@call_ids::uuid[]);
