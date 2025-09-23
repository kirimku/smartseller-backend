package dto

import (
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// AdminUserListRequest represents the request parameters for listing users
type AdminUserListRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	Limit    int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
	UserType string `form:"user_type" binding:"omitempty,oneof=personal bisnis agen"`
	UserTier string `form:"user_tier" binding:"omitempty,oneof=pendekar tuan_muda tuan_besar tuan_raja"`
	UserRole string `form:"user_role" binding:"omitempty,oneof=owner admin manager support user"`
}

// AdminUserSummary represents a summary of user information for admin lists
type AdminUserSummary struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	Phone            string          `json:"phone"`
	UserType         entity.UserType `json:"user_type"`
	UserTier         entity.UserTier `json:"user_tier"`
	UserRole         entity.UserRole `json:"user_role"`
	TransactionCount int             `json:"transaction_count"`
	IsAdmin          bool            `json:"is_admin"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// PaginationResponse represents pagination metadata
type PaginationResponse struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// AdminUserListResponse represents the response for user listing
type AdminUserListResponse struct {
	Users      []AdminUserSummary `json:"users"`
	Pagination PaginationResponse `json:"pagination"`
}

// ToAdminUserSummary converts a User entity to AdminUserSummary DTO
func ToAdminUserSummary(user *entity.User) AdminUserSummary {
	return AdminUserSummary{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		UserType:         user.UserType,
		UserTier:         user.UserTier,
		UserRole:         user.UserRole,
		TransactionCount: user.TransactionCount,
		IsAdmin:          user.IsAdmin,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}
}

// SetDefaults sets default values for the request
func (req *AdminUserListRequest) SetDefaults() {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
}

// ToGetUsersRequest converts AdminUserListRequest to repository.GetUsersRequest
func (req *AdminUserListRequest) ToGetUsersRequest() *entity.GetUsersRequest {
	getUsersReq := &entity.GetUsersRequest{
		Page:   req.Page,
		Limit:  req.Limit,
		Search: req.Search,
	}

	// Convert string filters to entity types if provided
	if req.UserType != "" {
		userType := entity.UserType(req.UserType)
		getUsersReq.UserType = &userType
	}

	if req.UserTier != "" {
		userTier := entity.UserTier(req.UserTier)
		getUsersReq.UserTier = &userTier
	}

	if req.UserRole != "" {
		userRole := entity.UserRole(req.UserRole)
		getUsersReq.UserRole = &userRole
	}

	return getUsersReq
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit, total int) PaginationResponse {
	totalPages := (total + limit - 1) / limit // Ceiling division

	return PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
