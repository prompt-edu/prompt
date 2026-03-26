CREATE TABLE IF NOT EXISTS application_question_file_upload
(
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_phase_id             UUID    NOT NULL,
    title                       TEXT    NOT NULL,
    description                 TEXT,
    is_required                 BOOLEAN NOT NULL DEFAULT false,
    allowed_file_types          TEXT,
    max_file_size_mb            INTEGER,
    order_num                   INTEGER NOT NULL,
    accessible_for_other_phases BOOLEAN NOT NULL DEFAULT false,
    access_key                  TEXT,
    CONSTRAINT fk_application_question_file_upload_course_phase FOREIGN KEY (course_phase_id) REFERENCES course_phase (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_application_question_file_upload_course_phase_id ON application_question_file_upload (course_phase_id);
CREATE INDEX IF NOT EXISTS idx_application_question_file_upload_order_num ON application_question_file_upload (course_phase_id, order_num);

CREATE TABLE IF NOT EXISTS application_answer_file_upload
(
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_question_id UUID NOT NULL,
    course_participation_id UUID NOT NULL,
    file_id                 UUID NOT NULL,
    CONSTRAINT fk_application_answer_file_upload_question FOREIGN KEY (application_question_id) REFERENCES application_question_file_upload (id) ON DELETE CASCADE,
    CONSTRAINT fk_application_answer_file_upload_participation FOREIGN KEY (course_participation_id) REFERENCES course_participation (id) ON DELETE CASCADE,
    CONSTRAINT fk_application_answer_file_upload_file FOREIGN KEY (file_id) REFERENCES files (id) ON DELETE CASCADE,
    CONSTRAINT unique_file_upload_answer UNIQUE (course_participation_id, application_question_id)
);

CREATE INDEX IF NOT EXISTS idx_application_answer_file_upload_question ON application_answer_file_upload (application_question_id);
CREATE INDEX IF NOT EXISTS idx_application_answer_file_upload_participation ON application_answer_file_upload (course_participation_id);
CREATE INDEX IF NOT EXISTS idx_application_answer_file_upload_file ON application_answer_file_upload (file_id);
