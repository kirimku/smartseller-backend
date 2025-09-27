package service

import (
	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new instance of UserService
func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetTierDistribution gets the distribution of users across different tiers
func (s *UserService) GetTierDistribution() (*dto.TierDistributionDTO, error) {
	// Get counts for each tier
	basicCount, err := s.userRepo.CountUsersByTier(entity.UserTierBasic)
	if err != nil {
		return nil, err
	}

	premiumCount, err := s.userRepo.CountUsersByTier(entity.UserTierPremium)
	if err != nil {
		return nil, err
	}

	proCount, err := s.userRepo.CountUsersByTier(entity.UserTierPro)
	if err != nil {
		return nil, err
	}

	enterpriseCount, err := s.userRepo.CountUsersByTier(entity.UserTierEnterprise)
	if err != nil {
		return nil, err
	}

	// Get users with no tier (default state)
	noTierCount, err := s.userRepo.CountUsersByTier("")
	if err != nil {
		return nil, err
	}

	// Total users
	totalUsers := basicCount + premiumCount + proCount + enterpriseCount + noTierCount

	// Build response
	return &dto.TierDistributionDTO{
		BasicCount:      basicCount,
		PremiumCount:    premiumCount,
		ProCount:        proCount,
		EnterpriseCount: enterpriseCount,
		NoTierCount:     noTierCount,
		TotalUsers:      totalUsers,
	}, nil
}
