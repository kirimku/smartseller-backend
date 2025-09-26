package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/tenant"
)

// TenantContextKey is the key used to store tenant context in the request context
type TenantContextKey string

const TenantContext TenantContextKey = "tenant_context"

// TenantMiddleware resolves tenant context from HTTP requests
type TenantMiddleware struct {
	tenantResolver tenant.TenantResolver
	defaultDomain  string
}

// NewTenantMiddleware creates a new tenant middleware instance
func NewTenantMiddleware(tenantResolver tenant.TenantResolver, defaultDomain string) *TenantMiddleware {
	return &TenantMiddleware{
		tenantResolver: tenantResolver,
		defaultDomain:  defaultDomain,
	}
}

// ResolveTenant middleware resolves tenant context from the request and adds it to the context
func (tm *TenantMiddleware) ResolveTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext, err := tm.extractTenantContext(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_resolution_failed",
				"message": "Unable to resolve tenant from request",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		if tenantContext == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_not_found",
				"message": "No valid tenant found for this request",
			})
			c.Abort()
			return
		}

		// Add tenant context to the request context
		ctx := context.WithValue(c.Request.Context(), TenantContext, tenantContext)
		c.Request = c.Request.WithContext(ctx)

		// Also add to Gin context for easier access
		c.Set("tenant_context", tenantContext)
		c.Set("storefront_id", tenantContext.StorefrontID.String())
		c.Set("storefront_slug", tenantContext.StorefrontSlug)
		c.Set("seller_id", tenantContext.SellerID.String())

		c.Next()
	}
}

// OptionalTenant middleware resolves tenant context but doesn't fail if no tenant is found
func (tm *TenantMiddleware) OptionalTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext, _ := tm.extractTenantContext(c)

		if tenantContext != nil {
			// Add tenant context to the request context
			ctx := context.WithValue(c.Request.Context(), TenantContext, tenantContext)
			c.Request = c.Request.WithContext(ctx)

			// Also add to Gin context
			c.Set("tenant_context", tenantContext)
			c.Set("storefront_id", tenantContext.StorefrontID.String())
			c.Set("storefront_slug", tenantContext.StorefrontSlug)
			c.Set("seller_id", tenantContext.SellerID.String())
		}

		c.Next()
	}
}

// RequireStorefront middleware ensures a specific storefront context is present
func (tm *TenantMiddleware) RequireStorefront() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "storefront_required",
				"message": "This endpoint requires a valid storefront context",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSellerAccess middleware ensures the current user has access to the storefront
func (tm *TenantMiddleware) RequireSellerAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized_access",
				"message": "No storefront context found",
			})
			c.Abort()
			return
		}

		// Get current user from context (assuming auth middleware sets this)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "authentication_required",
				"message": "User authentication required",
			})
			c.Abort()
			return
		}

		// Check if user is the seller of the storefront
		userUUID, err := uuid.Parse(userID.(string))
		if err != nil || userUUID != tenantContext.SellerID {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "access_denied",
				"message": "You don't have access to this storefront",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractTenantContext extracts tenant context from the HTTP request
func (tm *TenantMiddleware) extractTenantContext(c *gin.Context) (*tenant.TenantContext, error) {
	var storefront *entity.Storefront
	var err error

	// Strategy 1: Extract from X-Storefront-ID header
	if storefrontIDHeader := c.GetHeader("X-Storefront-ID"); storefrontIDHeader != "" {
		if _, err := uuid.Parse(storefrontIDHeader); err == nil {
			// We'd need to get storefront by ID - for now skip this approach
			// This could be implemented by adding GetStorefrontByID to tenant resolver
			// storefront, err = tm.tenantResolver.GetStorefrontByID(c.Request.Context(), storefrontID)
		}
	}

	// Strategy 2: Extract from X-Storefront-Slug header
	if storefront == nil {
		if storefrontSlug := c.GetHeader("X-Storefront-Slug"); storefrontSlug != "" {
			storefront, err = tm.tenantResolver.GetStorefrontBySlug(c.Request.Context(), storefrontSlug)
			if err != nil {
				return nil, err
			}
		}
	}

	// Strategy 3: Extract from subdomain
	if storefront == nil {
		if subdomain := tm.extractSubdomain(c.Request.Host); subdomain != "" {
			storefront, err = tm.tenantResolver.GetStorefrontBySlug(c.Request.Context(), subdomain)
			if err != nil {
				// Try by subdomain field
				// storefront, err = tm.tenantResolver.GetStorefrontBySubdomain(c.Request.Context(), subdomain)
			}
		}
	}

	// Strategy 4: Extract from custom domain
	if storefront == nil {
		host := c.Request.Host
		// Remove port if present
		if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
			host = host[:colonIndex]
		}

		// Skip if it's the default domain or localhost
		if host != tm.defaultDomain && host != "localhost" && !strings.HasSuffix(host, ".localhost") {
			storefront, err = tm.tenantResolver.GetStorefrontByDomain(c.Request.Context(), host)
			if err != nil {
				return nil, err
			}
		}
	}

	// Strategy 5: Extract from URL path (e.g., /api/v1/storefront/{slug}/...)
	if storefront == nil {
		if slug := tm.extractSlugFromPath(c.Request.URL.Path); slug != "" {
			storefront, err = tm.tenantResolver.GetStorefrontBySlug(c.Request.Context(), slug)
			if err != nil {
				return nil, err
			}
		}
	}

	// Strategy 6: Extract from query parameters
	if storefront == nil {
		if slug := c.Query("storefront_slug"); slug != "" {
			storefront, err = tm.tenantResolver.GetStorefrontBySlug(c.Request.Context(), slug)
			if err != nil {
				return nil, err
			}
		}
	}

	if storefront == nil {
		return nil, nil
	}

	// Create tenant context
	tenantContext := tm.tenantResolver.CreateTenantContext(storefront)

	return tenantContext, nil
}

// extractSubdomain extracts subdomain from host
func (tm *TenantMiddleware) extractSubdomain(host string) string {
	// Remove port if present
	if colonIndex := strings.LastIndex(host, ":"); colonIndex > 0 {
		host = host[:colonIndex]
	}

	// Skip if it's localhost or IP
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return ""
	}

	// Extract subdomain
	if tm.defaultDomain != "" && strings.HasSuffix(host, "."+tm.defaultDomain) {
		subdomain := strings.TrimSuffix(host, "."+tm.defaultDomain)
		if subdomain != "" && !strings.Contains(subdomain, ".") {
			return subdomain
		}
	}

	return ""
}

// extractSlugFromPath extracts storefront slug from URL path
func (tm *TenantMiddleware) extractSlugFromPath(path string) string {
	// Handle patterns like /api/v1/storefront/{slug}/...
	segments := strings.Split(strings.Trim(path, "/"), "/")

	for i, segment := range segments {
		if segment == "storefront" && i+1 < len(segments) {
			return segments[i+1]
		}
		if segment == "s" && i+1 < len(segments) {
			// Short URL pattern /s/{slug}/...
			return segments[i+1]
		}
	}

	return ""
}

// Helper functions

// GetTenantContext extracts tenant context from Gin context
func GetTenantContext(c *gin.Context) *tenant.TenantContext {
	if tenantContext, exists := c.Get("tenant_context"); exists {
		return tenantContext.(*tenant.TenantContext)
	}
	return nil
}

// GetTenantContextFromRequest extracts tenant context from request context
func GetTenantContextFromRequest(ctx context.Context) *tenant.TenantContext {
	if tenantContext := ctx.Value(TenantContext); tenantContext != nil {
		return tenantContext.(*tenant.TenantContext)
	}
	return nil
}

// GetStorefrontID gets the storefront ID from the context
func GetStorefrontID(c *gin.Context) (uuid.UUID, bool) {
	tenantContext := GetTenantContext(c)
	if tenantContext != nil {
		return tenantContext.StorefrontID, true
	}
	return uuid.Nil, false
}

// GetSellerID gets the seller ID from the context
func GetSellerID(c *gin.Context) (uuid.UUID, bool) {
	tenantContext := GetTenantContext(c)
	if tenantContext != nil {
		return tenantContext.SellerID, true
	}
	return uuid.Nil, false
}

// MustGetStorefrontID gets the storefront ID from the context or panics
func MustGetStorefrontID(c *gin.Context) uuid.UUID {
	if id, ok := GetStorefrontID(c); ok {
		return id
	}
	panic("storefront ID not found in context")
}

// MustGetSellerID gets the seller ID from the context or panics
func MustGetSellerID(c *gin.Context) uuid.UUID {
	if id, ok := GetSellerID(c); ok {
		return id
	}
	panic("seller ID not found in context")
}

// ValidateStorefrontAccess validates that the current user has access to the storefront
func ValidateStorefrontAccess(c *gin.Context, storefrontID uuid.UUID) error {
	tenantContext := GetTenantContext(c)
	if tenantContext == nil {
		return errors.New("no storefront context found")
	}

	if tenantContext.StorefrontID != storefrontID {
		return errors.New("access denied to this storefront")
	}

	return nil
}

// SetTenantHeaders sets tenant-related headers in the response
func SetTenantHeaders(c *gin.Context) {
	tenantContext := GetTenantContext(c)
	if tenantContext != nil {
		c.Header("X-Storefront-ID", tenantContext.StorefrontID.String())
		c.Header("X-Storefront-Slug", tenantContext.StorefrontSlug)
		c.Header("X-Seller-ID", tenantContext.SellerID.String())
		c.Header("X-Tenant-Type", string(tenantContext.TenantType))
	}
}

// TenantAwareHandler wraps a handler function with tenant context awareness
func TenantAwareHandler(handler func(*gin.Context, *tenant.TenantContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		handler(c, tenantContext)
	}
}

// RequireTenantHandler wraps a handler function that requires tenant context
func RequireTenantHandler(handler func(*gin.Context, *tenant.TenantContext)) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantContext := GetTenantContext(c)
		if tenantContext == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "tenant_required",
				"message": "This endpoint requires a valid tenant context",
			})
			return
		}
		handler(c, tenantContext)
	}
}
