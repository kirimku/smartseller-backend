package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
	infraRepo "github.com/kirimku/smartseller-backend/internal/infrastructure/repository"
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/middleware"
	"github.com/kirimku/smartseller-backend/pkg/utils"
	"github.com/rs/zerolog"
	"os"
)

// WarrantyBarcodeHandler handles warranty barcode-related HTTP requests
type WarrantyBarcodeHandler struct {
	logger              *slog.Logger
	barcodeService      service.BarcodeGeneratorService
	db                  *sqlx.DB
	tenantResolver      tenant.TenantResolver
}

// NewWarrantyBarcodeHandler creates a new warranty barcode handler
func NewWarrantyBarcodeHandler(logger *slog.Logger) *WarrantyBarcodeHandler {
	return &WarrantyBarcodeHandler{
		logger: logger,
	}
}

// NewWarrantyBarcodeHandlerWithDependencies creates a new warranty barcode handler with all dependencies
func NewWarrantyBarcodeHandlerWithDependencies(logger *slog.Logger, db *sqlx.DB, tenantResolver tenant.TenantResolver, repo repository.WarrantyBarcodeRepository) *WarrantyBarcodeHandler {
	zeroLogger := zerolog.New(os.Stdout).With().Str("component", "warranty_barcode").Timestamp().Logger()
	barcodeRepoAdapter := service.NewWarrantyBarcodeRepositoryAdapter(repo)
	barcodeService := service.NewBarcodeGeneratorService(
		barcodeRepoAdapter,
		nil, // collisionRepo - not implemented yet
		nil, // batchRepo - not implemented yet  
		zeroLogger,
	)
	
	return &WarrantyBarcodeHandler{
		logger:         logger,
		db:             db,
		tenantResolver: tenantResolver,
		barcodeService: barcodeService,
	}
}

// initializeBarcodeService initializes the barcode service lazily
func (h *WarrantyBarcodeHandler) initializeBarcodeService() {
	if h.barcodeService != nil {
		return
	}

	// Get database and tenant resolver from context or initialize
	// This is a temporary solution until proper dependency injection is set up
	// TODO: Move this to proper dependency injection in router
	if h.db == nil || h.tenantResolver == nil {
		h.logger.Error("Database or tenant resolver not initialized")
		return
	}

	// Create zerolog logger for repositories
	zeroLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Initialize warranty barcode repository
	warrantyBarcodeRepo := infraRepo.NewWarrantyBarcodeRepository(h.db, h.tenantResolver, zeroLogger)

	// Create adapter for service interface
	barcodeRepoAdapter := service.NewWarrantyBarcodeRepositoryAdapter(warrantyBarcodeRepo)

	// Initialize barcode service with nil for unimplemented repositories
	h.barcodeService = service.NewBarcodeGeneratorService(
		barcodeRepoAdapter,
		nil, // BarcodeCollisionRepository - not implemented yet
		nil, // BatchRepository - not implemented yet
		zeroLogger,
	)
}

// GenerateBarcodes handles barcode generation requests
// @Summary Generate warranty barcodes
// @Description Generate warranty barcodes for a product
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.WarrantyBarcodeRequest true "Barcode generation request"
// @Success 201 {object} dto.SuccessResponse{data=dto.BatchResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/generate [post]
func (h *WarrantyBarcodeHandler) GenerateBarcodes(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	var req dto.WarrantyBarcodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Parse product ID
	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		h.logger.Warn("Invalid product ID format", "product_id", req.ProductID)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	// Use the database and tenant resolver that were passed during handler initialization
	// These are available as handler fields and don't need to be retrieved from context

	// Initialize service components
	zeroLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	warrantyBarcodeRepo := infraRepo.NewWarrantyBarcodeRepository(
		h.db, 
		h.tenantResolver, 
		zeroLogger,
	)
	
	// Create adapter for service interface
	barcodeRepoAdapter := service.NewWarrantyBarcodeRepositoryAdapter(warrantyBarcodeRepo)
	
	// Initialize barcode generator service
	barcodeGeneratorService := service.NewBarcodeGeneratorService(
		barcodeRepoAdapter,
		nil, // collisionRepo - not needed for basic functionality
		nil, // batchRepo - not needed for basic functionality
		zeroLogger,
	)

	// Parse user ID
	createdBy, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Warn("Invalid user ID format", "user_id", userID)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	// Get storefront ID from context
	storefrontID, ok := middleware.GetStorefrontID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusBadRequest, "Storefront ID not found in context", nil)
		return
	}

	// Generate barcodes using the service
	startTime := time.Now()
	successCount := 0
	failureCount := 0

	for i := 0; i < req.Quantity; i++ {
		_, err := barcodeGeneratorService.GenerateBarcode(
			c.Request.Context(),
			productID,
			storefrontID,
			createdBy,
			req.ExpiryMonths,
		)
		if err != nil {
			h.logger.Error("Failed to generate barcode", "error", err.Error(), "attempt", i+1)
			failureCount++
		} else {
			successCount++
		}
	}

	processingTime := time.Since(startTime)
	
	batchResponse := &dto.BatchResponse{
		TotalProcessed: req.Quantity,
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		ProcessingTime: processingTime.String(),
		Timestamp:      time.Now(),
	}

	h.logger.Info("Successfully generated warranty barcodes",
		"product_id", productID,
		"requested", req.Quantity,
		"success", successCount,
		"failures", failureCount,
		"processing_time", processingTime,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusCreated, "Warranty barcodes generated successfully", batchResponse)
}

// ListBarcodes handles barcode listing requests
// @Summary List warranty barcodes
// @Description Get paginated list of warranty barcodes with filtering options
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param sort_by query string false "Sort field" default("created_at")
// @Param sort_dir query string false "Sort direction" Enums(asc,desc) default("desc")
// @Param product_id query string false "Filter by product ID"
// @Param batch_id query string false "Filter by batch ID"
// @Param status query string false "Filter by status" Enums(active,inactive,claimed)
// @Param search query string false "Search term"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyBarcodeListResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes [get]
func (h *WarrantyBarcodeHandler) ListBarcodes(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse query parameters
	var req dto.WarrantyBarcodeListRequest

	// Parse pagination
	req.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	req.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))
	req.SortBy = c.DefaultQuery("sort_by", "created_at")
	req.SortDir = c.DefaultQuery("sort_dir", "desc")

	// Parse filters
	if productID := c.Query("product_id"); productID != "" {
		req.ProductID = &productID
	}
	if batchID := c.Query("batch_id"); batchID != "" {
		req.BatchID = &batchID
	}
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}
	if search := c.Query("search"); search != "" {
		req.Search = &search
	}

	// TODO: Implement actual listing logic when usecase is ready
	// For now, return a mock response
	response := &dto.WarrantyBarcodeListResponse{
		Data: []dto.WarrantyBarcodeResponse{},
		Pagination: dto.PaginationResponse{
			Page:       req.Page,
			Limit:      req.PageSize,
			Total:      0,
			TotalPages: 1,
			HasNext:    false,
			HasPrev:    false,
		},
		Filters: dto.WarrantyBarcodeFilters{
			Statuses: []string{"active", "inactive", "claimed"},
			Products: []dto.WarrantyProductSummary{},
			Batches:  []dto.WarrantyBatchSummary{},
		},
	}

	h.logger.Info("Successfully listed warranty barcodes (mock)",
		"page", req.Page,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcodes retrieved successfully", response)
}

// GetBarcode handles individual barcode retrieval
// @Summary Get warranty barcode details
// @Description Get detailed information about a specific warranty barcode
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Barcode ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyBarcodeResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/{id} [get]
func (h *WarrantyBarcodeHandler) GetBarcode(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	barcodeIDStr := c.Param("id")
	barcodeID, err := uuid.Parse(barcodeIDStr)
	if err != nil {
		h.logger.Warn("Invalid barcode ID format", "barcode_id", barcodeIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid barcode ID format", nil)
		return
	}

	// TODO: Implement actual retrieval logic when usecase is ready
	// For now, return a mock response
	response := &dto.WarrantyBarcodeResponse{
		ID:           barcodeID.String(),
		ProductID:    "550e8400-e29b-41d4-a716-446655440001",
		ProductName:  "Sample Product",
		ProductSKU:   "SAMPLE-001",
		BarcodeValue: "REX24ABC123DEF456",
		Status:       "active",
		IsActive:     true,
		ExpiryDate:   time.Now().AddDate(2, 0, 0),  // 2 years from now
		CreatedAt:    time.Now().AddDate(0, -1, 0), // 1 month ago
		UpdatedAt:    time.Now(),
	}

	h.logger.Info("Successfully retrieved warranty barcode (mock)",
		"barcode_id", barcodeID,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcode retrieved successfully", response)
}

// ActivateBarcode handles barcode activation
// @Summary Activate warranty barcode
// @Description Activate a warranty barcode
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Barcode ID"
// @Param request body dto.WarrantyBarcodeActivationRequest true "Activation request"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyBarcodeResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/{id}/activate [post]
func (h *WarrantyBarcodeHandler) ActivateBarcode(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	barcodeIDStr := c.Param("id")
	barcodeID, err := uuid.Parse(barcodeIDStr)
	if err != nil {
		h.logger.Warn("Invalid barcode ID format", "barcode_id", barcodeIDStr)
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid barcode ID format", nil)
		return
	}

	var req dto.WarrantyBarcodeActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// TODO: Implement actual activation logic when usecase is ready
	// For now, return a mock response
	response := &dto.WarrantyBarcodeResponse{
		ID:           barcodeID.String(),
		ProductID:    "550e8400-e29b-41d4-a716-446655440001",
		ProductName:  "Sample Product",
		ProductSKU:   "SAMPLE-001",
		BarcodeValue: "REX24ABC123DEF456",
		Status:       "active",
		IsActive:     true,
		ExpiryDate:   time.Now().AddDate(2, 0, 0),  // 2 years from now
		CreatedAt:    time.Now().AddDate(0, -1, 0), // 1 month ago
		UpdatedAt:    time.Now(),
	}

	h.logger.Info("Successfully activated warranty barcode (mock)",
		"barcode_id", barcodeID,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcode activated successfully", response)
}

// BulkActivateBarcodes handles bulk barcode activation
// @Summary Bulk activate warranty barcodes
// @Description Activate multiple warranty barcodes at once
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.BulkWarrantyBarcodeActivationRequest true "Bulk activation request"
// @Success 200 {object} dto.SuccessResponse{data=dto.BatchResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/bulk-activate [post]
func (h *WarrantyBarcodeHandler) BulkActivateBarcodes(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	var req dto.BulkWarrantyBarcodeActivationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind request", "error", err.Error())
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// Parse barcode IDs
	barcodeIDs := make([]uuid.UUID, len(req.BarcodeIDs))
	for i, idStr := range req.BarcodeIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			h.logger.Warn("Invalid barcode ID format", "barcode_id", idStr)
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid barcode ID format", nil)
			return
		}
		barcodeIDs[i] = id
	}

	// TODO: Implement actual bulk activation logic when usecase is ready
	// For now, return a mock response
	batchResponse := &dto.BatchResponse{
		TotalProcessed: len(req.BarcodeIDs),
		SuccessCount:   len(req.BarcodeIDs),
		FailureCount:   0,
		ProcessingTime: "0.5s",
		Timestamp:      time.Now(),
	}

	h.logger.Info("Successfully bulk activated warranty barcodes (mock)",
		"total", len(barcodeIDs),
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcodes bulk activated", batchResponse)
}

// GetBarcodeStats handles barcode statistics requests
// @Summary Get warranty barcode statistics
// @Description Get comprehensive statistics about warranty barcodes
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id query string false "Filter by product ID"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyBarcodeStatsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/stats [get]
func (h *WarrantyBarcodeHandler) GetBarcodeStats(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Parse query parameters
	productIDStr := c.Query("product_id")
	_ = c.Query("date_from") // TODO: Implement date filtering
	_ = c.Query("date_to")   // TODO: Implement date filtering

	_ = productIDStr // TODO: Implement product filtering
	// var productID *uuid.UUID
	// if productIDStr != "" {
	// 	if pid, err := uuid.Parse(productIDStr); err != nil {
	// 		h.logger.Warn("Invalid product ID format", "product_id", productIDStr)
	// 		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid product ID format", nil)
	// 		return
	// 	} else {
	// 		productID = &pid
	// 	}
	// }

	// TODO: Implement actual statistics logic when usecase is ready
	// For now, return a mock response
	response := &dto.WarrantyBarcodeStatsResponse{
		TotalBarcodes:      1000,
		ActiveBarcodes:     850,
		InactiveBarcodes:   100,
		ClaimedBarcodes:    50,
		ExpiredBarcodes:    0,
		StatusBreakdown:    map[string]int{"active": 850, "inactive": 100, "claimed": 50},
		GeneratedToday:     25,
		GeneratedThisWeek:  150,
		GeneratedThisMonth: 600,
		ProductStats:       []dto.ProductWarrantyStats{},
		MonthlyGenerated:   []dto.MonthlyGenerationStats{},
		ExpiryBreakdown:    []dto.ExpiryBreakdownStats{},
		LastUpdated:        time.Now(),
	}

	h.logger.Info("Successfully retrieved barcode statistics (mock)",
		"total_barcodes", response.TotalBarcodes,
		"product_filter", productIDStr,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcode statistics retrieved successfully", response)
}

// ValidateBarcode handles barcode validation requests
// @Summary Validate warranty barcode
// @Description Validate a warranty barcode by its value
// @Tags warranty-barcodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param barcode_value path string true "Barcode value"
// @Success 200 {object} dto.SuccessResponse{data=dto.WarrantyBarcodeValidationResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/admin/warranty/barcodes/validate/{barcode_value} [get]
func (h *WarrantyBarcodeHandler) ValidateBarcode(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)
	if userID == "" {
		h.logger.Warn("User ID not found in context")
		utils.ErrorResponse(c, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	barcodeValue := strings.TrimSpace(c.Param("barcode_value"))
	if barcodeValue == "" {
		h.logger.Warn("Empty barcode value provided")
		utils.ErrorResponse(c, http.StatusBadRequest, "Barcode value is required", nil)
		return
	}

	// TODO: Implement actual validation logic when usecase is ready
	// For now, return a mock response
	expiryDate := time.Now().AddDate(2, 0, 0) // 2 years from now
	response := &dto.WarrantyBarcodeValidationResponse{
		IsValid:      true,
		BarcodeValue: barcodeValue,
		Status:       "active",
		IsExpired:    false,
		ExpiryDate:   &expiryDate,
		Product: &dto.WarrantyProductSummary{
			ID:   "550e8400-e29b-41d4-a716-446655440001",
			Name: "Sample Product",
			SKU:  "SAMPLE-001",
		},
		ValidatedAt: time.Now(),
	}

	h.logger.Info("Successfully validated warranty barcode (mock)",
		"barcode_value", barcodeValue,
		"is_valid", response.IsValid,
		"user_id", userID,
	)

	utils.SuccessResponse(c, http.StatusOK, "Warranty barcode validated successfully", response)
}
