-- SmartSeller Warranty System Database Schema
-- Based on Rexus Gaming warranty requirements adapted for multi-tenant SmartSeller platform

-- Warranty Barcodes/QR Codes Table
-- Stores unique warranty identifiers with secure generation tracking
CREATE TABLE IF NOT EXISTS warranty_barcodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Barcode Identification (REX[YY][RANDOM_12] format)
    barcode_number VARCHAR(17) UNIQUE NOT NULL,
    qr_code_data TEXT NOT NULL, -- URL for warranty claims
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    
    -- Multi-tenant support
    storefront_id UUID NOT NULL REFERENCES storefronts(id) ON DELETE CASCADE,
    
    -- Generation metadata for security tracking
    generated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    generation_method VARCHAR(20) NOT NULL DEFAULT 'CSPRNG',
    entropy_bits INTEGER NOT NULL DEFAULT 60,
    generation_attempt INTEGER DEFAULT 1,
    collision_checked BOOLEAN DEFAULT TRUE,
    
    -- Distribution tracking
    batch_id UUID,
    batch_number VARCHAR(100),
    distributed_at TIMESTAMP WITH TIME ZONE,
    distributed_to VARCHAR(255), -- Retailer/channel info
    distribution_notes TEXT,
    
    -- Activation tracking
    activated_at TIMESTAMP WITH TIME ZONE,
    customer_id UUID REFERENCES customers(id),
    purchase_date DATE,
    purchase_location VARCHAR(255),
    purchase_invoice VARCHAR(255),
    
    -- Status management
    status VARCHAR(20) NOT NULL DEFAULT 'generated' 
        CHECK (status IN ('generated', 'distributed', 'activated', 'used', 'expired')),
    
    -- Warranty period
    warranty_period_months INTEGER NOT NULL DEFAULT 12,
    expiry_date DATE,
    
    -- Audit fields
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT barcode_format_check CHECK (
        barcode_number ~ '^REX\d{2}[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{12}$'
    )
);

-- Warranty Claims Table
CREATE TABLE IF NOT EXISTS warranty_claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Claim identification
    claim_number VARCHAR(20) UNIQUE NOT NULL, -- WC-YYYY-MM-###
    barcode_id UUID NOT NULL REFERENCES warranty_barcodes(id),
    customer_id UUID NOT NULL REFERENCES customers(id),
    product_id UUID NOT NULL REFERENCES products(id),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    
    -- Issue details
    issue_description TEXT NOT NULL,
    issue_category VARCHAR(100) NOT NULL,
    issue_date DATE NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'medium'
        CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    
    -- Claim dates and timeline
    claim_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    validated_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Status management with comprehensive tracking
    status VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN (
            'pending', 'validated', 'rejected', 'assigned', 'in_repair', 
            'repaired', 'replaced', 'shipped', 'delivered', 'completed', 
            'cancelled', 'disputed'
        )),
    previous_status VARCHAR(20),
    status_updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status_updated_by UUID REFERENCES users(id),
    
    -- Processing assignment
    validated_by UUID REFERENCES users(id),
    assigned_technician_id UUID REFERENCES users(id),
    estimated_completion_date DATE,
    actual_completion_date DATE,
    
    -- Resolution details
    resolution_type VARCHAR(20)
        CHECK (resolution_type IN ('repair', 'replace', 'refund', 'rejected')),
    repair_notes TEXT,
    replacement_product_id UUID REFERENCES products(id),
    refund_amount DECIMAL(15,2),
    
    -- Cost tracking
    repair_cost DECIMAL(15,2) DEFAULT 0.00,
    shipping_cost DECIMAL(15,2) DEFAULT 0.00,
    replacement_cost DECIMAL(15,2) DEFAULT 0.00,
    total_cost DECIMAL(15,2) GENERATED ALWAYS AS (
        COALESCE(repair_cost, 0) + COALESCE(shipping_cost, 0) + COALESCE(replacement_cost, 0)
    ) STORED,
    
    -- Customer information (snapshot at claim time)
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(50),
    
    -- Pickup/delivery address (JSONB for flexibility)
    pickup_address JSONB NOT NULL,
    
    -- Logistics tracking
    shipping_provider VARCHAR(50),
    tracking_number VARCHAR(100),
    estimated_delivery_date DATE,
    actual_delivery_date DATE,
    delivery_status VARCHAR(20) DEFAULT 'not_shipped'
        CHECK (delivery_status IN (
            'not_shipped', 'preparing', 'picked_up', 'in_transit', 
            'out_for_delivery', 'delivered', 'failed_delivery', 'returned'
        )),
    
    -- Communication and notes
    customer_notes TEXT,
    admin_notes TEXT,
    rejection_reason TEXT,
    internal_notes TEXT, -- Only visible to staff
    
    -- Priority and categorization
    priority VARCHAR(20) NOT NULL DEFAULT 'normal'
        CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    tags TEXT[], -- Flexible categorization
    
    -- Quality metrics
    customer_satisfaction_rating INTEGER CHECK (customer_satisfaction_rating BETWEEN 1 AND 5),
    customer_feedback TEXT,
    processing_time_hours INTEGER,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Repair Tickets Table (detailed repair workflow)
CREATE TABLE IF NOT EXISTS repair_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationship to claim
    claim_id UUID NOT NULL REFERENCES warranty_claims(id) ON DELETE CASCADE,
    technician_id UUID NOT NULL REFERENCES users(id),
    
    -- Scheduling
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    start_date TIMESTAMP WITH TIME ZONE,
    target_completion_date TIMESTAMP WITH TIME ZONE,
    actual_completion_date TIMESTAMP WITH TIME ZONE,
    
    -- Repair process tracking
    status VARCHAR(20) NOT NULL DEFAULT 'assigned'
        CHECK (status IN ('assigned', 'in_progress', 'waiting_parts', 'completed', 'failed', 'cancelled')),
    
    -- Technical details
    diagnosis TEXT NOT NULL,
    repair_steps TEXT[], -- Array of repair steps completed
    
    -- Parts and labor
    parts_used JSONB DEFAULT '[]'::jsonb, -- Array of {part_number, quantity, cost, description}
    labor_hours DECIMAL(5,2) DEFAULT 0.00,
    hourly_rate DECIMAL(10,2),
    parts_cost DECIMAL(15,2) DEFAULT 0.00,
    labor_cost DECIMAL(15,2) GENERATED ALWAYS AS (
        labor_hours * COALESCE(hourly_rate, 0)
    ) STORED,
    total_cost DECIMAL(15,2) GENERATED ALWAYS AS (
        COALESCE(parts_cost, 0) + (labor_hours * COALESCE(hourly_rate, 0))
    ) STORED,
    
    -- Quality assurance
    quality_check_passed BOOLEAN,
    quality_notes TEXT,
    test_results JSONB DEFAULT '{}'::jsonb, -- Structured test results
    
    -- Documentation
    before_photos TEXT[], -- Array of photo URLs
    after_photos TEXT[], -- Array of photo URLs
    process_photos TEXT[], -- Work in progress photos
    
    -- Technical notes
    technician_notes TEXT,
    supervisor_notes TEXT,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Claim Attachments Table (files uploaded by customers)
CREATE TABLE IF NOT EXISTS claim_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationship
    claim_id UUID NOT NULL REFERENCES warranty_claims(id) ON DELETE CASCADE,
    uploaded_by UUID NOT NULL REFERENCES customers(id),
    
    -- File information
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    
    -- Categorization
    attachment_type VARCHAR(50) NOT NULL
        CHECK (attachment_type IN ('receipt', 'photo', 'invoice', 'document', 'video', 'other')),
    description TEXT,
    
    -- Processing status
    is_processed BOOLEAN DEFAULT FALSE,
    processing_notes TEXT,
    
    -- Security and validation
    checksum VARCHAR(64), -- SHA-256 hash
    virus_scan_status VARCHAR(20) DEFAULT 'pending'
        CHECK (virus_scan_status IN ('pending', 'clean', 'infected', 'failed')),
    virus_scan_date TIMESTAMP WITH TIME ZONE,
    
    -- Audit fields
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Claim Timeline Table (audit trail of all status changes)
CREATE TABLE IF NOT EXISTS claim_timeline (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Relationship
    claim_id UUID NOT NULL REFERENCES warranty_claims(id) ON DELETE CASCADE,
    
    -- Event details
    event_type VARCHAR(50) NOT NULL, -- status_change, note_added, attachment_uploaded, etc.
    from_status VARCHAR(20),
    to_status VARCHAR(20),
    
    -- Actor information
    actor_id UUID REFERENCES users(id), -- Who performed the action
    actor_type VARCHAR(20) NOT NULL -- customer, admin, technician, system
        CHECK (actor_type IN ('customer', 'admin', 'technician', 'system')),
    
    -- Event description and metadata
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb, -- Additional structured data
    
    -- Visibility control
    is_customer_visible BOOLEAN DEFAULT TRUE,
    
    -- Timestamp
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Barcode Generation Batches Table (for bulk generation tracking)
CREATE TABLE IF NOT EXISTS barcode_generation_batches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Batch information
    batch_number VARCHAR(100) UNIQUE NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    
    -- Generation details
    requested_quantity INTEGER NOT NULL,
    generated_quantity INTEGER DEFAULT 0,
    failed_quantity INTEGER DEFAULT 0,
    
    -- Batch metadata
    generation_started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    generation_completed_at TIMESTAMP WITH TIME ZONE,
    generation_status VARCHAR(20) DEFAULT 'in_progress'
        CHECK (generation_status IN ('in_progress', 'completed', 'failed', 'partial')),
    
    -- Performance tracking
    average_generation_time_ms INTEGER,
    collision_count INTEGER DEFAULT 0,
    retry_count INTEGER DEFAULT 0,
    
    -- Distribution details
    intended_recipient VARCHAR(255),
    distribution_notes TEXT,
    
    -- Audit
    requested_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Barcode Collision Log (for security monitoring)
CREATE TABLE IF NOT EXISTS barcode_collision_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Collision details
    attempted_barcode VARCHAR(17) NOT NULL,
    collision_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    generation_attempt INTEGER NOT NULL,
    batch_id UUID REFERENCES barcode_generation_batches(id),
    
    -- Resolution
    resolved_barcode VARCHAR(17),
    resolution_time_ms INTEGER,
    
    -- Security tracking
    generation_source VARCHAR(50), -- api, bulk_generation, migration
    client_ip INET,
    user_agent TEXT
);

-- Warranty Policy Templates (configurable warranty terms per product/category)
CREATE TABLE IF NOT EXISTS warranty_policy_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Template identification
    name VARCHAR(255) NOT NULL,
    description TEXT,
    storefront_id UUID NOT NULL REFERENCES storefronts(id),
    
    -- Applicability
    product_id UUID REFERENCES products(id), -- Specific product
    product_category_id UUID REFERENCES product_categories(id), -- Product category
    is_default BOOLEAN DEFAULT FALSE, -- Default policy for storefront
    
    -- Warranty terms
    warranty_period_months INTEGER NOT NULL,
    coverage_type VARCHAR(50) NOT NULL DEFAULT 'manufacturer_defect'
        CHECK (coverage_type IN ('manufacturer_defect', 'full_coverage', 'limited_warranty', 'extended_warranty')),
    
    -- Conditions and exclusions
    terms_and_conditions TEXT NOT NULL,
    exclusions TEXT,
    coverage_limitations TEXT,
    
    -- Claim processing rules
    requires_purchase_receipt BOOLEAN DEFAULT TRUE,
    max_claims_per_product INTEGER DEFAULT 3,
    claim_processing_sla_hours INTEGER DEFAULT 72,
    
    -- Status
    is_active BOOLEAN DEFAULT TRUE,
    effective_from DATE NOT NULL,
    effective_until DATE,
    
    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create comprehensive indexes for performance

-- Warranty Barcodes Indexes
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_barcode_number ON warranty_barcodes(barcode_number);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_product_id ON warranty_barcodes(product_id);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_storefront_id ON warranty_barcodes(storefront_id);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_customer_id ON warranty_barcodes(customer_id);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_status ON warranty_barcodes(status);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_batch_id ON warranty_barcodes(batch_id);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_expiry_date ON warranty_barcodes(expiry_date);
CREATE INDEX IF NOT EXISTS idx_warranty_barcodes_generated_at ON warranty_barcodes(generated_at);

-- Warranty Claims Indexes
CREATE INDEX IF NOT EXISTS idx_warranty_claims_claim_number ON warranty_claims(claim_number);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_barcode_id ON warranty_claims(barcode_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_customer_id ON warranty_claims(customer_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_storefront_id ON warranty_claims(storefront_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_status ON warranty_claims(status);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_priority ON warranty_claims(priority);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_technician ON warranty_claims(assigned_technician_id);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_created_at ON warranty_claims(created_at);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_completion_date ON warranty_claims(estimated_completion_date);
CREATE INDEX IF NOT EXISTS idx_warranty_claims_tags ON warranty_claims USING gin(tags);

-- Repair Tickets Indexes
CREATE INDEX IF NOT EXISTS idx_repair_tickets_claim_id ON repair_tickets(claim_id);
CREATE INDEX IF NOT EXISTS idx_repair_tickets_technician_id ON repair_tickets(technician_id);
CREATE INDEX IF NOT EXISTS idx_repair_tickets_status ON repair_tickets(status);
CREATE INDEX IF NOT EXISTS idx_repair_tickets_assigned_at ON repair_tickets(assigned_at);
CREATE INDEX IF NOT EXISTS idx_repair_tickets_completion_date ON repair_tickets(target_completion_date);

-- Claim Attachments Indexes
CREATE INDEX IF NOT EXISTS idx_claim_attachments_claim_id ON claim_attachments(claim_id);
CREATE INDEX IF NOT EXISTS idx_claim_attachments_type ON claim_attachments(attachment_type);
CREATE INDEX IF NOT EXISTS idx_claim_attachments_uploaded_at ON claim_attachments(uploaded_at);
CREATE INDEX IF NOT EXISTS idx_claim_attachments_processing ON claim_attachments(is_processed);

-- Claim Timeline Indexes
CREATE INDEX IF NOT EXISTS idx_claim_timeline_claim_id ON claim_timeline(claim_id);
CREATE INDEX IF NOT EXISTS idx_claim_timeline_event_type ON claim_timeline(event_type);
CREATE INDEX IF NOT EXISTS idx_claim_timeline_created_at ON claim_timeline(created_at);
CREATE INDEX IF NOT EXISTS idx_claim_timeline_visibility ON claim_timeline(is_customer_visible);

-- Batch Generation Indexes
CREATE INDEX IF NOT EXISTS idx_barcode_batches_batch_number ON barcode_generation_batches(batch_number);
CREATE INDEX IF NOT EXISTS idx_barcode_batches_product_id ON barcode_generation_batches(product_id);
CREATE INDEX IF NOT EXISTS idx_barcode_batches_storefront_id ON barcode_generation_batches(storefront_id);
CREATE INDEX IF NOT EXISTS idx_barcode_batches_status ON barcode_generation_batches(generation_status);
CREATE INDEX IF NOT EXISTS idx_barcode_batches_created_at ON barcode_generation_batches(created_at);

-- Collision Log Indexes
CREATE INDEX IF NOT EXISTS idx_collision_log_attempted_barcode ON barcode_collision_log(attempted_barcode);
CREATE INDEX IF NOT EXISTS idx_collision_log_collision_date ON barcode_collision_log(collision_date);
CREATE INDEX IF NOT EXISTS idx_collision_log_batch_id ON barcode_collision_log(batch_id);

-- Warranty Policy Indexes
CREATE INDEX IF NOT EXISTS idx_warranty_policy_storefront_id ON warranty_policy_templates(storefront_id);
CREATE INDEX IF NOT EXISTS idx_warranty_policy_product_id ON warranty_policy_templates(product_id);
CREATE INDEX IF NOT EXISTS idx_warranty_policy_category_id ON warranty_policy_templates(product_category_id);
CREATE INDEX IF NOT EXISTS idx_warranty_policy_is_active ON warranty_policy_templates(is_active);
CREATE INDEX IF NOT EXISTS idx_warranty_policy_effective_dates ON warranty_policy_templates(effective_from, effective_until);

-- Functions for warranty system

-- Function to generate sequential claim numbers
CREATE OR REPLACE FUNCTION generate_claim_number() RETURNS TEXT AS $$
DECLARE
    current_year TEXT;
    current_month TEXT;
    sequence_num INTEGER;
    claim_number TEXT;
BEGIN
    current_year := EXTRACT(YEAR FROM CURRENT_DATE)::TEXT;
    current_month := LPAD(EXTRACT(MONTH FROM CURRENT_DATE)::TEXT, 2, '0');
    
    -- Get next sequence number for current month
    SELECT COALESCE(MAX(
        CAST(SUBSTRING(claim_number FROM 'WC-\d{4}-\d{2}-(\d+)') AS INTEGER)
    ), 0) + 1
    INTO sequence_num
    FROM warranty_claims
    WHERE claim_number LIKE 'WC-' || current_year || '-' || current_month || '-%';
    
    claim_number := 'WC-' || current_year || '-' || current_month || '-' || LPAD(sequence_num::TEXT, 3, '0');
    RETURN claim_number;
END;
$$ LANGUAGE plpgsql;

-- Function to update warranty claim timeline
CREATE OR REPLACE FUNCTION update_claim_timeline() RETURNS TRIGGER AS $$
BEGIN
    -- Insert timeline entry for status changes
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO claim_timeline (
            claim_id, event_type, from_status, to_status, actor_id, actor_type, description
        ) VALUES (
            NEW.id, 'status_change', OLD.status, NEW.status, NEW.status_updated_by, 
            CASE 
                WHEN NEW.status_updated_by IS NULL THEN 'system'
                ELSE 'admin'
            END,
            'Claim status changed from ' || COALESCE(OLD.status, 'none') || ' to ' || NEW.status
        );
        
        -- Update status change tracking
        NEW.previous_status := OLD.status;
        NEW.status_updated_at := CURRENT_TIMESTAMP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to calculate warranty expiry date
CREATE OR REPLACE FUNCTION calculate_warranty_expiry(
    purchase_date DATE, 
    warranty_period_months INTEGER
) RETURNS DATE AS $$
BEGIN
    RETURN purchase_date + (warranty_period_months || ' months')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Function to validate barcode format
CREATE OR REPLACE FUNCTION validate_barcode_format(barcode TEXT) RETURNS BOOLEAN AS $$
BEGIN
    RETURN barcode ~ '^REX\d{2}[ABCDEFGHJKLMNPQRSTUVWXYZ23456789]{12}$';
END;
$$ LANGUAGE plpgsql;

-- Triggers

-- Auto-generate claim numbers
CREATE TRIGGER generate_claim_number_trigger
    BEFORE INSERT ON warranty_claims
    FOR EACH ROW
    WHEN (NEW.claim_number IS NULL OR NEW.claim_number = '')
    EXECUTE FUNCTION generate_claim_number();

-- Update claim timeline on status changes
CREATE TRIGGER update_claim_timeline_trigger
    AFTER UPDATE ON warranty_claims
    FOR EACH ROW
    EXECUTE FUNCTION update_claim_timeline();

-- Auto-update expiry date when purchase date or warranty period changes
CREATE OR REPLACE FUNCTION update_warranty_expiry() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.purchase_date IS NOT NULL AND NEW.warranty_period_months IS NOT NULL THEN
        NEW.expiry_date := calculate_warranty_expiry(NEW.purchase_date, NEW.warranty_period_months);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_warranty_expiry_trigger
    BEFORE INSERT OR UPDATE ON warranty_barcodes
    FOR EACH ROW
    EXECUTE FUNCTION update_warranty_expiry();

-- Update timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_warranty_claims_updated_at BEFORE UPDATE ON warranty_claims FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_warranty_barcodes_updated_at BEFORE UPDATE ON warranty_barcodes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_repair_tickets_updated_at BEFORE UPDATE ON repair_tickets FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_barcode_batches_updated_at BEFORE UPDATE ON barcode_generation_batches FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_warranty_policy_updated_at BEFORE UPDATE ON warranty_policy_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();