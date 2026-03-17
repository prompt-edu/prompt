-- Add application_question_file_upload table for file upload questions
CREATE TABLE IF NOT EXISTS application_question_file_upload (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id UUID NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_required BOOLEAN NOT NULL DEFAULT false,
    allowed_file_types TEXT,
    max_file_size_mb INTEGER,
    order_num INTEGER NOT NULL,
    accessible_for_other_phases BOOLEAN NOT NULL DEFAULT false,
    access_key TEXT,
    CONSTRAINT fk_application_question_file_upload_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE
);

CREATE INDEX idx_application_question_file_upload_course_phase_id ON application_question_file_upload(course_phase_id);
CREATE INDEX idx_application_question_file_upload_order_num ON application_question_file_upload(course_phase_id, order_num);

-- Add sample file upload question for testing
INSERT INTO application_question_file_upload (id, course_phase_id, title, description, is_required, allowed_file_types, max_file_size_mb, order_num, accessible_for_other_phases, access_key)
VALUES 
    ('b1b04042-95d1-4765-8592-caf9560c8c3d', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Resume Upload', 'Please upload your resume', true, '.pdf,.doc,.docx', 10, 3, false, null),
    ('c2c04042-95d1-4765-8592-caf9560c8c3e', '4179d58a-d00d-4fa7-94a5-397bc69fab02', 'Portfolio', 'Upload your portfolio (optional)', false, '.pdf,.zip', 20, 4, false, null);
