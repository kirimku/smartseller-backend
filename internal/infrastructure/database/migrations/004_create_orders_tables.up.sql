-- Create orders and related tables for SmartSeller
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Order Identification
    order_number VARCHAR(50) UNIQUE NOT NULL,
    
    -- Customer Information
    customer_id UUID REFERENCES customers(id),
    customer_email VARCHAR(255),
    customer_phone VARCHAR(20),
    
    -- Order Status
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending', 'confirmed', 'processing', 'shipped', 'delivered', 
        'cancelled', 'refunded', 'returned'
    )),
    
    -- Financial Information
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    shipping_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    discount_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    
    -- Currency
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    
    -- Payment Information
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (payment_status IN (
        'pending', 'paid', 'partially_paid', 'refunded', 'failed'
    )),
    payment_method VARCHAR(50),
    payment_reference VARCHAR(255),
    
    -- Shipping Information
    shipping_method VARCHAR(100),
    shipping_tracking_number VARCHAR(255),
    shipping_carrier VARCHAR(100),
    
    -- Billing Address
    billing_first_name VARCHAR(255),
    billing_last_name VARCHAR(255),
    billing_company VARCHAR(255),
    billing_address_line_1 VARCHAR(500),
    billing_address_line_2 VARCHAR(500),
    billing_city VARCHAR(255),
    billing_state_province VARCHAR(255),
    billing_postal_code VARCHAR(20),
    billing_country VARCHAR(100),
    billing_phone VARCHAR(20),
    
    -- Shipping Address
    shipping_first_name VARCHAR(255),
    shipping_last_name VARCHAR(255),
    shipping_company VARCHAR(255),
    shipping_address_line_1 VARCHAR(500),
    shipping_address_line_2 VARCHAR(500),
    shipping_city VARCHAR(255),
    shipping_state_province VARCHAR(255),
    shipping_postal_code VARCHAR(20),
    shipping_country VARCHAR(100),
    shipping_phone VARCHAR(20),
    
    -- Order Metadata
    notes TEXT,
    internal_notes TEXT,
    tags TEXT[],
    
    -- Channel Information (where the order came from)
    channel VARCHAR(50) DEFAULT 'direct' CHECK (channel IN (
        'direct', 'marketplace', 'social', 'mobile_app', 'pos', 'api'
    )),
    channel_reference VARCHAR(255), -- External order ID from marketplace
    
    -- Important Dates
    order_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    confirmed_at TIMESTAMP WITH TIME ZONE,
    shipped_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,
    
    -- Ownership
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Order Items
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Product Information
    product_id UUID REFERENCES products(id),
    product_variant_id UUID REFERENCES product_variants(id),
    
    -- Item Details (snapshot at time of order)
    product_name VARCHAR(500) NOT NULL,
    product_sku VARCHAR(100),
    variant_name VARCHAR(255),
    
    -- Pricing
    unit_price DECIMAL(15,2) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    total_price DECIMAL(15,2) NOT NULL,
    
    -- Product Metadata at Time of Order
    product_weight DECIMAL(10,3),
    product_image_url TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Order Status History
CREATE TABLE IF NOT EXISTS order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    
    -- Status Change
    from_status VARCHAR(20),
    to_status VARCHAR(20) NOT NULL,
    
    -- Change Information
    reason VARCHAR(500),
    notes TEXT,
    changed_by UUID REFERENCES users(id),
    
    -- Timestamp
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for orders
CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number);
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_customer_email ON orders(customer_email);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_payment_status ON orders(payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_channel ON orders(channel);
CREATE INDEX IF NOT EXISTS idx_orders_created_by ON orders(created_by);
CREATE INDEX IF NOT EXISTS idx_orders_order_date ON orders(order_date);
CREATE INDEX IF NOT EXISTS idx_orders_total_amount ON orders(total_amount);
CREATE INDEX IF NOT EXISTS idx_orders_tags ON orders USING gin(tags);
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at ON orders(deleted_at) WHERE deleted_at IS NOT NULL;

-- Indexes for order_items
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_id ON order_items(product_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_variant_id ON order_items(product_variant_id);
CREATE INDEX IF NOT EXISTS idx_order_items_product_sku ON order_items(product_sku);

-- Indexes for order_status_history
CREATE INDEX IF NOT EXISTS idx_order_status_history_order_id ON order_status_history(order_id);
CREATE INDEX IF NOT EXISTS idx_order_status_history_to_status ON order_status_history(to_status);
CREATE INDEX IF NOT EXISTS idx_order_status_history_changed_at ON order_status_history(changed_at);

