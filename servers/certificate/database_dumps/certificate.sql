-- Certificate test data, loaded on top of the migrated schema

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
