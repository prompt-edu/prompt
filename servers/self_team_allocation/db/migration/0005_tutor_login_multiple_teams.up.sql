BEGIN;

-- A tutor is identified by their university login and can lead several teams of a course phase
-- (a project lead with two teams, for example): one row per (tutor, team).
ALTER TABLE tutor ADD COLUMN university_login TEXT;

ALTER TABLE tutor DROP CONSTRAINT tutor_pkey;
ALTER TABLE tutor ADD PRIMARY KEY (course_phase_id, course_participation_id, team_id);

CREATE UNIQUE INDEX idx_tutor_phase_login_team
    ON tutor (course_phase_id, university_login, team_id)
    WHERE university_login IS NOT NULL;

COMMIT;
