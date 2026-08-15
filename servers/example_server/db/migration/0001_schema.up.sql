BEGIN;

CREATE TABLE
    template_table (
        course_phase_id uuid PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );

COMMIT;