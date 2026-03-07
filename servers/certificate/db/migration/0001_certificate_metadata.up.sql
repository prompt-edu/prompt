BEGIN;

-- Course phase configuration for certificate settings
CREATE TABLE course_phase_config (
    course_phase_id     uuid PRIMARY KEY,
    template_content    text,  -- Typst template content stored directly
    created_at          timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at          timestamp with time zone NOT NULL DEFAULT NOW()
);

-- Certificate download tracking per student per course phase
CREATE TABLE certificate_download (
    id                  SERIAL PRIMARY KEY,
    student_id          uuid NOT NULL,
    course_phase_id     uuid NOT NULL,
    first_download      timestamp with time zone NOT NULL DEFAULT NOW(),
    last_download       timestamp with time zone NOT NULL DEFAULT NOW(),
    download_count      integer NOT NULL DEFAULT 1,
    CONSTRAINT idx_certificate_download_student_phase UNIQUE (student_id, course_phase_id)
);

-- Indexes
CREATE INDEX idx_certificate_download_student_id ON certificate_download (student_id);
CREATE INDEX idx_certificate_download_course_phase_id ON certificate_download (course_phase_id);

COMMIT;
