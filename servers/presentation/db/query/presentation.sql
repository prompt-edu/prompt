-- name: EnsureCoursePhaseConfig :one
INSERT INTO course_phase_config (course_phase_id)
VALUES ($1)
ON CONFLICT (course_phase_id) DO UPDATE SET course_phase_id = EXCLUDED.course_phase_id
RETURNING *;

-- name: GetCoursePhaseConfig :one
SELECT * FROM course_phase_config WHERE course_phase_id = $1;

-- name: UpdateCoursePhaseConfig :one
UPDATE course_phase_config
SET target_mode = $2, feedback_mode = $3, updated_at = now()
WHERE course_phase_id = $1
RETURNING *;

-- name: CountPresentationsByPhase :one
SELECT count(*) FROM presentation WHERE course_phase_id = $1;

-- name: CountFeedbackFormsByPhase :one
SELECT count(*) FROM feedback_form f
JOIN presentation p ON p.id = f.presentation_id
WHERE p.course_phase_id = $1;

-- name: DeletePresentationsByPhase :exec
DELETE FROM presentation WHERE course_phase_id = $1;

-- name: DeletePresentationSlotsByPhase :exec
DELETE FROM presentation_slot WHERE course_phase_id = $1;

-- name: DeleteCoursePhaseConfig :exec
DELETE FROM course_phase_config WHERE course_phase_id = $1;

-- name: ListFeedbackCategories :many
SELECT * FROM feedback_category WHERE course_phase_id = $1 ORDER BY position, id;

-- name: GetFeedbackCategory :one
SELECT * FROM feedback_category WHERE id = $1 AND course_phase_id = $2;

-- name: CreateFeedbackCategory :one
INSERT INTO feedback_category (course_phase_id, name, description, position)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateFeedbackCategory :one
UPDATE feedback_category
SET name = $3, description = $4, position = $5, updated_at = now()
WHERE id = $1 AND course_phase_id = $2
RETURNING *;

-- name: DeleteFeedbackCategory :execrows
DELETE FROM feedback_category WHERE id = $1 AND course_phase_id = $2;

-- name: DeleteFeedbackCategoriesByPhase :exec
DELETE FROM feedback_category WHERE course_phase_id = $1;

-- name: DeleteFeedbackFormsByPhase :exec
DELETE FROM feedback_form
WHERE presentation_id IN (SELECT id FROM presentation WHERE course_phase_id = $1);

-- name: ClearFeedbackReleasesByPhase :exec
UPDATE presentation
SET feedback_release_name = NULL,
    feedback_released_at = NULL,
    feedback_released_by_user_id = NULL,
    feedback_released_by_name = NULL,
    updated_at = now()
WHERE course_phase_id = $1;

-- name: CreatePresentationSlot :one
INSERT INTO presentation_slot (course_phase_id, start_time, end_time, location)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPresentationSlots :many
SELECT * FROM presentation_slot WHERE course_phase_id = $1 ORDER BY start_time, id;

-- name: GetPresentationSlot :one
SELECT * FROM presentation_slot WHERE id = $1 AND course_phase_id = $2;

-- name: GetPresentationSlotForUpdate :one
SELECT * FROM presentation_slot WHERE id = $1 AND course_phase_id = $2 FOR UPDATE;

-- name: UpdatePresentationSlot :one
UPDATE presentation_slot
SET start_time = $3, end_time = $4, location = $5, updated_at = now()
WHERE id = $1 AND course_phase_id = $2
RETURNING *;

-- name: DeletePresentationSlot :execrows
DELETE FROM presentation_slot s
WHERE s.id = $1 AND s.course_phase_id = $2
  AND NOT EXISTS (SELECT 1 FROM presentation p WHERE p.slot_id = s.id);

-- name: CreatePresentation :one
INSERT INTO presentation (course_phase_id, slot_id, target_type, target_id, target_name)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPresentation :one
SELECT * FROM presentation WHERE id = $1 AND course_phase_id = $2;

-- name: GetPresentationForUpdate :one
SELECT * FROM presentation WHERE id = $1 AND course_phase_id = $2 FOR UPDATE;

-- name: GetPresentationByTarget :one
SELECT * FROM presentation WHERE course_phase_id = $1 AND target_type = $2 AND target_id = $3;

-- name: GetPresentationBySlot :one
SELECT * FROM presentation WHERE course_phase_id = $1 AND slot_id = $2;

-- name: ListPresentations :many
SELECT p.*, s.start_time, s.end_time, s.location
FROM presentation p
JOIN presentation_slot s ON s.id = p.slot_id
WHERE p.course_phase_id = $1
ORDER BY s.start_time, p.target_name;

-- name: CountPresentationDependencies :one
SELECT
  (SELECT count(*) FROM presentation_material m WHERE m.presentation_id = $1) AS material_count,
  (SELECT count(*) FROM feedback_form f WHERE f.presentation_id = $1) AS feedback_count;

-- name: CountSubmittedFeedbackForms :one
SELECT count(*) FROM feedback_form
WHERE presentation_id = $1 AND status = 'submitted';

-- name: DeletePresentation :execrows
DELETE FROM presentation WHERE id = $1 AND course_phase_id = $2;

-- name: SetFeedbackRelease :one
UPDATE presentation
SET feedback_release_name = $3, feedback_released_at = now(), feedback_released_by_user_id = $4,
    feedback_released_by_name = $5, updated_at = now()
WHERE id = $1 AND course_phase_id = $2
RETURNING *;

-- name: ClearFeedbackRelease :one
UPDATE presentation
SET feedback_release_name = NULL, feedback_released_at = NULL, feedback_released_by_user_id = NULL,
    feedback_released_by_name = NULL, updated_at = now()
WHERE id = $1 AND course_phase_id = $2
RETURNING *;

-- name: CreatePendingMaterial :one
INSERT INTO presentation_material (
  presentation_id, original_filename, content_type, storage_key,
  uploader_user_id, uploader_name, uploader_email, expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetPresentationMaterial :one
SELECT m.* FROM presentation_material m
JOIN presentation p ON p.id = m.presentation_id
WHERE m.id = $1 AND p.course_phase_id = $2;

-- name: CompletePresentationMaterial :one
UPDATE presentation_material
SET size_bytes = $3, content_type = $4, state = 'ready', expires_at = NULL, updated_at = now()
WHERE id = $1 AND presentation_id = $2 AND state = 'pending' AND expires_at > now()
RETURNING *;

-- name: ListPresentationMaterials :many
SELECT * FROM presentation_material
WHERE presentation_id = $1 AND state = 'ready'
ORDER BY created_at, id;

-- name: DeletePresentationMaterial :execrows
DELETE FROM presentation_material WHERE id = $1 AND presentation_id = $2;

-- name: ListExpiredPendingMaterials :many
SELECT * FROM presentation_material
WHERE state = 'pending' AND expires_at <= now()
ORDER BY expires_at
LIMIT $1;

-- name: GetFeedbackFormByScope :one
SELECT * FROM feedback_form WHERE presentation_id = $1 AND scope_key = $2;

-- name: GetFeedbackForm :one
SELECT f.* FROM feedback_form f
JOIN presentation p ON p.id = f.presentation_id
WHERE f.id = $1 AND p.course_phase_id = $2;

-- name: CreateFeedbackForm :one
INSERT INTO feedback_form (
  presentation_id, scope_key, evaluator_user_id, evaluator_name, evaluator_email
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (presentation_id, scope_key) DO UPDATE
SET presentation_id = EXCLUDED.presentation_id
RETURNING *;

-- name: ListFeedbackForms :many
SELECT * FROM feedback_form WHERE presentation_id = $1 ORDER BY created_at, id;

-- name: ListSubmittedFeedbackForms :many
SELECT * FROM feedback_form WHERE presentation_id = $1 AND status = 'submitted' ORDER BY submitted_at, id;

-- name: SetFeedbackFormStatus :one
UPDATE feedback_form
SET status = $3,
    submitted_at = CASE WHEN $3 = 'submitted' THEN now() ELSE NULL END,
    updated_at = now()
WHERE id = $1 AND presentation_id = $2
RETURNING *;

-- name: DeleteDraftFeedbackForm :execrows
DELETE FROM feedback_form
WHERE id = $1 AND presentation_id = $2 AND status = 'draft';

-- name: CountFeedbackFormsByStatus :one
SELECT count(*) FROM feedback_form WHERE presentation_id = $1 AND status = $2;

-- name: CountDraftFeedbackForms :one
SELECT count(*) FROM feedback_form WHERE presentation_id = $1 AND status = 'draft';

-- name: ListFeedbackAnswers :many
SELECT * FROM feedback_answer WHERE feedback_form_id = $1 ORDER BY category_id;

-- name: PutFeedbackAnswer :one
INSERT INTO feedback_answer (
  feedback_form_id, category_id, value, revision, updated_by_user_id, updated_by_name
)
SELECT sqlc.arg(feedback_form_id), sqlc.arg(category_id), sqlc.arg(value), 1,
       sqlc.arg(updated_by_user_id), sqlc.arg(updated_by_name)
WHERE sqlc.arg(expected_revision)::bigint = 0
ON CONFLICT (feedback_form_id, category_id) DO UPDATE
SET value = EXCLUDED.value,
    revision = feedback_answer.revision + 1,
    updated_by_user_id = EXCLUDED.updated_by_user_id,
    updated_by_name = EXCLUDED.updated_by_name,
    updated_at = now()
WHERE feedback_answer.revision = sqlc.arg(expected_revision)::bigint
RETURNING *;

-- name: GetFeedbackAnswer :one
SELECT * FROM feedback_answer WHERE feedback_form_id = $1 AND category_id = $2;

-- name: UpsertFeedbackContributor :one
INSERT INTO feedback_contributor (feedback_form_id, user_id, name, email)
VALUES ($1, $2, $3, $4)
ON CONFLICT (feedback_form_id, user_id) DO UPDATE
SET name = EXCLUDED.name, email = EXCLUDED.email, last_contributed_at = now()
RETURNING *;

-- name: ListFeedbackContributors :many
SELECT * FROM feedback_contributor WHERE feedback_form_id = $1 ORDER BY first_contributed_at;

-- name: ResetPresentationFeedback :exec
DELETE FROM feedback_form WHERE presentation_id = $1;

-- name: NotifyFeedbackEvent :exec
SELECT pg_notify('presentation_feedback', $1);

-- name: UpsertFeedbackPresence :one
INSERT INTO feedback_presence (presentation_id, connection_id, user_id, name, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (presentation_id, connection_id) DO UPDATE
SET expires_at = EXCLUDED.expires_at, name = EXCLUDED.name
RETURNING *;

-- name: DeleteFeedbackPresence :exec
DELETE FROM feedback_presence WHERE presentation_id = $1 AND connection_id = $2;

-- name: DeleteExpiredFeedbackPresence :exec
DELETE FROM feedback_presence WHERE expires_at <= now();

-- name: ListFeedbackPresence :many
SELECT * FROM feedback_presence WHERE presentation_id = $1 AND expires_at > now() ORDER BY name, connection_id;

-- name: GetPrivacyMaterials :many
SELECT m.* FROM presentation_material m
JOIN presentation p ON p.id = m.presentation_id
WHERE m.uploader_user_id = sqlc.arg(uploader_user_id)
   OR (p.target_type = 'individual' AND p.target_id = ANY(sqlc.arg(course_participation_ids)::uuid[]));

-- name: ListMaterialStorageKeysByPhase :many
SELECT m.storage_key
FROM presentation_material m
JOIN presentation p ON p.id = m.presentation_id
WHERE p.course_phase_id = $1;

-- name: GetPrivacyFeedbackForms :many
SELECT * FROM feedback_form WHERE evaluator_user_id = $1;

-- name: DeletePrivacyDrafts :exec
DELETE FROM feedback_form WHERE evaluator_user_id = $1 AND status = 'draft';

-- name: AnonymizePrivacyFeedback :exec
UPDATE feedback_form SET evaluator_user_id = NULL, evaluator_name = 'Deleted user', evaluator_email = ''
WHERE evaluator_user_id = $1;

-- name: AnonymizePrivacyContributions :exec
UPDATE feedback_contributor SET user_id = 'deleted:' || md5(user_id), name = 'Deleted user', email = ''
WHERE user_id = $1;

-- name: AnonymizePrivacyAnswers :exec
UPDATE feedback_answer SET updated_by_user_id = 'deleted:' || md5(updated_by_user_id), updated_by_name = 'Deleted user'
WHERE updated_by_user_id = $1;

-- name: AnonymizePrivacyMaterials :exec
UPDATE presentation_material SET uploader_user_id = 'deleted:' || md5(uploader_user_id), uploader_email = ''
WHERE uploader_user_id = $1;

-- name: AnonymizePrivacyPresentationTargets :exec
UPDATE presentation
SET target_name = 'Deleted student'
WHERE target_type = 'individual' AND target_id = ANY($1::uuid[]);
