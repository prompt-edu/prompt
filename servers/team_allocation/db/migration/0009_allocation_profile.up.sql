CREATE TABLE allocation_profile (
    course_phase_id UUID NOT NULL PRIMARY KEY,
    profile VARCHAR(64) NOT NULL DEFAULT 'standard'
);
