-- Returns the survey timeframe (survey_start and survey_deadline) for a course phase.
-- name: GetSurveyTimeframe :one
SELECT survey_start, survey_deadline
FROM survey_timeframe
WHERE course_phase_id = $1;

-- Returns the student’s team preference responses.
-- name: GetStudentTeamPreferences :many
SELECT team_id, preference
FROM student_team_preference_response
JOIN team ON team.id = student_team_preference_response.team_id
WHERE course_participation_id = $1
AND team.course_phase_id = $2;

-- Returns the student’s skill responses.
-- name: GetStudentSkillResponses :many
SELECT skill_id, skill_level
FROM student_skill_response
JOIN skill ON skill.id = student_skill_response.skill_id
WHERE course_participation_id = $1
AND skill.course_phase_id = $2;

-- Deletes all student team preference responses (for overwriting answers).
-- name: DeleteStudentTeamPreferences :exec
DELETE FROM student_team_preference_response
WHERE course_participation_id = $1
AND team_id IN (
    SELECT id
    FROM team
    WHERE course_phase_id = $2
);

-- Deletes all student skill responses (for overwriting answers).
-- name: DeleteStudentSkillResponses :exec
DELETE FROM student_skill_response
WHERE course_participation_id = $1
AND skill_id IN (
    SELECT id
    FROM skill
    WHERE course_phase_id = $2
);

-- Inserts a new student team preference.
-- name: InsertStudentTeamPreference :exec
INSERT INTO student_team_preference_response (course_participation_id, team_id, preference)
VALUES ($1, $2, $3);

-- Inserts a new student skill response.
-- name: InsertStudentSkillResponse :exec
INSERT INTO student_skill_response (course_participation_id, skill_id, skill_level)
VALUES ($1, $2, $3);

-- Returns team popularity as average preference rank per team (lower = more popular).
-- name: GetTeamPopularityStatistics :many
SELECT
    t.id AS team_id,
    t.name AS team_name,
    AVG(r.preference)::float8 AS avg_preference,
    COUNT(r.course_participation_id) AS response_count
FROM student_team_preference_response r
JOIN team t ON t.id = r.team_id
WHERE t.course_phase_id = $1
GROUP BY t.id, t.name
ORDER BY avg_preference ASC;

-- Returns per-rank student counts per team for tooltip breakdown.
-- name: GetTeamPreferenceCounts :many
SELECT
    t.id AS team_id,
    r.preference,
    COUNT(r.course_participation_id)::int AS count
FROM student_team_preference_response r
JOIN team t ON t.id = r.team_id
WHERE t.course_phase_id = $1
GROUP BY t.id, r.preference
ORDER BY t.id, r.preference;

-- Returns skill distribution statistics aggregated by skill and proficiency level.
-- name: GetSkillDistributionStatistics :many
SELECT
    s.id AS skill_id,
    s.name AS skill_name,
    r.skill_level,
    COUNT(r.course_participation_id) AS count
FROM student_skill_response r
JOIN skill s ON s.id = r.skill_id
WHERE s.course_phase_id = $1
GROUP BY s.id, s.name, r.skill_level
ORDER BY s.name, r.skill_level;

-- Upsert the survey timeframe for a given course phase.
-- name: SetSurveyTimeframe :exec
INSERT INTO survey_timeframe (course_phase_id, survey_start, survey_deadline)
VALUES ($1, $2, $3)
ON CONFLICT (course_phase_id)
DO UPDATE SET survey_start = EXCLUDED.survey_start,
              survey_deadline = EXCLUDED.survey_deadline;