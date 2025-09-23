-- Create products table for SmartSeller e-commerce
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Product Identification
    sku VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    
    -- Product Organization
    category_id UUID,
    brand VARCHAR(255),
    tags TEXT[], -- Array of tags for flexible categorization
    
    -- Pricing
    base_price DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    sale_price DECIMAL(15,2),
    cost_price DECIMAL(15,2),
    
    -- Inventory
    track_inventory BOOLEAN NOT NULL DEFAULT TRUE,
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    low_stock_threshold INTEGER DEFAULT 10,
    
    -- Product Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'inactive', 'archived')),
    
    -- SEO and Marketing
    meta_title VARCHAR(255),
    meta_description TEXT,
    slug VARCHAR(255) UNIQUE,
    
    -- Product Attributes
    weight DECIMAL(10,3), -- in kg
    dimensions_length DECIMAL(10,2), -- in cm
    dimensions_width DECIMAL(10,2),  -- in cm
    dimensions_height DECIMAL(10,2), -- in cm
    
    -- Ownership
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Product Categories
CREATE TABLE IF NOT EXISTS product_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    slug VARCHAR(255) UNIQUE NOT NULL,
    parent_id UUID REFERENCES product_categories(id),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Images
CREATE TABLE IF NOT EXISTS product_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    alt_text VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Product Variant Options (define what options are available for a product)
CREATE TABLE IF NOT EXISTS product_variant_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Option Definition
    option_name VARCHAR(100) NOT NULL, -- e.g., "Color", "Size", "Material", "Style"
    option_values TEXT[] NOT NULL, -- e.g., ["Red", "Blue", "Green"] or ["S", "M", "L", "XL"]
    
    -- Display Properties
    display_name VARCHAR(255), -- user-friendly name
    sort_order INTEGER DEFAULT 0,
    is_required BOOLEAN DEFAULT true,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one option name per product
    UNIQUE(product_id, option_name)
);

-- Product Variants (specific combinations of options)
CREATE TABLE IF NOT EXISTS product_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Variant Identification
    sku VARCHAR(100) UNIQUE NOT NULL,
    variant_name VARCHAR(255), -- auto-generated or manual: "Red - Large - Cotton"
    
    -- Dynamic Variant Options (JSON object with option_name: value pairs)
    -- Example: {"Color": "Red", "Size": "Large", "Material": "Cotton"}
    variant_options JSONB NOT NULL,
    
    -- Pricing Override (if NULL, use product's base_price)
    price DECIMAL(15,2), -- Variant-specific price (overrides product base_price)
    cost_price DECIMAL(15,2), -- Variant-specific cost price
    
    -- Sale Price for Variants
    sale_price DECIMAL(15,2), -- Variant-specific sale price
    
    -- Inventory Override
    stock_quantity INTEGER NOT NULL DEFAULT 0,
    
    -- Physical Properties Override
    weight DECIMAL(10,3),
    dimensions_length DECIMAL(10,2),
    dimensions_width DECIMAL(10,2),
    dimensions_height DECIMAL(10,2),
    
    -- Media Override
    image_url TEXT, -- specific image for this variant
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Unique constraint: one variant per option combination per product
    UNIQUE(product_id, variant_options)
);

-- Add foreign key for category
ALTER TABLE products ADD CONSTRAINT fk_products_category 
    FOREIGN KEY (category_id) REFERENCES product_categories(id);

-- Indexes for products
CREATE INDEX IF NOT EXISTS idx_products_sku ON products(sku);
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_created_by ON products(created_by);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);
CREATE INDEX IF NOT EXISTS idx_products_tags ON products USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
CREATE INDEX IF NOT EXISTS idx_products_deleted_at ON products(deleted_at) WHERE deleted_at IS NOT NULL;

-- Indexes for product_categories
CREATE INDEX IF NOT EXISTS idx_product_categories_slug ON product_categories(slug);
CREATE INDEX IF NOT EXISTS idx_product_categories_parent_id ON product_categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_product_categories_is_active ON product_categories(is_active);

-- Indexes for product_images
CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);
CREATE INDEX IF NOT EXISTS idx_product_images_is_primary ON product_images(is_primary);

-- Indexes for product_variant_options
CREATE INDEX IF NOT EXISTS idx_product_variant_options_product_id ON product_variant_options(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variant_options_option_name ON product_variant_options(option_name);
CREATE INDEX IF NOT EXISTS idx_product_variant_options_sort_order ON product_variant_options(sort_order);

-- Indexes for product_variants
CREATE INDEX IF NOT EXISTS idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX IF NOT EXISTS idx_product_variants_sku ON product_variants(sku);
CREATE INDEX IF NOT EXISTS idx_product_variants_is_active ON product_variants(is_active);
CREATE INDEX IF NOT EXISTS idx_product_variants_options ON product_variants USING gin(variant_options);

-- Triggers for updated_at
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_categories_updated_at 
    BEFORE UPDATE ON product_categories 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_variant_options_updated_at 
    BEFORE UPDATE ON product_variant_options 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_product_variants_updated_at 
    BEFORE UPDATE ON product_variants 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Function to auto-generate variant name from options
CREATE OR REPLACE FUNCTION generate_variant_name(variant_options JSONB) RETURNS TEXT AS $$
DECLARE
    key TEXT;
    value TEXT;
    result TEXT := '';
    first BOOLEAN := TRUE;
BEGIN
    -- Sort keys to ensure consistent naming
    FOR key IN SELECT jsonb_object_keys(variant_options) ORDER BY jsonb_object_keys(variant_options) LOOP
        value := variant_options ->> key;
        
        IF NOT first THEN
            result := result || ' - ';
        END IF;
        
        result := result || value;
        first := FALSE;
    END LOOP;
    
    RETURN result;
END;
$$ LANGUAGE plpgsql;

-- Auto-generate variant name trigger
CREATE OR REPLACE FUNCTION set_variant_name() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.variant_name IS NULL OR NEW.variant_name = '' THEN
        NEW.variant_name := generate_variant_name(NEW.variant_options);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_variant_name_trigger
    BEFORE INSERT OR UPDATE ON product_variants
    FOR EACH ROW
    EXECUTE FUNCTION set_variant_name();

-- Function to validate variant options against product variant options
CREATE OR REPLACE FUNCTION validate_variant_options() RETURNS TRIGGER AS $$
DECLARE
    option_record RECORD;
    provided_value TEXT;
BEGIN
    -- Check that all required options are provided
    FOR option_record IN 
        SELECT option_name, option_values, is_required 
        FROM product_variant_options 
        WHERE product_id = NEW.product_id
    LOOP
        provided_value := NEW.variant_options ->> option_record.option_name;
        
        -- Check if required option is missing
        IF option_record.is_required AND (provided_value IS NULL OR provided_value = '') THEN
            RAISE EXCEPTION 'Required variant option "%" is missing', option_record.option_name;
        END IF;
        
        -- Check if provided value is in allowed values
        IF provided_value IS NOT NULL AND NOT (provided_value = ANY(option_record.option_values)) THEN
            RAISE EXCEPTION 'Invalid value "%" for option "%". Allowed values: %', 
                provided_value, option_record.option_name, array_to_string(option_record.option_values, ', ');
        END IF;
    END LOOP;
    
    -- Check that no extra options are provided
    FOR provided_value IN SELECT jsonb_object_keys(NEW.variant_options) LOOP
        IF NOT EXISTS (
            SELECT 1 FROM product_variant_options 
            WHERE product_id = NEW.product_id AND option_name = provided_value
        ) THEN
            RAISE EXCEPTION 'Unknown variant option "%"', provided_value;
        END IF;
    END LOOP;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER validate_variant_options_trigger
    BEFORE INSERT OR UPDATE ON product_variants
    FOR EACH ROW
    EXECUTE FUNCTION validate_variant_options();

-- Function to get effective price for product/variant (helper for queries)
CREATE OR REPLACE FUNCTION get_effective_price(
    product_base_price DECIMAL(15,2),
    product_sale_price DECIMAL(15,2),
    variant_price DECIMAL(15,2),
    variant_sale_price DECIMAL(15,2)
) RETURNS DECIMAL(15,2) AS $$
BEGIN
    -- Priority: variant_sale_price > variant_price > product_sale_price > product_base_price
    IF variant_sale_price IS NOT NULL THEN
        RETURN variant_sale_price;
    ELSIF variant_price IS NOT NULL THEN
        RETURN variant_price;
    ELSIF product_sale_price IS NOT NULL THEN
        RETURN product_sale_price;
    ELSE
        RETURN product_base_price;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- View for easy querying of products with effective pricing
CREATE OR REPLACE VIEW product_variants_with_pricing AS
SELECT 
    pv.*,
    p.name as product_name,
    p.base_price as product_base_price,
    p.sale_price as product_sale_price,
    get_effective_price(p.base_price, p.sale_price, pv.price, pv.sale_price) as effective_price,
    CASE 
        WHEN pv.sale_price IS NOT NULL THEN 'variant_sale'
        WHEN pv.price IS NOT NULL THEN 'variant_regular'
        WHEN p.sale_price IS NOT NULL THEN 'product_sale'
        ELSE 'product_regular'
    END as price_source
FROM product_variants pv
JOIN products p ON pv.product_id = p.id
WHERE pv.deleted_at IS NULL AND p.deleted_at IS NULL;
