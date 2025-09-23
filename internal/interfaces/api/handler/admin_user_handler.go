package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/application/usecase"
	"github.com/kirimku/smartseller-backend/pkg/utils"
)

// AdminUserHandler handles HTTP requests for admin user management
type AdminUserHandler struct {
	adminUserUseCase usecase.AdminUserUseCase
}

// NewAdminUserHandler creates a new instance of AdminUserHandler
func NewAdminUserHandler(adminUserUseCase usecase.AdminUserUseCase) *AdminUserHandler {
	return &AdminUserHandler{
		adminUserUseCase: adminUserUseCase,
	}
}

// GetUsers handles the request to list users with pagination, search, and filtering
// @Summary List users with filters
// @Description Get a paginated list of users with optional search and filtering
// @Tags Admin
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search by email or phone"
// @Param user_type query string false "Filter by user type" Enums(personal, bisnis, agen)
// @Param user_tier query string false "Filter by user tier" Enums(pendekar, tuan_muda, tuan_besar, tuan_raja)
// @Param user_role query string false "Filter by user role" Enums(owner, admin, manager, support, user)
// @Success 200 {object} utils.APIResponse{data=dto.AdminUserListResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
// @Failure 403 {object} utils.APIResponse
// @Failure 500 {object} utils.APIResponse
// @Router /admin/users [get]
// @Security BearerAuth
func (h *AdminUserHandler) GetUsers(c *gin.Context) {
	var req dto.AdminUserListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters", err)
		return
	}

	// Call use case
	response, err := h.adminUserUseCase.GetUsers(c.Request.Context(), &req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve users", err)
		return
	}

	// Return success response
	utils.SuccessResponse(c, http.StatusOK, "Users retrieved successfully", response)
}
