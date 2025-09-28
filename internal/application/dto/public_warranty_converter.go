package dto

import (
	"fmt"
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// ConvertWarrantyBarcodeToPublicValidationResponse converts a WarrantyBarcode entity to PublicWarrantyValidationResponse
func ConvertWarrantyBarcodeToPublicValidationResponse(barcode *entity.WarrantyBarcode, product *entity.Product) *PublicWarrantyValidationResponse {
	if barcode == nil {
		return &PublicWarrantyValidationResponse{
			Valid:          false,
			Status:         "not_found",
			Message:        "Warranty barcode not found",
			ValidationTime: time.Now(),
		}
	}

	barcode.ComputeFields() // Ensure computed fields are updated

	response := &PublicWarrantyValidationResponse{
		Valid:          barcode.IsActive && !barcode.IsExpired,
		BarcodeValue:   barcode.BarcodeNumber,
		Status:         barcode.Status.String(),
		ValidationTime: time.Now(),
	}

	// Set appropriate message based on status
	if !barcode.IsActive {
		response.Message = "Warranty barcode is inactive"
	} else if barcode.IsExpired {
		response.Message = "Warranty has expired"
	} else {
		response.Message = "Warranty is valid and active"
	}

	// Convert product information if available
	if product != nil {
		response.Product = ConvertProductToPublicInfo(product)
	}

	// Convert warranty information
	response.Warranty = ConvertWarrantyBarcodeToPublicInfo(barcode)

	// Convert coverage information (mock data for now)
	response.Coverage = &PublicWarrantyCoverage{
		CoverageType:        "comprehensive",
		CoveredComponents:   []string{"hardware", "software", "battery", "screen"},
		ExcludedComponents:  []string{"water_damage", "physical_abuse"},
		RepairCoverage:      true,
		ReplacementCoverage: true,
		LaborCoverage:       true,
		PartsCoverage:       true,
		Terms:               []string{"Must provide proof of purchase", "Damage must be reported within 30 days"},
	}

	return response
}

// ConvertProductToPublicInfo converts a Product entity to PublicProductInfo
func ConvertProductToPublicInfo(product *entity.Product) *PublicProductInfo {
	if product == nil {
		return nil
	}

	info := &PublicProductInfo{
		ID:   product.ID.String(),
		SKU:  product.SKU,
		Name: product.Name,
	}

	// Handle optional fields
	if product.Brand != nil {
		info.Brand = *product.Brand
	}

	// Set category as "General" for now since Product doesn't have a direct category field
	info.Category = "General"

	if product.Description != nil {
		info.Description = product.Description
	}

	// Mock image URL for now
	imageURL := "https://example.com/images/" + product.SKU + ".jpg"
	info.ImageURL = &imageURL

	return info
}

// ConvertWarrantyBarcodeToPublicInfo converts a WarrantyBarcode entity to PublicWarrantyInfo
func ConvertWarrantyBarcodeToPublicInfo(barcode *entity.WarrantyBarcode) *PublicWarrantyInfo {
	if barcode == nil {
		return nil
	}

	barcode.ComputeFields() // Ensure computed fields are updated

	info := &PublicWarrantyInfo{
		ID:           barcode.ID.String(),
		BarcodeValue: barcode.BarcodeNumber,
		Status:       barcode.Status.String(),
		IsActive:     barcode.IsActive,
		IsExpired:    barcode.IsExpired,
		CanClaim:     barcode.IsActive && !barcode.IsExpired,
	}

	// Handle activation date
	if barcode.ActivatedAt != nil {
		info.ActivatedAt = barcode.ActivatedAt
	}

	// Handle expiry date
	if barcode.ExpiryDate != nil {
		info.ExpiryDate = *barcode.ExpiryDate
		
		// Calculate days remaining
		now := time.Now()
		if barcode.ExpiryDate.After(now) {
			info.DaysRemaining = int(barcode.ExpiryDate.Sub(now).Hours() / 24)
		} else {
			info.DaysRemaining = 0
		}

		// Calculate warranty period
		if barcode.ActivatedAt != nil {
			duration := barcode.ExpiryDate.Sub(*barcode.ActivatedAt)
			months := int(duration.Hours() / (24 * 30))
			if months > 0 {
				info.WarrantyPeriod = fmt.Sprintf("%d months", months)
			} else {
				days := int(duration.Hours() / 24)
				info.WarrantyPeriod = fmt.Sprintf("%d days", days)
			}
		}
	}

	return info
}

// ConvertWarrantyBarcodesToPublicLookupResponse converts multiple WarrantyBarcode entities to PublicWarrantyLookupResponse
func ConvertWarrantyBarcodesToPublicLookupResponse(barcodes []*entity.WarrantyBarcode, product *entity.Product) *PublicWarrantyLookupResponse {
	response := &PublicWarrantyLookupResponse{
		Found:      len(barcodes) > 0,
		Warranties: make([]PublicWarrantyInfo, 0, len(barcodes)),
		SearchTime: time.Now(),
	}

	if len(barcodes) == 0 {
		response.Message = "No warranties found for the specified criteria"
		return response
	}

	// Convert warranties
	for _, barcode := range barcodes {
		if warrantyInfo := ConvertWarrantyBarcodeToPublicInfo(barcode); warrantyInfo != nil {
			response.Warranties = append(response.Warranties, *warrantyInfo)
		}
	}

	// Convert product information
	if product != nil {
		response.Product = ConvertProductToPublicInfo(product)
	}

	// Set appropriate message
	if len(response.Warranties) == 1 {
		response.Message = "Found 1 warranty for this product"
	} else {
		response.Message = fmt.Sprintf("Found %d warranties for this product", len(response.Warranties))
	}

	return response
}

// ConvertProductToPublicInfoResponse converts a Product entity to PublicProductInfoResponse
func ConvertProductToPublicInfoResponse(product *entity.Product) *PublicProductInfoResponse {
	if product == nil {
		return nil
	}

	response := &PublicProductInfoResponse{
		Product:     ConvertProductToPublicInfo(product),
		RetrievedAt: time.Now(),
	}

	// Add warranty options (mock data for now)
	response.WarrantyOptions = []PublicWarrantyOption{
		{
			Type:        "standard",
			Name:        "Standard Warranty",
			Duration:    "24 months",
			Coverage:    "Hardware defects and manufacturing issues",
			Price:       &decimal.Zero,
			IsDefault:   true,
			Description: "Covers all manufacturing defects for 2 years",
		},
		{
			Type:        "extended",
			Name:        "Extended Warranty",
			Duration:    "36 months",
			Coverage:    "Hardware defects, manufacturing issues, and accidental damage",
			Price:       func() *decimal.Decimal { d := decimal.NewFromFloat(99.99); return &d }(),
			IsDefault:   false,
			Description: "Extended coverage including accidental damage for 3 years",
		},
	}

	// Add specifications (mock data for now)
	response.Specifications = map[string]interface{}{
		"screen_size": "6.7 inches",
		"storage":     "256GB",
		"color":       "Space Gray",
		"weight":      "240g",
		"dimensions":  "160.8 x 78.1 x 7.4 mm",
	}

	// Add documentation (mock data for now)
	response.Documentation = []PublicDocumentInfo{
		{
			Type:        "manual",
			Title:       "User Manual",
			Description: "Complete user guide and setup instructions",
			URL:         "https://example.com/docs/user-manual.pdf",
			Language:    "en",
			FileSize:    "2.5MB",
		},
		{
			Type:        "quick_start",
			Title:       "Quick Start Guide",
			Description: "Quick setup and basic usage guide",
			URL:         "https://example.com/docs/quick-start.pdf",
			Language:    "en",
			FileSize:    "1.2MB",
		},
	}

	// Add support information (mock data for now)
	response.SupportInfo = &PublicSupportInfo{
		SupportEmail:     "support@techbrand.com",
		SupportPhone:     "+1-800-123-4567",
		SupportHours:     "Mon-Fri 9AM-6PM EST",
		OnlineSupport:    "https://support.techbrand.com",
		ChatSupport:      true,
		SupportLanguages: []string{"en", "es", "fr"},
		FAQ:              "https://support.techbrand.com/faq",
	}

	return response
}

// ConvertToCoverageCheckResponse creates a PublicWarrantyCoverageCheckResponse
func ConvertToCoverageCheckResponse(barcode *entity.WarrantyBarcode, issueType, issueCategory, description string, covered bool) *PublicWarrantyCoverageCheckResponse {
	response := &PublicWarrantyCoverageCheckResponse{
		Covered:      covered,
		BarcodeValue: barcode.BarcodeNumber,
		IssueType:    issueType,
		CheckedAt:    time.Now(),
	}

	if covered {
		response.CoverageType = "full"
		response.EstimatedCost = &decimal.Zero
		response.Message = "This issue is fully covered under your warranty"
		response.NextSteps = []string{
			"Submit warranty claim",
			"Schedule repair appointment",
		}
		response.Recommendations = []string{
			"Contact authorized service center",
			"Backup your data before repair",
		}
	} else {
		response.CoverageType = "not_covered"
		estimatedCost := decimal.NewFromFloat(150.00)
		response.EstimatedCost = &estimatedCost
		response.Message = "This issue is not covered under your warranty"
		response.NextSteps = []string{
			"Contact customer service for paid repair options",
			"Get quote from authorized service center",
		}
		response.Recommendations = []string{
			"Consider extended warranty for future coverage",
			"Review warranty terms and conditions",
		}
	}

	// Add coverage details
	response.Coverage = &PublicWarrantyCoverage{
		CoverageType:        "comprehensive",
		CoveredComponents:   []string{"hardware", "software", "battery", "screen"},
		ExcludedComponents:  []string{"water_damage", "physical_abuse"},
		RepairCoverage:      true,
		ReplacementCoverage: true,
		LaborCoverage:       true,
		PartsCoverage:       true,
		Terms:               []string{"Must provide proof of purchase", "Damage must be reported within 30 days"},
	}

	return response
}