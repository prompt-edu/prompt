BEGIN;

ALTER TABLE survey_timeframe
    ALTER COLUMN survey_start TYPE TIMESTAMP USING survey_start AT TIME ZONE 'UTC',
    ALTER COLUMN survey_deadline TYPE TIMESTAMP USING survey_deadline AT TIME ZONE 'UTC';

COMMIT;
