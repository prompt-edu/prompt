BEGIN;

-- Store the phase on each assignment so "one interview slot per participation and phase"
-- can be enforced by the database instead of a check-then-insert in the service.
ALTER TABLE interview_assignment
    ADD COLUMN course_phase_id uuid;

UPDATE interview_assignment ia
SET course_phase_id = s.course_phase_id
FROM interview_slot s
WHERE s.id = ia.interview_slot_id;

-- Concurrent bookings could leave a participation with several slots in one phase.
-- Keep the earliest booking and drop the rest so the constraint below can be added.
-- The count is reported, since the removed bookings cannot be restored.
DO $$
DECLARE
    removed_bookings integer;
BEGIN
    DELETE FROM interview_assignment
    WHERE id IN (
        SELECT id
        FROM (
            SELECT id,
                   ROW_NUMBER() OVER (
                       PARTITION BY course_participation_id, course_phase_id
                       ORDER BY assigned_at ASC NULLS LAST, id
                   ) AS booking_rank
            FROM interview_assignment
        ) ranked
        WHERE booking_rank > 1
    );
    GET DIAGNOSTICS removed_bookings = ROW_COUNT;
    RAISE NOTICE 'interview_assignment: removed % duplicate booking(s) per participation and phase', removed_bookings;
END $$;

ALTER TABLE interview_slot
    ADD CONSTRAINT interview_slot_id_course_phase_uk UNIQUE (id, course_phase_id);

ALTER TABLE interview_assignment
    ALTER COLUMN course_phase_id SET NOT NULL,
    ADD CONSTRAINT interview_assignment_slot_phase_fk
        FOREIGN KEY (interview_slot_id, course_phase_id)
        REFERENCES interview_slot (id, course_phase_id) ON DELETE CASCADE,
    ADD CONSTRAINT interview_assignment_participation_phase_uk
        UNIQUE (course_participation_id, course_phase_id);

COMMIT;
