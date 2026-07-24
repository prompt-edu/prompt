-- name: GetStudentSkillResponseByCourseParticipationIDs :many
SELECT sr.*, s.*
FROM student_skill_response sr, skill s
WHERE s.id = sr.skill_id
AND sr.course_participation_id = ANY($1::uuid[]);

-- name: GetAllocationByCourseParticipationIDs :many
SELECT a.*, t.*
FROM allocations a, team t
WHERE a.course_participation_id = ANY($1::uuid[])
AND a.team_id = t.id;

-- name: GetStudentTeamPreferenceResponseByCourseParticipationIDs :many
SELECT spr.*, t.*
FROM student_team_preference_response spr, team t
WHERE course_participation_id = ANY($1::uuid[])
AND spr.team_id = t.id;

-- name: GetTutorByCourseParticipationIDs :many
SELECT t.*
FROM tutor t
WHERE t.course_participation_id = ANY($1::uuid[]);

-- name: DeleteAllocationsByCourseParticipationIDs :exec
DELETE FROM allocations
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteStudentTeamPreferenceResponseByCourseParticipationIDs :exec
DELETE FROM student_team_preference_response
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteStudentSkillResponseByCourseParticipationIDs :exec
DELETE FROM student_skill_response
WHERE course_participation_id = ANY($1::uuid[]);

-- name: DeleteTutorByCourseParticipationIDs :exec
DELETE FROM tutor
WHERE course_participation_id = ANY($1::uuid[]);
