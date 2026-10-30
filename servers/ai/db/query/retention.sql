-- name: ListCallFeatures :many
SELECT DISTINCT feature
FROM ai_call;

-- name: PurgeCallContents :execrows
WITH expired AS (SELECT id FROM ai_call WHERE feature = @feature AND requested_at < @cutoff),
     deleted_subjects AS (DELETE FROM ai_call_subject WHERE call_id IN (SELECT id FROM expired)),
     deleted_restrictions AS (DELETE FROM ai_call_content_restriction WHERE call_id IN (SELECT id FROM expired))
DELETE
FROM ai_call_content
WHERE call_id IN (SELECT id FROM expired);

-- name: PurgeCalls :execrows
DELETE
FROM ai_call
WHERE requested_at < @cutoff;

-- name: AbandonPendingCalls :execrows
UPDATE ai_call
SET outcome      = 'error',
    error_code   = 'abandoned',
    completed_at = now()
WHERE completed_at IS NULL
  AND requested_at < @cutoff;
