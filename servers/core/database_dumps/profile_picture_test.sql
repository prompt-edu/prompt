-- Minimal schema for the profile picture tests: students, their course participations, the
-- files table, and profile_picture (matching migration 0030).

CREATE TYPE gender AS ENUM ('male', 'female', 'diverse', 'prefer_not_to_say');

CREATE TABLE student (
    id uuid PRIMARY KEY,
    first_name character varying(50),
    last_name character varying(50),
    email character varying(255) UNIQUE,
    matriculation_number character varying(30),
    university_login character varying(20),
    has_university_account boolean,
    gender gender NOT NULL
);

CREATE TABLE course_participation (
    id uuid PRIMARY KEY,
    course_id uuid NOT NULL,
    student_id uuid NOT NULL REFERENCES student (id) ON DELETE CASCADE
);

CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename VARCHAR(500) NOT NULL,
    original_filename VARCHAR(500) NOT NULL,
    content_type VARCHAR(200) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    storage_key VARCHAR(500) NOT NULL UNIQUE,
    storage_provider VARCHAR(50) NOT NULL DEFAULT 'seaweedfs',
    uploaded_by_user_id VARCHAR(200) NOT NULL,
    uploaded_by_email VARCHAR(200),
    course_phase_id UUID,
    description TEXT,
    tags VARCHAR(100)[],
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE TABLE profile_picture (
    user_id UUID PRIMARY KEY,
    university_login VARCHAR(20),
    file_id UUID NOT NULL UNIQUE REFERENCES files (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Ada has a picture (uploaded with her university login), Grace has none, Alan has no login.
INSERT INTO student (id, first_name, last_name, email, matriculation_number, university_login, has_university_account, gender) VALUES
('aaaaaaaa-0000-0000-0000-000000000001', 'Ada', 'Lovelace', 'ada@tum.de', '03700001', 'ab12cde', true, 'female'),
('aaaaaaaa-0000-0000-0000-000000000002', 'Grace', 'Hopper', 'grace@tum.de', '03700002', 'gh34ijk', true, 'female'),
('aaaaaaaa-0000-0000-0000-000000000003', 'Alan', 'Turing', 'alan@example.com', NULL, NULL, false, 'male');

INSERT INTO course_participation (id, course_id, student_id) VALUES
('cccccccc-0000-0000-0000-000000000001', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000001'),
('cccccccc-0000-0000-0000-000000000002', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000002'),
('cccccccc-0000-0000-0000-000000000003', 'dddddddd-0000-0000-0000-000000000001', 'aaaaaaaa-0000-0000-0000-000000000003');

-- Ada's account and an instructor account without a university login.
INSERT INTO files (id, filename, original_filename, content_type, size_bytes, storage_key, uploaded_by_user_id) VALUES
('ffffffff-0000-0000-0000-000000000001', 'ada.jpg', 'profile-picture.jpg', 'image/jpeg', 2048,
 'profile-picture/bbbbbbbb-0000-0000-0000-000000000001/ada.jpg', 'bbbbbbbb-0000-0000-0000-000000000001'),
('ffffffff-0000-0000-0000-000000000002', 'instructor.jpg', 'profile-picture.jpg', 'image/jpeg', 2048,
 'profile-picture/bbbbbbbb-0000-0000-0000-000000000002/instructor.jpg', 'bbbbbbbb-0000-0000-0000-000000000002');

INSERT INTO profile_picture (user_id, university_login, file_id) VALUES
('bbbbbbbb-0000-0000-0000-000000000001', 'ab12cde', 'ffffffff-0000-0000-0000-000000000001'),
('bbbbbbbb-0000-0000-0000-000000000002', NULL, 'ffffffff-0000-0000-0000-000000000002');
