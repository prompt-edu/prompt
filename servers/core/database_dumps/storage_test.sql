-- Storage test database dump
-- This includes schema and test data for storage module tests

-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    keycloak_user_id VARCHAR(200) UNIQUE NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    university_login VARCHAR(100),
    matriculation_number VARCHAR(50),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Courses table (minimal for FK constraints)
CREATE TABLE IF NOT EXISTS courses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Course iterations table (minimal for FK constraints)
CREATE TABLE IF NOT EXISTS course_iterations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    semester_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Applications table
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    course_iteration_id UUID NOT NULL REFERENCES course_iterations(id) ON DELETE CASCADE,
    assessment_score_threshold NUMERIC(5, 2),
    accepted_at TIMESTAMP,
    rejected_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Course phase table (minimal for FK constraints)
CREATE TABLE IF NOT EXISTS course_phase (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_iteration_id UUID NOT NULL REFERENCES course_iterations(id) ON DELETE CASCADE,
    phase_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Files table
CREATE TABLE IF NOT EXISTS files (
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
    deleted_at TIMESTAMP,
    CONSTRAINT fk_files_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_files_storage_key ON files(storage_key);
CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by_user_id);
CREATE INDEX IF NOT EXISTS idx_files_course_phase_id ON files(course_phase_id) WHERE course_phase_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files(deleted_at) WHERE deleted_at IS NULL;

-- Test data
INSERT INTO users (user_id, keycloak_user_id, email, university_login, matriculation_number, first_name, last_name) VALUES
('11111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'test.user@tum.de', 'testuser', '12345678', 'Test', 'User'),
('99999999-9999-9999-9999-999999999999', '99999999-9999-9999-9999-999999999999', 'admin@tum.de', 'admin', '99999999', 'Admin', 'User');

INSERT INTO courses (id, course_name) VALUES
('33333333-3333-3333-3333-333333333333', 'Test Course');

INSERT INTO course_iterations (id, course_id, semester_name) VALUES
('44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333', 'WS2024');

INSERT INTO applications (id, student_id, course_iteration_id) VALUES
('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', '44444444-4444-4444-4444-444444444444');

INSERT INTO course_phase (id, course_iteration_id, phase_name) VALUES
('55555555-5555-5555-5555-555555555555', '44444444-4444-4444-4444-444444444444', 'Test Phase');
