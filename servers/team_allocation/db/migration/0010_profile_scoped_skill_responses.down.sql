ALTER TABLE student_skill_response
DROP CONSTRAINT student_skill_response_pkey;

ALTER TABLE student_skill_response
ADD PRIMARY KEY (course_participation_id, skill_id);

ALTER TABLE student_skill_response
DROP COLUMN preference_mode;
