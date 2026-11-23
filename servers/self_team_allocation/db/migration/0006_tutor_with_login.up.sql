BEGIN;

-- The tutor scope reads this view instead of the table: tutor rows stored before 0005 or added
-- manually have no login, so no editor can match them, and counting them as tutors of the phase
-- would lock every editor out of its allocations. It lists only the columns the scope reads, so
-- dropping any other tutor column does not have to recreate it.
CREATE VIEW tutor_with_login AS
SELECT course_phase_id, university_login, team_id
FROM tutor
WHERE university_login IS NOT NULL;

COMMIT;
