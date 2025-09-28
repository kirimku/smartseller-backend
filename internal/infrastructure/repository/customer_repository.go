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

// PostgreSQLCustomerRepository implements the CustomerRepository interface using PostgreSQL
type PostgreSQLCustomerRepository struct {
	*BaseRepository
	metricsCollector MetricsCollector
}

// NewPostgreSQLCustomerRepository creates a new PostgreSQL customer repository
func NewPostgreSQLCustomerRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	metricsCollector MetricsCollector,
) repository.CustomerRepository {
	if metricsCollector == nil {
		metricsCollector = &NoOpMetricsCollector{}
	}

	return &PostgreSQLCustomerRepository{
		BaseRepository:   NewBaseRepository(db, tenantResolver),
		metricsCollector: metricsCollector,
	}
}

// Create creates a new customer in the database
func (r *PostgreSQLCustomerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	return WithMetrics(r.metricsCollector, "CREATE", "customers", func() error {
		// Validate customer before creating
		if err := customer.Validate(); err != nil {
			return fmt.Errorf("customer validation failed: %w", err)
		}

		// Validate storefront access
		if err := r.ValidateStorefrontAccess(ctx, customer.StorefrontID); err != nil {
			return err
		}

		// Check for duplicate email/phone within storefront
		if customer.Email != nil && *customer.Email != "" {
			if err := r.ValidateUniqueEmail(ctx, customer.StorefrontID, *customer.Email, nil); err != nil {
				return err
			}
		}

		if customer.Phone != nil && *customer.Phone != "" {
			if err := r.ValidateUniquePhone(ctx, customer.StorefrontID, *customer.Phone, nil); err != nil {
				return err
			}
		}

		// Ensure ID is set
		if customer.ID == uuid.Nil {
			customer.ID = uuid.New()
		}

		// Set timestamps
		now := time.Now()
		customer.CreatedAt = now
		customer.UpdatedAt = now

		// Normalize data
		customer.NormalizeEmail()
		customer.NormalizePhone()

		// Set defaults if not provided
		if customer.Status == "" {
			customer.Status = entity.CustomerStatusActive
		}
		if customer.CustomerType == "" {
			customer.CustomerType = entity.CustomerTypeRegular
		}
		if customer.Preferences.Language == "" {
			customer.SetDefaultPreferences()
		}

		// Prepare query
		query := `
			INSERT INTO customers (
				id, storefront_id, email, phone, first_name, last_name, full_name,
				date_of_birth, gender, password_hash, email_verified_at, 
				email_verification_token, phone_verified_at, phone_verification_token,
				password_reset_token, password_reset_expires_at, refresh_token, 
				refresh_token_expires_at, last_login_at, failed_login_attempts,
				locked_until, status, customer_type, tags, preferences,
				accepts_marketing, marketing_opt_in_date, total_orders, total_spent,
				average_order_value, last_order_date, notes, internal_notes,
				created_by, created_at, updated_at
			) VALUES (
				:id, :storefront_id, :email, :phone, :first_name, :last_name, :full_name,
				:date_of_birth, :gender, :password_hash, :email_verified_at,
				:email_verification_token, :phone_verified_at, :phone_verification_token,
				:password_reset_token, :password_reset_expires_at, :refresh_token,
				:refresh_token_expires_at, :last_login_at, :failed_login_attempts,
				:locked_until, :status, :customer_type, :tags, :preferences,
				:accepts_marketing, :marketing_opt_in_date, :total_orders, :total_spent,
				:average_order_value, :last_order_date, :notes, :internal_notes,
				:created_by, :created_at, :updated_at
			)
		`

		// Get tenant-specific database
		db, err := r.GetDB(ctx, customer.StorefrontID)
		if err != nil {
			return err
		}

		_, err = db.NamedExecContext(ctx, query, customer)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505": // unique violation
					if strings.Contains(pqErr.Detail, "email") {
						return errors.ErrEmailAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "phone") {
						return errors.ErrPhoneAlreadyExists
					}
					return errors.ErrDuplicateCustomerData
				case "23503": // foreign key violation
					if strings.Contains(pqErr.Detail, "storefront") {
						return errors.ErrStorefrontNotFound
					}
				}
			}
			return fmt.Errorf("failed to create customer: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a customer by ID with tenant isolation
func (r *PostgreSQLCustomerRepository) GetByID(ctx context.Context, storefrontID, customerID uuid.UUID) (*entity.Customer, error) {
	var customer entity.Customer

	return &customer, WithMetrics(r.metricsCollector, "GET_BY_ID", "customers", func() error {
		// Get tenant context for query building
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			Select("*").
			From("customers").
			TenantWhere(storefrontID).
			Where("id = $1", customerID).
			Where("deleted_at IS NULL").
			Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		err = db.GetContext(ctx, &customer, query, args...)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrCustomerNotFound
			}
			return fmt.Errorf("failed to get customer by ID: %w", err)
		}

		return nil
	})
}

// GetByEmail retrieves a customer by email with tenant isolation
func (r *PostgreSQLCustomerRepository) GetByEmail(ctx context.Context, storefrontID uuid.UUID, email string) (*entity.Customer, error) {
	var customer entity.Customer

	return &customer, WithMetrics(r.metricsCollector, "GET_BY_EMAIL", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			Select("*").
			From("customers").
			TenantWhere(storefrontID).
			Where("LOWER(email) = LOWER($1)", email).
			Where("deleted_at IS NULL").
			Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		err = db.GetContext(ctx, &customer, query, args...)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrCustomerNotFound
			}
			return fmt.Errorf("failed to get customer by email: %w", err)
		}

		return nil
	})
}

// GetByPhone retrieves a customer by phone with tenant isolation
func (r *PostgreSQLCustomerRepository) GetByPhone(ctx context.Context, storefrontID uuid.UUID, phone string) (*entity.Customer, error) {
	var customer entity.Customer

	return &customer, WithMetrics(r.metricsCollector, "GET_BY_PHONE", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			Select("*").
			From("customers").
			TenantWhere(storefrontID).
			Where("phone = $1", phone).
			Where("deleted_at IS NULL").
			Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		err = db.GetContext(ctx, &customer, query, args...)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrCustomerNotFound
			}
			return fmt.Errorf("failed to get customer by phone: %w", err)
		}

		return nil
	})
}

// Update updates an existing customer
func (r *PostgreSQLCustomerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		// Validate customer
		if err := customer.Validate(); err != nil {
			return fmt.Errorf("customer validation failed: %w", err)
		}

		// Check if customer exists
		existing, err := r.GetByID(ctx, customer.StorefrontID, customer.ID)
		if err != nil {
			return err
		}

		// Check for unique email/phone (excluding current customer)
		if customer.Email != nil && *customer.Email != "" &&
			(existing.Email == nil || *existing.Email != *customer.Email) {
			if err := r.ValidateUniqueEmail(ctx, customer.StorefrontID, *customer.Email, &customer.ID); err != nil {
				return err
			}
		}

		if customer.Phone != nil && *customer.Phone != "" &&
			(existing.Phone == nil || *existing.Phone != *customer.Phone) {
			if err := r.ValidateUniquePhone(ctx, customer.StorefrontID, *customer.Phone, &customer.ID); err != nil {
				return err
			}
		}

		// Update timestamp
		customer.UpdatedAt = time.Now()

		// Normalize data
		customer.NormalizeEmail()
		customer.NormalizePhone()

		query := `
			UPDATE customers SET
				email = :email, phone = :phone, first_name = :first_name,
				last_name = :last_name, full_name = :full_name, date_of_birth = :date_of_birth,
				gender = :gender, password_hash = :password_hash, email_verified_at = :email_verified_at,
				email_verification_token = :email_verification_token, phone_verified_at = :phone_verified_at,
				phone_verification_token = :phone_verification_token, password_reset_token = :password_reset_token,
				password_reset_expires_at = :password_reset_expires_at, refresh_token = :refresh_token,
				refresh_token_expires_at = :refresh_token_expires_at, last_login_at = :last_login_at,
				failed_login_attempts = :failed_login_attempts, locked_until = :locked_until,
				status = :status, customer_type = :customer_type, tags = :tags,
				preferences = :preferences, accepts_marketing = :accepts_marketing,
				marketing_opt_in_date = :marketing_opt_in_date, total_orders = :total_orders,
				total_spent = :total_spent, average_order_value = :average_order_value,
				last_order_date = :last_order_date, notes = :notes, internal_notes = :internal_notes,
				updated_at = :updated_at
			WHERE id = :id AND storefront_id = :storefront_id AND deleted_at IS NULL
		`

		db, err := r.GetDB(ctx, customer.StorefrontID)
		if err != nil {
			return err
		}

		result, err := db.NamedExecContext(ctx, query, customer)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505": // unique violation
					if strings.Contains(pqErr.Detail, "email") {
						return errors.ErrEmailAlreadyExists
					}
					if strings.Contains(pqErr.Detail, "phone") {
						return errors.ErrPhoneAlreadyExists
					}
				}
			}
			return fmt.Errorf("failed to update customer: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

// SoftDelete marks a customer as deleted
func (r *PostgreSQLCustomerRepository) SoftDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "SOFT_DELETE", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			From("customers").
			TenantWhere(storefrontID).
			Where("id = $1", customerID).
			Where("deleted_at IS NULL").
			Build()

		// Rebuild as UPDATE
		updateQuery := strings.Replace(query, "SELECT * FROM", "UPDATE", 1)
		updateQuery = strings.Replace(updateQuery, " WHERE", " SET deleted_at = NOW() WHERE", 1)

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, updateQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to soft delete customer: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

// HardDelete permanently deletes a customer
func (r *PostgreSQLCustomerRepository) HardDelete(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "HARD_DELETE", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			From("customers").
			TenantWhere(storefrontID).
			Where("id = $1", customerID).
			Build()

		// Rebuild as DELETE
		deleteQuery := strings.Replace(query, "SELECT * FROM", "DELETE FROM", 1)

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, deleteQuery, args...)
		if err != nil {
			return fmt.Errorf("failed to hard delete customer: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

// GetByStorefront retrieves customers for a storefront with filtering and pagination
func (r *PostgreSQLCustomerRepository) GetByStorefront(ctx context.Context, req *repository.GetCustomersRequest) (*repository.CustomerListResponse, error) {
	var customers []*entity.Customer
	var total int

	err := WithMetrics(r.metricsCollector, "GET_BY_STOREFRONT", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, req.StorefrontID)
		if err != nil {
			return err
		}

		db, err := r.GetDB(ctx, req.StorefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		qb.From("customers").
			TenantWhere(req.StorefrontID).
			Where("deleted_at IS NULL")

		// Apply filters
		if req.Search != "" {
			searchPattern := "%" + req.Search + "%"
			qb.Where("(LOWER(first_name) LIKE LOWER($1) OR LOWER(last_name) LIKE LOWER($1) OR LOWER(email) LIKE LOWER($1) OR phone LIKE $1)", searchPattern)
		}

		if req.Status != nil {
			qb.Where("status = $1", *req.Status)
		}

		if req.CustomerType != nil {
			qb.Where("customer_type = $1", *req.CustomerType)
		}

		if len(req.Tags) > 0 {
			qb.Where("tags && $1", pq.Array(req.Tags))
		}

		if req.DateFrom != nil {
			qb.Where("created_at >= $1", *req.DateFrom)
		}

		if req.DateTo != nil {
			qb.Where("created_at <= $1", *req.DateTo)
		}

		if req.HasEmail != nil {
			if *req.HasEmail {
				qb.Where("email IS NOT NULL AND email != ''")
			} else {
				qb.Where("email IS NULL OR email = ''")
			}
		}

		if req.HasPhone != nil {
			if *req.HasPhone {
				qb.Where("phone IS NOT NULL AND phone != ''")
			} else {
				qb.Where("phone IS NULL OR phone = ''")
			}
		}

		if req.IsVerified != nil {
			if *req.IsVerified {
				qb.Where("email_verified_at IS NOT NULL")
			} else {
				qb.Where("email_verified_at IS NULL")
			}
		}

		// Get total count
		countQuery, countArgs := qb.BuildCount()
		err = db.GetContext(ctx, &total, countQuery, countArgs...)
		if err != nil {
			return fmt.Errorf("failed to get customer count: %w", err)
		}

		// Apply sorting
		orderBy := "created_at"
		if req.OrderBy != "" {
			orderBy = req.OrderBy
		}
		direction := "DESC"
		if !req.SortDesc {
			direction = "ASC"
		}
		qb.OrderBy(orderBy, direction)

		// Apply pagination
		pagination := ValidatePagination(req.Page, req.PageSize)
		qb.Limit(pagination.GetLimit()).Offset(pagination.GetOffset())

		// Execute query
		query, args := qb.Select("*").Build()
		err = db.SelectContext(ctx, &customers, query, args...)
		if err != nil {
			return fmt.Errorf("failed to get customers: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	pagination := ValidatePagination(req.Page, req.PageSize)
	return &repository.CustomerListResponse{
		Customers:  customers,
		Total:      total,
		Page:       pagination.Page,
		PageSize:   pagination.PageSize,
		TotalPages: pagination.CalculateTotalPages(total),
	}, nil
}

// ValidateUniqueEmail checks if email is unique within a storefront
func (r *PostgreSQLCustomerRepository) ValidateUniqueEmail(ctx context.Context, storefrontID uuid.UUID, email string, excludeCustomerID *uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "VALIDATE_UNIQUE_EMAIL", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		qb.Select("id").
			From("customers").
			TenantWhere(storefrontID).
			Where("LOWER(email) = LOWER($1)", email).
			Where("deleted_at IS NULL")

		if excludeCustomerID != nil {
			qb.Where("id != $1", *excludeCustomerID)
		}

		query, args := qb.Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		var existingID uuid.UUID
		err = db.GetContext(ctx, &existingID, query, args...)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to validate unique email: %w", err)
		}

		if err != sql.ErrNoRows {
			return errors.ErrEmailAlreadyExists
		}

		return nil
	})
}

// ValidateUniquePhone checks if phone is unique within a storefront
func (r *PostgreSQLCustomerRepository) ValidateUniquePhone(ctx context.Context, storefrontID uuid.UUID, phone string, excludeCustomerID *uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "VALIDATE_UNIQUE_PHONE", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		qb.Select("id").
			From("customers").
			TenantWhere(storefrontID).
			Where("phone = $1", phone).
			Where("deleted_at IS NULL")

		if excludeCustomerID != nil {
			qb.Where("id != $1", *excludeCustomerID)
		}

		query, args := qb.Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		var existingID uuid.UUID
		err = db.GetContext(ctx, &existingID, query, args...)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to validate unique phone: %w", err)
		}

		if err != sql.ErrNoRows {
			return errors.ErrPhoneAlreadyExists
		}

		return nil
	})
}

// Search searches customers with full-text search capabilities
func (r *PostgreSQLCustomerRepository) Search(ctx context.Context, storefrontID uuid.UUID, req *repository.SearchCustomersRequest) (*repository.CustomerSearchResult, error) {
	// Implementation similar to GetByStorefront but with full-text search
	// This is a simplified version - full implementation would use PostgreSQL's full-text search
	getReq := &repository.GetCustomersRequest{
		StorefrontID: storefrontID,
		Search:       req.Query,
		Status:       req.Status,
		CustomerType: req.CustomerType,
		Page:         req.Page,
		PageSize:     req.PageSize,
		OrderBy:      "created_at",
		SortDesc:     true,
	}

	result, err := r.GetByStorefront(ctx, getReq)
	if err != nil {
		return nil, err
	}

	return &repository.CustomerSearchResult{
		Customers: result.Customers,
		Total:     result.Total,
		Query:     req.Query,
	}, nil
}

// GetTopCustomers gets top customers by spending
func (r *PostgreSQLCustomerRepository) GetTopCustomers(ctx context.Context, storefrontID uuid.UUID, limit int) ([]*entity.Customer, error) {
	var customers []*entity.Customer

	err := WithMetrics(r.metricsCollector, "GET_TOP_CUSTOMERS", "customers", func() error {
		tenantCtx, err := r.GetTenantContext(ctx, storefrontID)
		if err != nil {
			return err
		}

		qb := NewQueryBuilder(tenantCtx)
		query, args := qb.
			Select("*").
			From("customers").
			TenantWhere(storefrontID).
			Where("deleted_at IS NULL").
			Where("total_spent > 0").
			OrderBy("total_spent", "DESC").
			Limit(limit).
			Build()

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		err = db.SelectContext(ctx, &customers, query, args...)
		if err != nil {
			return fmt.Errorf("failed to get top customers: %w", err)
		}

		return nil
	})

	return customers, err
}

// GetCustomerStats retrieves customer statistics for a storefront
func (r *PostgreSQLCustomerRepository) GetCustomerStats(ctx context.Context, storefrontID uuid.UUID) (*repository.CustomerStats, error) {
	var stats repository.CustomerStats

	err := WithMetrics(r.metricsCollector, "GET_CUSTOMER_STATS", "customers", func() error {
		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		// Basic stats query
		query := `
			SELECT 
				COUNT(*) as total_customers,
				COUNT(*) FILTER (WHERE status = 'active') as active_customers,
				COUNT(*) FILTER (WHERE email_verified_at IS NOT NULL) as verified_customers,
				COUNT(*) FILTER (WHERE created_at >= date_trunc('month', CURRENT_DATE)) as new_this_month,
				COUNT(*) FILTER (WHERE created_at >= date_trunc('week', CURRENT_DATE)) as new_this_week,
				COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE) as new_today,
				COALESCE(SUM(total_spent), 0) as total_revenue,
				COALESCE(AVG(average_order_value), 0) as avg_order_value
			FROM customers 
			WHERE storefront_id = $1 AND deleted_at IS NULL
		`

		err = db.GetContext(ctx, &stats, query, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to get customer stats: %w", err)
		}

		// Initialize maps
		stats.CustomersByType = make(map[entity.CustomerType]int)
		stats.CustomersByStatus = make(map[entity.CustomerStatus]int)

		return nil
	})

	return &stats, err
}

// Authentication methods
func (r *PostgreSQLCustomerRepository) GetByEmailVerificationToken(ctx context.Context, token string) (*entity.Customer, error) {
	var customer entity.Customer

	return &customer, WithMetrics(r.metricsCollector, "GET_BY_EMAIL_TOKEN", "customers", func() error {
		query := `
			SELECT * FROM customers 
			WHERE email_verification_token = $1 AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &customer, query, token)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrInvalidEmailToken
			}
			return fmt.Errorf("failed to get customer by email token: %w", err)
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) GetByPasswordResetToken(ctx context.Context, token string) (*entity.Customer, error) {
	var customer entity.Customer

	return &customer, WithMetrics(r.metricsCollector, "GET_BY_PASSWORD_TOKEN", "customers", func() error {
		query := `
			SELECT * FROM customers 
			WHERE password_reset_token = $1 
			AND password_reset_expires_at > NOW() 
			AND deleted_at IS NULL
		`

		err := r.db.GetContext(ctx, &customer, query, token)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrInvalidPasswordToken
			}
			return fmt.Errorf("failed to get customer by password token: %w", err)
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) UpdateLastLogin(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	return r.ExecuteInTransaction(ctx, storefrontID, func(tx *sqlx.Tx) error {
		query := `
			UPDATE customers 
			SET last_login_at = NOW(), failed_login_attempts = 0, updated_at = NOW()
			WHERE id = $1 AND storefront_id = $2 AND deleted_at IS NULL
		`

		result, err := tx.ExecContext(ctx, query, customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to update last login: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) UpdateRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string, expiresAt *time.Time) error {
	return r.ExecuteInTransaction(ctx, storefrontID, func(tx *sqlx.Tx) error {
		query := `
			UPDATE customers 
			SET refresh_token = $1, refresh_token_expires_at = $2, updated_at = NOW()
			WHERE id = $3 AND storefront_id = $4 AND deleted_at IS NULL
		`

		result, err := tx.ExecContext(ctx, query, token, expiresAt, customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to update refresh token: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

// Placeholder implementations for missing methods - these would need full implementation
func (r *PostgreSQLCustomerRepository) ClearRefreshToken(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	return r.UpdateRefreshToken(ctx, storefrontID, customerID, "", nil)
}

func (r *PostgreSQLCustomerRepository) UpdateEmailVerification(ctx context.Context, storefrontID, customerID uuid.UUID, verified bool) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		var emailVerifiedAt *time.Time
		if verified {
			now := time.Now()
			emailVerifiedAt = &now
		}

		query := `
			UPDATE customers 
			SET email_verified_at = $1, email_verification_token = NULL, updated_at = $2
			WHERE id = $3 AND storefront_id = $4 AND deleted_at IS NULL`

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, query, emailVerifiedAt, time.Now(), customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to update email verification: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) UpdatePhoneVerification(ctx context.Context, storefrontID, customerID uuid.UUID, verified bool) error {
	// Implementation would update phone_verified_at field
	return nil
}

func (r *PostgreSQLCustomerRepository) UpdateFailedLoginAttempts(ctx context.Context, storefrontID, customerID uuid.UUID, attempts int) error {
	// Implementation would update failed_login_attempts field
	return nil
}

func (r *PostgreSQLCustomerRepository) LockAccount(ctx context.Context, storefrontID, customerID uuid.UUID, until *time.Time) error {
	// Implementation would set locked_until field
	return nil
}

func (r *PostgreSQLCustomerRepository) UnlockAccount(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	// Implementation would clear locked_until field
	return nil
}

func (r *PostgreSQLCustomerRepository) UpdatePassword(ctx context.Context, storefrontID, customerID uuid.UUID, passwordHash string) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		query := `
			UPDATE customers 
			SET password_hash = $1, updated_at = $2
			WHERE id = $3 AND storefront_id = $4 AND deleted_at IS NULL`

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, query, passwordHash, time.Now(), customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to update password: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) SetPasswordResetToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string, expiresAt time.Time) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		query := `
			UPDATE customers 
			SET password_reset_token = $1, password_reset_expires_at = $2, updated_at = $3
			WHERE id = $4 AND storefront_id = $5 AND deleted_at IS NULL`

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, query, token, expiresAt, time.Now(), customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to set password reset token: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) ClearPasswordResetToken(ctx context.Context, storefrontID, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		query := `
			UPDATE customers 
			SET password_reset_token = NULL, password_reset_expires_at = NULL, updated_at = $1
			WHERE id = $2 AND storefront_id = $3 AND deleted_at IS NULL`

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, query, time.Now(), customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to clear password reset token: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) SetEmailVerificationToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customers", func() error {
		query := `
			UPDATE customers 
			SET email_verification_token = $1, updated_at = $2
			WHERE id = $3 AND storefront_id = $4 AND deleted_at IS NULL`

		db, err := r.GetDB(ctx, storefrontID)
		if err != nil {
			return err
		}

		result, err := db.ExecContext(ctx, query, token, time.Now(), customerID, storefrontID)
		if err != nil {
			return fmt.Errorf("failed to set email verification token: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrCustomerNotFound
		}

		return nil
	})
}

func (r *PostgreSQLCustomerRepository) SetPhoneVerificationToken(ctx context.Context, storefrontID, customerID uuid.UUID, token string) error {
	// Implementation would set phone verification token
	return nil
}

func (r *PostgreSQLCustomerRepository) GetCustomerSegments(ctx context.Context, storefrontID uuid.UUID) ([]*repository.CustomerSegment, error) {
	// Implementation would return customer segments based on various criteria
	return []*repository.CustomerSegment{}, nil
}

func (r *PostgreSQLCustomerRepository) UpdateCustomerMetrics(ctx context.Context, storefrontID, customerID uuid.UUID, metrics repository.CustomerMetricsUpdate) error {
	// Implementation would update customer metrics
	return nil
}

func (r *PostgreSQLCustomerRepository) GetCustomerActivity(ctx context.Context, storefrontID, customerID uuid.UUID, limit int) ([]*repository.CustomerActivity, error) {
	// Implementation would return customer activity log
	return []*repository.CustomerActivity{}, nil
}

func (r *PostgreSQLCustomerRepository) GetCustomersByIDs(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID) ([]*entity.Customer, error) {
	// Implementation would get customers by list of IDs
	return []*entity.Customer{}, nil
}

func (r *PostgreSQLCustomerRepository) GetCustomersByStatus(ctx context.Context, storefrontID uuid.UUID, status entity.CustomerStatus) ([]*entity.Customer, error) {
	// Implementation would get customers by status
	return []*entity.Customer{}, nil
}

func (r *PostgreSQLCustomerRepository) GetCustomersByType(ctx context.Context, storefrontID uuid.UUID, customerType entity.CustomerType) ([]*entity.Customer, error) {
	// Implementation would get customers by type
	return []*entity.Customer{}, nil
}

func (r *PostgreSQLCustomerRepository) GetCustomersWithTags(ctx context.Context, storefrontID uuid.UUID, tags []string) ([]*entity.Customer, error) {
	// Implementation would get customers with specific tags
	return []*entity.Customer{}, nil
}

func (r *PostgreSQLCustomerRepository) BulkUpdateStatus(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, status entity.CustomerStatus) error {
	// Implementation would bulk update customer status
	return nil
}

func (r *PostgreSQLCustomerRepository) BulkAddTags(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, tags []string) error {
	// Implementation would bulk add tags to customers
	return nil
}

func (r *PostgreSQLCustomerRepository) BulkRemoveTags(ctx context.Context, storefrontID uuid.UUID, customerIDs []uuid.UUID, tags []string) error {
	// Implementation would bulk remove tags from customers
	return nil
}

func (r *PostgreSQLCustomerRepository) CleanupExpiredTokens(ctx context.Context) (int, error) {
	// Implementation would clean up expired tokens across all storefronts
	return 0, nil
}

func (r *PostgreSQLCustomerRepository) CleanupExpiredSessions(ctx context.Context) (int, error) {
	// Implementation would clean up expired sessions across all storefronts
	return 0, nil
}
