BEGIN;

-- Best-effort reverse. Data note: student_readable_data was derived (icon / bg-color) and
-- is dropped rather than merged back into the renamed column.
ALTER TABLE course_phase_participation DROP COLUMN IF EXISTS student_readable_data;
ALTER TABLE course_phase_participation RENAME COLUMN restricted_data TO meta_data;

ALTER TABLE course_phase DROP COLUMN IF EXISTS student_readable_data;
ALTER TABLE course_phase RENAME COLUMN restricted_data TO meta_data;

ALTER TABLE course DROP COLUMN IF EXISTS student_readable_data;
ALTER TABLE course RENAME COLUMN restricted_data TO meta_data;

COMMIT;
