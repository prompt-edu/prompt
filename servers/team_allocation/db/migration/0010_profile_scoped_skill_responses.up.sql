ALTER TABLE student_skill_response
ADD COLUMN preference_mode VARCHAR(16) NOT NULL DEFAULT 'teams';

ALTER TABLE student_skill_response
DROP CONSTRAINT student_skill_response_pkey;

ALTER TABLE student_skill_response
ADD PRIMARY KEY (course_participation_id, skill_id, preference_mode);
