package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
)

// ProductVariantHandler handles product variant-related HTTP requests
type ProductVariantHandler struct {
	variantUseCase *usecase.ProductVariantUseCase
}

// NewProductVariantHandler creates a new product variant handler
func NewProductVariantHandler(variantUseCase *usecase.ProductVariantUseCase) *ProductVariantHandler {
	return &ProductVariantHandler{
		variantUseCase: variantUseCase,
	}
}

// CreateVariantOptions creates variant options for a product
// @Summary Create variant options for a product
// @Description Create variant options (like Color, Size) for a specific product
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID"
// @Param request body dto.CreateVariantOptionRequest true "Variant option creation request"
// @Success 201 {object} dto.VariantOptionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/products/{product_id}/variant-options [post]
func (h *ProductVariantHandler) CreateVariantOptions(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_PRODUCT_ID",
				Message: "Invalid product ID format",
				Details: map[string]interface{}{
					"product_id": productIDStr,
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	var req usecase.CreateVariantOptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	// Set the product ID from the URL parameter
	req.ProductID = productID

	response, err := h.variantUseCase.CreateVariantOption(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VARIANT_OPTION_CREATION_FAILED",
				Message: "Failed to create variant option",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// CreateVariant creates a specific variant for a product
// @Summary Create a product variant
// @Description Create a specific variant with option combinations for a product
// @Tags Product Variants
// @Accept json
// @Produce json
// @Param product_id path string true "Product ID"
// @Param request body dto.CreateVariantRequest true "Variant creation request"
// @Success 201 {object} dto.VariantResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/products/{product_id}/variants [post]
func (h *ProductVariantHandler) CreateVariant(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_PRODUCT_ID",
				Message: "Invalid product ID format",
				Details: map[string]interface{}{
					"product_id": productIDStr,
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	var req usecase.CreateVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INVALID_REQUEST",
				Message: "Invalid request body",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	// Set the product ID from the URL parameter
	req.ProductID = productID

	response, err := h.variantUseCase.CreateVariant(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VARIANT_CREATION_FAILED",
				Message: "Failed to create variant",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			},
			RequestID: c.GetString("request_id"),
			Timestamp: time.Now(),
			Path:      c.Request.URL.Path,
			Method:    c.Request.Method,
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}