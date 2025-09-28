package dto

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// ConvertToCustomerWarrantyRegistrationResponse converts entities to customer warranty registration response
func ConvertToCustomerWarrantyRegistrationResponse(
	warranty *entity.WarrantyBarcode,
	product *entity.Product,
	customerID uuid.UUID,
	customerInfo CustomerRegistrationInfo,
) CustomerWarrantyRegistrationResponse {
	// Calculate days remaining
	daysRemaining := 0
	if warranty.ExpiryDate != nil {
		daysRemaining = int(time.Until(*warranty.ExpiryDate).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	}

	// Calculate warranty period
	warrantyPeriod := "Unknown"
	if warranty.WarrantyPeriodMonths > 0 {
		if warranty.WarrantyPeriodMonths == 12 {
			warrantyPeriod = "12 months"
		} else if warranty.WarrantyPeriodMonths%12 == 0 {
			warrantyPeriod = fmt.Sprintf("%d years", warranty.WarrantyPeriodMonths/12)
		} else {
			warrantyPeriod = fmt.Sprintf("%d months", warranty.WarrantyPeriodMonths)
		}
	}

	// Convert product
	productInfo := CustomerProductInfo{
		ID:          product.ID,
		SKU:         product.SKU,
		Name:        product.Name,
		Category:    "General", // Default category
		Description: "",
		ImageURL:    "https://example.com/images/" + product.SKU + ".jpg",
	}

	if product.Brand != nil {
		productInfo.Brand = *product.Brand
	}
	if product.Description != nil {
		productInfo.Description = *product.Description
	}
	if !product.BasePrice.IsZero() {
		productInfo.Price = &product.BasePrice
	}

	// Convert customer info
	customer := CustomerInfo{
		ID:          customerID,
		FirstName:   customerInfo.FirstName,
		LastName:    customerInfo.LastName,
		Email:       customerInfo.Email,
		PhoneNumber: customerInfo.PhoneNumber,
	}

	// Create coverage info
	coverage := CustomerWarrantyCoverage{
		CoverageType:        "comprehensive",
		CoveredComponents:   []string{"hardware", "software", "battery", "screen"},
		ExcludedComponents:  []string{"water_damage", "physical_abuse", "normal_wear"},
		RepairCoverage:      true,
		ReplacementCoverage: true,
		LaborCoverage:       true,
		PartsCoverage:       true,
		Terms: []string{
			"Must provide proof of purchase",
			"Damage must be reported within 30 days",
			"Warranty void if tampered with",
		},
		Limitations: []string{
			"Does not cover accidental damage",
			"Limited to original purchaser",
		},
	}

	// Create next steps
	nextSteps := []string{
		"Keep your warranty registration confirmation safe",
		"Register for online warranty portal access",
		"Download the mobile app for easy claim submission",
		"Contact support if you have any questions",
	}

	return CustomerWarrantyRegistrationResponse{
		Success:          true,
		RegistrationID:   uuid.New(), // Generate new registration ID
		WarrantyID:       warranty.ID,
		BarcodeValue:     warranty.BarcodeNumber,
		Status:           string(warranty.Status),
		ActivationDate:   time.Now(),
		ExpiryDate:       *warranty.ExpiryDate,
		WarrantyPeriod:   warrantyPeriod,
		Product:          productInfo,
		Customer:         customer,
		Coverage:         coverage,
		NextSteps:        nextSteps,
		RegistrationTime: time.Now(),
	}
}

// ConvertToCustomerWarrantySummary converts warranty entity to customer warranty summary
func ConvertToCustomerWarrantySummary(warranty *entity.WarrantyBarcode, product *entity.Product) CustomerWarrantySummary {
	// Calculate days remaining
	daysRemaining := 0
	if warranty.ExpiryDate != nil {
		daysRemaining = int(time.Until(*warranty.ExpiryDate).Hours() / 24)
		if daysRemaining < 0 {
			daysRemaining = 0
		}
	}

	// Calculate warranty period
	warrantyPeriod := "Unknown"
	if warranty.WarrantyPeriodMonths > 0 {
		if warranty.WarrantyPeriodMonths == 12 {
			warrantyPeriod = "12 months"
		} else if warranty.WarrantyPeriodMonths%12 == 0 {
			warrantyPeriod = fmt.Sprintf("%d years", warranty.WarrantyPeriodMonths/12)
		} else {
			warrantyPeriod = fmt.Sprintf("%d months", warranty.WarrantyPeriodMonths)
		}
	}

	// Convert product
	productInfo := CustomerProductInfo{
		ID:          product.ID,
		SKU:         product.SKU,
		Name:        product.Name,
		Category:    "General",
		Description: "",
		ImageURL:    "https://example.com/images/" + product.SKU + ".jpg",
	}

	if product.Brand != nil {
		productInfo.Brand = *product.Brand
	}
	if product.Description != nil {
		productInfo.Description = *product.Description
	}
	if !product.BasePrice.IsZero() {
		productInfo.Price = &product.BasePrice
	}

	activationDate := time.Now()
	if warranty.ActivatedAt != nil {
		activationDate = *warranty.ActivatedAt
	}

	return CustomerWarrantySummary{
		ID:             warranty.ID,
		BarcodeValue:   warranty.BarcodeNumber,
		Status:         string(warranty.Status),
		Product:        productInfo,
		ActivationDate: activationDate,
		ExpiryDate:     *warranty.ExpiryDate,
		DaysRemaining:  daysRemaining,
		WarrantyPeriod: warrantyPeriod,
		IsExpired:      warranty.IsExpired,
		CanClaim:       warranty.IsActive && !warranty.IsExpired,
		ClaimsCount:    0, // TODO: Get from claims service
	}
}

// ConvertToCustomerWarrantyDetailResponse converts entities to detailed warranty response
func ConvertToCustomerWarrantyDetailResponse(
	warranty *entity.WarrantyBarcode,
	product *entity.Product,
	customerID uuid.UUID,
	customerInfo CustomerRegistrationInfo,
) CustomerWarrantyDetailResponse {
	summary := ConvertToCustomerWarrantySummary(warranty, product)

	// Convert customer info
	customer := CustomerInfo{
		ID:          customerID,
		FirstName:   customerInfo.FirstName,
		LastName:    customerInfo.LastName,
		Email:       customerInfo.Email,
		PhoneNumber: customerInfo.PhoneNumber,
	}

	// Create coverage info
	coverage := CustomerWarrantyCoverage{
		CoverageType:        "comprehensive",
		CoveredComponents:   []string{"hardware", "software", "battery", "screen"},
		ExcludedComponents:  []string{"water_damage", "physical_abuse", "normal_wear"},
		RepairCoverage:      true,
		ReplacementCoverage: true,
		LaborCoverage:       true,
		PartsCoverage:       true,
		Terms: []string{
			"Must provide proof of purchase",
			"Damage must be reported within 30 days",
			"Warranty void if tampered with",
		},
		Limitations: []string{
			"Does not cover accidental damage",
			"Limited to original purchaser",
		},
	}

	// Create purchase info
	purchasePrice := decimal.NewFromFloat(999.99)
	purchaseInfo := PurchaseInfo{
		PurchaseDate:    time.Now().AddDate(0, -1, 0), // Mock: 1 month ago
		PurchasePrice:   &purchasePrice,
		RetailerName:    "TechStore Inc",
		RetailerAddress: "123 Main St, City, State",
		InvoiceNumber:   "INV-2024-001",
		SerialNumber:    "SN123456789",
	}

	// Create support info
	supportInfo := CustomerSupportInfo{
		SupportEmail: "support@techbrand.com",
		SupportPhone: "+1-800-SUPPORT",
		SupportHours: "Mon-Fri 9AM-6PM EST",
		OnlinePortal: "https://support.techbrand.com",
		ChatSupport:  true,
		ServiceCenters: []string{
			"New York Service Center - 123 Tech Ave, NY 10001",
			"Los Angeles Service Center - 456 Innovation Blvd, CA 90210",
		},
	}

	return CustomerWarrantyDetailResponse{
		ID:             summary.ID,
		BarcodeValue:   summary.BarcodeValue,
		Status:         summary.Status,
		Product:        summary.Product,
		Customer:       customer,
		ActivationDate: summary.ActivationDate,
		ExpiryDate:     summary.ExpiryDate,
		DaysRemaining:  summary.DaysRemaining,
		WarrantyPeriod: summary.WarrantyPeriod,
		IsExpired:      summary.IsExpired,
		CanClaim:       summary.CanClaim,
		Coverage:       coverage,
		PurchaseInfo:   purchaseInfo,
		ClaimsHistory:  []CustomerClaimSummary{}, // TODO: Get from claims service
		Documents:      []WarrantyDocument{},     // TODO: Get from document service
		SupportInfo:    supportInfo,
		RetrievalTime:  time.Now(),
	}
}

// ConvertToCustomerWarrantyListResponse converts warranty entities to list response
func ConvertToCustomerWarrantyListResponse(
	warranties []*entity.WarrantyBarcode,
	products map[uuid.UUID]*entity.Product,
	totalCount int,
	page, limit int,
) CustomerWarrantyListResponse {
	summaries := make([]CustomerWarrantySummary, 0, len(warranties))

	for _, warranty := range warranties {
		if product, exists := products[warranty.ProductID]; exists {
			summary := ConvertToCustomerWarrantySummary(warranty, product)
			summaries = append(summaries, summary)
		}
	}

	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrevious := page > 1

	return CustomerWarrantyListResponse{
		Warranties:  summaries,
		TotalCount:  totalCount,
		Page:        page,
		Limit:       limit,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
		RequestTime: time.Now(),
	}
}