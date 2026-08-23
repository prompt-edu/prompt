BEGIN;

-- Which uploads a phase asks its presenters for. The catalog of valid keys lives in
-- presentation/materialTypes.go, so growing it does not need a migration.
ALTER TABLE course_phase_config
    ADD COLUMN required_material_types text[] NOT NULL DEFAULT ARRAY['slides']::text[];

-- Existing uploads predate the typed slots and are all slide decks in practice, so the
-- default backfills them into the slot every phase asks for by default.
ALTER TABLE presentation_material
    ADD COLUMN material_type text NOT NULL DEFAULT 'slides'
    CHECK (length(btrim(material_type)) > 0);

CREATE INDEX idx_material_presentation_type ON presentation_material(presentation_id, material_type);

COMMIT;
