-- Not wrapped in a transaction: a new enum value must be committed before it can be used.
ALTER TYPE resource_status ADD VALUE IF NOT EXISTS 'partial';
