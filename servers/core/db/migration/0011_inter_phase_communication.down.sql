BEGIN;

-- Best-effort reverse. Data note: metadata migrated into the DTO tables and the
-- interviewScore -> score JSON rewrite are not restored beyond the schema shape.

-- Restore the course_phase_type columns (base_url added, the two meta_data columns dropped
-- by the up). initial_phase is owned by 0003, not this migration.
ALTER TABLE course_phase_type
    ADD COLUMN required_input_meta_data jsonb DEFAULT '[]' NOT NULL,
    ADD COLUMN provided_output_meta_data jsonb DEFAULT '[]' NOT NULL,
    DROP COLUMN IF EXISTS base_url;

-- Drop the 0011 dependency graph (it references the DTO tables) and recreate the
-- 0009-version two-column graph.
DROP TABLE IF EXISTS meta_data_dependency_graph;
CREATE TABLE meta_data_dependency_graph (
    from_phase_id uuid NOT NULL,
    to_phase_id   uuid NOT NULL,
    PRIMARY KEY (from_phase_id, to_phase_id),
    CONSTRAINT fk_from_phase
      FOREIGN KEY (from_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE,
    CONSTRAINT fk_to_phase
      FOREIGN KEY (to_phase_id) REFERENCES course_phase(id) ON DELETE CASCADE
);

DROP TABLE IF EXISTS course_phase_type_required_input_dto;
DROP TABLE IF EXISTS course_phase_type_provided_output_dto;

COMMIT;
