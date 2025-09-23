package dto

import (
	"time"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// UserProfileResponse represents the current user's profile information
type UserProfileResponse struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	Email            string          `json:"email"`
	Phone            string          `json:"phone"`
	Picture          string          `json:"picture"`
	UserType         entity.UserType `json:"user_type"`
	UserTier         entity.UserTier `json:"user_tier"`
	UserRole         entity.UserRole `json:"user_role"`
	TransactionCount int             `json:"transaction_count"`
	WalletBalance    float64         `json:"wallet_balance"`
	WalletID         string          `json:"wallet_id"`
	IsAdmin          bool            `json:"is_admin"`
	AcceptTerms      bool            `json:"accept_terms"`
	AcceptPromos     bool            `json:"accept_promos"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// BuildUserProfileResponse converts a User entity and wallet info to UserProfileResponse
func BuildUserProfileResponse(user *entity.User, walletBalance float64, walletID string) UserProfileResponse {
	return UserProfileResponse{
		ID:               user.ID,
		Name:             user.Name,
		Email:            user.Email,
		Phone:            user.Phone,
		Picture:          user.Picture,
		UserType:         user.UserType,
		UserTier:         user.UserTier,
		UserRole:         user.UserRole,
		TransactionCount: user.TransactionCount,
		WalletBalance:    walletBalance,
		WalletID:         walletID,
		IsAdmin:          user.IsAdmin,
		AcceptTerms:      user.AcceptTerms,
		AcceptPromos:     user.AcceptPromos,
		CreatedAt:        user.CreatedAt,
		UpdatedAt:        user.UpdatedAt,
	}
}
