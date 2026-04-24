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

-- name: GetStudentPreferencesByTeamType :many
SELECT team_id, preference
FROM student_team_preference_response
JOIN team ON team.id = student_team_preference_response.team_id
WHERE course_participation_id = $1
AND team.course_phase_id = $2
AND team.team_type = $3;

-- Returns the student’s skill responses.
-- name: GetStudentSkillResponses :many
SELECT skill_id, skill_level
FROM student_skill_response
JOIN skill ON skill.id = student_skill_response.skill_id
WHERE course_participation_id = $1
AND skill.course_phase_id = $2;

-- name: GetStudentSkillResponsesByPreferenceMode :many
SELECT skill_id, skill_level
FROM student_skill_response
JOIN skill ON skill.id = student_skill_response.skill_id
WHERE course_participation_id = $1
AND skill.course_phase_id = $2
AND preference_mode = $3;

-- Deletes all student team preference responses (for overwriting answers).
-- name: DeleteStudentTeamPreferences :exec
DELETE FROM student_team_preference_response
WHERE course_participation_id = $1
AND team_id IN (
    SELECT id
    FROM team
    WHERE course_phase_id = $2
);

-- name: DeleteStudentPreferencesByTeamType :exec
DELETE FROM student_team_preference_response
WHERE course_participation_id = $1
AND team_id IN (
    SELECT id
    FROM team
    WHERE course_phase_id = $2
    AND team_type = $3
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

-- name: DeleteStudentSkillResponsesByPreferenceMode :exec
DELETE FROM student_skill_response
WHERE course_participation_id = $1
AND preference_mode = $3
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
INSERT INTO student_skill_response (course_participation_id, skill_id, skill_level, preference_mode)
VALUES ($1, $2, $3, $4);

-- Upsert the survey timeframe for a given course phase.
-- name: SetSurveyTimeframe :exec
INSERT INTO survey_timeframe (course_phase_id, survey_start, survey_deadline)
VALUES ($1, $2, $3)
ON CONFLICT (course_phase_id)
DO UPDATE SET survey_start = EXCLUDED.survey_start,
              survey_deadline = EXCLUDED.survey_deadline;
