package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// ProductCategoryHandler handles HTTP requests for product categories
type ProductCategoryHandler struct {
	categoryUseCase *usecase.ProductCategoryUseCase
}

// NewProductCategoryHandler creates a new ProductCategoryHandler
func NewProductCategoryHandler(categoryUseCase *usecase.ProductCategoryUseCase) *ProductCategoryHandler {
	return &ProductCategoryHandler{
		categoryUseCase: categoryUseCase,
	}
}

// CreateCategory creates a new product category
// @Summary Create a new product category
// @Description Create a new product category with optional parent hierarchy
// @Tags Categories
// @Accept json
// @Produce json
// @Param category body dto.CreateCategoryRequest true "Category creation data"
// @Success 201 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /api/v1/categories [post]
func (h *ProductCategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Convert DTO to usecase request
	usecaseReq := usecase.CreateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentID,
		IsActive:    req.IsActive != nil && *req.IsActive,
		SortOrder:   0, // Default sort order
	}
	if req.SortOrder != nil {
		usecaseReq.SortOrder = *req.SortOrder
	}

	category, err := h.categoryUseCase.CreateCategory(c.Request.Context(), usecaseReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusCreated, response)
}

// GetCategory retrieves a category by ID
// @Summary Get category by ID
// @Description Retrieve a single category with its details
// @Tags Categories
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *ProductCategoryHandler) GetCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	category, err := h.categoryUseCase.GetCategoryByID(c.Request.Context(), categoryID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Category not found", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusOK, response)
}

// UpdateCategory updates an existing category
// @Summary Update category
// @Description Update an existing category's information
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body dto.UpdateCategoryRequest true "Category update data"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id} [put]
func (h *ProductCategoryHandler) UpdateCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Convert DTO to usecase request
	usecaseReq := usecase.UpdateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	}

	category, err := h.categoryUseCase.UpdateCategory(c.Request.Context(), categoryID, usecaseReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update category", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusOK, response)
}

// DeleteCategory deletes a category
// @Summary Delete category
// @Description Delete a category and optionally reassign its products
// @Tags Categories
// @Param id path string true "Category ID"
// @Param reassign_to query string false "Category ID to reassign products to"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id} [delete]
func (h *ProductCategoryHandler) DeleteCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	var reassignToID *uuid.UUID
	if reassignTo := c.Query("reassign_to"); reassignTo != "" {
		id, err := uuid.Parse(reassignTo)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid reassign_to category ID", err)
			return
		}
		reassignToID = &id
	}

	err = h.categoryUseCase.DeleteCategory(c.Request.Context(), categoryID, reassignToID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete category", err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ListCategories lists categories with pagination and filtering
// @Summary List categories
// @Description Get a paginated list of categories with optional filtering
// @Tags Categories
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param parent_id query string false "Filter by parent category ID"
// @Param is_active query bool false "Filter by active status"
// @Param search query string false "Search in category names"
// @Success 200 {object} dto.CategoryListResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories [get]
func (h *ProductCategoryHandler) ListCategories(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filters
	filters := dto.CategoryFilters{}
	
	if parentID := c.Query("parent_id"); parentID != "" {
		id, err := uuid.Parse(parentID)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid parent_id", err)
			return
		}
		filters.ParentID = &id
	}

	if isActiveStr := c.Query("is_active"); isActiveStr != "" {
		isActive, err := strconv.ParseBool(isActiveStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid is_active value", err)
			return
		}
		filters.IsActive = &isActive
	}

	if search := c.Query("search"); search != "" {
		filters.Search = &search
	}

	// Get categories from usecase
	filter := repository.ProductCategoryFilter{}
	categories, err := h.categoryUseCase.GetAllCategories(c.Request.Context(), filter)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get categories", err)
		return
	}

	// Apply filters and pagination (simplified implementation)
	filteredCategories := h.applyFilters(categories, filters)
	
	// Calculate pagination
	total := len(filteredCategories)
	offset := (page - 1) * limit
	end := offset + limit
	if end > total {
		end = total
	}
	
	var paginatedCategories []*entity.ProductCategory
	if offset < total {
		paginatedCategories = filteredCategories[offset:end]
	}

	// Convert to response format
	categorySummaries := make([]dto.CategorySummary, len(paginatedCategories))
	for i, cat := range paginatedCategories {
		categorySummaries[i] = h.convertToSummary(cat)
	}

	response := dto.CategoryListResponse{
		Categories: categorySummaries,
		Pagination: dto.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: (total + limit - 1) / limit,
		},
		Summary: dto.CategoryListSummary{
			TotalCategories:  total,
			ActiveCategories: h.countActiveCategories(filteredCategories),
			RootCategories:   h.countRootCategories(filteredCategories),
			MaxDepthLevel:    h.calculateMaxDepth(categories),
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetCategoryTree retrieves the category hierarchy tree
// @Summary Get category tree
// @Description Get the hierarchical tree structure of categories
// @Tags Categories
// @Produce json
// @Param root_id query string false "Root category ID to start from"
// @Success 200 {object} dto.CategoryTreeResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories/tree [get]
func (h *ProductCategoryHandler) GetCategoryTree(c *gin.Context) {
	var rootID *uuid.UUID
	if rootIDStr := c.Query("root_id"); rootIDStr != "" {
		id, err := uuid.Parse(rootIDStr)
		if err != nil {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid root_id", err)
			return
		}
		rootID = &id
	}

	tree, err := h.categoryUseCase.GetCategoryTree(c.Request.Context(), rootID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get category tree", err)
		return
	}

	treeNodes := make([]dto.CategoryTreeNode, len(tree))
	for i, node := range tree {
		treeNodes[i] = h.convertToTreeNode(node)
	}

	response := dto.CategoryTreeResponse{
		Categories: treeNodes,
		TotalCount: len(treeNodes),
	}

	c.JSON(http.StatusOK, response)
}

// MoveCategory moves a category to a new parent
// @Summary Move category
// @Description Move a category to a different parent in the hierarchy
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param move body dto.MoveCategoryRequest true "Move category data"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id}/move [post]
func (h *ProductCategoryHandler) MoveCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	var req dto.MoveCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Override the category ID from the URL
	req.CategoryID = categoryID

	// Convert to usecase request
	usecaseReq := usecase.MoveCategoryRequest{
		CategoryID:   req.CategoryID,
		NewParentID:  req.NewParentID,
		NewSortOrder: req.NewSortOrder,
	}

	category, err := h.categoryUseCase.MoveCategoryToParent(c.Request.Context(), usecaseReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to move category", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusOK, response)
}

// ActivateCategory activates a category
// @Summary Activate category
// @Description Activate a category to make it visible
// @Tags Categories
// @Param id path string true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id}/activate [post]
func (h *ProductCategoryHandler) ActivateCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	isActive := true
	usecaseReq := usecase.UpdateCategoryRequest{
		IsActive: &isActive,
	}

	category, err := h.categoryUseCase.UpdateCategory(c.Request.Context(), categoryID, usecaseReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to activate category", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusOK, response)
}

// DeactivateCategory deactivates a category
// @Summary Deactivate category
// @Description Deactivate a category to hide it
// @Tags Categories
// @Param id path string true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories/{id}/deactivate [post]
func (h *ProductCategoryHandler) DeactivateCategory(c *gin.Context) {
	categoryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid category ID", err)
		return
	}

	isActive := false
	usecaseReq := usecase.UpdateCategoryRequest{
		IsActive: &isActive,
	}

	category, err := h.categoryUseCase.UpdateCategory(c.Request.Context(), categoryID, usecaseReq)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to deactivate category", err)
		return
	}

	response := h.convertToResponse(category)
	c.JSON(http.StatusOK, response)
}

// BulkOperations performs bulk operations on categories
// @Summary Bulk operations on categories
// @Description Perform bulk operations (activate, deactivate, delete) on multiple categories
// @Tags Categories
// @Accept json
// @Produce json
// @Param operation body dto.CategoryBulkOperationRequest true "Bulk operation data"
// @Success 200 {object} dto.CategoryBulkOperationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/v1/categories/bulk [post]
func (h *ProductCategoryHandler) BulkOperations(c *gin.Context) {
	var req dto.CategoryBulkOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	startTime := time.Now()
	successCount := 0
	var failures []dto.CategoryOperationFailure

	for _, categoryID := range req.CategoryIDs {
		var err error
		switch req.Operation {
		case "activate":
			isActive := true
			usecaseReq := usecase.UpdateCategoryRequest{IsActive: &isActive}
			_, err = h.categoryUseCase.UpdateCategory(c.Request.Context(), categoryID, usecaseReq)
		case "deactivate":
			isActive := false
			usecaseReq := usecase.UpdateCategoryRequest{IsActive: &isActive}
			_, err = h.categoryUseCase.UpdateCategory(c.Request.Context(), categoryID, usecaseReq)
		case "delete":
			err = h.categoryUseCase.DeleteCategory(c.Request.Context(), categoryID, nil)
		default:
			err = fmt.Errorf("unsupported operation: %s", req.Operation)
		}

		if err != nil {
			failures = append(failures, dto.CategoryOperationFailure{
				CategoryID: categoryID,
				Error:      err.Error(),
				ErrorCode:  "OPERATION_FAILED",
			})
		} else {
			successCount++
		}
	}

	response := dto.CategoryBulkOperationResponse{
		SuccessCount: successCount,
		FailureCount: len(failures),
		Failures:     failures,
		Metadata: dto.CategoryBulkOperationMetadata{
			TotalRequested: len(req.CategoryIDs),
			ProcessingTime: time.Since(startTime).String(),
			ProcessedAt:    time.Now().Format(time.RFC3339),
		},
	}

	c.JSON(http.StatusOK, response)
}

// Helper methods for conversion and filtering

func (h *ProductCategoryHandler) convertToResponse(category *entity.ProductCategory) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:              category.ID,
		Name:            category.Name,
		Description:     category.Description,
		ParentID:        category.ParentID,
		Path:            category.GetFullPath(),
		Level:           category.GetLevel(),
		Slug:            &category.Slug,
		ImageURL:        nil, // TODO: Add ImageURL field to entity if needed
		IsActive:        category.IsActive,
		SortOrder:       category.SortOrder,
		ProductCount:    0, // TODO: Implement product count
		MetaTitle:       nil, // TODO: Add MetaTitle field to entity if needed
		MetaDescription: nil, // TODO: Add MetaDescription field to entity if needed
		CreatedAt:       category.CreatedAt,
		UpdatedAt:       category.UpdatedAt,
	}
}

func (h *ProductCategoryHandler) convertToSummary(category *entity.ProductCategory) dto.CategorySummary {
	return dto.CategorySummary{
		ID:       category.ID,
		Name:     category.Name,
		Path:     category.GetFullPath(),
		ParentID: category.ParentID,
	}
}

func (h *ProductCategoryHandler) convertToTreeNode(node *usecase.CategoryTreeNode) dto.CategoryTreeNode {
	children := make([]dto.CategoryTreeNode, len(node.Children))
	for i, child := range node.Children {
		children[i] = h.convertToTreeNode(child)
	}

	return dto.CategoryTreeNode{
		ID:           node.ID,
		Name:         node.Name,
		Description:  node.Description,
		Path:         node.GetFullPath(),
		Level:        node.Level,
		ProductCount: 0, // TODO: Implement product count
		IsActive:     node.IsActive,
		SortOrder:    node.SortOrder,
		Children:     children,
	}
}

func (h *ProductCategoryHandler) applyFilters(categories []*entity.ProductCategory, filters dto.CategoryFilters) []*entity.ProductCategory {
	var filtered []*entity.ProductCategory
	
	for _, cat := range categories {
		// Apply parent ID filter
		if filters.ParentID != nil {
			if cat.ParentID == nil || *cat.ParentID != *filters.ParentID {
				continue
			}
		}

		// Apply active status filter
		if filters.IsActive != nil && cat.IsActive != *filters.IsActive {
			continue
		}

		// Apply search filter
		if filters.Search != nil && !strings.Contains(strings.ToLower(cat.Name), strings.ToLower(*filters.Search)) {
			continue
		}

		filtered = append(filtered, cat)
	}

	return filtered
}

func (h *ProductCategoryHandler) countActiveCategories(categories []*entity.ProductCategory) int {
	count := 0
	for _, cat := range categories {
		if cat.IsActive {
			count++
		}
	}
	return count
}

func (h *ProductCategoryHandler) countRootCategories(categories []*entity.ProductCategory) int {
	count := 0
	for _, cat := range categories {
		if cat.ParentID == nil {
			count++
		}
	}
	return count
}

func (h *ProductCategoryHandler) calculateMaxDepth(categories []*entity.ProductCategory) int {
	maxDepth := 0
	for _, cat := range categories {
		level := cat.GetLevel()
		if level > maxDepth {
			maxDepth = level
		}
	}
	return maxDepth
}