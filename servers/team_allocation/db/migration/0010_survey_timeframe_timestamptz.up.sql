BEGIN;

-- Existing values are UTC wall clocks: the browser serializes Date as RFC3339 with a Z suffix and
-- the deployed service containers run in UTC, so pgx stored the instant unshifted.
ALTER TABLE survey_timeframe
    ALTER COLUMN survey_start TYPE TIMESTAMPTZ USING survey_start AT TIME ZONE 'UTC',
    ALTER COLUMN survey_deadline TYPE TIMESTAMPTZ USING survey_deadline AT TIME ZONE 'UTC';

COMMIT;
