-- Revert google_id unique constraint fix
-- This migration reverts the changes made to allow multiple NULL and empty google_id values

-- Drop the partial unique index
DROP INDEX IF EXISTS users_google_id_unique_idx;

-- Recreate the original unique constraint
-- Note: This may fail if there are multiple users with NULL or empty google_id values
-- In that case, you would need to clean up the data first
ALTER TABLE users ADD CONSTRAINT users_google_id_key UNIQUE (google_id);