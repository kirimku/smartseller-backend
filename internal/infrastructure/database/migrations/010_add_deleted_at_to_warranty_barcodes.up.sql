-- Add deleted_at column to warranty_barcodes table for soft delete functionality
ALTER TABLE warranty_barcodes 
ADD COLUMN deleted_at TIMESTAMP WITH TIME ZONE DEFAULT NULL;

-- Create index on deleted_at for better query performance
CREATE INDEX idx_warranty_barcodes_deleted_at ON warranty_barcodes(deleted_at);

-- Create partial index for active (non-deleted) records
CREATE INDEX idx_warranty_barcodes_active ON warranty_barcodes(storefront_id, batch_id) 
WHERE deleted_at IS NULL;