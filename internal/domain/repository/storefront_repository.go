package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// StorefrontRepository defines the interface for storefront data operations
type StorefrontRepository interface {
	// Core CRUD operations
	Create(ctx context.Context, storefront *entity.Storefront) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Storefront, error)
	GetBySellerID(ctx context.Context, sellerID uuid.UUID) ([]*entity.Storefront, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Storefront, error)
	GetByDomain(ctx context.Context, domain string) (*entity.Storefront, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*entity.Storefront, error)
	Update(ctx context.Context, storefront *entity.Storefront) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	HardDelete(ctx context.Context, id uuid.UUID) error

	// Business queries
	List(ctx context.Context, req *ListStorefrontsRequest) (*StorefrontListResponse, error)
	Search(ctx context.Context, req *SearchStorefrontsRequest) (*StorefrontSearchResult, error)
	GetActiveStorefronts(ctx context.Context, limit int) ([]*entity.Storefront, error)
	GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*StorefrontStats, error)

	// Slug and domain management
	IsSlugAvailable(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
	IsDomainAvailable(ctx context.Context, domain string, excludeID *uuid.UUID) (bool, error)
	IsSubdomainAvailable(ctx context.Context, subdomain string, excludeID *uuid.UUID) (bool, error)

	// Bulk operations
	GetStorefrontsByStatus(ctx context.Context, status entity.StorefrontStatus) ([]*entity.Storefront, error)
	UpdateStorefrontStatus(ctx context.Context, id uuid.UUID, status entity.StorefrontStatus) error
}

// Request/Response types for Storefront operations
type ListStorefrontsRequest struct {
	SellerID *uuid.UUID               `json:"seller_id,omitempty"`
	Status   *entity.StorefrontStatus `json:"status,omitempty"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
	OrderBy  string                   `json:"order_by"`
	SortDesc bool                     `json:"sort_desc"`
	Search   string                   `json:"search,omitempty"`
}

type StorefrontListResponse struct {
	Storefronts []*entity.Storefront `json:"storefronts"`
	Total       int                  `json:"total"`
	Page        int                  `json:"page"`
	PageSize    int                  `json:"page_size"`
	TotalPages  int                  `json:"total_pages"`
}

type SearchStorefrontsRequest struct {
	Query    string                   `json:"query"`
	SellerID *uuid.UUID               `json:"seller_id,omitempty"`
	Status   *entity.StorefrontStatus `json:"status,omitempty"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}

type StorefrontSearchResult struct {
	Storefronts []*entity.Storefront `json:"storefronts"`
	Total       int                  `json:"total"`
	Query       string               `json:"query"`
}

type StorefrontStats struct {
	CustomerCount  int        `json:"customer_count"`
	OrderCount     int        `json:"order_count"`
	TotalRevenue   float64    `json:"total_revenue"`
	AvgQueryTime   int64      `json:"avg_query_time_ms"`
	ActiveSessions int        `json:"active_sessions"`
	ConversionRate float64    `json:"conversion_rate"`
	LastOrderDate  *time.Time `json:"last_order_date,omitempty"`
	TopProducts    []struct {
		ProductID uuid.UUID `json:"product_id"`
		Name      string    `json:"name"`
		Sales     int       `json:"sales"`
	} `json:"top_products"`
}
