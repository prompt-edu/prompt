BEGIN;

-- A tutor can lead several teams of a course phase (a project lead with two teams, for example).
-- One row per (tutor, team): the identity of a tutor is their university login, and the import
-- keeps one course_participation_id per login.
ALTER TABLE tutor DROP CONSTRAINT tutor_pkey;
ALTER TABLE tutor ADD PRIMARY KEY (course_phase_id, course_participation_id, team_id);

DROP INDEX IF EXISTS idx_tutor_phase_login;
CREATE UNIQUE INDEX idx_tutor_phase_login_team
    ON tutor (course_phase_id, university_login, team_id)
    WHERE university_login IS NOT NULL;

COMMIT;
