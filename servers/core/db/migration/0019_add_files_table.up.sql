CREATE TABLE IF NOT EXISTS files
(
    id                  UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    filename            VARCHAR(500) NOT NULL,
    original_filename   VARCHAR(500) NOT NULL,
    content_type        VARCHAR(200) NOT NULL,
    size_bytes          BIGINT       NOT NULL CHECK (size_bytes >= 0),
    storage_key         VARCHAR(500) NOT NULL UNIQUE,
    storage_provider    VARCHAR(50)  NOT NULL DEFAULT 'seaweedfs',
    uploaded_by_user_id VARCHAR(200) NOT NULL,
    uploaded_by_email   VARCHAR(200),
    course_phase_id     UUID,
    description         TEXT,
    tags                VARCHAR(100)[],
    created_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMP,
    CONSTRAINT fk_files_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase (id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_files_storage_key ON files (storage_key);
CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files (uploaded_by_user_id);
CREATE INDEX IF NOT EXISTS idx_files_course_phase_id ON files (course_phase_id) WHERE course_phase_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files (deleted_at) WHERE deleted_at IS NULL;

CREATE OR REPLACE FUNCTION update_files_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_files_updated_at
    BEFORE UPDATE
    ON files
    FOR EACH ROW
EXECUTE FUNCTION update_files_updated_at();
