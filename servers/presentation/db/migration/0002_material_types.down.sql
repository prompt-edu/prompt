BEGIN;

DROP INDEX IF EXISTS idx_material_presentation_type;

ALTER TABLE presentation_material
    DROP COLUMN IF EXISTS material_type;

ALTER TABLE course_phase_config
    DROP COLUMN IF EXISTS required_material_types;

COMMIT;
