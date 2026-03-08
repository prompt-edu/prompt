-- Certificate test database seed
-- Creates the schema and populates test data

-- Course phase configuration for certificate settings
CREATE TABLE IF NOT EXISTS course_phase_config (
    course_phase_id uuid PRIMARY KEY,
    template_content text,
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_by text,
    release_date timestamp with time zone
);

-- Certificate download tracking per student per course phase
CREATE TABLE IF NOT EXISTS certificate_download (
    id SERIAL PRIMARY KEY,
    student_id uuid NOT NULL,
    course_phase_id uuid NOT NULL,
    first_download timestamp with time zone NOT NULL DEFAULT NOW(),
    last_download timestamp with time zone NOT NULL DEFAULT NOW(),
    download_count integer NOT NULL DEFAULT 1,
    CONSTRAINT idx_certificate_download_student_phase UNIQUE (student_id, course_phase_id)
);

CREATE INDEX IF NOT EXISTS idx_certificate_download_student_id ON certificate_download (student_id);

CREATE INDEX IF NOT EXISTS idx_certificate_download_course_phase_id ON certificate_download (course_phase_id);

-- Seed: course phase config with template
INSERT INTO
    course_phase_config (
        course_phase_id,
        template_content,
        created_at,
        updated_at,
        updated_by
    )
VALUES (
        '10000000-0000-0000-0000-000000000001',
        '#let data = json("data.json")
#set page(paper: "a4")
= Certificate of Completion
#v(2cm)
This certifies that *#data.studentName* has successfully completed the course *#data.courseName*.
#v(1cm)
Date: #data.date',
        NOW(),
        NOW(),
        'Test Admin'
    );

-- Seed: course phase config without template
INSERT INTO
    course_phase_config (
        course_phase_id,
        template_content,
        created_at,
        updated_at
    )
VALUES (
        '10000000-0000-0000-0000-000000000002',
        NULL,
        NOW(),
        NOW()
    );

-- Seed: course phase config with invalid template (for testing compilation errors)
INSERT INTO
    course_phase_config (
        course_phase_id,
        template_content,
        created_at,
        updated_at
    )
VALUES (
        '10000000-0000-0000-0000-000000000003',
        '#let data = json("nonexistent.json")
#data.studentName',
        NOW(),
        NOW()
    );

-- Seed: certificate download records
INSERT INTO
    certificate_download (
        student_id,
        course_phase_id,
        first_download,
        last_download,
        download_count
    )
VALUES (
        '30000000-0000-0000-0000-000000000001',
        '10000000-0000-0000-0000-000000000001',
        '2025-01-15T10:00:00Z',
        '2025-02-01T14:30:00Z',
        3
    );

INSERT INTO
    certificate_download (
        student_id,
        course_phase_id,
        first_download,
        last_download,
        download_count
    )
VALUES (
        '30000000-0000-0000-0000-000000000002',
        '10000000-0000-0000-0000-000000000001',
        '2025-01-20T09:00:00Z',
        '2025-01-20T09:00:00Z',
        1
    );