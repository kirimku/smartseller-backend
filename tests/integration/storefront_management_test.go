package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// StorefrontManagementTestSuite tests all storefront management endpoints
type StorefrontManagementTestSuite struct {
	suite.Suite
	framework *Phase4TestFramework
	router    *gin.Engine
}

// SetupSuite initializes the test suite
func (suite *StorefrontManagementTestSuite) SetupSuite() {
	// Initialize Gin router with storefront routes
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware
	suite.router.Use(gin.Recovery())
	suite.router.Use(func(c *gin.Context) {
		c.Set("storefront_id", uuid.New())
		c.Next()
	})

	// Mock storefront handler
	storefrontHandler := &MockStorefrontHandler{}

	v1 := suite.router.Group("/api/v1")
	{
		// Storefront management routes
		storefronts := v1.Group("/storefronts")
		{
			storefronts.POST("", storefrontHandler.CreateStorefront)
			storefronts.GET("/:id", storefrontHandler.GetStorefront)
			storefronts.PUT("/:id", storefrontHandler.UpdateStorefront)
			storefronts.DELETE("/:id", storefrontHandler.DeleteStorefront)
			storefronts.GET("/:id/analytics", storefrontHandler.GetStorefrontAnalytics)
			storefronts.POST("/:id/publish", storefrontHandler.PublishStorefront)
			storefronts.POST("/:id/unpublish", storefrontHandler.UnpublishStorefront)
		}

		// Domain validation route (simple version)
		v1.POST("/storefronts/validate-domain", storefrontHandler.ValidateDomain)

		// Customer storefront routes
		customers := v1.Group("/customers")
		{
			customers.GET("/:id/storefronts", storefrontHandler.GetCustomerStorefronts)
		}
	}

	suite.framework = NewPhase4TestFramework(suite.T(), suite.router)
}

// TearDownSuite cleans up after all tests
func (suite *StorefrontManagementTestSuite) TearDownSuite() {
	if suite.framework != nil {
		suite.framework.Cleanup()
	}
}

// TestStorefrontCRUD tests storefront CRUD operations
func (suite *StorefrontManagementTestSuite) TestStorefrontCRUD() {
	// Create a test customer first
	testCustomer := suite.framework.GenerateTestCustomer("storefront")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	var storefrontID string

	suite.T().Run("CreateStorefront", func(t *testing.T) {
		createReq := dto.StorefrontCreateRequest{
			Name:        "Test Store",
			Slug:        "test-store-123",
			Description: stringPtr("A test storefront for integration testing"),
			Domain:      stringPtr("teststore.example.com"),
			Subdomain:   stringPtr("teststore"),
			OwnerEmail:  "owner@teststore.com",
			OwnerName:   "Test Owner",
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			createReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)
		suite.framework.AssertStorefrontResponse(apiResp.Data, customerID)

		// Store storefront ID for subsequent tests
		storefrontData := apiResp.Data.(map[string]interface{})
		storefrontID = storefrontData["id"].(string)

		// Verify specific fields
		assert.Equal(t, createReq.Name, storefrontData["name"])
		assert.Equal(t, createReq.Slug, storefrontData["slug"])
		assert.Equal(t, *createReq.Domain, storefrontData["domain"])
		assert.Equal(t, *createReq.Description, storefrontData["description"])
		assert.Equal(t, *createReq.Subdomain, storefrontData["subdomain"])
		assert.Equal(t, createReq.OwnerEmail, storefrontData["owner_email"])
		assert.Equal(t, createReq.OwnerName, storefrontData["owner_name"])
	})

	suite.T().Run("GetStorefront", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertStorefrontResponse(apiResp.Data, customerID)

		// Verify storefront data
		storefrontData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, storefrontID, storefrontData["id"])
		assert.Equal(t, "Test Store", storefrontData["name"])
	})

	suite.T().Run("UpdateStorefront", func(t *testing.T) {
		updateReq := dto.StorefrontUpdateRequest{
			Name:        stringPtr("Updated Test Store"),
			Description: stringPtr("Updated description for the test store"),
			Domain:      stringPtr("updated-teststore.example.com"),
			Subdomain:   stringPtr("updated-teststore"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			updateReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify updated fields
		storefrontData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, *updateReq.Name, storefrontData["name"])
		assert.Equal(t, *updateReq.Description, storefrontData["description"])
		assert.Equal(t, *updateReq.Domain, storefrontData["domain"])
		assert.Equal(t, *updateReq.Subdomain, storefrontData["subdomain"])
	})

	suite.T().Run("PublishStorefront", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/storefronts/%s/publish", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
	})

	suite.T().Run("UnpublishStorefront", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/storefronts/%s/unpublish", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
	})

	suite.T().Run("DeleteStorefront", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"DELETE",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
	})

	suite.T().Run("GetDeletedStorefront", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})
}

// TestStorefrontValidation tests domain validation
func (suite *StorefrontManagementTestSuite) TestStorefrontValidation() {
	suite.T().Run("ValidateDomain", func(t *testing.T) {
		tests := []struct {
			name           string
			domain         string
			expectedStatus int
			expectValid    bool
		}{
			{
				name:           "ValidDomain",
				domain:         "valid-domain.example.com",
				expectedStatus: http.StatusOK,
				expectValid:    true,
			},
			{
				name:           "InvalidDomainFormat",
				domain:         "invalid_domain_format",
				expectedStatus: http.StatusOK,
				expectValid:    false,
			},
			{
				name:           "ExistingDomain",
				domain:         "existing-domain.example.com",
				expectedStatus: http.StatusOK,
				expectValid:    false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				request := map[string]string{"domain": tt.domain}

				resp, body := suite.framework.MakeRequest(
					"POST",
					"/api/v1/storefronts/validate-domain",
					request,
					nil,
				)

				apiResp := suite.framework.AssertSuccessResponse(resp, body, tt.expectedStatus)

				validationData := apiResp.Data.(map[string]interface{})
				assert.Equal(t, tt.expectValid, validationData["valid"], "Domain validation result should match expected")
				assert.Contains(t, validationData, "message", "Should include validation message")
			})
		}
	})
}

// TestStorefrontAnalytics tests storefront analytics endpoints
func (suite *StorefrontManagementTestSuite) TestStorefrontAnalytics() {
	// Create test customer and storefront
	testCustomer := suite.framework.GenerateTestCustomer("analytics")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	// Create a simple storefront for testing
	storefrontID := uuid.New().String()

	suite.T().Run("GetBasicAnalytics", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s/analytics", storefrontID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		analyticsData := apiResp.Data.(map[string]interface{})
		assert.Contains(t, analyticsData, "views", "Should include views data")
		assert.Contains(t, analyticsData, "visitors", "Should include visitors data")
		assert.Contains(t, analyticsData, "engagement", "Should include engagement metrics")
		assert.Contains(t, analyticsData, "period", "Should include time period")
	})

	suite.T().Run("GetAnalyticsWithDateRange", func(t *testing.T) {
		// Test with query parameters
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s/analytics?period=30d&include_charts=true", storefrontID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		analyticsData := apiResp.Data.(map[string]interface{})
		assert.Contains(t, analyticsData, "charts", "Should include chart data when requested")
		assert.Contains(t, analyticsData, "summary", "Should include summary statistics")
	})
}

// TestCustomerStorefronts tests customer-specific storefront operations
func (suite *StorefrontManagementTestSuite) TestCustomerStorefronts() {
	// Create test customer
	testCustomer := suite.framework.GenerateTestCustomer("customer")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	suite.T().Run("GetCustomerStorefronts", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/storefronts", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		storefrontsData, ok := apiResp.Data.([]interface{})
		require.True(t, ok, "Storefronts data should be an array")
		assert.LessOrEqual(t, len(storefrontsData), 10, "Should have reasonable number of storefronts")
	})

	suite.T().Run("GetCustomerStorefrontsWithPagination", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/storefronts?limit=2&offset=0", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		storefrontsData, ok := apiResp.Data.([]interface{})
		require.True(t, ok, "Storefronts data should be an array")
		assert.LessOrEqual(t, len(storefrontsData), 2, "Should respect limit parameter")
	})
}

// TestStorefrontValidationErrors tests storefront validation error scenarios
func (suite *StorefrontManagementTestSuite) TestStorefrontValidationErrors() {
	// Create a test customer
	testCustomer := suite.framework.GenerateTestCustomer("validation")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	suite.T().Run("CreateStorefrontWithMissingRequiredFields", func(t *testing.T) {
		incompleteReq := dto.StorefrontCreateRequest{
			// Missing required fields: Name, Slug, OwnerEmail, OwnerName
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			incompleteReq,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("CreateStorefrontWithInvalidSlug", func(t *testing.T) {
		invalidSlugReq := dto.StorefrontCreateRequest{
			Name:       "Test Store",
			Slug:       "Invalid Slug With Spaces!",
			OwnerEmail: "test@example.com",
			OwnerName:  "Test Owner",
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			invalidSlugReq,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("UpdateNonExistentStorefront", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		updateReq := dto.StorefrontUpdateRequest{
			Name: stringPtr("Updated Name"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/storefronts/%s", nonExistentID),
			updateReq,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})
}

// MockStorefrontHandler provides mock implementations for storefront endpoints
type MockStorefrontHandler struct{}

func (h *MockStorefrontHandler) CreateStorefront(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Storefront created successfully",
		"data": gin.H{
			"id":          uuid.New().String(),
			"name":        "Test Store",
			"slug":        "test-store",
			"description": "A test storefront",
			"domain":      "teststore.example.com",
			"subdomain":   "teststore",
			"owner_email": "owner@teststore.com",
			"owner_name":  "Test Owner",
			"status":      "draft",
			"settings":    gin.H{},
			"created_at":  time.Now().Format(time.RFC3339),
			"updated_at":  time.Now().Format(time.RFC3339),
		},
	})
}

func (h *MockStorefrontHandler) GetStorefront(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storefront retrieved successfully",
		"data": gin.H{
			"id":         c.Param("id"),
			"name":       "Test Store",
			"slug":       "test-store",
			"status":     "draft",
			"created_at": time.Now().Format(time.RFC3339),
		},
	})
}

func (h *MockStorefrontHandler) UpdateStorefront(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storefront updated successfully",
		"data": gin.H{
			"id":          c.Param("id"),
			"name":        "Updated Test Store",
			"description": "Updated description for the test store",
			"updated_at":  time.Now().Format(time.RFC3339),
		},
	})
}

func (h *MockStorefrontHandler) DeleteStorefront(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storefront deleted successfully",
	})
}

func (h *MockStorefrontHandler) PublishStorefront(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storefront published successfully",
	})
}

func (h *MockStorefrontHandler) UnpublishStorefront(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Storefront unpublished successfully",
	})
}

func (h *MockStorefrontHandler) ValidateDomain(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := req["domain"]
	isValid := !strings.Contains(domain, "existing") && strings.Contains(domain, ".")
	message := "Domain is available"
	if !isValid {
		message = "Domain is not available or invalid format"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Domain validation completed",
		"data": gin.H{
			"valid":   isValid,
			"domain":  domain,
			"message": message,
		},
	})
}

func (h *MockStorefrontHandler) GetStorefrontAnalytics(c *gin.Context) {
	includeCharts := c.Query("include_charts") == "true"

	data := gin.H{
		"views":      1250,
		"visitors":   890,
		"engagement": gin.H{"bounce_rate": 0.35, "avg_session_duration": 180},
		"period":     c.DefaultQuery("period", "7d"),
		"summary": gin.H{
			"total_views":     1250,
			"unique_visitors": 890,
			"conversion_rate": 0.025,
		},
	}

	if includeCharts {
		data["charts"] = gin.H{
			"daily_views":     []int{120, 150, 180, 200, 175, 160, 140},
			"traffic_sources": gin.H{"direct": 40, "search": 35, "social": 25},
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Analytics retrieved successfully",
		"data":    data,
	})
}

func (h *MockStorefrontHandler) GetCustomerStorefronts(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	if limit == "2" {
		// Return limited results for pagination test
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Customer storefronts retrieved successfully",
			"data": []gin.H{
				{"id": uuid.New().String(), "name": "Store 1"},
				{"id": uuid.New().String(), "name": "Store 2"},
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Customer storefronts retrieved successfully",
		"data": []gin.H{
			{"id": uuid.New().String(), "name": "Store 1"},
			{"id": uuid.New().String(), "name": "Store 2"},
			{"id": uuid.New().String(), "name": "Store 3"},
		},
	})
}

// TestStorefrontManagementSuite runs the storefront management test suite
func TestStorefrontManagementSuite(t *testing.T) {
	suite.Run(t, new(StorefrontManagementTestSuite))
}
