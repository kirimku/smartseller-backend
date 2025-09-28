package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/shopspring/decimal"
)

// PublicWarrantyHandler handles public warranty validation endpoints
type PublicWarrantyHandler struct {
	// TODO: Add usecase dependencies when available
	// warrantyUsecase usecase.WarrantyUsecase
	// productUsecase  usecase.ProductUsecase
}

// NewPublicWarrantyHandler creates a new public warranty handler
func NewPublicWarrantyHandler() *PublicWarrantyHandler {
	return &PublicWarrantyHandler{}
}

// ValidateWarranty validates a warranty barcode
// @Summary Validate warranty barcode
// @Description Validates a warranty barcode and returns warranty information
// @Tags Public Warranty
// @Accept json
// @Produce json
// @Param request body dto.PublicWarrantyValidationRequest true "Warranty validation request"
// @Success 200 {object} dto.PublicWarrantyValidationResponse "Warranty validation successful"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid request"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "Warranty not found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/validate [post]
func (h *PublicWarrantyHandler) ValidateWarranty(c *gin.Context) {
	var req dto.PublicWarrantyValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_request",
			Message:   "Invalid request format: " + err.Error(),
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// warranty, product, err := h.warrantyUsecase.ValidateWarrantyBarcode(req.BarcodeValue, req.ProductSKU)
	
	// Mock response for now
	mockWarranty := &entity.WarrantyBarcode{
		ID:                   uuid.New(),
		BarcodeNumber:        req.BarcodeValue,
		Status:               entity.BarcodeStatusActivated,
		IsActive:             true,
		IsExpired:            false,
		ActivatedAt:          func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
		ExpiryDate:           func() *time.Time { t := time.Now().AddDate(2, 0, 0); return &t }(),
		WarrantyPeriodMonths: 24,
	}

	mockProduct := &entity.Product{
		ID:   uuid.New(),
		SKU:  req.ProductSKU,
		Name: "Smartphone Pro Max 256GB",
		Brand: func() *string { s := "TechBrand"; return &s }(),
		Description: func() *string { s := "Latest flagship smartphone with advanced features"; return &s }(),
	}

	response := dto.ConvertWarrantyBarcodeToPublicValidationResponse(mockWarranty, mockProduct)
	c.JSON(http.StatusOK, response)
}

// LookupWarranty looks up warranties by product information
// @Summary Lookup warranty by product
// @Description Looks up warranties associated with a product
// @Tags Public Warranty
// @Accept json
// @Produce json
// @Param request body dto.PublicWarrantyLookupRequest true "Warranty lookup request"
// @Success 200 {object} dto.PublicWarrantyLookupResponse "Warranty lookup successful"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid request"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "No warranties found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/lookup [post]
func (h *PublicWarrantyHandler) LookupWarranty(c *gin.Context) {
	var req dto.PublicWarrantyLookupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_request",
			Message:   "Invalid request format: " + err.Error(),
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// warranties, product, err := h.warrantyUsecase.LookupWarrantiesByProduct(req.ProductSKU, req.SerialNumber, req.PurchaseDate, req.CustomerEmail)

	// Mock response for now
	mockWarranties := []*entity.WarrantyBarcode{
		{
			ID:                   uuid.New(),
			BarcodeNumber:        "WB-2024-001234567",
			Status:               entity.BarcodeStatusActivated,
			IsActive:             true,
			IsExpired:            false,
			ActivatedAt:          func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
			ExpiryDate:           func() *time.Time { t := time.Now().AddDate(2, 0, 0); return &t }(),
			WarrantyPeriodMonths: 24,
		},
		{
			ID:                   uuid.New(),
			BarcodeNumber:        "WB-2024-001234568",
			Status:               entity.BarcodeStatusActivated,
			IsActive:             true,
			IsExpired:            false,
			ActivatedAt:          func() *time.Time { t := time.Now().AddDate(0, -2, 0); return &t }(),
			ExpiryDate:           func() *time.Time { t := time.Now().AddDate(1, 10, 0); return &t }(),
			WarrantyPeriodMonths: 24,
		},
	}

	mockProduct := &entity.Product{
		ID:   uuid.New(),
		SKU:  req.ProductSKU,
		Name: "Smartphone Pro Max 256GB",
		Brand: func() *string { s := "TechBrand"; return &s }(),
		Description: func() *string { s := "Latest flagship smartphone with advanced features"; return &s }(),
	}

	response := dto.ConvertWarrantyBarcodesToPublicLookupResponse(mockWarranties, mockProduct)
	c.JSON(http.StatusOK, response)
}

// GetProductInfo gets product information and warranty options
// @Summary Get product information
// @Description Gets detailed product information including warranty options
// @Tags Public Warranty
// @Accept json
// @Produce json
// @Param request body dto.PublicProductInfoRequest true "Product info request"
// @Success 200 {object} dto.PublicProductInfoResponse "Product information retrieved"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid request"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "Product not found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/product-info [post]
func (h *PublicWarrantyHandler) GetProductInfo(c *gin.Context) {
	var req dto.PublicProductInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_request",
			Message:   "Invalid request format: " + err.Error(),
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// product, err := h.productUsecase.GetProductInfo(req.ProductSKU, req.ProductID, req.BarcodeValue)

	// Mock response for now
	mockProduct := &entity.Product{
		ID:   uuid.New(),
		SKU:  "SKU-PHONE-001",
		Name: "Smartphone Pro Max 256GB",
		Brand: func() *string { s := "TechBrand"; return &s }(),
		Description: func() *string { s := "Latest flagship smartphone with advanced features"; return &s }(),
		BasePrice: decimal.NewFromFloat(999.99),
	}

	response := dto.ConvertProductToPublicInfoResponse(mockProduct)
	c.JSON(http.StatusOK, response)
}

// CheckCoverage checks warranty coverage for a specific issue
// @Summary Check warranty coverage
// @Description Checks if a specific issue is covered under warranty
// @Tags Public Warranty
// @Accept json
// @Produce json
// @Param request body dto.PublicWarrantyCoverageCheckRequest true "Coverage check request"
// @Success 200 {object} dto.PublicWarrantyCoverageCheckResponse "Coverage check completed"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid request"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "Warranty not found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/check-coverage [post]
func (h *PublicWarrantyHandler) CheckCoverage(c *gin.Context) {
	var req dto.PublicWarrantyCoverageCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_request",
			Message:   "Invalid request format: " + err.Error(),
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// warranty, covered, err := h.warrantyUsecase.CheckCoverage(req.BarcodeValue, req.IssueType, req.IssueCategory, req.Description)

	// Mock response for now
	mockWarranty := &entity.WarrantyBarcode{
		ID:                   uuid.New(),
		BarcodeNumber:        req.BarcodeValue,
		Status:               entity.BarcodeStatusActivated,
		IsActive:             true,
		IsExpired:            false,
		ActivatedAt:          func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
		ExpiryDate:           func() *time.Time { t := time.Now().AddDate(2, 0, 0); return &t }(),
		WarrantyPeriodMonths: 24,
	}

	// Determine coverage based on issue type
	covered := true
	if req.IssueType == "water_damage" || req.IssueType == "physical_abuse" {
		covered = false
	}

	response := dto.ConvertToCoverageCheckResponse(mockWarranty, req.IssueType, req.IssueCategory, req.Description, covered)
	c.JSON(http.StatusOK, response)
}

// GetWarrantyByBarcode gets warranty information by barcode (GET endpoint)
// @Summary Get warranty by barcode
// @Description Gets warranty information using barcode value as URL parameter
// @Tags Public Warranty
// @Produce json
// @Param barcode path string true "Warranty barcode value"
// @Success 200 {object} dto.PublicWarrantyValidationResponse "Warranty information retrieved"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid barcode"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "Warranty not found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/{barcode} [get]
func (h *PublicWarrantyHandler) GetWarrantyByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	if barcode == "" {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_barcode",
			Message:   "Barcode parameter is required",
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// warranty, product, err := h.warrantyUsecase.GetWarrantyByBarcode(barcode)

	// Mock response for now
	mockWarranty := &entity.WarrantyBarcode{
		ID:                   uuid.New(),
		BarcodeNumber:        barcode,
		Status:               entity.BarcodeStatusActivated,
		IsActive:             true,
		IsExpired:            false,
		ActivatedAt:          func() *time.Time { t := time.Now().AddDate(0, -1, 0); return &t }(),
		ExpiryDate:           func() *time.Time { t := time.Now().AddDate(2, 0, 0); return &t }(),
		WarrantyPeriodMonths: 24,
	}

	mockProduct := &entity.Product{
		ID:   uuid.New(),
		SKU:  "SKU-PHONE-001",
		Name: "Smartphone Pro Max 256GB",
		Brand: func() *string { s := "TechBrand"; return &s }(),
		Description: func() *string { s := "Latest flagship smartphone with advanced features"; return &s }(),
	}

	response := dto.ConvertWarrantyBarcodeToPublicValidationResponse(mockWarranty, mockProduct)
	c.JSON(http.StatusOK, response)
}

// GetProductByBarcode gets product information by warranty barcode
// @Summary Get product by warranty barcode
// @Description Gets product information using warranty barcode value
// @Tags Public Warranty
// @Produce json
// @Param barcode path string true "Warranty barcode value"
// @Success 200 {object} dto.PublicProductInfoResponse "Product information retrieved"
// @Failure 400 {object} dto.PublicWarrantyErrorResponse "Invalid barcode"
// @Failure 404 {object} dto.PublicWarrantyErrorResponse "Product not found"
// @Failure 500 {object} dto.PublicWarrantyErrorResponse "Internal server error"
// @Router /api/v1/public/warranty/{barcode}/product [get]
func (h *PublicWarrantyHandler) GetProductByBarcode(c *gin.Context) {
	barcode := c.Param("barcode")
	if barcode == "" {
		c.JSON(http.StatusBadRequest, dto.PublicWarrantyErrorResponse{
			Error:     "invalid_barcode",
			Message:   "Barcode parameter is required",
			Code:      "WAR_400",
			Timestamp: time.Now(),
			RequestID: c.GetString("request_id"),
		})
		return
	}

	// TODO: Replace with actual usecase call
	// product, err := h.productUsecase.GetProductByWarrantyBarcode(barcode)

	// Mock response for now
	mockProduct := &entity.Product{
		ID:   uuid.New(),
		SKU:  "SKU-PHONE-001",
		Name: "Smartphone Pro Max 256GB",
		Brand: func() *string { s := "TechBrand"; return &s }(),
		Description: func() *string { s := "Latest flagship smartphone with advanced features"; return &s }(),
		BasePrice: decimal.NewFromFloat(999.99),
	}

	response := dto.ConvertProductToPublicInfoResponse(mockProduct)
	c.JSON(http.StatusOK, response)
}