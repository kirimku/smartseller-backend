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
	pendekarCount, err := s.userRepo.CountUsersByTier(entity.UserTierPendekar)
	if err != nil {
		return nil, err
	}

	tuanMudaCount, err := s.userRepo.CountUsersByTier(entity.UserTierTuanMuda)
	if err != nil {
		return nil, err
	}

	tuanBesarCount, err := s.userRepo.CountUsersByTier(entity.UserTierTuanBesar)
	if err != nil {
		return nil, err
	}

	tuanRajaCount, err := s.userRepo.CountUsersByTier(entity.UserTierTuanRaja)
	if err != nil {
		return nil, err
	}

	// Get users with no tier (default state)
	noTierCount, err := s.userRepo.CountUsersByTier("")
	if err != nil {
		return nil, err
	}

	// Total users
	totalUsers := pendekarCount + tuanMudaCount + tuanBesarCount + tuanRajaCount + noTierCount

	// Calculate percentages
	calculatePercentage := func(count int) float64 {
		if totalUsers == 0 {
			return 0
		}
		return float64(count) / float64(totalUsers) * 100
	}

	// Build response
	return &dto.TierDistributionDTO{
		TotalUsers: totalUsers,
		Distribution: map[string]struct {
			Count      int     `json:"count"`
			Percentage float64 `json:"percentage"`
		}{
			"no_tier": {
				Count:      noTierCount,
				Percentage: calculatePercentage(noTierCount),
			},
			string(entity.UserTierPendekar): {
				Count:      pendekarCount,
				Percentage: calculatePercentage(pendekarCount),
			},
			string(entity.UserTierTuanMuda): {
				Count:      tuanMudaCount,
				Percentage: calculatePercentage(tuanMudaCount),
			},
			string(entity.UserTierTuanBesar): {
				Count:      tuanBesarCount,
				Percentage: calculatePercentage(tuanBesarCount),
			},
			string(entity.UserTierTuanRaja): {
				Count:      tuanRajaCount,
				Percentage: calculatePercentage(tuanRajaCount),
			},
		},
	}, nil
}
