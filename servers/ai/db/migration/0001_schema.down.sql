BEGIN;

DROP TABLE IF EXISTS ai_call_event;
DROP TABLE IF EXISTS ai_call_content_restriction;
DROP TABLE IF EXISTS ai_call_subject;
DROP TABLE IF EXISTS ai_call_content;
DROP TABLE IF EXISTS ai_call;

DROP FUNCTION IF EXISTS ai_audit_no_update();
DROP FUNCTION IF EXISTS ai_call_content_response_only();
DROP FUNCTION IF EXISTS ai_call_complete_only();

COMMIT;
