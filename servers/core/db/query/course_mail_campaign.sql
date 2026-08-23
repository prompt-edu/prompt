-- name: CreateMailCampaign :one
INSERT INTO mail_campaign (
    course_id, name, subject, body, target_course_phase_id, target_pass_statuses,
    reply_to_override, cc_override, bcc_override,
    created_by_id, created_by_email, created_by_name,
    updated_by_id, updated_by_email, updated_by_name
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9,
    $10, $11, $12,
    $13, $14, $15
)
RETURNING *;

-- name: GetMailCampaignBase :one
SELECT * FROM mail_campaign WHERE id = $1 AND course_id = $2;

-- name: ListMailCampaignsForCourse :many
SELECT
    mc.*,
    COUNT(mcr.id) AS recipient_count,
    COUNT(mcr.id) FILTER (WHERE mcr.status = 'sent') AS sent_count,
    COUNT(mcr.id) FILTER (WHERE mcr.status = 'failed') AS failed_count,
    COUNT(mcr.id) FILTER (WHERE mcr.status = 'pending') AS pending_count
FROM mail_campaign mc
LEFT JOIN mail_campaign_recipient mcr ON mcr.campaign_id = mc.id
WHERE mc.course_id = $1
GROUP BY mc.id
ORDER BY mc.created_at DESC;

-- name: UpdateMailCampaign :one
UPDATE mail_campaign
SET name = $3,
    subject = $4,
    body = $5,
    target_course_phase_id = $6,
    target_pass_statuses = $7,
    reply_to_override = $8,
    cc_override = $9,
    bcc_override = $10,
    updated_at = now(),
    updated_by_id = $11,
    updated_by_email = $12,
    updated_by_name = $13
WHERE id = $1 AND course_id = $2
RETURNING *;

-- name: DeleteMailCampaign :exec
DELETE FROM mail_campaign WHERE id = $1 AND course_id = $2;

-- name: TrySetMailCampaignSending :one
UPDATE mail_campaign
SET status = 'sending', updated_at = now()
WHERE id = $1 AND course_id = $2 AND status <> 'sending'
RETURNING id;

-- name: SetMailCampaignSentMeta :exec
UPDATE mail_campaign
SET status = $2, sent_at = $3, sent_by_id = $4, sent_by_email = $5, sent_by_name = $6
WHERE id = $1;

-- name: SetMailCampaignStatus :exec
UPDATE mail_campaign SET status = $2 WHERE id = $1;

-- name: GetSendingCampaignIDs :many
SELECT id FROM mail_campaign WHERE status = 'sending';

-- name: DeleteCampaignRecipients :exec
DELETE FROM mail_campaign_recipient WHERE campaign_id = $1;

-- name: InsertCampaignRecipient :exec
INSERT INTO mail_campaign_recipient (campaign_id, course_id, course_participation_id, email, status)
VALUES ($1, $2, $3, $4, 'pending')
ON CONFLICT (campaign_id, course_participation_id) DO UPDATE
SET email = EXCLUDED.email, status = 'pending', error_message = '', sent_at = NULL;

-- name: ListCampaignRecipients :many
SELECT * FROM mail_campaign_recipient WHERE campaign_id = $1 ORDER BY email ASC;

-- name: ListCampaignRecipientsWithStudent :many
SELECT
    mcr.id,
    mcr.course_participation_id,
    mcr.email,
    mcr.status,
    mcr.error_message,
    mcr.sent_at,
    s.first_name,
    s.last_name
FROM mail_campaign_recipient mcr
LEFT JOIN course_participation cp ON mcr.course_participation_id = cp.id
LEFT JOIN student s ON cp.student_id = s.id
WHERE mcr.campaign_id = $1
ORDER BY mcr.email ASC;

-- name: ListFailedCampaignRecipients :many
SELECT * FROM mail_campaign_recipient WHERE campaign_id = $1 AND status = 'failed' ORDER BY email ASC;

-- name: SetRecipientStatus :exec
UPDATE mail_campaign_recipient
SET status = $2, error_message = $3, sent_at = $4
WHERE id = $1;

-- name: FailStalePendingRecipients :exec
UPDATE mail_campaign_recipient
SET status = 'failed', error_message = $2
WHERE campaign_id = $1 AND status = 'pending';

-- name: ResetFailedRecipientsToPending :exec
UPDATE mail_campaign_recipient
SET status = 'pending', error_message = '', sent_at = NULL
WHERE campaign_id = $1 AND status = 'failed';

-- name: GetParticipantMailingInformationForCampaign :many
SELECT
    cpp.course_participation_id,
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.study_degree,
    s.current_semester,
    s.study_program
FROM
    course_phase p
JOIN course_phase_participation cpp ON p.id = cpp.course_phase_id
JOIN course_participation cp ON cpp.course_participation_id = cp.id
JOIN student s ON cp.student_id = s.id
WHERE
    p.id = $1
AND cpp.pass_status::text = ANY($2::text[]);

-- name: GetCampaignRecipientMailingInfoByIDs :many
SELECT
    cpp.course_participation_id,
    s.first_name,
    s.last_name,
    s.email,
    s.matriculation_number,
    s.university_login,
    s.study_degree,
    s.current_semester,
    s.study_program
FROM
    course_phase p
JOIN course_phase_participation cpp ON p.id = cpp.course_phase_id
JOIN course_participation cp ON cpp.course_participation_id = cp.id
JOIN student s ON cp.student_id = s.id
WHERE
    p.id = $1
AND cpp.course_participation_id = ANY($2::uuid[])
AND cpp.pass_status::text = ANY($3::text[]);
