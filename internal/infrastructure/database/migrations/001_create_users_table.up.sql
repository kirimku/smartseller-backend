-- Create users table for SmartSeller
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- OAuth Integration
    google_id VARCHAR(255) UNIQUE,
    
    -- Basic User Information
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20) UNIQUE,
    picture TEXT,
    
    -- Authentication
    password_hash TEXT,
    password_salt TEXT,
    
    -- Password Reset
    password_reset_token TEXT,
    password_reset_expires TIMESTAMP WITH TIME ZONE,
    
    -- User Classification for SmartSeller
    user_type VARCHAR(20) NOT NULL DEFAULT 'individual' CHECK (user_type IN ('individual', 'business', 'enterprise')),
    user_tier VARCHAR(20) NOT NULL DEFAULT 'basic' CHECK (user_tier IN ('basic', 'premium', 'pro', 'enterprise')),
    
    -- Authorization
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    user_role VARCHAR(20) NOT NULL DEFAULT 'user' CHECK (user_role IN ('owner', 'admin', 'manager', 'support', 'user')),
    
    -- OAuth Tokens
    access_token TEXT,
    refresh_token TEXT,
    token_expiry TIMESTAMP WITH TIME ZONE,
    
    -- Terms and Marketing
    accept_terms BOOLEAN NOT NULL DEFAULT FALSE,
    accept_promos BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT users_contact_info CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id) WHERE google_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_user_tier ON users(user_tier);
CREATE INDEX IF NOT EXISTS idx_users_user_role ON users(user_role);
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
