-- name: GetInterviewAssignmentsByParticipationIDs :many
SELECT
    ia.id,
    ia.course_participation_id,
    ia.assigned_at,
    s.course_phase_id,
    s.start_time,
    s.end_time,
    s.location
FROM interview_assignment ia
JOIN interview_slot s ON ia.interview_slot_id = s.id
WHERE ia.course_participation_id = ANY(@course_participation_ids::uuid[])
ORDER BY s.start_time ASC;
