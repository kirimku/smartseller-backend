package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// AdminTenantMiddleware provides admin-level tenant management capabilities
type AdminTenantMiddleware struct {
	tenantResolver tenant.TenantResolver
}

// NewAdminTenantMiddleware creates a new admin tenant middleware
func NewAdminTenantMiddleware(tenantResolver tenant.TenantResolver) *AdminTenantMiddleware {
	return &AdminTenantMiddleware{
		tenantResolver: tenantResolver,
	}
}

// RequireAdminAccess middleware ensures the current user has admin access
func (atm *AdminTenantMiddleware) RequireAdminAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current user role from context (assuming auth middleware sets this)
		userRole, exists := c.Get("user_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication required",
			})
			c.Abort()
			return
		}

		// Check if user has admin role
		if userRole != "admin" && userRole != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "admin_access_required",
				"message": "Administrative access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TenantStats returns comprehensive tenant statistics
func (atm *AdminTenantMiddleware) TenantStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		storefrontIDParam := c.Param("storefront_id")
		storefrontID, err := uuid.Parse(storefrontIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_storefront_id",
				"message": "Invalid storefront ID format",
			})
			return
		}

		stats, err := atm.tenantResolver.GetTenantStats(c.Request.Context(), storefrontID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "stats_fetch_failed",
				"message": "Failed to fetch tenant statistics",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"storefront_id": storefrontID,
			"stats":         stats,
		})
	}
}

// CheckMigrationEligibility checks if a tenant can be migrated
func (atm *AdminTenantMiddleware) CheckMigrationEligibility() gin.HandlerFunc {
	return func(c *gin.Context) {
		storefrontIDParam := c.Param("storefront_id")
		storefrontID, err := uuid.Parse(storefrontIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_storefront_id",
				"message": "Invalid storefront ID format",
			})
			return
		}

		canMigrate, targetType, err := atm.tenantResolver.CanMigrateTenant(c.Request.Context(), storefrontID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "migration_check_failed",
				"message": "Failed to check migration eligibility",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"storefront_id": storefrontID,
			"can_migrate":   canMigrate,
			"target_type":   targetType,
			"current_type":  "", // Would need to get current type
		})
	}
}

// MigrateTenant initiates tenant migration
func (atm *AdminTenantMiddleware) MigrateTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		storefrontIDParam := c.Param("storefront_id")
		storefrontID, err := uuid.Parse(storefrontIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_storefront_id",
				"message": "Invalid storefront ID format",
			})
			return
		}

		var req struct {
			TargetType tenant.TenantType `json:"target_type" binding:"required"`
			Force      bool              `json:"force"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		// Validate target type
		if req.TargetType != tenant.TenantTypeShared &&
			req.TargetType != tenant.TenantTypeSchema &&
			req.TargetType != tenant.TenantTypeDatabase {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_target_type",
				"message": "Invalid target tenant type",
			})
			return
		}

		// Check migration eligibility unless forced
		if !req.Force {
			canMigrate, _, err := atm.tenantResolver.CanMigrateTenant(c.Request.Context(), storefrontID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "migration_check_failed",
					"message": "Failed to check migration eligibility",
					"details": err.Error(),
				})
				return
			}

			if !canMigrate {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":      "migration_not_eligible",
					"message":    "Tenant is not eligible for migration",
					"suggestion": "Use force=true to override eligibility check",
				})
				return
			}
		}

		// Perform migration
		err = atm.tenantResolver.MigrateTenant(c.Request.Context(), storefrontID, req.TargetType)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "migration_failed",
				"message": "Failed to migrate tenant",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"storefront_id": storefrontID,
			"target_type":   req.TargetType,
			"status":        "migration_initiated",
			"message":       "Tenant migration has been initiated successfully",
		})
	}
}

// InvalidateCache invalidates tenant cache
func (atm *AdminTenantMiddleware) InvalidateCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			StorefrontSlug string    `json:"storefront_slug"`
			StorefrontID   uuid.UUID `json:"storefront_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body",
				"details": err.Error(),
			})
			return
		}

		if req.StorefrontSlug != "" {
			atm.tenantResolver.InvalidateStorefront(req.StorefrontSlug)
		}

		if req.StorefrontID != uuid.Nil {
			atm.tenantResolver.InvalidateStorefrontByID(req.StorefrontID)
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "cache_invalidated",
			"message": "Tenant cache has been invalidated successfully",
		})
	}
}

// TenantHealthCheck provides health check for tenant-related services
func (atm *AdminTenantMiddleware) TenantHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would check the health of tenant-related services
		// For now, return a simple health status
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": "2024-01-01T00:00:00Z", // Would use actual time
			"components": gin.H{
				"tenant_resolver": "healthy",
				"cache":           "healthy",
				"database":        "healthy",
			},
		})
	}
}

// ListTenants provides a list of all tenants (for admin purposes)
func (atm *AdminTenantMiddleware) ListTenants() gin.HandlerFunc {
	return func(c *gin.Context) {
		// This would implement tenant listing logic
		// For now, return a placeholder response
		c.JSON(http.StatusOK, gin.H{
			"tenants": []gin.H{},
			"total":   0,
			"page":    1,
		})
	}
}
