package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/errors"
	"github.com/kirimku/smartseller-backend/internal/domain/repository"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// PostgreSQLStorefrontRepository implements the StorefrontRepository interface using PostgreSQL
type PostgreSQLStorefrontRepository struct {
	*BaseRepository
	metricsCollector MetricsCollector
}

// NewPostgreSQLStorefrontRepository creates a new PostgreSQL storefront repository
func NewPostgreSQLStorefrontRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	metricsCollector MetricsCollector,
) repository.StorefrontRepository {
	if metricsCollector == nil {
		metricsCollector = &NoOpMetricsCollector{}
	}

	return &PostgreSQLStorefrontRepository{
		BaseRepository:   NewBaseRepository(db, tenantResolver),
		metricsCollector: metricsCollector,
	}
}

// Create creates a new storefront in the database
func (r *PostgreSQLStorefrontRepository) Create(ctx context.Context, storefront *entity.Storefront) error {
	return WithMetrics(r.metricsCollector, "CREATE", "storefronts", func() error {
		// Validate storefront before creating
		if err := storefront.Validate(); err != nil {
			return fmt.Errorf("storefront validation failed: %w", err)
		}

		// Check for duplicate slug
		if available, err := r.IsSlugAvailable(ctx, storefront.Slug, nil); err != nil {
			return err
		} else if !available {
			return errors.ErrSlugAlreadyExists
		}

		// Check for duplicate domain if provided
		if storefront.Domain != nil && *storefront.Domain != "" {
			if available, err := r.IsDomainAvailable(ctx, *storefront.Domain, nil); err != nil {
				return err
			} else if !available {
				return errors.ErrDomainAlreadyExists
			}
		}

		// Check for duplicate subdomain if provided
		if storefront.Subdomain != nil && *storefront.Subdomain != "" {
			if available, err := r.IsSubdomainAvailable(ctx, *storefront.Subdomain, nil); err != nil {
				return err
			} else if !available {
				return errors.ErrSubdomainAlreadyExists
			}
		}

		// Ensure ID is set
		if storefront.ID == uuid.Nil {
			storefront.ID = uuid.New()
		}

		// Set timestamps
		now := time.Now()
		storefront.CreatedAt = now
		storefront.UpdatedAt = now

		// Normalize slug
		storefront.NormalizeSlug()

		// Set default settings if empty
		if storefront.Settings.Currency == "" {
			storefront.SetDefaultSettings()
		}

		// Set default status if empty
		if storefront.Status == "" {
			storefront.Status = entity.StorefrontStatusActive
		}

		query := `
			INSERT INTO storefronts (
				id, seller_id, name, slug, description, domain, subdomain,
				status, settings, logo_url, favicon_url, primary_color,
				secondary_color, business_name, business_email, business_phone,
				business_address, tax_id, created_at, updated_at
			) VALUES (
				:id, :seller_id, :name, :slug, :description, :domain, :subdomain,
				:status, :settings, :logo_url, :favicon_url, :primary_color,
				:secondary_color, :business_name, :business_email, :business_phone,
				:business_address, :tax_id, :created_at, :updated_at
			)
		`

		_, err := r.db.NamedExecContext(ctx, query, storefront)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505": // unique violation
					if strings.Contains(pqErr.Detail, "slug") {
						return errors.ErrSlugAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "domain") {
						return errors.ErrDomainAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "subdomain") {
						return errors.ErrSubdomainAlreadyExists
					}
				case "23503": // foreign key violation
					if strings.Contains(pqErr.Detail, "seller") {
						return fmt.Errorf("seller not found")
					}
				}
			}
			return fmt.Errorf("failed to create storefront: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a storefront by ID
func (r *PostgreSQLStorefrontRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Storefront, error) {
	var storefront entity.Storefront

	return &storefront, WithMetrics(r.metricsCollector, "GET_BY_ID", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE id = $1 AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &storefront, query, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrStorefrontNotFound
			}
			return fmt.Errorf("failed to get storefront by ID: %w", err)
		}

		return nil
	})
}

// GetBySellerID retrieves all storefronts for a seller
func (r *PostgreSQLStorefrontRepository) GetBySellerID(ctx context.Context, sellerID uuid.UUID) ([]*entity.Storefront, error) {
	var storefronts []*entity.Storefront

	err := WithMetrics(r.metricsCollector, "GET_BY_SELLER_ID", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE seller_id = $1 AND deleted_at IS NULL 
			ORDER BY created_at DESC
		`

		err := r.db.SelectContext(ctx, &storefronts, query, sellerID)
		if err != nil {
			return fmt.Errorf("failed to get storefronts by seller ID: %w", err)
		}

		return nil
	})

	return storefronts, err
}

// GetBySlug retrieves a storefront by slug
func (r *PostgreSQLStorefrontRepository) GetBySlug(ctx context.Context, slug string) (*entity.Storefront, error) {
	var storefront entity.Storefront

	return &storefront, WithMetrics(r.metricsCollector, "GET_BY_SLUG", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE slug = $1 AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &storefront, query, slug)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrStorefrontNotFound
			}
			return fmt.Errorf("failed to get storefront by slug: %w", err)
		}

		return nil
	})
}

// GetByDomain retrieves a storefront by domain
func (r *PostgreSQLStorefrontRepository) GetByDomain(ctx context.Context, domain string) (*entity.Storefront, error) {
	var storefront entity.Storefront

	return &storefront, WithMetrics(r.metricsCollector, "GET_BY_DOMAIN", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE domain = $1 AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &storefront, query, domain)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrStorefrontNotFound
			}
			return fmt.Errorf("failed to get storefront by domain: %w", err)
		}

		return nil
	})
}

// GetBySubdomain retrieves a storefront by subdomain
func (r *PostgreSQLStorefrontRepository) GetBySubdomain(ctx context.Context, subdomain string) (*entity.Storefront, error) {
	var storefront entity.Storefront

	return &storefront, WithMetrics(r.metricsCollector, "GET_BY_SUBDOMAIN", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE subdomain = $1 AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &storefront, query, subdomain)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrStorefrontNotFound
			}
			return fmt.Errorf("failed to get storefront by subdomain: %w", err)
		}

		return nil
	})
}

// Update updates an existing storefront
func (r *PostgreSQLStorefrontRepository) Update(ctx context.Context, storefront *entity.Storefront) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "storefronts", func() error {
		// Validate storefront
		if err := storefront.Validate(); err != nil {
			return fmt.Errorf("storefront validation failed: %w", err)
		}

		// Check if storefront exists
		existing, err := r.GetByID(ctx, storefront.ID)
		if err != nil {
			return err
		}

		// Check for unique slug (excluding current storefront)
		if existing.Slug != storefront.Slug {
			if available, err := r.IsSlugAvailable(ctx, storefront.Slug, &storefront.ID); err != nil {
				return err
			} else if !available {
				return errors.ErrSlugAlreadyExists
			}
		}

		// Check for unique domain
		if (existing.Domain == nil && storefront.Domain != nil) ||
			(existing.Domain != nil && storefront.Domain != nil && *existing.Domain != *storefront.Domain) {
			if storefront.Domain != nil && *storefront.Domain != "" {
				if available, err := r.IsDomainAvailable(ctx, *storefront.Domain, &storefront.ID); err != nil {
					return err
				} else if !available {
					return errors.ErrDomainAlreadyExists
				}
			}
		}

		// Check for unique subdomain
		if (existing.Subdomain == nil && storefront.Subdomain != nil) ||
			(existing.Subdomain != nil && storefront.Subdomain != nil && *existing.Subdomain != *storefront.Subdomain) {
			if storefront.Subdomain != nil && *storefront.Subdomain != "" {
				if available, err := r.IsSubdomainAvailable(ctx, *storefront.Subdomain, &storefront.ID); err != nil {
					return err
				} else if !available {
					return errors.ErrSubdomainAlreadyExists
				}
			}
		}

		// Update timestamp
		storefront.UpdatedAt = time.Now()

		// Normalize slug
		storefront.NormalizeSlug()

		query := `
			UPDATE storefronts SET
				name = :name, slug = :slug, description = :description,
				domain = :domain, subdomain = :subdomain, status = :status,
				settings = :settings, logo_url = :logo_url, favicon_url = :favicon_url,
				primary_color = :primary_color, secondary_color = :secondary_color,
				business_name = :business_name, business_email = :business_email,
				business_phone = :business_phone, business_address = :business_address,
				tax_id = :tax_id, updated_at = :updated_at
			WHERE id = :id AND deleted_at IS NULL
		`

		result, err := r.db.NamedExecContext(ctx, query, storefront)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505": // unique violation
					if strings.Contains(pqErr.Detail, "slug") {
						return errors.ErrSlugAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "domain") {
						return errors.ErrDomainAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "subdomain") {
						return errors.ErrSubdomainAlreadyExists
					}
				}
			}
			return fmt.Errorf("failed to update storefront: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrStorefrontNotFound
		}

		return nil
	})
}

// SoftDelete marks a storefront as deleted
func (r *PostgreSQLStorefrontRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "SOFT_DELETE", "storefronts", func() error {
		query := `
			UPDATE storefronts 
			SET deleted_at = NOW(), updated_at = NOW()
			WHERE id = $1 AND deleted_at IS NULL
		`

		result, err := r.db.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to soft delete storefront: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrStorefrontNotFound
		}

		return nil
	})
}

// HardDelete permanently deletes a storefront
func (r *PostgreSQLStorefrontRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "HARD_DELETE", "storefronts", func() error {
		query := `DELETE FROM storefronts WHERE id = $1`

		result, err := r.db.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to hard delete storefront: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrStorefrontNotFound
		}

		return nil
	})
}

// List retrieves storefronts with filtering and pagination
func (r *PostgreSQLStorefrontRepository) List(ctx context.Context, req *repository.ListStorefrontsRequest) (*repository.StorefrontListResponse, error) {
	var storefronts []*entity.Storefront
	var total int

	err := WithMetrics(r.metricsCollector, "LIST", "storefronts", func() error {
		// Build base query
		var whereClauses []string
		var args []interface{}
		argCounter := 0

		whereClauses = append(whereClauses, "deleted_at IS NULL")

		// Apply filters
		if req.SellerID != nil {
			argCounter++
			whereClauses = append(whereClauses, fmt.Sprintf("seller_id = $%d", argCounter))
			args = append(args, *req.SellerID)
		}

		if req.Status != nil {
			argCounter++
			whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCounter))
			args = append(args, *req.Status)
		}

		if req.Search != "" {
			argCounter++
			searchPattern := "%" + req.Search + "%"
			whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(name) LIKE LOWER($%d) OR LOWER(slug) LIKE LOWER($%d) OR LOWER(business_name) LIKE LOWER($%d))", argCounter, argCounter, argCounter))
			args = append(args, searchPattern)
		}

		whereClause := ""
		if len(whereClauses) > 0 {
			whereClause = " WHERE " + strings.Join(whereClauses, " AND ")
		}

		// Get total count
		countQuery := "SELECT COUNT(*) FROM storefronts" + whereClause
		err := r.db.GetContext(ctx, &total, countQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to get storefront count: %w", err)
		}

		// Build main query with pagination and sorting
		orderBy := "created_at"
		if req.OrderBy != "" {
			orderBy = req.OrderBy
		}
		direction := "DESC"
		if !req.SortDesc {
			direction = "ASC"
		}

		pagination := ValidatePagination(req.Page, req.PageSize)

		mainQuery := fmt.Sprintf(
			"SELECT * FROM storefronts%s ORDER BY %s %s LIMIT %d OFFSET %d",
			whereClause, orderBy, direction, pagination.GetLimit(), pagination.GetOffset(),
		)

		err = r.db.SelectContext(ctx, &storefronts, mainQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to get storefronts: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	pagination := ValidatePagination(req.Page, req.PageSize)
	return &repository.StorefrontListResponse{
		Storefronts: storefronts,
		Total:       total,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalPages:  pagination.CalculateTotalPages(total),
	}, nil
}

// Search searches storefronts with full-text capabilities
func (r *PostgreSQLStorefrontRepository) Search(ctx context.Context, req *repository.SearchStorefrontsRequest) (*repository.StorefrontSearchResult, error) {
	// Convert search request to list request
	listReq := &repository.ListStorefrontsRequest{
		SellerID: req.SellerID,
		Status:   req.Status,
		Search:   req.Query,
		Page:     req.Page,
		PageSize: req.PageSize,
		OrderBy:  "name",
		SortDesc: false,
	}

	result, err := r.List(ctx, listReq)
	if err != nil {
		return nil, err
	}

	return &repository.StorefrontSearchResult{
		Storefronts: result.Storefronts,
		Total:       result.Total,
		Query:       req.Query,
	}, nil
}

// GetActiveStorefronts retrieves active storefronts
func (r *PostgreSQLStorefrontRepository) GetActiveStorefronts(ctx context.Context, limit int) ([]*entity.Storefront, error) {
	var storefronts []*entity.Storefront

	err := WithMetrics(r.metricsCollector, "GET_ACTIVE", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE status = $1 AND deleted_at IS NULL 
			ORDER BY created_at DESC 
			LIMIT $2
		`

		err := r.db.SelectContext(ctx, &storefronts, query, entity.StorefrontStatusActive, limit)
		if err != nil {
			return fmt.Errorf("failed to get active storefronts: %w", err)
		}

		return nil
	})

	return storefronts, err
}

// GetStorefrontStats retrieves comprehensive statistics for a storefront
func (r *PostgreSQLStorefrontRepository) GetStorefrontStats(ctx context.Context, storefrontID uuid.UUID) (*repository.StorefrontStats, error) {
	var stats repository.StorefrontStats

	err := WithMetrics(r.metricsCollector, "GET_STATS", "storefronts", func() error {
		// Basic customer stats
		customerQuery := `
			SELECT 
				COUNT(*) as customer_count,
				COUNT(*) FILTER (WHERE created_at >= date_trunc('month', CURRENT_DATE)) as new_customers_this_month
			FROM customers 
			WHERE storefront_id = $1 AND deleted_at IS NULL
		`

		var customerCount, newCustomersThisMonth int
		err := r.db.QueryRowContext(ctx, customerQuery, storefrontID).Scan(&customerCount, &newCustomersThisMonth)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to get customer stats: %w", err)
		}

		stats.CustomerCount = customerCount

		// Placeholder for other stats - would need to implement based on actual business requirements
		stats.OrderCount = 0       // Would query from orders table
		stats.TotalRevenue = 0.0   // Would sum from orders
		stats.AvgQueryTime = 50    // Would collect from metrics
		stats.ActiveSessions = 0   // Would query from customer_sessions
		stats.ConversionRate = 0.0 // Would calculate from metrics

		return nil
	})

	return &stats, err
}

// IsSlugAvailable checks if a slug is available
func (r *PostgreSQLStorefrontRepository) IsSlugAvailable(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	var exists bool

	err := WithMetrics(r.metricsCollector, "CHECK_SLUG", "storefronts", func() error {
		query := `SELECT EXISTS(SELECT 1 FROM storefronts WHERE slug = $1 AND deleted_at IS NULL`
		args := []interface{}{slug}

		if excludeID != nil {
			query += ` AND id != $2`
			args = append(args, *excludeID)
		}

		query += `)`

		err := r.db.GetContext(ctx, &exists, query, args...)
		if err != nil {
			return fmt.Errorf("failed to check slug availability: %w", err)
		}

		return nil
	})

	return !exists, err
}

// IsDomainAvailable checks if a domain is available
func (r *PostgreSQLStorefrontRepository) IsDomainAvailable(ctx context.Context, domain string, excludeID *uuid.UUID) (bool, error) {
	var exists bool

	err := WithMetrics(r.metricsCollector, "CHECK_DOMAIN", "storefronts", func() error {
		query := `SELECT EXISTS(SELECT 1 FROM storefronts WHERE domain = $1 AND deleted_at IS NULL`
		args := []interface{}{domain}

		if excludeID != nil {
			query += ` AND id != $2`
			args = append(args, *excludeID)
		}

		query += `)`

		err := r.db.GetContext(ctx, &exists, query, args...)
		if err != nil {
			return fmt.Errorf("failed to check domain availability: %w", err)
		}

		return nil
	})

	return !exists, err
}

// IsSubdomainAvailable checks if a subdomain is available
func (r *PostgreSQLStorefrontRepository) IsSubdomainAvailable(ctx context.Context, subdomain string, excludeID *uuid.UUID) (bool, error) {
	var exists bool

	err := WithMetrics(r.metricsCollector, "CHECK_SUBDOMAIN", "storefronts", func() error {
		query := `SELECT EXISTS(SELECT 1 FROM storefronts WHERE subdomain = $1 AND deleted_at IS NULL`
		args := []interface{}{subdomain}

		if excludeID != nil {
			query += ` AND id != $2`
			args = append(args, *excludeID)
		}

		query += `)`

		err := r.db.GetContext(ctx, &exists, query, args...)
		if err != nil {
			return fmt.Errorf("failed to check subdomain availability: %w", err)
		}

		return nil
	})

	return !exists, err
}

// GetStorefrontsByStatus retrieves storefronts by status
func (r *PostgreSQLStorefrontRepository) GetStorefrontsByStatus(ctx context.Context, status entity.StorefrontStatus) ([]*entity.Storefront, error) {
	var storefronts []*entity.Storefront

	err := WithMetrics(r.metricsCollector, "GET_BY_STATUS", "storefronts", func() error {
		query := `
			SELECT * FROM storefronts 
			WHERE status = $1 AND deleted_at IS NULL 
			ORDER BY created_at DESC
		`

		err := r.db.SelectContext(ctx, &storefronts, query, status)
		if err != nil {
			return fmt.Errorf("failed to get storefronts by status: %w", err)
		}

		return nil
	})

	return storefronts, err
}

// UpdateStorefrontStatus updates the status of a storefront
func (r *PostgreSQLStorefrontRepository) UpdateStorefrontStatus(ctx context.Context, id uuid.UUID, status entity.StorefrontStatus) error {
	return WithMetrics(r.metricsCollector, "UPDATE_STATUS", "storefronts", func() error {
		query := `
			UPDATE storefronts 
			SET status = $1, updated_at = NOW() 
			WHERE id = $2 AND deleted_at IS NULL
		`

		result, err := r.db.ExecContext(ctx, query, status, id)
		if err != nil {
			return fmt.Errorf("failed to update storefront status: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrStorefrontNotFound
		}

		return nil
	})
}
