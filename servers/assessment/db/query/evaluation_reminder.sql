-- name: GetEvaluationRemindersForCoursePhase :many
SELECT *
FROM evaluation_reminder
WHERE course_phase_id = $1;

-- name: GetEvaluationReminder :one
SELECT *
FROM evaluation_reminder
WHERE course_phase_id = $1
  AND evaluation_type = $2;

-- name: UpsertEvaluationReminderLastSentAt :exec
INSERT INTO evaluation_reminder (course_phase_id, evaluation_type, last_sent_at)
VALUES ($1, $2, $3)
ON CONFLICT (course_phase_id, evaluation_type) DO UPDATE SET last_sent_at = GREATEST(evaluation_reminder.last_sent_at, EXCLUDED.last_sent_at);
