ALTER TABLE course_phase_type_participation_required_input_dto
ADD COLUMN optional boolean NOT NULL DEFAULT false;

ALTER TABLE course_phase_type_phase_required_input_dto
ADD COLUMN optional boolean NOT NULL DEFAULT false;
