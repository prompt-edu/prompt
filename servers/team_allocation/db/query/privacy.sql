
-- name: GetStudentSkillResponseByCourseParticipationID :many
SELECT sr.*, s.* 
FROM student_skill_response sr, skill s 
WHERE s.id = sr.skill_id 
AND sr.course_participation_id = ANY($1::uuid[]);

-- name: GetAllocationByCourseParticipationID :many
SELECT a.*, t.*
FROM allocations a, team t
WHERE a.course_participation_id = ANY($1::uuid[])
AND a.team_id = t.id;

-- name: GetStudentTeamPreferenceResponseByCourseParticipationID :many
SELECT spr.*, t.*
FROM student_team_preference_response spr, team t
WHERE course_participation_id = ANY($1::uuid[])
AND spr.team_id = t.id;

