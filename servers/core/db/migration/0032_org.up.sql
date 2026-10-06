BEGIN;

-- An org is an organizational unit such as a university chair. Orgs nest through
-- parent_org_id. The slug is embedded in the org's Keycloak group and role names
-- (/Orgs/<slug>, org-<slug>-Admin), so it is globally unique and never changes.
-- A slug must not contain a "cg" segment: a course in semester "org" mints custom
-- group roles "org-<course>-cg-<name>", which could otherwise equal an org role.
CREATE TABLE org (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  parent_org_id uuid,
  name          text        NOT NULL,
  slug          text        NOT NULL,
  school        text,
  university    text,
  website       text,
  contact_email text,
  created_at    timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT fk_parent_org
    FOREIGN KEY (parent_org_id) REFERENCES org (id) ON DELETE RESTRICT,
  CONSTRAINT check_org_not_own_parent CHECK (parent_org_id <> id),
  CONSTRAINT unique_org_slug UNIQUE (slug),
  CONSTRAINT check_org_slug_format CHECK (
    char_length(slug) BETWEEN 2 AND 50
    AND slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'
    AND position('-cg-' IN '-' || slug || '-') = 0
  )
);

CREATE INDEX idx_org_parent_org_id ON org (parent_org_id);

-- Courses without an org keep working unchanged. An org cannot be deleted while
-- courses are assigned to it.
ALTER TABLE course
  ADD COLUMN org_id uuid,
  ADD CONSTRAINT fk_course_org
    FOREIGN KEY (org_id) REFERENCES org (id) ON DELETE RESTRICT;

CREATE INDEX idx_course_org_id ON course (org_id);

COMMIT;
