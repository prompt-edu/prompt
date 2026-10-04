-- name: DeleteTeamsByCoursePhase :exec
-- Deletes every team of the course phase. The ON DELETE CASCADE constraints declared in
-- db/migration/ clean up student_team_preference_response (team_id), allocations (team_id and
-- (team_id, course_phase_id)) and tutor ((team_id, course_phase_id)) along with it. If a future
-- migration drops one of those cascades, delete that table explicitly here.
DELETE
FROM team
WHERE course_phase_id = $1;

-- name: DeleteSkillsByCoursePhase :exec
-- Deletes every skill of the course phase. student_skill_response (skill_id) is removed by its
-- ON DELETE CASCADE constraint.
DELETE
FROM skill
WHERE course_phase_id = $1;

-- name: DeleteSurveyTimeframeByCoursePhase :exec
DELETE
FROM survey_timeframe
WHERE course_phase_id = $1;

-- name: DeleteTeaseWorkspaceByCoursePhase :exec
DELETE
FROM tease_workspace
WHERE course_phase_id = $1;
