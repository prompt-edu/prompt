BEGIN;

-- Back to one team per tutor: keep one assignment per tutor and drop the others.
DELETE
FROM tutor t
    USING tutor d
WHERE t.course_phase_id = d.course_phase_id
  AND t.course_participation_id = d.course_participation_id
  AND t.ctid > d.ctid;

DELETE
FROM tutor t
    USING tutor d
WHERE t.course_phase_id = d.course_phase_id
  AND t.university_login = d.university_login
  AND t.ctid > d.ctid;

DROP INDEX IF EXISTS idx_tutor_phase_login_team;
CREATE UNIQUE INDEX idx_tutor_phase_login
    ON tutor (course_phase_id, university_login)
    WHERE university_login IS NOT NULL;

ALTER TABLE tutor DROP CONSTRAINT tutor_pkey;
ALTER TABLE tutor ADD PRIMARY KEY (course_phase_id, course_participation_id);

COMMIT;
