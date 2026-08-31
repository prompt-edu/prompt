BEGIN;

-- Existing values are UTC wall clocks: the browser serializes Date as RFC3339 with a Z suffix and
-- the deployed service containers run in UTC, so pgx stored the instant unshifted.
ALTER TABLE timeframe
    ALTER COLUMN starttime TYPE TIMESTAMPTZ USING starttime AT TIME ZONE 'UTC',
    ALTER COLUMN endtime TYPE TIMESTAMPTZ USING endtime AT TIME ZONE 'UTC';

COMMIT;
