package dto

// TierDistributionDTO represents the distribution of users across tiers
type TierDistributionDTO struct {
	BasicCount      int `json:"basic_count"`
	PremiumCount    int `json:"premium_count"`
	ProCount        int `json:"pro_count"`
	EnterpriseCount int `json:"enterprise_count"`
	NoTierCount     int `json:"no_tier_count"`
	TotalUsers      int `json:"total_users"`
}
