-- Remove indexes first
DROP INDEX IF EXISTS idx_warranty_barcodes_active;
DROP INDEX IF EXISTS idx_warranty_barcodes_deleted_at;

-- Remove deleted_at column from warranty_barcodes table
ALTER TABLE warranty_barcodes 
DROP COLUMN IF EXISTS deleted_at;