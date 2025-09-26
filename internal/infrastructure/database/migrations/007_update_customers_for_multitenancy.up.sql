-- Update customers table for multi-tenant storefront support
-- Add storefront_id and authentication fields

-- Add new columns to existing customers table
ALTER TABLE customers ADD COLUMN IF NOT EXISTS storefront_id UUID REFERENCES storefronts(id) ON DELETE CASCADE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS email_verification_token VARCHAR(255);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS phone_verified_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS phone_verification_token VARCHAR(10);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS password_reset_token VARCHAR(255);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS password_reset_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS refresh_token VARCHAR(500);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS refresh_token_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;
ALTER TABLE customers ADD COLUMN IF NOT EXISTS preferences JSONB DEFAULT '{}';

-- Update constraints
-- Remove old unique constraint on email and add composite unique constraint with storefront_id
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_email_key;
ALTER TABLE customers ADD CONSTRAINT customers_storefront_email_unique 
    UNIQUE(storefront_id, email) DEFERRABLE INITIALLY DEFERRED;

-- Add composite unique constraint for phone within storefront
ALTER TABLE customers ADD CONSTRAINT customers_storefront_phone_unique 
    UNIQUE(storefront_id, phone) DEFERRABLE INITIALLY DEFERRED;

-- Update the contact info check constraint
ALTER TABLE customers DROP CONSTRAINT IF EXISTS customers_contact_info;
ALTER TABLE customers ADD CONSTRAINT customers_contact_info 
    CHECK (email IS NOT NULL OR phone IS NOT NULL);

-- Add constraint to ensure storefront_id is set for new records (we'll make it required)
-- For existing records, we'll need a data migration
ALTER TABLE customers ALTER COLUMN storefront_id SET NOT NULL;

-- New indexes for multi-tenant queries
CREATE INDEX IF NOT EXISTS idx_customers_storefront_id ON customers(storefront_id);
CREATE INDEX IF NOT EXISTS idx_customers_storefront_email ON customers(storefront_id, email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_storefront_phone ON customers(storefront_id, phone) WHERE phone IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_email_verification_token ON customers(email_verification_token) WHERE email_verification_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_phone_verification_token ON customers(phone_verification_token) WHERE phone_verification_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_password_reset_token ON customers(password_reset_token) WHERE password_reset_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_refresh_token ON customers(refresh_token) WHERE refresh_token IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_failed_login_attempts ON customers(failed_login_attempts) WHERE failed_login_attempts > 0;
CREATE INDEX IF NOT EXISTS idx_customers_locked_until ON customers(locked_until) WHERE locked_until IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_preferences ON customers USING gin(preferences);

-- Update existing indexes to include storefront_id for better performance
DROP INDEX IF EXISTS idx_customers_status;
CREATE INDEX idx_customers_storefront_status ON customers(storefront_id, status);

-- Create customer sessions table for authentication
CREATE TABLE IF NOT EXISTS customer_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    storefront_id UUID NOT NULL REFERENCES storefronts(id) ON DELETE CASCADE,
    
    -- Session Information
    session_token VARCHAR(500) NOT NULL UNIQUE,
    refresh_token VARCHAR(500) NOT NULL UNIQUE,
    user_agent TEXT,
    ip_address INET,
    
    -- Session Timing
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for customer_sessions
CREATE INDEX IF NOT EXISTS idx_customer_sessions_customer_id ON customer_sessions(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_sessions_storefront_id ON customer_sessions(storefront_id);
CREATE INDEX IF NOT EXISTS idx_customer_sessions_session_token ON customer_sessions(session_token) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_sessions_refresh_token ON customer_sessions(refresh_token) WHERE revoked_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_customer_sessions_expires_at ON customer_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_customer_sessions_last_used ON customer_sessions(last_used_at);

-- Trigger for customer_sessions
CREATE TRIGGER update_customer_sessions_last_used_at 
    BEFORE UPDATE ON customer_sessions
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to clean up expired sessions
CREATE OR REPLACE FUNCTION cleanup_expired_customer_sessions()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM customer_sessions 
    WHERE expires_at < CURRENT_TIMESTAMP 
    AND revoked_at IS NULL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;