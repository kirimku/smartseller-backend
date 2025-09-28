-- Fix google_id unique constraint to allow multiple NULL and empty string values
-- This migration addresses the issue where regular user registration fails
-- due to google_id unique constraint violation when multiple users have NULL or empty google_id

-- Drop the existing unique constraint on google_id
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_google_id_key;

-- Drop any existing indexes on google_id to start fresh
DROP INDEX IF EXISTS idx_users_google_id;
DROP INDEX IF EXISTS users_google_id_unique_idx;

-- Create a proper partial unique index that excludes both NULL and empty strings
-- This allows multiple users with NULL or empty google_id (regular registration)
-- while ensuring uniqueness for actual Google OAuth users
CREATE UNIQUE INDEX users_google_id_unique_idx 
ON users (google_id) 
WHERE google_id IS NOT NULL AND google_id != '';

-- Add a comment to document the change
COMMENT ON INDEX users_google_id_unique_idx IS 'Partial unique index on google_id excluding NULL and empty strings, allowing multiple regular users while ensuring uniqueness for OAuth users';