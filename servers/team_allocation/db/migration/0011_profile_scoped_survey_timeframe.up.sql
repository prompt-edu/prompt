CREATE TABLE survey_timeframe_profile (
    course_phase_id UUID NOT NULL,
    profile VARCHAR(64) NOT NULL,
    survey_start TIMESTAMP NOT NULL,
    survey_deadline TIMESTAMP NOT NULL,
    PRIMARY KEY (course_phase_id, profile)
);

INSERT INTO survey_timeframe_profile (course_phase_id, profile, survey_start, survey_deadline)
SELECT course_phase_id, 'standard', survey_start, survey_deadline
FROM survey_timeframe
ON CONFLICT (course_phase_id, profile) DO NOTHING;
