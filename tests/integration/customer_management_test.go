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
	"github.com/kirimku/smartseller-backend/internal/interfaces/api/handler"
)

// CustomerManagementTestSuite tests all customer management endpoints
type CustomerManagementTestSuite struct {
	suite.Suite
	framework *Phase4TestFramework
	router    *gin.Engine
}

// SetupSuite initializes the test suite
func (suite *CustomerManagementTestSuite) SetupSuite() {
	// Initialize Gin router with customer routes
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Add middleware
	suite.router.Use(gin.Recovery())
	suite.router.Use(func(c *gin.Context) {
		c.Set("storefront_id", uuid.New())
		c.Next()
	})

	// Initialize customer handler and routes
	// Note: In a real implementation, you would inject dependencies properly
	customerHandler := &handler.CustomerHandler{} // This would be properly initialized

	v1 := suite.router.Group("/api/v1")
	{
		// Customer routes
		customers := v1.Group("/customers")
		{
			customers.POST("/register", customerHandler.RegisterCustomer)
			customers.GET("/:id", customerHandler.GetCustomer)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.GET("/by-email", customerHandler.GetCustomerByEmail)
			customers.POST("/:id/deactivate", customerHandler.DeactivateCustomer)
			customers.POST("/:id/reactivate", customerHandler.ReactivateCustomer)
			customers.GET("/search", customerHandler.SearchCustomers)
			customers.GET("/stats", customerHandler.GetCustomerStats)

			// Customer address routes
			customers.GET("/:id/addresses", customerHandler.GetCustomerAddresses)
			customers.POST("/:id/addresses", customerHandler.CreateCustomerAddress)
			customers.POST("/:customer_id/addresses/:address_id/default", customerHandler.SetDefaultAddress)
			customers.GET("/:id/addresses/default", customerHandler.GetDefaultAddress)
		}
	}

	suite.framework = NewPhase4TestFramework(suite.T(), suite.router)
}

// TearDownSuite cleans up after all tests
func (suite *CustomerManagementTestSuite) TearDownSuite() {
	if suite.framework != nil {
		suite.framework.Cleanup()
	}
}

// TestCustomerRegistration tests customer registration endpoint
func (suite *CustomerManagementTestSuite) TestCustomerRegistration() {
	tests := []struct {
		name           string
		request        dto.CustomerRegistrationRequest
		expectedStatus int
		expectSuccess  bool
		description    string
	}{
		{
			name: "ValidRegistration",
			request: dto.CustomerRegistrationRequest{
				Email:       "newcustomer@test.com",
				Password:    "SecurePassword123",
				FirstName:   "John",
				LastName:    "Doe",
				Phone:       stringPtr("+628123456789"),
				DateOfBirth: timePtr("1990-01-15"),
			},
			expectedStatus: http.StatusCreated,
			expectSuccess:  true,
			description:    "Should successfully register a new customer",
		},
		{
			name: "DuplicateEmail",
			request: dto.CustomerRegistrationRequest{
				Email:     "newcustomer@test.com", // Same as above
				Password:  "SecurePassword123",
				FirstName: "Jane",
				LastName:  "Smith",
			},
			expectedStatus: http.StatusConflict,
			expectSuccess:  false,
			description:    "Should reject duplicate email registration",
		},
		{
			name: "InvalidEmail",
			request: dto.CustomerRegistrationRequest{
				Email:     "invalid-email",
				Password:  "SecurePassword123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			description:    "Should reject invalid email format",
		},
		{
			name: "WeakPassword",
			request: dto.CustomerRegistrationRequest{
				Email:     "weakpass@test.com",
				Password:  "123",
				FirstName: "John",
				LastName:  "Doe",
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			description:    "Should reject weak password",
		},
		{
			name: "MissingRequiredFields",
			request: dto.CustomerRegistrationRequest{
				Email:    "incomplete@test.com",
				Password: "SecurePassword123",
				// Missing FirstName and LastName
			},
			expectedStatus: http.StatusBadRequest,
			expectSuccess:  false,
			description:    "Should reject registration with missing required fields",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, body := suite.framework.MakeRequest("POST", "/api/v1/customers/register", tt.request, nil)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code should match. Response: %s", string(body))

			if tt.expectSuccess {
				apiResp := suite.framework.AssertSuccessResponse(resp, body, tt.expectedStatus)
				suite.framework.AssertCustomerResponse(apiResp.Data, tt.request.Email)

				// Verify specific fields
				customerData := apiResp.Data.(map[string]interface{})
				assert.Equal(t, tt.request.FirstName, customerData["first_name"])
				assert.Equal(t, tt.request.LastName, customerData["last_name"])
				if tt.request.Phone != nil {
					assert.Equal(t, *tt.request.Phone, customerData["phone"])
				}
			} else {
				suite.framework.AssertErrorResponse(resp, body, tt.expectedStatus)
			}
		})
	}
}

// TestCustomerProfileManagement tests customer profile operations
func (suite *CustomerManagementTestSuite) TestCustomerProfileManagement() {
	// Create a test customer
	testCustomer := suite.framework.GenerateTestCustomer("profile")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)

	customerID := customerResp.ID

	suite.T().Run("GetCustomerProfile", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s", customerID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertCustomerResponse(apiResp.Data, testCustomer.Email)
	})

	suite.T().Run("UpdateCustomerProfile", func(t *testing.T) {
		updateReq := dto.CustomerUpdateRequest{
			FirstName:   stringPtr("Updated John"),
			LastName:    stringPtr("Updated Doe"),
			Phone:       stringPtr("+628123456790"),
			DateOfBirth: timePtr("1985-05-20"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/customers/%s", customerID),
			updateReq,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify updated fields
		customerData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, *updateReq.FirstName, customerData["first_name"])
		assert.Equal(t, *updateReq.LastName, customerData["last_name"])
		assert.Equal(t, *updateReq.Phone, customerData["phone"])
		assert.Equal(t, *updateReq.DateOfBirth, customerData["date_of_birth"])
	})

	suite.T().Run("GetCustomerByEmail", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/by-email?email=%s", testCustomer.Email),
			nil,
			nil,
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertCustomerResponse(apiResp.Data, testCustomer.Email)
	})
}

// TestCustomerAccountManagement tests account activation/deactivation
func (suite *CustomerManagementTestSuite) TestCustomerAccountManagement() {
	// Create a test customer
	testCustomer := suite.framework.GenerateTestCustomer("account")
	customerResp := suite.framework.RegisterTestCustomer(testCustomer)

	customerID := customerResp.ID

	suite.T().Run("DeactivateCustomer", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/deactivate", customerID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify customer is deactivated
		customerData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, "inactive", customerData["status"])
	})

	suite.T().Run("ReactivateCustomer", func(t *testing.T) {
		resp, body := suite.framework.MakeAuthenticatedRequest(
			"POST",
			fmt.Sprintf("/api/v1/customers/%s/reactivate", customerID),
			nil,
			customerID.String(),
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

		// Verify customer is reactivated
		customerData := apiResp.Data.(map[string]interface{})
		assert.Equal(t, "active", customerData["status"])
	})
}

// TestCustomerSearch tests customer search functionality
func (suite *CustomerManagementTestSuite) TestCustomerSearch() {
	// Create multiple test customers for search testing
	customers := make([]*TestCustomer, 3)
	for i := 0; i < 3; i++ {
		customer := suite.framework.GenerateTestCustomer(fmt.Sprintf("search%d", i))
		suite.framework.RegisterTestCustomer(customer)
		customers[i] = customer
	}

	suite.framework.WaitForAsyncOperations() // Allow time for indexing

	suite.T().Run("SearchCustomersWithQuery", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			"/api/v1/customers/search?query=search&page=1&page_size=10",
			nil,
			nil,
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertPaginatedResponse(apiResp.Data, 1) // At least 1 customer should match
	})

	suite.T().Run("SearchCustomersWithStatus", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			"/api/v1/customers/search?status=active&page_size=20",
			nil,
			nil,
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertPaginatedResponse(apiResp.Data, 0)
	})

	suite.T().Run("SearchCustomersWithSorting", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			"/api/v1/customers/search?sort_by=created_at&sort_dir=desc&page_size=5",
			nil,
			nil,
		)

		apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)
		suite.framework.AssertPaginatedResponse(apiResp.Data, 0)
	})
}

// TestCustomerStats tests customer statistics endpoint
func (suite *CustomerManagementTestSuite) TestCustomerStats() {
	// Create some test customers to generate statistics
	for i := 0; i < 2; i++ {
		customer := suite.framework.GenerateTestCustomer(fmt.Sprintf("stats%d", i))
		suite.framework.RegisterTestCustomer(customer)
	}

	suite.framework.WaitForAsyncOperations() // Allow time for stats calculation

	tests := []struct {
		name   string
		params string
		desc   string
	}{
		{
			name:   "DefaultStats",
			params: "",
			desc:   "Should return default 30-day statistics",
		},
		{
			name:   "WeeklyStats",
			params: "?period=7d",
			desc:   "Should return 7-day statistics",
		},
		{
			name: "CustomDateRange",
			params: fmt.Sprintf("?start_date=%s&end_date=%s",
				time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
				time.Now().Format("2006-01-02")),
			desc: "Should return statistics for custom date range",
		},
	}

	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			resp, body := suite.framework.MakeRequest(
				"GET",
				"/api/v1/customers/stats"+tt.params,
				nil,
				nil,
			)

			apiResp := suite.framework.AssertSuccessResponse(resp, body, http.StatusOK)

			// Verify statistics structure
			statsData, ok := apiResp.Data.(map[string]interface{})
			require.True(t, ok, "Stats data should be an object")

			assert.Contains(t, statsData, "total_customers", "Should include total customers")
			assert.Contains(t, statsData, "active_customers", "Should include active customers")
			assert.Contains(t, statsData, "new_customers_today", "Should include new customers today")
			assert.Contains(t, statsData, "customer_growth_rate", "Should include growth rate")
		})
	}
}

// TestCustomerValidation tests various validation scenarios
func (suite *CustomerManagementTestSuite) TestCustomerValidation() {
	suite.T().Run("InvalidCustomerID", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			"/api/v1/customers/invalid-uuid",
			nil,
			nil,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("NonExistentCustomer", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		resp, body := suite.framework.MakeRequest(
			"GET",
			fmt.Sprintf("/api/v1/customers/%s", nonExistentID),
			nil,
			nil,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})

	suite.T().Run("UpdateNonExistentCustomer", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		updateReq := dto.CustomerUpdateRequest{
			FirstName: stringPtr("Test"),
		}

		resp, body := suite.framework.MakeAuthenticatedRequest(
			"PUT",
			fmt.Sprintf("/api/v1/customers/%s", nonExistentID),
			updateReq,
			nonExistentID,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusNotFound)
	})
}

// TestCustomerErrorHandling tests error handling scenarios
func (suite *CustomerManagementTestSuite) TestCustomerErrorHandling() {
	suite.T().Run("MalformedRequestBody", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"POST",
			"/api/v1/customers/register",
			"invalid json",
			nil,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("EmptyRequestBody", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"POST",
			"/api/v1/customers/register",
			map[string]interface{}{},
			nil,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})

	suite.T().Run("InvalidSearchParameters", func(t *testing.T) {
		resp, body := suite.framework.MakeRequest(
			"GET",
			"/api/v1/customers/search?page=-1&page_size=0",
			nil,
			nil,
		)

		suite.framework.AssertErrorResponse(resp, body, http.StatusBadRequest)
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Helper function to create time pointers from date strings
func timePtr(dateStr string) *time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return &t
}

// TestCustomerManagementSuite runs the customer management test suite
func TestCustomerManagementSuite(t *testing.T) {
	suite.Run(t, new(CustomerManagementTestSuite))
}
