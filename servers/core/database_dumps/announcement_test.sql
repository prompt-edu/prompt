INSERT INTO announcement (id, severity, title, message, starts_at, expires_at, enabled) VALUES
    ('a0000000-0000-0000-0000-000000000001', 'info', 'Active info', 'Visible announcement', now() - interval '1 day', now() + interval '1 day', true),
    ('a0000000-0000-0000-0000-000000000002', 'critical', 'Active critical', 'Visible without schedule', NULL, NULL, true),
    ('a0000000-0000-0000-0000-000000000003', 'warning', 'Disabled', 'Not visible because disabled', NULL, NULL, false),
    ('a0000000-0000-0000-0000-000000000004', 'warning', 'Expired', 'Not visible because expired', now() - interval '2 days', now() - interval '1 day', true),
    ('a0000000-0000-0000-0000-000000000005', 'info', 'Scheduled', 'Not visible yet', now() + interval '1 day', NULL, true);
