BEGIN;

DROP TRIGGER IF EXISTS trg_audit_log_no_update ON audit_log;
DROP FUNCTION IF EXISTS audit_log_no_update();
DROP TABLE IF EXISTS audit_log;

COMMIT;
