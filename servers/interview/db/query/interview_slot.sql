-- name: CreateInterviewSlot :one
INSERT INTO interview_slot (
    course_phase_id,
    start_time,
    end_time,
    location,
    capacity
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetInterviewSlot :one
SELECT * FROM interview_slot
WHERE id = $1;

-- name: GetInterviewSlotForUpdate :one
SELECT * FROM interview_slot
WHERE id = $1
FOR UPDATE;

-- name: GetInterviewSlotsByCoursePhase :many
SELECT * FROM interview_slot
WHERE course_phase_id = $1
ORDER BY start_time ASC;

-- name: UpdateInterviewSlot :one
UPDATE interview_slot
SET start_time = $2,
    end_time = $3,
    location = $4,
    capacity = $5,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteInterviewSlot :exec
DELETE FROM interview_slot
WHERE id = $1;

-- name: CreateInterviewAssignment :one
INSERT INTO interview_assignment (
    interview_slot_id,
    course_participation_id
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetInterviewAssignment :one
SELECT * FROM interview_assignment
WHERE id = $1;

-- name: GetInterviewAssignmentByParticipation :one
SELECT ia.* FROM interview_assignment ia
JOIN interview_slot s ON ia.interview_slot_id = s.id
WHERE ia.course_participation_id = $1 AND s.course_phase_id = $2;

-- name: GetInterviewAssignmentsBySlot :many
SELECT * FROM interview_assignment
WHERE interview_slot_id = $1;

-- name: DeleteInterviewAssignment :exec
DELETE FROM interview_assignment
WHERE id = $1;

-- name: DeleteInterviewAssignmentByParticipation :exec
DELETE FROM interview_assignment
WHERE course_participation_id = $1;

-- name: GetInterviewSlotWithAssignments :many
SELECT 
    s.id as slot_id,
    s.course_phase_id,
    s.start_time,
    s.end_time,
    s.location,
    s.capacity,
    s.created_at,
    s.updated_at,
    a.id as assignment_id,
    a.course_participation_id,
    a.assigned_at
FROM interview_slot s
LEFT JOIN interview_assignment a ON s.id = a.interview_slot_id
WHERE s.course_phase_id = $1
ORDER BY s.start_time ASC;

-- name: CountAssignmentsBySlot :one
SELECT COUNT(*) FROM interview_assignment
WHERE interview_slot_id = $1;
