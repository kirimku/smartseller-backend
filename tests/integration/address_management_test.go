package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
)

// AddressManagementTestSuite tests all address management endpoints
type AddressManagementTestSuite struct {
	suite.Suite
	framework *Phase4TestFramework
	router    *gin.Engine
}

// SetupSuite initializes the test suite
func (suite *AddressManagementTestSuite) SetupSuite() {
	// Initialize Gin router with address routes
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware
	suite.router.Use(gin.Recovery())
	suite.router.Use(func(c *gin.Context) {
		c.Set("storefront_id", uuid.New())
		c.Next()
	})

	// Mock address handler
	addressHandler := &MockAddressHandler{}

	v1 := suite.router.Group("/api/v1")
	{
		// Customer address routes
		customers := v1.Group("/customers")
		{
			customers.GET("/:id/addresses", addressHandler.GetCustomerAddresses)
			customers.POST("/:id/addresses", addressHandler.CreateAddress)
			customers.POST("/:customer_id/addresses/:address_id/default", addressHandler.SetDefaultAddress)
			customers.GET("/:id/addresses/default", addressHandler.GetDefaultAddress)
		}

		// Individual address routes
		addresses := v1.Group("/addresses")
		{
			addresses.GET("/:id", addressHandler.GetAddress)
			addresses.PUT("/:id", addressHandler.UpdateAddress)
			addresses.DELETE("/:id", addressHandler.DeleteAddress)
			addresses.POST("/validate", addressHandler.ValidateAddress)
			addresses.POST("/geocode", addressHandler.GeocodeAddress)
		}
	}

	suite.framework = NewPhase4TestFramework(suite.T(), suite.router)
}

// TearDownSuite cleans up after all tests
func (suite *AddressManagementTestSuite) TearDownSuite() {
	if suite.framework != nil {
		suite.framework.Cleanup()
	}
}

// TestAddressCRUD tests address CRUD operations
func (suite *AddressManagementTestSuite) TestAddressCRUD() {
	// Create a test customer first
	testCustomer := suite.framework.GenerateTestCustomer("address")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	var addressID string

	suite.T().Run("CreateAddress", func(t *testing.T) {
		createReq := dto.CreateAddressRequest{
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
			IsDefault:     false,
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			createReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)
		suite.framework.AssertAddressResponse(apiResp.Data, customerID)

		// Store address ID for subsequent tests
		addressData := apiResp.Data.(map[string]interface{})
		addressID = addressData["id"].(string)

		// Verify specific fields
		assert.Equal(t, *createReq.Label, addressData["label"])
		assert.Equal(t, *createReq.FirstName, addressData["first_name"])
		assert.Equal(t, *createReq.LastName, addressData["last_name"])
		assert.Equal(t, *createReq.Phone, addressData["phone"])
		assert.Equal(t, createReq.AddressLine1, addressData["address_line1"])
		assert.Equal(t, *createReq.AddressLine2, addressData["address_line2"])
		assert.Equal(t, createReq.City, addressData["city"])
		assert.Equal(t, *createReq.StateProvince, addressData["state_province"])
		assert.Equal(t, createReq.PostalCode, addressData["postal_code"])
		assert.Equal(t, createReq.Country, addressData["country"])
	})

	suite.T().Run("GetAddress", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/addresses/%s", addressID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertAddressResponse(apiResp.Data, customerID)
	})

	suite.T().Run("UpdateAddress", func(t *testing.T) {
		updateReq := dto.UpdateAddressRequest{
			Label:         stringPtr("Office"),
			FirstName:     stringPtr("Jane"),
			LastName:      stringPtr("Doe"),
			Phone:         stringPtr("+628123456790"),
			AddressLine1:  stringPtr("Jl. Thamrin No. 456"),
			City:          stringPtr("Jakarta"),
			StateProvince: stringPtr("DKI Jakarta"),
			PostalCode:    stringPtr("10350"),
			Country:       stringPtr("Indonesia"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/addresses/%s", addressID),
			updateReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify updated fields
		addressData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, *updateReq.Label, addressData["label"])
		assert.Equal(t, *updateReq.FirstName, addressData["first_name"])
		assert.Equal(t, *updateReq.LastName, addressData["last_name"])
		assert.Equal(t, *updateReq.Phone, addressData["phone"])
		assert.Equal(t, *updateReq.AddressLine1, addressData["address_line1"])
		assert.Equal(t, *updateReq.City, addressData["city"])
		assert.Equal(t, *updateReq.PostalCode, addressData["postal_code"])
	})

	suite.T().Run("DeleteAddress", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"DELETE",
			fmt.Sprintf("/api/v1/addresses/%s", addressID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
	})

	suite.T().Run("GetDeletedAddress", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/addresses/%s", addressID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})
}

// TestCustomerAddressManagement tests customer-specific address operations
func (suite *AddressManagementTestSuite) TestCustomerAddressManagement() {
	// Create a test customer
	testCustomer := suite.framework.GenerateTestCustomer("custaddr")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	// Create multiple addresses
	addresses := make([]string, 3)
	for i := 0; i < 3; i++ {
		createReq := dto.CreateAddressRequest{
			CustomerID:    customerID,
			Type:          entity.AddressTypeShipping,
			Label:         stringPtr(fmt.Sprintf("Address %d", i+1)),
			FirstName:     stringPtr(fmt.Sprintf("First %d", i+1)),
			LastName:      stringPtr(fmt.Sprintf("Last %d", i+1)),
			Phone:         stringPtr(fmt.Sprintf("+62812345%04d", i+6789)),
			AddressLine1:  fmt.Sprintf("Jl. Test Street No. %d", i+1),
			City:          "Jakarta",
			StateProvince: stringPtr("DKI Jakarta"),
			PostalCode:    "12190",
			Country:       "Indonesia",
			IsDefault:     i == 0, // First address is default
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			createReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusCreated)
		addressData := apiResp.Data.(map[string]interface{})
		addresses[i] = addressData["id"].(string)
	}

	suite.T().Run("GetCustomerAddresses", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify addresses array
		addressesData, ok := apiResp.Data.([]interface{})
		require.True(t, ok, "Addresses data should be an array")
		assert.Len(t, addressesData, 3, "Should have 3 addresses")

		// Verify first address is marked as default
		firstAddress := addressesData[0].(map[string]interface{})
		assert.True(t, firstAddress["is_default"].(bool), "First address should be default")
	})

	suite.T().Run("GetDefaultAddress", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/addresses/default", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify it's the default address
		addressData := apiResp.Data.(map[string]interface{})
		assert.True(t, addressData["is_default"].(bool), "Should be marked as default")
		assert.Equal(t, addresses[0], addressData["id"], "Should be the first address we created")
	})

	suite.T().Run("SetDefaultAddress", func(t *testing.T) {
		// Set second address as default
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses/%s/default", customerID.String(), addresses[1]),
			nil,
			customerID.String(),
		)

		suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify new default address
		resp, body = suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s/addresses/default", customerID.String()),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		addressData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, addresses[1], addressData["id"], "Second address should now be default")
	})
}

// TestAddressValidation tests address validation endpoint
func (suite *AddressManagementTestSuite) TestAddressValidation() {
	tests := []struct {
		name           string
		request        dto.AddressValidationRequest
		expectedStatus int
		expectValid    bool
		description    string
	}{
		{
			name: "ValidAddress",
			request: dto.AddressValidationRequest{
				AddressLine1: "Jl. Sudirman No. 123",
				AddressLine2: stringPtr("Apartment 4B"),
				City:         "Jakarta",
				State:        "DKI Jakarta",
				PostalCode:   "12190",
				Country:      "ID",
			},
			expectedStatus: http.StatusOK,
			expectValid:    true,
			description:    "Should validate a complete, valid address",
		},
		{
			name: "IncompleteAddress",
			request: dto.AddressValidationRequest{
				AddressLine1: "Jl. Sudirman",
				City:         "",
				State:        "DKI Jakarta",
				PostalCode:   "12190",
				Country:      "ID",
			},
			expectedStatus: http.StatusOK,
			expectValid:    false,
			description:    "Should identify missing city as invalid",
		},
		{
			name: "InvalidPostalCode",
			request: dto.AddressValidationRequest{
				AddressLine1: "Jl. Sudirman No. 123",
				City:         "Jakarta",
				State:        "DKI Jakarta",
				PostalCode:   "INVALID",
				Country:      "ID",
			},
			expectedStatus: http.StatusOK,
			expectValid:    false,
			description:    "Should identify invalid postal code format",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, body := suite.framework.MakeRequest("POST", "/api/v1/addresses/validate", tt.request, nil)

			apiResp := suite.framework.AssertSuccessResponse(resp, body, tt.expectedStatus)

			// Verify validation result
			validationData := apiResp.Data.(map[string]interface{})
			assert.Equal(t, tt.expectValid, validationData["valid"], "Validation result should match expected")
			assert.Contains(t, validationData, "formatted_address", "Should include formatted address")
			assert.Contains(t, validationData, "suggestions", "Should include suggestions")
		})
	}
}

// TestAddressGeocoding tests address geocoding endpoint
func (suite *AddressManagementTestSuite) TestAddressGeocoding() {
	tests := []struct {
		name           string
		request        dto.GeocodeRequest
		expectedStatus int
		expectSuccess  bool
		description    string
	}{
		{
			name: "ValidAddressGeocoding",
			request: dto.GeocodeRequest{
				Address: "Jl. Sudirman No. 123, Jakarta",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			description:    "Should geocode a valid address",
		},
		{
			name: "VagueAddressGeocoding",
			request: dto.GeocodeRequest{
				Address: "Jakarta",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			description:    "Should geocode city-level address with lower accuracy",
		},
		{
			name: "EmptyAddress",
			request: dto.GeocodeRequest{
				Address: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			description:    "Should reject empty address",
		},
		{
			name: "InvalidAddress",
			request: dto.GeocodeRequest{
				Address: "NonExistentPlace12345XYZ",
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
			description:    "Should handle address that cannot be geocoded",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, body := suite.framework.MakeRequest("POST", "/api/v1/addresses/geocode", tt.request, nil)

			if tt.expectSuccess {
				apiResp := suite.framework.AssertSuccessResponse(resp, body, tt.expectedStatus)

				// Verify geocoding result structure
				geocodeData := apiResp.Data.(map[string]interface{})
				assert.Contains(t, geocodeData, "latitude", "Should include latitude")
				assert.Contains(t, geocodeData, "longitude", "Should include longitude")
				assert.Contains(t, geocodeData, "formatted_address", "Should include formatted address")
				assert.Contains(t, geocodeData, "accuracy", "Should include accuracy level")
			} else {
				suite.framework.AssertErrorResponse(resp, body, tt.expectedStatus)
			}
		})
	}
}

// TestAddressValidationErrors tests address validation error scenarios
func (suite *AddressManagementTestSuite) TestAddressValidationErrors() {
	// Create a test customer
	testCustomer := suite.framework.GenerateTestCustomer("addrerr")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)
	customerID := customerResp.ID

	suite.T().Run("InvalidAddressID", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			"/api/v1/addresses/invalid-uuid",
			nil,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("NonExistentAddress", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/addresses/%s", nonExistentID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})

	suite.T().Run("CreateAddressWithMissingFields", func(t *testing.T) {
		incompleteReq := dto.CreateAddressRequest{
			CustomerID: customerID,
			Type:       entity.AddressTypeShipping,
			Label:      stringPtr("Incomplete"),
			// Missing required fields: AddressLine1, City, PostalCode, Country
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses", customerID.String()),
			incompleteReq,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("SetDefaultForNonExistentAddress", func(t *testing.T) {
		nonExistentAddressID := uuid.New().String()
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/addresses/%s/default", customerID.String(), nonExistentAddressID),
			nil,
			customerID.String(),
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})
}

// MockAddressHandler provides mock implementations for address endpoints
type MockAddressHandler struct{}

func (h *MockAddressHandler) GetCustomerAddresses(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Addresses retrieved successfully",
		"data":    []interface{}{},
	})
}

func (h *MockAddressHandler) CreateAddress(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Address created successfully",
		"data": gin.H{
			"id":             uuid.New().String(),
			"customer_id":    c.Param("id"),
			"label":          "Home",
			"recipient_name": "John Doe",
			"phone":          "+628123456789",
			"address_line_1": "Jl. Sudirman No. 123",
			"address_line_2": "Apartment 4B",
			"city":           "Jakarta",
			"state":          "DKI Jakarta",
			"postal_code":    "12190",
			"country":        "Indonesia",
			"is_default":     true,
			"created_at":     "2023-01-15T10:30:00Z",
			"updated_at":     "2023-01-15T10:30:00Z",
		},
	})
}

func (h *MockAddressHandler) GetAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address retrieved successfully",
		"data": gin.H{
			"id":             c.Param("id"),
			"customer_id":    uuid.New().String(),
			"label":          "Home",
			"recipient_name": "John Doe",
			"created_at":     "2023-01-15T10:30:00Z",
		},
	})
}

func (h *MockAddressHandler) UpdateAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address updated successfully",
		"data": gin.H{
			"id":             c.Param("id"),
			"label":          "Office",
			"recipient_name": "Jane Doe",
			"updated_at":     "2023-01-15T11:30:00Z",
		},
	})
}

func (h *MockAddressHandler) DeleteAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address deleted successfully",
	})
}

func (h *MockAddressHandler) SetDefaultAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Default address set successfully",
	})
}

func (h *MockAddressHandler) GetDefaultAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Default address retrieved successfully",
		"data": gin.H{
			"id":         uuid.New().String(),
			"is_default": true,
			"created_at": "2023-01-15T10:30:00Z",
		},
	})
}

func (h *MockAddressHandler) ValidateAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address validation completed",
		"data": gin.H{
			"valid":             true,
			"formatted_address": "Jl. Sudirman No. 123, Apartment 4B, Jakarta, DKI Jakarta 12190, Indonesia",
			"suggestions":       []string{},
			"issues":            []string{},
		},
	})
}

func (h *MockAddressHandler) GeocodeAddress(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Address geocoded successfully",
		"data": gin.H{
			"latitude":          -6.2088,
			"longitude":         106.8456,
			"formatted_address": "Jl. Sudirman No. 123, Jakarta, DKI Jakarta, Indonesia",
			"accuracy":          "precise",
		},
	})
}

// TestAddressManagementSuite runs the address management test suite
func TestAddressManagementSuite(t *testing.T) {
	suite.Run(t, new(AddressManagementTestSuite))
}
