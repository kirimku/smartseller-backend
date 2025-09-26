-- Create storefronts table for multi-tenant customer management
CREATE TABLE IF NOT EXISTS storefronts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationship to seller (users table)
    seller_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Storefront Identity
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE, -- URL-friendly identifier
    description TEXT,
    
    -- Domain Configuration
    domain VARCHAR(255), -- Custom domain (optional)
    subdomain VARCHAR(100) UNIQUE, -- Subdomain on our platform
    
    -- Storefront Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended')),
    
    -- Configuration (JSONB for flexible settings)
    settings JSONB DEFAULT '{}',
    
    -- SEO and Branding
    logo_url VARCHAR(500),
    favicon_url VARCHAR(500),
    primary_color VARCHAR(7), -- Hex color
    secondary_color VARCHAR(7),
    
    -- Business Information
    business_name VARCHAR(255),
    business_email VARCHAR(255),
    business_phone VARCHAR(20),
    business_address TEXT,
    tax_id VARCHAR(50),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_storefronts_seller_id ON storefronts(seller_id);
CREATE INDEX IF NOT EXISTS idx_storefronts_slug ON storefronts(slug) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_storefronts_domain ON storefronts(domain) WHERE domain IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_storefronts_subdomain ON storefronts(subdomain) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_storefronts_status ON storefronts(status);
CREATE INDEX IF NOT EXISTS idx_storefronts_settings ON storefronts USING gin(settings);
CREATE INDEX IF NOT EXISTS idx_storefronts_created_at ON storefronts(created_at);
CREATE INDEX IF NOT EXISTS idx_storefronts_deleted_at ON storefronts(deleted_at) WHERE deleted_at IS NOT NULL;

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_storefronts_updated_at 
    BEFORE UPDATE ON storefronts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to validate slug format
CREATE OR REPLACE FUNCTION validate_slug(slug_value TEXT) 
RETURNS BOOLEAN AS $$
BEGIN
    -- Slug should be lowercase, alphanumeric with hyphens, 3-50 chars
    RETURN slug_value ~ '^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$' OR LENGTH(slug_value) = 1;
END;
$$ LANGUAGE plpgsql;

-- Add constraint for slug validation
ALTER TABLE storefronts ADD CONSTRAINT chk_storefronts_slug_format 
    CHECK (validate_slug(slug));