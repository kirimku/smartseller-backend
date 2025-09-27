package dto

import (
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// ConvertWarrantyBarcodeToResponse converts a WarrantyBarcode entity to WarrantyBarcodeResponse DTO
func ConvertWarrantyBarcodeToResponse(barcode *entity.WarrantyBarcode, productName, productSKU, batchName string) *WarrantyBarcodeResponse {
	if barcode == nil {
		return nil
	}

	barcode.ComputeFields() // Ensure computed fields are updated

	response := &WarrantyBarcodeResponse{
		ID:           barcode.ID.String(),
		ProductID:    barcode.ProductID.String(),
		ProductName:  productName,
		ProductSKU:   productSKU,
		BarcodeValue: barcode.BarcodeNumber,
		Status:       barcode.Status.String(),
		IsActive:     barcode.IsActive,
		CreatedAt:    barcode.CreatedAt,
		UpdatedAt:    barcode.UpdatedAt,
	}

	// Handle nullable expiry date
	if barcode.ExpiryDate != nil {
		response.ExpiryDate = *barcode.ExpiryDate
	} else {
		// Set to zero time if not available
		response.ExpiryDate = time.Time{}
	}

	// Handle batch information
	if barcode.BatchID != nil {
		batchIDStr := barcode.BatchID.String()
		response.BatchID = &batchIDStr
		if batchName != "" {
			response.BatchName = &batchName
		}
	}

	return response
}

// ConvertWarrantyBarcodesToResponses converts a slice of WarrantyBarcode entities to response DTOs
func ConvertWarrantyBarcodesToResponses(barcodes []*entity.WarrantyBarcode, productData map[string]struct{ Name, SKU string }, batchData map[string]string) []*WarrantyBarcodeResponse {
	responses := make([]*WarrantyBarcodeResponse, len(barcodes))

	for i, barcode := range barcodes {
		var productName, productSKU, batchName string

		productIDStr := barcode.ProductID.String()
		if product, exists := productData[productIDStr]; exists {
			productName = product.Name
			productSKU = product.SKU
		}

		if barcode.BatchID != nil {
			batchIDStr := barcode.BatchID.String()
			if name, exists := batchData[batchIDStr]; exists {
				batchName = name
			}
		}

		responses[i] = ConvertWarrantyBarcodeToResponse(barcode, productName, productSKU, batchName)
	}

	return responses
}

// ConvertBarcodeGenerationBatchToResponse converts a BarcodeGenerationBatch entity to response DTO
func ConvertBarcodeGenerationBatchToResponse(batch *entity.BarcodeGenerationBatch, productName, createdByName string) *BarcodeGenerationBatchResponse {
	if batch == nil {
		return nil
	}

	batch.ComputeFields() // Ensure computed fields are updated

	response := &BarcodeGenerationBatchResponse{
		ID:             batch.ID.String(),
		BatchName:      batch.BatchNumber,
		ProductID:      batch.ProductID.String(),
		ProductName:    productName,
		RequestedCount: batch.RequestedQuantity,
		GeneratedCount: batch.GeneratedQuantity,
		Status:         batch.GenerationStatus,
		CreatedBy:      batch.RequestedBy.String(),
		CreatedByName:  createdByName,
		CreatedAt:      batch.CreatedAt,
	}

	// Handle optional fields
	if batch.DistributionNotes != nil {
		response.Notes = batch.DistributionNotes
	}

	if batch.ProcessingTime != "" {
		response.ProcessingTime = &batch.ProcessingTime
	}

	if batch.GenerationCompletedAt != nil {
		response.CompletedAt = batch.GenerationCompletedAt
	}

	return response
}

// ConvertBarcodeGenerationBatchesToResponses converts a slice of BarcodeGenerationBatch entities to response DTOs
func ConvertBarcodeGenerationBatchesToResponses(batches []*entity.BarcodeGenerationBatch, productData map[string]string, userData map[string]string) []*BarcodeGenerationBatchResponse {
	responses := make([]*BarcodeGenerationBatchResponse, len(batches))

	for i, batch := range batches {
		var productName, createdByName string

		productIDStr := batch.ProductID.String()
		if name, exists := productData[productIDStr]; exists {
			productName = name
		}

		createdByStr := batch.RequestedBy.String()
		if name, exists := userData[createdByStr]; exists {
			createdByName = name
		}

		responses[i] = ConvertBarcodeGenerationBatchToResponse(batch, productName, createdByName)
	}

	return responses
}
