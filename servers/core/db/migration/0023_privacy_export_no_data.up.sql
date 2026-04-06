-- no_data status means an export succeeded but there was no data
-- useful for the export docs
ALTER TYPE export_status ADD VALUE IF NOT EXISTS 'no_data';
-- archived status means the corresponding s3 files have been deleted
-- and the actual export recors are only kept for archiving purposes
ALTER TYPE export_status ADD VALUE IF NOT EXISTS 'archived';
