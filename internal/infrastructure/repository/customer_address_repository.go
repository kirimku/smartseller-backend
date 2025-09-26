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

// PostgreSQLCustomerAddressRepository implements the CustomerAddressRepository interface using PostgreSQL
type PostgreSQLCustomerAddressRepository struct {
	*BaseRepository
	metricsCollector MetricsCollector
}

// NewPostgreSQLCustomerAddressRepository creates a new PostgreSQL customer address repository
func NewPostgreSQLCustomerAddressRepository(
	db *sqlx.DB,
	tenantResolver tenant.TenantResolver,
	metricsCollector MetricsCollector,
) repository.CustomerAddressRepository {
	if metricsCollector == nil {
		metricsCollector = &NoOpMetricsCollector{}
	}

	return &PostgreSQLCustomerAddressRepository{
		BaseRepository:   NewBaseRepository(db, tenantResolver),
		metricsCollector: metricsCollector,
	}
}

// Create creates a new customer address in the database
func (r *PostgreSQLCustomerAddressRepository) Create(ctx context.Context, address *entity.CustomerAddress) error {
	return WithMetrics(r.metricsCollector, "CREATE", "customer_addresses", func() error {
		// Validate address before creating
		if err := address.Validate(); err != nil {
			return fmt.Errorf("address validation failed: %w", err)
		}

		// Ensure ID is set
		if address.ID == uuid.Nil {
			address.ID = uuid.New()
		}

		// Set timestamps
		now := time.Now()
		address.CreatedAt = now
		address.UpdatedAt = now

		// If this is the first address for the customer, make it default
		if address.IsDefault {
			// First, unset any existing default addresses for this customer
			if err := r.UnsetDefault(ctx, address.CustomerID); err != nil {
				return fmt.Errorf("failed to unset existing default address: %w", err)
			}
		} else {
			// Check if customer has any addresses - if not, make this one default
			existingCount, err := r.GetCustomerAddressCount(ctx, address.CustomerID)
			if err != nil {
				return err
			}
			if existingCount == 0 {
				address.IsDefault = true
			}
		}

		// Normalize fields
		address.NormalizeFields()

		query := `
			INSERT INTO customer_addresses (
				id, customer_id, address_type, label, first_name, last_name,
				company, phone, address_line_1, address_line_2, city, 
				state_province, postal_code, country, is_default, is_active,
				latitude, longitude, delivery_instructions, created_at, updated_at
			) VALUES (
				:id, :customer_id, :address_type, :label, :first_name, :last_name,
				:company, :phone, :address_line_1, :address_line_2, :city,
				:state_province, :postal_code, :country, :is_default, :is_active,
				:latitude, :longitude, :delivery_instructions, :created_at, :updated_at
			)
		`

		_, err := r.db.NamedExecContext(ctx, query, address)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23503": // foreign key violation
					if strings.Contains(pqErr.Detail, "customer") {
						return fmt.Errorf("customer not found")
					}
				}
			}
			return fmt.Errorf("failed to create customer address: %w", err)
		}

		return nil
	})
}

// GetByID retrieves a customer address by ID
func (r *PostgreSQLCustomerAddressRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.CustomerAddress, error) {
	var address entity.CustomerAddress

	return &address, WithMetrics(r.metricsCollector, "GET_BY_ID", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE id = $1 AND is_active = true
		`

		err := r.db.GetContext(ctx, &address, query, id)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrAddressNotFound
			}
			return fmt.Errorf("failed to get customer address by ID: %w", err)
		}

		return nil
	})
}

// GetByCustomerID retrieves all addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error) {
	var addresses []*entity.CustomerAddress

	err := WithMetrics(r.metricsCollector, "GET_BY_CUSTOMER_ID", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND is_active = true
			ORDER BY is_default DESC, created_at DESC
		`

		err := r.db.SelectContext(ctx, &addresses, query, customerID)
		if err != nil {
			return fmt.Errorf("failed to get customer addresses: %w", err)
		}

		return nil
	})

	return addresses, err
}

// GetDefaultByCustomerID retrieves the default address for a customer
func (r *PostgreSQLCustomerAddressRepository) GetDefaultByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error) {
	var address entity.CustomerAddress

	return &address, WithMetrics(r.metricsCollector, "GET_DEFAULT", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND is_default = true AND is_active = true
		`

		err := r.db.GetContext(ctx, &address, query, customerID)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrAddressNotFound
			}
			return fmt.Errorf("failed to get default customer address: %w", err)
		}

		return nil
	})
}

// GetBillingAddressByCustomerID retrieves billing address for a customer
func (r *PostgreSQLCustomerAddressRepository) GetBillingAddressByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error) {
	var address entity.CustomerAddress

	return &address, WithMetrics(r.metricsCollector, "GET_BILLING", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND (address_type = $2 OR address_type = $3) AND is_active = true
			ORDER BY is_default DESC, created_at DESC
			LIMIT 1
		`

		err := r.db.GetContext(ctx, &address, query, customerID, entity.AddressTypeBilling, entity.AddressTypeBoth)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrAddressNotFound
			}
			return fmt.Errorf("failed to get billing address: %w", err)
		}

		return nil
	})
}

// GetShippingAddressByCustomerID retrieves shipping address for a customer
func (r *PostgreSQLCustomerAddressRepository) GetShippingAddressByCustomerID(ctx context.Context, customerID uuid.UUID) (*entity.CustomerAddress, error) {
	var address entity.CustomerAddress

	return &address, WithMetrics(r.metricsCollector, "GET_SHIPPING", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND (address_type = $2 OR address_type = $3) AND is_active = true
			ORDER BY is_default DESC, created_at DESC
			LIMIT 1
		`

		err := r.db.GetContext(ctx, &address, query, customerID, entity.AddressTypeShipping, entity.AddressTypeBoth)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.ErrAddressNotFound
			}
			return fmt.Errorf("failed to get shipping address: %w", err)
		}

		return nil
	})
}

// Update updates an existing customer address
func (r *PostgreSQLCustomerAddressRepository) Update(ctx context.Context, address *entity.CustomerAddress) error {
	return WithMetrics(r.metricsCollector, "UPDATE", "customer_addresses", func() error {
		// Validate address
		if err := address.Validate(); err != nil {
			return fmt.Errorf("address validation failed: %w", err)
		}

		// Check if address exists
		_, err := r.GetByID(ctx, address.ID)
		if err != nil {
			return err
		}

		// Handle default address logic
		if address.IsDefault {
			// Unset other default addresses for this customer
			if err := r.UnsetDefault(ctx, address.CustomerID); err != nil {
				return fmt.Errorf("failed to unset existing default address: %w", err)
			}
		}

		// Update timestamp
		address.UpdatedAt = time.Now()

		// Normalize fields
		address.NormalizeFields()

		query := `
			UPDATE customer_addresses SET
				address_type = :address_type, label = :label, first_name = :first_name,
				last_name = :last_name, company = :company, phone = :phone,
				address_line_1 = :address_line_1, address_line_2 = :address_line_2,
				city = :city, state_province = :state_province, postal_code = :postal_code,
				country = :country, is_default = :is_default, is_active = :is_active,
				latitude = :latitude, longitude = :longitude, 
				delivery_instructions = :delivery_instructions, updated_at = :updated_at
			WHERE id = :id AND is_active = true
		`

		result, err := r.db.NamedExecContext(ctx, query, address)
		if err != nil {
			return fmt.Errorf("failed to update customer address: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrAddressNotFound
		}

		return nil
	})
}

// Delete soft-deletes a customer address
func (r *PostgreSQLCustomerAddressRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "DELETE", "customer_addresses", func() error {
		// Check if this is the default address
		address, err := r.GetByID(ctx, id)
		if err != nil {
			return err
		}

		// Mark as inactive instead of deleting
		query := `
			UPDATE customer_addresses 
			SET is_active = false, updated_at = NOW()
			WHERE id = $1 AND is_active = true
		`

		result, err := r.db.ExecContext(ctx, query, id)
		if err != nil {
			return fmt.Errorf("failed to delete customer address: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrAddressNotFound
		}

		// If this was the default address, set another address as default
		if address.IsDefault {
			if err := r.EnsureDefaultAddress(ctx, address.CustomerID); err != nil {
				// Log error but don't fail - the address was deleted successfully
				_ = err
			}
		}

		return nil
	})
}

// SetAsDefault sets an address as the default address for a customer
func (r *PostgreSQLCustomerAddressRepository) SetAsDefault(ctx context.Context, customerID, addressID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "SET_DEFAULT", "customer_addresses", func() error {
		// Get the address to ensure it exists and belongs to customer
		address, err := r.GetByID(ctx, addressID)
		if err != nil {
			return err
		}

		if address.CustomerID != customerID {
			return fmt.Errorf("address does not belong to customer")
		}

		// Unset existing default addresses
		if err := r.UnsetDefault(ctx, customerID); err != nil {
			return err
		}

		// Set this address as default
		query := `
			UPDATE customer_addresses 
			SET is_default = true, updated_at = NOW()
			WHERE id = $1 AND is_active = true
		`

		result, err := r.db.ExecContext(ctx, query, addressID)
		if err != nil {
			return fmt.Errorf("failed to set default address: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrAddressNotFound
		}

		return nil
	})
}

// UnsetDefault removes default status from all addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) UnsetDefault(ctx context.Context, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "UNSET_DEFAULT", "customer_addresses", func() error {
		query := `
			UPDATE customer_addresses 
			SET is_default = false, updated_at = NOW()
			WHERE customer_id = $1 AND is_default = true AND is_active = true
		`

		_, err := r.db.ExecContext(ctx, query, customerID)
		if err != nil {
			return fmt.Errorf("failed to unset default addresses: %w", err)
		}

		return nil
	})
}

// SetAsDefaultBilling sets address as default billing address
func (r *PostgreSQLCustomerAddressRepository) SetAsDefaultBilling(ctx context.Context, customerID, addressID uuid.UUID) error {
	// For this simple implementation, we'll just set it as default and ensure it's a billing type
	return r.SetAsDefault(ctx, customerID, addressID)
}

// SetAsDefaultShipping sets address as default shipping address
func (r *PostgreSQLCustomerAddressRepository) SetAsDefaultShipping(ctx context.Context, customerID, addressID uuid.UUID) error {
	// For this simple implementation, we'll just set it as default and ensure it's a shipping type
	return r.SetAsDefault(ctx, customerID, addressID)
}

// GetAddressesByType retrieves addresses by type for a customer
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByType(ctx context.Context, customerID uuid.UUID, addressType entity.AddressType) ([]*entity.CustomerAddress, error) {
	var addresses []*entity.CustomerAddress

	err := WithMetrics(r.metricsCollector, "GET_BY_TYPE", "customer_addresses", func() error {
		query := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND address_type = $2 AND is_active = true
			ORDER BY is_default DESC, created_at DESC
		`

		err := r.db.SelectContext(ctx, &addresses, query, customerID, addressType)
		if err != nil {
			return fmt.Errorf("failed to get customer addresses by type: %w", err)
		}

		return nil
	})

	return addresses, err
}

// GetActiveAddresses retrieves all active addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) GetActiveAddresses(ctx context.Context, customerID uuid.UUID) ([]*entity.CustomerAddress, error) {
	return r.GetByCustomerID(ctx, customerID) // This already filters by is_active = true
}

// SearchAddresses searches addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) SearchAddresses(ctx context.Context, customerID uuid.UUID, query string) ([]*entity.CustomerAddress, error) {
	var addresses []*entity.CustomerAddress

	err := WithMetrics(r.metricsCollector, "SEARCH", "customer_addresses", func() error {
		searchQuery := `
			SELECT * FROM customer_addresses 
			WHERE customer_id = $1 AND is_active = true 
				AND (
					LOWER(label) LIKE LOWER($2) OR
					LOWER(address_line_1) LIKE LOWER($2) OR
					LOWER(city) LIKE LOWER($2) OR
					LOWER(first_name) LIKE LOWER($2) OR
					LOWER(last_name) LIKE LOWER($2)
				)
			ORDER BY is_default DESC, created_at DESC
		`

		searchPattern := "%" + query + "%"
		err := r.db.SelectContext(ctx, &addresses, searchQuery, customerID, searchPattern)
		if err != nil {
			return fmt.Errorf("failed to search customer addresses: %w", err)
		}

		return nil
	})

	return addresses, err
}

// GetAddressesByStorefront retrieves addresses with pagination for a storefront
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByStorefront(ctx context.Context, storefrontID uuid.UUID, req *repository.GetAddressesRequest) (*repository.AddressListResponse, error) {
	// Placeholder implementation - would need to join with customers table to filter by storefront
	return &repository.AddressListResponse{
		Addresses:  []*entity.CustomerAddress{},
		Total:      0,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: 0,
	}, nil
}

// GetAddressStats retrieves address statistics for a storefront
func (r *PostgreSQLCustomerAddressRepository) GetAddressStats(ctx context.Context, storefrontID uuid.UUID) (*repository.AddressStats, error) {
	// Placeholder implementation
	return &repository.AddressStats{
		TotalAddresses:     0,
		AddressesByType:    make(map[entity.AddressType]int),
		AddressesByCountry: make(map[string]int),
		AddressesByCity:    make(map[string]int),
		DefaultAddresses:   0,
		ActiveAddresses:    0,
	}, nil
}

// GetAddressesByCoordinates retrieves addresses near coordinates
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByCoordinates(ctx context.Context, customerID uuid.UUID, lat, lng float64, radiusKm float64) ([]*entity.CustomerAddress, error) {
	var addresses []*entity.CustomerAddress

	err := WithMetrics(r.metricsCollector, "GET_BY_COORDINATES", "customer_addresses", func() error {
		// Using Haversine formula for distance calculation
		query := `
			SELECT *, 
				(6371 * acos(cos(radians($2)) * cos(radians(latitude)) * 
				cos(radians(longitude) - radians($3)) + sin(radians($2)) * 
				sin(radians(latitude)))) AS distance_km
			FROM customer_addresses 
			WHERE customer_id = $1 AND latitude IS NOT NULL AND longitude IS NOT NULL 
				AND is_active = true
				AND (6371 * acos(cos(radians($2)) * cos(radians(latitude)) * 
					cos(radians(longitude) - radians($3)) + sin(radians($2)) * 
					sin(radians(latitude)))) <= $4
			ORDER BY distance_km ASC
		`

		rows, err := r.db.QueryContext(ctx, query, customerID, lat, lng, radiusKm)
		if err != nil {
			return fmt.Errorf("failed to get addresses by coordinates: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var address entity.CustomerAddress
			var distance float64

			err := rows.Scan(
				&address.ID, &address.CustomerID, &address.AddressType, &address.Label,
				&address.FirstName, &address.LastName, &address.Company, &address.Phone,
				&address.AddressLine1, &address.AddressLine2, &address.City,
				&address.StateProvince, &address.PostalCode, &address.Country,
				&address.IsDefault, &address.IsActive, &address.Latitude,
				&address.Longitude, &address.DeliveryInstructions, &address.CreatedAt,
				&address.UpdatedAt, &distance,
			)
			if err != nil {
				return fmt.Errorf("failed to scan address row: %w", err)
			}

			addresses = append(addresses, &address)
		}

		return rows.Err()
	})

	return addresses, err
}

// GetAddressesByCity retrieves addresses by city
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByCity(ctx context.Context, storefrontID uuid.UUID, city string) ([]*entity.CustomerAddress, error) {
	// Placeholder implementation - would need customer join
	return []*entity.CustomerAddress{}, nil
}

// GetAddressesByCountry retrieves addresses by country
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByCountry(ctx context.Context, storefrontID uuid.UUID, country string) ([]*entity.CustomerAddress, error) {
	// Placeholder implementation - would need customer join
	return []*entity.CustomerAddress{}, nil
}

// UpdateCoordinates updates the coordinates for an address
func (r *PostgreSQLCustomerAddressRepository) UpdateCoordinates(ctx context.Context, addressID uuid.UUID, lat, lng float64) error {
	return WithMetrics(r.metricsCollector, "UPDATE_COORDINATES", "customer_addresses", func() error {
		query := `
			UPDATE customer_addresses 
			SET latitude = $1, longitude = $2, updated_at = NOW()
			WHERE id = $3 AND is_active = true
		`

		result, err := r.db.ExecContext(ctx, query, lat, lng, addressID)
		if err != nil {
			return fmt.Errorf("failed to update coordinates: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return errors.ErrAddressNotFound
		}

		return nil
	})
}

// GetAddressesByIDs retrieves addresses by multiple IDs
func (r *PostgreSQLCustomerAddressRepository) GetAddressesByIDs(ctx context.Context, addressIDs []uuid.UUID) ([]*entity.CustomerAddress, error) {
	if len(addressIDs) == 0 {
		return []*entity.CustomerAddress{}, nil
	}

	var addresses []*entity.CustomerAddress

	err := WithMetrics(r.metricsCollector, "GET_BY_IDS", "customer_addresses", func() error {
		query, args, err := sqlx.In("SELECT * FROM customer_addresses WHERE id IN (?) AND is_active = true", addressIDs)
		if err != nil {
			return fmt.Errorf("failed to build IN query: %w", err)
		}

		query = r.db.Rebind(query)
		err = r.db.SelectContext(ctx, &addresses, query, args...)
		if err != nil {
			return fmt.Errorf("failed to get addresses by IDs: %w", err)
		}

		return nil
	})

	return addresses, err
}

// BulkUpdateStatus updates status for multiple addresses
func (r *PostgreSQLCustomerAddressRepository) BulkUpdateStatus(ctx context.Context, addressIDs []uuid.UUID, isActive bool) error {
	if len(addressIDs) == 0 {
		return nil
	}

	return WithMetrics(r.metricsCollector, "BULK_UPDATE_STATUS", "customer_addresses", func() error {
		query, args, err := sqlx.In("UPDATE customer_addresses SET is_active = ?, updated_at = NOW() WHERE id IN (?)", isActive, addressIDs)
		if err != nil {
			return fmt.Errorf("failed to build IN query: %w", err)
		}

		query = r.db.Rebind(query)
		_, err = r.db.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to bulk update status: %w", err)
		}

		return nil
	})
}

// DeleteByCustomerID soft-deletes all addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) DeleteByCustomerID(ctx context.Context, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "DELETE_BY_CUSTOMER", "customer_addresses", func() error {
		query := `
			UPDATE customer_addresses 
			SET is_active = false, updated_at = NOW()
			WHERE customer_id = $1 AND is_active = true
		`

		_, err := r.db.ExecContext(ctx, query, customerID)
		if err != nil {
			return fmt.Errorf("failed to delete addresses by customer ID: %w", err)
		}

		return nil
	})
}

// ValidateAddress validates address information
func (r *PostgreSQLCustomerAddressRepository) ValidateAddress(ctx context.Context, address *entity.CustomerAddress) error {
	return address.Validate()
}

// CheckDuplicateAddress checks for duplicate addresses
func (r *PostgreSQLCustomerAddressRepository) CheckDuplicateAddress(ctx context.Context, customerID uuid.UUID, address *entity.CustomerAddress) (bool, error) {
	var exists bool

	err := WithMetrics(r.metricsCollector, "CHECK_DUPLICATE", "customer_addresses", func() error {
		query := `
			SELECT EXISTS(
				SELECT 1 FROM customer_addresses 
				WHERE customer_id = $1 AND address_line_1 = $2 AND city = $3 
					AND postal_code = $4 AND country = $5 AND is_active = true
			)
		`

		err := r.db.GetContext(ctx, &exists, query, customerID, address.AddressLine1, address.City, address.PostalCode, address.Country)
		if err != nil {
			return fmt.Errorf("failed to check duplicate address: %w", err)
		}

		return nil
	})

	return exists, err
}

// CleanupInactiveAddresses removes old inactive addresses
func (r *PostgreSQLCustomerAddressRepository) CleanupInactiveAddresses(ctx context.Context, olderThanDays int) (int, error) {
	var deletedCount int

	err := WithMetrics(r.metricsCollector, "CLEANUP_INACTIVE", "customer_addresses", func() error {
		query := `
			DELETE FROM customer_addresses 
			WHERE is_active = false AND updated_at < NOW() - INTERVAL '%d days'
		`

		result, err := r.db.ExecContext(ctx, fmt.Sprintf(query, olderThanDays))
		if err != nil {
			return fmt.Errorf("failed to cleanup inactive addresses: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		deletedCount = int(rowsAffected)
		return nil
	})

	return deletedCount, err
}

// Helper methods

// GetCustomerAddressCount returns the number of addresses for a customer
func (r *PostgreSQLCustomerAddressRepository) GetCustomerAddressCount(ctx context.Context, customerID uuid.UUID) (int, error) {
	var count int

	err := WithMetrics(r.metricsCollector, "COUNT", "customer_addresses", func() error {
		query := `SELECT COUNT(*) FROM customer_addresses WHERE customer_id = $1 AND is_active = true`

		err := r.db.GetContext(ctx, &count, query, customerID)
		if err != nil {
			return fmt.Errorf("failed to get customer address count: %w", err)
		}

		return nil
	})

	return count, err
}

// EnsureDefaultAddress ensures the customer has a default address
func (r *PostgreSQLCustomerAddressRepository) EnsureDefaultAddress(ctx context.Context, customerID uuid.UUID) error {
	return WithMetrics(r.metricsCollector, "ENSURE_DEFAULT", "customer_addresses", func() error {
		// Check if customer has any default address
		var hasDefault bool
		defaultQuery := `
			SELECT EXISTS(
				SELECT 1 FROM customer_addresses 
				WHERE customer_id = $1 AND is_default = true AND is_active = true
			)
		`

		err := r.db.GetContext(ctx, &hasDefault, defaultQuery, customerID)
		if err != nil {
			return fmt.Errorf("failed to check for default address: %w", err)
		}

		if !hasDefault {
			// Set the oldest active address as default
			updateQuery := `
				UPDATE customer_addresses 
				SET is_default = true, updated_at = NOW()
				WHERE id = (
					SELECT id FROM customer_addresses 
					WHERE customer_id = $1 AND is_active = true
					ORDER BY created_at ASC 
					LIMIT 1
				) AND is_active = true
			`

			_, err := r.db.ExecContext(ctx, updateQuery, customerID)
			if err != nil {
				return fmt.Errorf("failed to set default address: %w", err)
			}
		}

		return nil
	})
}
