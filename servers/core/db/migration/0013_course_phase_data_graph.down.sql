BEGIN;

-- Drop the phase-variant tables created by the up.
DROP TABLE IF EXISTS phase_data_dependency_graph;
DROP TABLE IF EXISTS course_phase_type_phase_required_input_dto;
DROP TABLE IF EXISTS course_phase_type_phase_provided_output_dto;

-- Rename the participation-variant tables back to their 0011 names.
ALTER TABLE participation_data_dependency_graph
    RENAME TO meta_data_dependency_graph;
ALTER TABLE course_phase_type_participation_required_input_dto
    RENAME TO course_phase_type_required_input_dto;
ALTER TABLE course_phase_type_participation_provided_output_dto
    RENAME TO course_phase_type_provided_output_dto;

COMMIT;
