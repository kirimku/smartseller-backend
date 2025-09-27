-- Drop warranty system tables and related objects
-- This is the down migration for the warranty system

-- Drop triggers first
DROP TRIGGER IF EXISTS update_warranty_policy_updated_at ON warranty_policy_templates;
DROP TRIGGER IF EXISTS update_barcode_batches_updated_at ON barcode_generation_batches;
DROP TRIGGER IF EXISTS update_repair_tickets_updated_at ON repair_tickets;
DROP TRIGGER IF EXISTS update_warranty_barcodes_updated_at ON warranty_barcodes;
DROP TRIGGER IF EXISTS update_warranty_claims_updated_at ON warranty_claims;

DROP TRIGGER IF EXISTS update_warranty_expiry_trigger ON warranty_barcodes;
DROP TRIGGER IF EXISTS update_claim_timeline_trigger ON warranty_claims;
DROP TRIGGER IF EXISTS generate_claim_number_trigger ON warranty_claims;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP FUNCTION IF EXISTS update_warranty_expiry();
DROP FUNCTION IF EXISTS update_claim_timeline();
DROP FUNCTION IF EXISTS generate_claim_number();
DROP FUNCTION IF EXISTS calculate_warranty_expiry(DATE, INTEGER);
DROP FUNCTION IF EXISTS validate_barcode_format(TEXT);

-- Drop indexes (they will be dropped automatically with tables, but listing for clarity)
DROP INDEX IF EXISTS idx_warranty_policy_effective_dates;
DROP INDEX IF EXISTS idx_warranty_policy_is_active;
DROP INDEX IF EXISTS idx_warranty_policy_category_id;
DROP INDEX IF EXISTS idx_warranty_policy_product_id;
DROP INDEX IF EXISTS idx_warranty_policy_storefront_id;

DROP INDEX IF EXISTS idx_collision_log_batch_id;
DROP INDEX IF EXISTS idx_collision_log_collision_date;
DROP INDEX IF EXISTS idx_collision_log_attempted_barcode;

DROP INDEX IF EXISTS idx_barcode_batches_created_at;
DROP INDEX IF EXISTS idx_barcode_batches_status;
DROP INDEX IF EXISTS idx_barcode_batches_storefront_id;
DROP INDEX IF EXISTS idx_barcode_batches_product_id;
DROP INDEX IF EXISTS idx_barcode_batches_batch_number;

DROP INDEX IF EXISTS idx_claim_timeline_visibility;
DROP INDEX IF EXISTS idx_claim_timeline_created_at;
DROP INDEX IF EXISTS idx_claim_timeline_event_type;
DROP INDEX IF EXISTS idx_claim_timeline_claim_id;

DROP INDEX IF EXISTS idx_claim_attachments_processing;
DROP INDEX IF EXISTS idx_claim_attachments_uploaded_at;
DROP INDEX IF EXISTS idx_claim_attachments_type;
DROP INDEX IF EXISTS idx_claim_attachments_claim_id;

DROP INDEX IF EXISTS idx_repair_tickets_completion_date;
DROP INDEX IF EXISTS idx_repair_tickets_assigned_at;
DROP INDEX IF EXISTS idx_repair_tickets_status;
DROP INDEX IF EXISTS idx_repair_tickets_technician_id;
DROP INDEX IF EXISTS idx_repair_tickets_claim_id;

DROP INDEX IF EXISTS idx_warranty_claims_tags;
DROP INDEX IF EXISTS idx_warranty_claims_completion_date;
DROP INDEX IF EXISTS idx_warranty_claims_created_at;
DROP INDEX IF EXISTS idx_warranty_claims_technician;
DROP INDEX IF EXISTS idx_warranty_claims_priority;
DROP INDEX IF EXISTS idx_warranty_claims_status;
DROP INDEX IF EXISTS idx_warranty_claims_storefront_id;
DROP INDEX IF EXISTS idx_warranty_claims_customer_id;
DROP INDEX IF EXISTS idx_warranty_claims_barcode_id;
DROP INDEX IF EXISTS idx_warranty_claims_claim_number;

DROP INDEX IF EXISTS idx_warranty_barcodes_generated_at;
DROP INDEX IF EXISTS idx_warranty_barcodes_expiry_date;
DROP INDEX IF EXISTS idx_warranty_barcodes_batch_id;
DROP INDEX IF EXISTS idx_warranty_barcodes_status;
DROP INDEX IF EXISTS idx_warranty_barcodes_customer_id;
DROP INDEX IF EXISTS idx_warranty_barcodes_storefront_id;
DROP INDEX IF EXISTS idx_warranty_barcodes_product_id;
DROP INDEX IF EXISTS idx_warranty_barcodes_barcode_number;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS warranty_policy_templates;
DROP TABLE IF EXISTS barcode_collision_log;
DROP TABLE IF EXISTS barcode_generation_batches;
DROP TABLE IF EXISTS claim_timeline;
DROP TABLE IF EXISTS claim_attachments;
DROP TABLE IF EXISTS repair_tickets;
DROP TABLE IF EXISTS warranty_claims;
DROP TABLE IF EXISTS warranty_barcodes;