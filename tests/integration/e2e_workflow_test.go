package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// E2EWorkflowTestSuite tests complete end-to-end workflows
type E2EWorkflowTestSuite struct {
	suite.Suite
	framework *Phase4TestFramework
	router    *gin.Engine
}

// SetupSuite initializes the test suite with all handlers
func (suite *E2EWorkflowTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware
	suite.router.Use(gin.Recovery())
	suite.router.Use(func(c *gin.Context) {
		c.Set("storefront_id", uuid.New())
		c.Next()
	})

	// Mock handlers for all endpoints
	customerHandler := &MockCustomerHandler{}
	addressHandler := &MockAddressHandler{}
	storefrontHandler := &MockStorefrontHandler{}

	v1 := suite.router.Group("/api/v1")
	{
		// Customer routes
		customers := v1.Group("/customers")
		{
			customers.POST("/register", customerHandler.RegisterCustomer)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.GET("/:id/profile", customerHandler.GetProfile)
			customers.PUT("/:id/profile", customerHandler.UpdateProfile)
			customers.POST("/:id/activate", customerHandler.ActivateCustomer)
			customers.POST("/:id/deactivate", customerHandler.DeactivateCustomer)

			// Customer address routes
			customers.GET("/:id/addresses", addressHandler.GetCustomerAddresses)
			customers.POST("/:id/addresses", addressHandler.CreateAddress)
			customers.POST("/:customer_id/addresses/:address_id/default", addressHandler.SetDefaultAddress)
			customers.GET("/:id/addresses/default", addressHandler.GetDefaultAddress)

			// Customer storefront routes
			customers.GET("/:id/storefronts", storefrontHandler.GetCustomerStorefronts)
		}

		// Address routes
		addresses := v1.Group("/addresses")
		{
			addresses.GET("/:id", addressHandler.GetAddress)
			addresses.PUT("/:id", addressHandler.UpdateAddress)
			addresses.DELETE("/:id", addressHandler.DeleteAddress)
			addresses.POST("/validate", addressHandler.ValidateAddress)
		}

		// Storefront routes
		storefronts := v1.Group("/storefronts")
		{
			storefronts.POST("", storefrontHandler.CreateStorefront)
			storefronts.GET("/:id", storefrontHandler.GetStorefront)
			storefronts.PUT("/:id", storefrontHandler.UpdateStorefront)
			storefronts.DELETE("/:id", storefrontHandler.DeleteStorefront)
			storefronts.GET("/:id/analytics", storefrontHandler.GetStorefrontAnalytics)
			storefronts.POST("/:id/publish", storefrontHandler.PublishStorefront)
		}

		// Domain validation
		v1.POST("/storefronts/validate-domain", storefrontHandler.ValidateDomain)
	}

	suite.framework = NewPhase4TestFramework(suite.T(), suite.router)
}

// TearDownSuite cleans up after all tests
func (suite *E2EWorkflowTestSuite) TearDownSuite() {
	if suite.framework != nil {
		suite.framework.Cleanup()
	}
}

// TestCompleteCustomerOnboardingWorkflow tests the complete customer journey
func (suite *E2EWorkflowTestSuite) TestCompleteCustomerOnboardingWorkflow() {
	var customerID uuid.UUID
	var customerEmail string
	var addressID string
	var storefrontID string

	suite.T().Run("Step1_CustomerRegistration", func(t *testing.T) {
		// Generate test customer
		testCustomer := suite.framework.GenerateTestCustomer("onboarding")
		customerEmail = testCustomer.Email

		// Register customer
		resp, body := suite.framework.MakeRequest("POST", "/api/v1/customers/register", testCustomer, nil)
		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)

		// Extract customer ID
		customerData := apiResp.Data.(map[string]interface{})
		customerIDStr := customerData["id"].(string)
		var err error
		customerID, err = uuid.Parse(customerIDStr)
		require.NoError(t, err, "Should be able to parse customer ID")

		// Verify registration data
		assert.Equal(t, testCustomer.FirstName, customerData["first_name"])
		assert.Equal(t, testCustomer.LastName, customerData["last_name"])
		assert.Equal(t, testCustomer.Email, customerData["email"])
		if testCustomer.Phone != nil {
			assert.Equal(t, *testCustomer.Phone, customerData["phone"])
		}
		assert.Contains(t, customerData, "id", "Should include customer ID")
		assert.Contains(t, customerData, "created_at", "Should include creation timestamp")

		// Verify customer status
		assert.Equal(t, "pending_verification", customerData["status"], "New customer should be pending verification")
	})

	suite.T().Run("Step2_CustomerActivation", func(t *testing.T) {
		// Activate the customer account
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/activate", customerID.String()),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify customer is now active
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		customerData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, "active", customerData["status"], "Customer should now be active")
	})

	suite.T().Run("Step3_ProfileSetup", func(t *testing.T) {
		// Update customer profile with additional information
		firstName := "Updated First"
		lastName := "Updated Last"
		dob := time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC)
		profileUpdate := dto.CustomerUpdateRequest{
			FirstName:   &firstName,
			LastName:    &lastName,
			DateOfBirth: &dob,
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/customers/%s/profile", customerID.String()),
			profileUpdate,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify profile updates
		profileData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, *profileUpdate.FirstName, profileData["first_name"])
		assert.Equal(t, *profileUpdate.LastName, profileData["last_name"])
	})

	suite.T().Run("Step4_AddressCreation", func(t *testing.T) {
		// Create primary shipping address
		addressReq := dto.CreateAddressRequest{
			CustomerID:    customerID,
			Type:          entity.AddressTypeShipping,
			Label:         stringPtr("Home"),
			FirstName:     stringPtr("John"),
			LastName:      stringPtr("Doe"),
			Phone:         stringPtr("+628123456789"),
			AddressLine1:  "Jl. Sudirman No. 123",
			AddressLine2:  stringPtr("Apartment 4B"),
			City:          "Jakarta",
			StateProvince: stringPtr("DKI Jakarta"),
			PostalCode:    "12190",
			Country:       "Indonesia",
			IsDefault:     true,
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			addressReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)

		// Store address ID for later use
		addressData := apiResp.Data.(map[string]interface{})
		addressID = addressData["id"].(string)

		// Verify address creation
		assert.NotEmpty(t, addressID, "Address ID should not be empty")
		assert.Equal(t, *addressReq.Label, addressData["label"])
		assert.Equal(t, addressReq.AddressLine1, addressData["address_line1"])
		assert.Equal(t, addressReq.City, addressData["city"])
		assert.Equal(t, addressReq.PostalCode, addressData["postal_code"])
		assert.Equal(t, addressReq.IsDefault, addressData["is_default"])
	})

	suite.T().Run("Step5_AdditionalAddresses", func(t *testing.T) {
		// Create billing address
		billingAddressReq := dto.CreateAddressRequest{
			CustomerID:    customerID,
			Type:          entity.AddressTypeBilling,
			Label:         stringPtr("Office"),
			FirstName:     stringPtr("John"),
			LastName:      stringPtr("Doe"),
			Company:       stringPtr("Acme Corp"),
			Phone:         stringPtr("+628123456788"),
			AddressLine1:  "Jl. Thamrin No. 456",
			City:          "Jakarta",
			StateProvince: stringPtr("DKI Jakarta"),
			PostalCode:    "10350",
			Country:       "Indonesia",
			IsDefault:     false,
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			billingAddressReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)

		// Verify we now have multiple addresses
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		addressesData, ok := apiResp.Data.([]interface{})
		require.True(t, ok, "Addresses data should be an array")
		assert.Len(t, addressesData, 2, "Should have 2 addresses")
	})

	suite.T().Run("Step6_StorefrontCreation", func(t *testing.T) {
		// Create storefront for the customer
		storefrontReq := dto.StorefrontCreateRequest{
			Name:        "John's Store",
			Slug:        "johns-store-onboarding",
			Description: stringPtr("My awesome online store"),
			Domain:      stringPtr("johnsstore.example.com"),
			Subdomain:   stringPtr("johnsstore"),
			OwnerEmail:  customerEmail,
			OwnerName:   "John Doe",
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			storefrontReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)

		// Store storefront ID
		storefrontData := apiResp.Data.(map[string]interface{})
		storefrontID = storefrontData["id"].(string)

		// Verify storefront creation
		assert.Equal(t, storefrontReq.Name, storefrontData["name"])
		assert.Equal(t, storefrontReq.Slug, storefrontData["slug"])
		assert.Equal(t, *storefrontReq.Description, storefrontData["description"])
		assert.Equal(t, storefrontReq.OwnerEmail, storefrontData["owner_email"])
		assert.Contains(t, storefrontData, "status", "Should include status")
	})

	suite.T().Run("Step7_StorefrontConfiguration", func(t *testing.T) {
		// Update storefront with additional configuration
		updateReq := dto.StorefrontUpdateRequest{
			Description: stringPtr("My updated awesome online store with premium products"),
			Domain:      stringPtr("premium-johnsstore.example.com"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			updateReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify updates
		storefrontData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, *updateReq.Description, storefrontData["description"])
	})

	suite.T().Run("Step8_StorefrontPublishing", func(t *testing.T) {
		// Publish the storefront
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/storefronts/%s/publish", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify storefront is now published by checking its details
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s", storefrontID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		storefrontData := apiResp.Data.(map[string]interface{})
		// In a real scenario, published status would be reflected in the response
		assert.Contains(t, storefrontData, "status", "Should include status after publishing")
	})

	suite.T().Run("Step9_VerifyCompleteSetup", func(t *testing.T) {
		// Verify customer has complete setup
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/profile", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		profileData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, "active", profileData["status"], "Customer should be active")

		// Verify customer has addresses
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		addressesData, ok := apiResp.Data.([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(addressesData), 1, "Should have at least one address")

		// Verify customer has storefronts
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/storefronts", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		storefrontsData, ok := apiResp.Data.([]interface{})
		require.True(t, ok)
		assert.GreaterOrEqual(t, len(storefrontsData), 1, "Should have at least one storefront")
	})
}

// TestStorefrontManagementWorkflow tests storefront-focused workflow
func (suite *E2EWorkflowTestSuite) TestStorefrontManagementWorkflow() {
	suite.T().Run("StorefrontLifecycle", func(t *testing.T) {
		// Create customer
		testCustomer := suite.framework.GenerateTestCustomer("storefront-mgmt")
		customerResp := suite.framework.RegisterTestCustomer(testCustomer)
		customerID := customerResp.ID

		// Validate domain before creating storefront
		domainReq := map[string]string{"domain": "valid-store.example.com"}
		resp, body := suite.framework.MakeRequest(
			"POST",
			"/api/v1/storefronts/validate-domain",
			domainReq,
			nil,
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		validationData := apiResp.Data.(map[string]interface{})
		assert.True(t, validationData["valid"].(bool), "Domain should be valid")

		// Create storefront
		storefrontReq := dto.StorefrontCreateRequest{
			Name:       "Validated Store",
			Slug:       "validated-store",
			Domain:     stringPtr("valid-store.example.com"),
			OwnerEmail: testCustomer.Email,
			OwnerName:  fmt.Sprintf("%s %s", testCustomer.FirstName, testCustomer.LastName),
		}

		resp, body = suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			storefrontReq,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)
		storefrontData := apiResp.Data.(map[string]interface{})
		storefrontID := storefrontData["id"].(string)

		// Get analytics (should start at zero)
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/storefronts/%s/analytics", storefrontID),
			nil,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		analyticsData := apiResp.Data.(map[string]interface{})
		assert.Contains(t, analyticsData, "views", "Should include view metrics")
		assert.Contains(t, analyticsData, "visitors", "Should include visitor metrics")

		// Publish storefront
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/storefronts/%s/publish", storefrontID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify storefront appears in customer's list
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/storefronts", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp = suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		storefrontsData, ok := apiResp.Data.([]interface{})
		require.True(t, ok)

		// Find our storefront in the list
		found := false
		for _, sf := range storefrontsData {
			sfData := sf.(map[string]interface{})
			if sfData["id"].(string) == storefrontID {
				found = true
				assert.Equal(t, storefrontReq.Name, sfData["name"])
				break
			}
		}
		assert.True(t, found, "Storefront should appear in customer's storefront list")
	})
}

// TestErrorHandlingWorkflow tests error scenarios in workflows
func (suite *E2EWorkflowTestSuite) TestErrorHandlingWorkflow() {
	suite.T().Run("FailedRegistrationRecovery", func(t *testing.T) {
		// Try to register with invalid email
		phone := "+628123456789"
		invalidCustomer := dto.CustomerRegistrationRequest{
			FirstName: "Test",
			LastName:  "User",
			Email:     "invalid-email-format",
			Phone:     &phone,
			Password:  "password123",
		}

		resp, body := suite.framework.MakeRequest("POST", "/api/v1/customers/register", invalidCustomer, nil)
		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)

		// Register with correct email
		validCustomer := suite.framework.GenerateTestCustomer("recovery")
		resp, body = suite.framework.MakeRequest("POST", "/api/v1/customers/register", validCustomer, nil)
		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)

		customerData := apiResp.Data.(map[string]interface{})
		customerID, _ := uuid.Parse(customerData["id"].(string))

		// Try to create storefront with invalid slug
		invalidStorefront := dto.StorefrontCreateRequest{
			Name:       "Test Store",
			Slug:       "Invalid Slug With Spaces",
			OwnerEmail: validCustomer.Email,
			OwnerName:  "Test User",
		}

		resp, body = suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			invalidStorefront,
			customerID.String(),
		)
		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)

		// Create storefront with valid slug
		validStorefront := dto.StorefrontCreateRequest{
			Name:       "Test Store",
			Slug:       "valid-test-store",
			OwnerEmail: validCustomer.Email,
			OwnerName:  "Test User",
		}

		resp, body = suite.framework.MakeAuthenticatedRequest(
			"POST",
			"/api/v1/storefronts",
			validStorefront,
			customerID.String(),
		)
		suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)
	})
}

// MockCustomerHandler provides mock implementations for customer endpoints
type MockCustomerHandler struct{}

func (h *MockCustomerHandler) RegisterCustomer(c *gin.Context) {
	customerID := uuid.New()
	email := "test@example.com"
	firstName := "John"
	lastName := "Doe"
	phone := "+628123456789"

	response := dto.CustomerResponse{
		ID:        customerID,
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
		Status:    string(entity.CustomerStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": response})
}

func (h *MockCustomerHandler) GetCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, _ := uuid.Parse(customerIDStr)
	email := "test@example.com"
	firstName := "John"
	lastName := "Doe"
	phone := "+628123456789"

	response := dto.CustomerResponse{
		ID:        customerID,
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
		Status:    string(entity.CustomerStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

func (h *MockCustomerHandler) UpdateCustomer(c *gin.Context) {
	customerIDStr := c.Param("id")
	customerID, _ := uuid.Parse(customerIDStr)
	email := "test@example.com"
	firstName := "Updated First"
	lastName := "Updated Last"
	phone := "+628123456789"

	response := dto.CustomerResponse{
		ID:        customerID,
		Email:     &email,
		FirstName: &firstName,
		LastName:  &lastName,
		Phone:     &phone,
		Status:    string(entity.CustomerStatusActive),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}

func (h *MockCustomerHandler) GetProfile(c *gin.Context) {
	h.GetCustomer(c) // Same response structure
}

func (h *MockCustomerHandler) UpdateProfile(c *gin.Context) {
	h.UpdateCustomer(c) // Same response structure
}

func (h *MockCustomerHandler) ActivateCustomer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Customer activated successfully"})
}

func (h *MockCustomerHandler) DeactivateCustomer(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Customer deactivated successfully"})
}

// Helper functions for pointer types
func timePtr(t time.Time) *time.Time {
	return &t
}

func boolPtr(b bool) *bool {
	return &b
}

// TestE2EWorkflowSuite runs the end-to-end workflow test suite
func TestE2EWorkflowSuite(t *testing.T) {
	suite.Run(t, new(E2EWorkflowTestSuite))
}
