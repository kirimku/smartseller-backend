-- Revert customers table multi-tenancy changes
-- Drop new tables
DROP FUNCTION IF EXISTS cleanup_expired_customer_sessions();
DROP TRIGGER IF EXISTS update_customer_sessions_last_used_at ON customer_sessions;
DROP TABLE IF EXISTS customer_sessions CASCADE;

-- Drop new indexes
DROP INDEX IF EXISTS idx_customers_storefront_id;
DROP INDEX IF EXISTS idx_customers_storefront_email;
DROP INDEX IF EXISTS idx_customers_storefront_phone;
DROP INDEX IF EXISTS idx_customers_email_verification_token;
DROP INDEX IF EXISTS idx_customers_phone_verification_token;
DROP INDEX IF EXISTS idx_customers_password_reset_token;
DROP INDEX IF EXISTS idx_customers_refresh_token;
DROP INDEX IF EXISTS idx_customers_failed_login_attempts;
DROP INDEX IF EXISTS idx_customers_locked_until;
DROP INDEX IF EXISTS idx_customers_preferences;
DROP INDEX IF EXISTS idx_customers_storefront_status;

-- Recreate old status index
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);

-- Drop new constraints
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_storefront_email_unique;
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_storefront_phone_unique;

-- Restore old unique constraint on email
ALTER TABLE customers ADD CONSTRAINT customers_email_key UNIQUE(email);

-- Remove new columns
ALTER TABLE customers DROP COLUMN IF EXISTS storefront_id;
ALTER TABLE customers DROP COLUMN IF EXISTS password_hash;
ALTER TABLE customers DROP COLUMN IF EXISTS email_verified_at;
ALTER TABLE customers DROP COLUMN IF EXISTS email_verification_token;
ALTER TABLE customers DROP COLUMN IF EXISTS phone_verified_at;
ALTER TABLE customers DROP COLUMN IF EXISTS phone_verification_token;
ALTER TABLE customers DROP COLUMN IF EXISTS password_reset_token;
ALTER TABLE customers DROP COLUMN IF EXISTS password_reset_expires_at;
ALTER TABLE customers DROP COLUMN IF EXISTS refresh_token;
ALTER TABLE customers DROP COLUMN IF EXISTS refresh_token_expires_at;
ALTER TABLE customers DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE customers DROP COLUMN IF EXISTS failed_login_attempts;
ALTER TABLE customers DROP COLUMN IF EXISTS locked_until;
ALTER TABLE customers DROP COLUMN IF EXISTS preferences;