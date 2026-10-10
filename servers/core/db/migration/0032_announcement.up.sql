CREATE TYPE announcement_severity AS ENUM ('info', 'warning', 'critical');

CREATE TABLE announcement (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  severity   announcement_severity NOT NULL,
  title      text NOT NULL DEFAULT '',
  message    text NOT NULL,
  link_url   text NOT NULL DEFAULT '',
  link_label text NOT NULL DEFAULT '',
  starts_at  timestamptz,
  expires_at timestamptz,
  enabled    boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT check_expires_at_after_starts_at CHECK (expires_at > starts_at)
);
