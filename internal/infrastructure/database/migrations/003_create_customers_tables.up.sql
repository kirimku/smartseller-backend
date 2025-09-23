-- Create customers table for SmartSeller CRM
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Customer Identification
    email VARCHAR(255) UNIQUE,
    phone VARCHAR(20),
    
    -- Personal Information
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    full_name VARCHAR(500), -- Computed or manually set
    date_of_birth DATE,
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other', 'prefer_not_to_say')),
    
    -- Customer Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'blocked')),
    
    -- Customer Segmentation
    customer_type VARCHAR(20) DEFAULT 'regular' CHECK (customer_type IN ('regular', 'vip', 'wholesale')),
    tags TEXT[], -- Array of tags for segmentation
    
    -- Marketing Preferences
    accepts_marketing BOOLEAN DEFAULT FALSE,
    marketing_opt_in_date TIMESTAMP WITH TIME ZONE,
    
    -- Customer Metrics
    total_orders INTEGER DEFAULT 0,
    total_spent DECIMAL(15,2) DEFAULT 0.00,
    average_order_value DECIMAL(15,2) DEFAULT 0.00,
    last_order_date TIMESTAMP WITH TIME ZONE,
    
    -- Notes and Internal Info
    notes TEXT,
    internal_notes TEXT, -- Private notes not visible to customer
    
    -- Ownership (which seller this customer belongs to)
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Constraints
    CONSTRAINT customers_contact_info CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

-- Customer Addresses
CREATE TABLE IF NOT EXISTS customer_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    
    -- Address Type
    address_type VARCHAR(20) NOT NULL DEFAULT 'shipping' CHECK (address_type IN ('billing', 'shipping', 'both')),
    
    -- Address Information
    label VARCHAR(100), -- e.g., "Home", "Office"
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    company VARCHAR(255),
    address_line_1 VARCHAR(500) NOT NULL,
    address_line_2 VARCHAR(500),
    city VARCHAR(255) NOT NULL,
    state_province VARCHAR(255),
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    phone VARCHAR(20),
    
    -- Address Status
    is_default BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Customer Groups (for segmentation and bulk operations)
CREATE TABLE IF NOT EXISTS customer_groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    color VARCHAR(7), -- Hex color code for UI
    
    -- Group Rules (JSON for flexible criteria)
    criteria JSONB, -- Store segmentation rules
    
    -- Ownership
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Customer Group Memberships
CREATE TABLE IF NOT EXISTS customer_group_memberships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES customer_groups(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    added_by UUID REFERENCES users(id),
    
    -- Prevent duplicate memberships
    UNIQUE(customer_id, group_id)
);

-- Indexes for customers
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email) WHERE email IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone) WHERE phone IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers(status);
CREATE INDEX IF NOT EXISTS idx_customers_customer_type ON customers(customer_type);
CREATE INDEX IF NOT EXISTS idx_customers_created_by ON customers(created_by);
CREATE INDEX IF NOT EXISTS idx_customers_tags ON customers USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_customers_total_spent ON customers(total_spent);
CREATE INDEX IF NOT EXISTS idx_customers_total_orders ON customers(total_orders);
CREATE INDEX IF NOT EXISTS idx_customers_last_order_date ON customers(last_order_date);
CREATE INDEX IF NOT EXISTS idx_customers_created_at ON customers(created_at);
CREATE INDEX IF NOT EXISTS idx_customers_deleted_at ON customers(deleted_at) WHERE deleted_at IS NOT NULL;

-- Indexes for customer_addresses
CREATE INDEX IF NOT EXISTS idx_customer_addresses_customer_id ON customer_addresses(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_type ON customer_addresses(address_type);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_is_default ON customer_addresses(is_default);
CREATE INDEX IF NOT EXISTS idx_customer_addresses_country ON customer_addresses(country);

-- Indexes for customer_groups
CREATE INDEX IF NOT EXISTS idx_customer_groups_created_by ON customer_groups(created_by);
CREATE INDEX IF NOT EXISTS idx_customer_groups_criteria ON customer_groups USING gin(criteria);

-- Indexes for customer_group_memberships
CREATE INDEX IF NOT EXISTS idx_customer_group_memberships_customer_id ON customer_group_memberships(customer_id);
CREATE INDEX IF NOT EXISTS idx_customer_group_memberships_group_id ON customer_group_memberships(group_id);
