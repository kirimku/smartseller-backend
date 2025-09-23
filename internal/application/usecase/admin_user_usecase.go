package usecase

import (
	"context"
	"fmt"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// AdminUserUseCase defines the interface for admin user management operations
type AdminUserUseCase interface {
	GetUsers(ctx context.Context, req *dto.AdminUserListRequest) (*dto.AdminUserListResponse, error)
}

// AdminUserUseCaseImpl implements the AdminUserUseCase interface
type AdminUserUseCaseImpl struct {
	userRepo repository.UserRepository
}

// NewAdminUserUseCase creates a new instance of AdminUserUseCaseImpl
func NewAdminUserUseCase(userRepo repository.UserRepository) AdminUserUseCase {
	return &AdminUserUseCaseImpl{
		userRepo: userRepo,
	}
}

// GetUsers retrieves users with pagination, search, and filtering
func (uc *AdminUserUseCaseImpl) GetUsers(ctx context.Context, req *dto.AdminUserListRequest) (*dto.AdminUserListResponse, error) {
	// Set default values
	req.SetDefaults()

	// Convert to repository request
	getUsersReq := req.ToGetUsersRequest()

	// Get users from repository
	users, err := uc.userRepo.GetAllUsersWithFilters(ctx, getUsersReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// Get total count for pagination
	totalCount, err := uc.userRepo.CountUsersWithFilters(ctx, getUsersReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	// Convert users to summary DTOs
	userSummaries := make([]dto.AdminUserSummary, len(users))
	for i, user := range users {
		userSummaries[i] = dto.ToAdminUserSummary(user)
	}

	// Calculate pagination
	pagination := dto.CalculatePagination(req.Page, req.Limit, totalCount)

	// Build response
	response := &dto.AdminUserListResponse{
		Users:      userSummaries,
		Pagination: pagination,
	}

	return response, nil
}
