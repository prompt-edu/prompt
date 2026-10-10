SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', 'public', false);

CREATE TYPE public.announcement_severity AS ENUM ('info', 'warning', 'critical');

CREATE TABLE public.announcement (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    severity public.announcement_severity NOT NULL,
    title text NOT NULL DEFAULT '',
    message text NOT NULL,
    link_url text NOT NULL DEFAULT '',
    link_label text NOT NULL DEFAULT '',
    starts_at timestamptz,
    expires_at timestamptz,
    enabled boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT check_expires_at_after_starts_at CHECK (expires_at > starts_at)
);

INSERT INTO public.announcement (id, severity, title, message, starts_at, expires_at, enabled) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'info', 'Active info', 'Visible announcement', now() - interval '1 day', now() + interval '1 day', true),
    ('a0000000-0000-0000-0000-000000000002', 'critical', 'Active critical', 'Visible without schedule', NULL, NULL, true),
    ('a0000000-0000-0000-0000-000000000003', 'warning', 'Disabled', 'Not visible because disabled', NULL, NULL, false),
    ('a0000000-0000-0000-0000-000000000004', 'warning', 'Expired', 'Not visible because expired', now() - interval '2 days', now() - interval '1 day', true),
    ('a0000000-0000-0000-0000-000000000005', 'info', 'Scheduled', 'Not visible yet', now() + interval '1 day', NULL, true);
