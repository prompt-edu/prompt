BEGIN;

DROP TABLE IF EXISTS mail_campaign_recipient;
DROP TABLE IF EXISTS mail_campaign;
DROP TYPE IF EXISTS mail_recipient_status;
DROP TYPE IF EXISTS mail_campaign_status;

COMMIT;
