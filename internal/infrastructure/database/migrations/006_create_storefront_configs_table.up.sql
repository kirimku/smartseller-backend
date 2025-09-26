-- Create storefront_configs table for additional configuration
CREATE TABLE IF NOT EXISTS storefront_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    storefront_id UUID NOT NULL REFERENCES storefronts(id) ON DELETE CASCADE,
    
    -- Configuration key-value pairs
    config_key VARCHAR(100) NOT NULL,
    config_value JSONB,
    data_type VARCHAR(20) NOT NULL DEFAULT 'json' CHECK (data_type IN ('string', 'number', 'boolean', 'json', 'array')),
    
    -- Metadata
    description TEXT,
    is_sensitive BOOLEAN DEFAULT FALSE, -- For encryption later
    is_required BOOLEAN DEFAULT FALSE,
    default_value JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint to prevent duplicate keys per storefront
    UNIQUE(storefront_id, config_key)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_storefront_configs_storefront_id ON storefront_configs(storefront_id);
CREATE INDEX IF NOT EXISTS idx_storefront_configs_key ON storefront_configs(config_key);
CREATE INDEX IF NOT EXISTS idx_storefront_configs_value ON storefront_configs USING gin(config_value);
CREATE INDEX IF NOT EXISTS idx_storefront_configs_is_sensitive ON storefront_configs(is_sensitive);

-- Update trigger
CREATE TRIGGER update_storefront_configs_updated_at 
    BEFORE UPDATE ON storefront_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();