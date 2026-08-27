BEGIN;

DROP TABLE IF EXISTS mail_campaign_recipient;
DROP TABLE IF EXISTS mail_campaign;
ALTER TABLE course_participation DROP CONSTRAINT IF EXISTS unique_course_participation_id_course_id;
ALTER TABLE course_phase DROP CONSTRAINT IF EXISTS unique_course_phase_id_course_id;
DROP TYPE IF EXISTS mail_recipient_status;
DROP TYPE IF EXISTS mail_campaign_status;

COMMIT;
