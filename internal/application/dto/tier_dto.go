package dto

// TierDistributionDTO represents the distribution of users across tiers
type TierDistributionDTO struct {
	PendekarCount  int `json:"pendekar_count"`
	TuanMudaCount  int `json:"tuan_muda_count"`
	TuanBesarCount int `json:"tuan_besar_count"`
	TuanRajaCount  int `json:"tuan_raja_count"`
	NoTierCount    int `json:"no_tier_count"`
	TotalUsers     int `json:"total_users"`
}
