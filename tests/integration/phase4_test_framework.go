package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kirimku/smartseller-backend/internal/application/dto"
)

// Phase4TestFramework provides utilities for testing Phase 4 API endpoints
type Phase4TestFramework struct {
	t      *testing.T
	router *gin.Engine
	server *httptest.Server

	// Test data
	testStorefrontID uuid.UUID
	testCustomers    []*TestCustomer
	testAddresses    []*TestAddress
	testStorefronts  []*TestStorefront
	authTokens       map[string]string // customer_id -> token
}

// TestCustomer represents a test customer with expected data
type TestCustomer struct {
	ID          uuid.UUID  `json:"id"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Phone       *string    `json:"phone,omitempty"`
	DateOfBirth *time.Time `json:"date_of_birth,omitempty"`
	Password    string     `json:"-"` // Only for testing, not in responses
}

// TestAddress represents a test address
type TestAddress struct {
	ID            uuid.UUID `json:"id"`
	CustomerID    uuid.UUID `json:"customer_id"`
	Label         string    `json:"label"`
	RecipientName string    `json:"recipient_name"`
	Phone         string    `json:"phone"`
	AddressLine1  string    `json:"address_line_1"`
	AddressLine2  *string   `json:"address_line_2,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	PostalCode    string    `json:"postal_code"`
	Country       string    `json:"country"`
	IsDefault     bool      `json:"is_default"`
}

// TestStorefront represents a test storefront
type TestStorefront struct {
	ID          uuid.UUID `json:"id"`
	CustomerID  uuid.UUID `json:"customer_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Domain      string    `json:"domain"`
	Description *string   `json:"description,omitempty"`
	Theme       string    `json:"theme"`
	Status      string    `json:"status"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    interface{}            `json:"data,omitempty"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// APIError represents an API error response
type APIError struct {
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	Error            string                 `json:"error"`
	ErrorDetail      string                 `json:"error_detail,omitempty"`
	ValidationErrors []string               `json:"validation_errors,omitempty"`
	Meta             map[string]interface{} `json:"meta,omitempty"`
}

// NewPhase4TestFramework creates a new test framework
func NewPhase4TestFramework(t *testing.T, router *gin.Engine) *Phase4TestFramework {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	server := httptest.NewServer(router)

	return &Phase4TestFramework{
		t:                t,
		router:           router,
		server:           server,
		authTokens:       make(map[string]string),
		testCustomers:    make([]*TestCustomer, 0),
		testAddresses:    make([]*TestAddress, 0),
		testStorefronts:  make([]*TestStorefront, 0),
		testStorefrontID: uuid.New(), // Generate a test storefront ID
	}
}

// Cleanup cleans up test resources
func (f *Phase4TestFramework) Cleanup() {
	if f.server != nil {
		f.server.Close()
	}

	// Clean up test data
	f.testCustomers = f.testCustomers[:0]
	f.testAddresses = f.testAddresses[:0]
	f.testStorefronts = f.testStorefronts[:0]
	f.authTokens = make(map[string]string)
}

// MakeRequest makes an HTTP request to the test server
func (f *Phase4TestFramework) MakeRequest(method, path string, body interface{}, headers map[string]string) (*http.Response, []byte) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(f.t, err, "Failed to marshal request body")
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, f.server.URL+path, reqBody)
	require.NoError(f.t, err, "Failed to create request")

	// Set default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(f.t, err, "Failed to make request")

	responseBody, err := io.ReadAll(resp.Body)
	require.NoError(f.t, err, "Failed to read response body")
	resp.Body.Close()

	return resp, responseBody
}

// MakeAuthenticatedRequest makes an authenticated request using stored token
func (f *Phase4TestFramework) MakeAuthenticatedRequest(method, path string, body interface{}, customerID string) (*http.Response, []byte) {
	headers := make(map[string]string)

	if token, exists := f.authTokens[customerID]; exists {
		headers["Authorization"] = "Bearer " + token
	}

	return f.MakeRequest(method, path, body, headers)
}

// AssertSuccessResponse asserts that the response is successful and returns parsed data
func (f *Phase4TestFramework) AssertSuccessResponse(resp *http.Response, body []byte, expectedStatus int) *APIResponse {
	assert.Equal(f.t, expectedStatus, resp.StatusCode, "Unexpected status code. Response body: %s", string(body))

	var apiResp APIResponse
	err := json.Unmarshal(body, &apiResp)
	require.NoError(f.t, err, "Failed to unmarshal success response")

	assert.True(f.t, apiResp.Success, "Expected success=true in response")
	assert.NotEmpty(f.t, apiResp.Message, "Expected non-empty message in response")

	return &apiResp
}

// AssertErrorResponse asserts that the response is an error and returns parsed error
func (f *Phase4TestFramework) AssertErrorResponse(resp *http.Response, body []byte, expectedStatus int) *APIError {
	assert.Equal(f.t, expectedStatus, resp.StatusCode, "Unexpected status code. Response body: %s", string(body))

	var apiErr APIError
	err := json.Unmarshal(body, &apiErr)
	require.NoError(f.t, err, "Failed to unmarshal error response")

	assert.False(f.t, apiErr.Success, "Expected success=false in error response")
	assert.NotEmpty(f.t, apiErr.Message, "Expected non-empty message in error response")

	return &apiErr
}

// GenerateTestCustomer generates test customer data
func (f *Phase4TestFramework) GenerateTestCustomer(suffix string) *TestCustomer {
	phone := "+62812345" + fmt.Sprintf("%04d", len(f.testCustomers)+1000)
	dateOfBirth := time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC)

	customer := &TestCustomer{
		ID:          uuid.New(),
		Email:       fmt.Sprintf("customer_%s@test.com", suffix),
		FirstName:   "John" + suffix,
		LastName:    "Doe" + suffix,
		Phone:       &phone,
		DateOfBirth: &dateOfBirth,
		Password:    "SecurePassword123",
	}

	f.testCustomers = append(f.testCustomers, customer)
	return customer
}

// RegisterTestCustomer registers a customer and stores auth token
func (f *Phase4TestFramework) RegisterTestCustomer(customer *TestCustomer) *dto.CustomerResponse {
	registrationReq := &dto.CustomerRegistrationRequest{
		Email:       customer.Email,
		Password:    customer.Password,
		FirstName:   customer.FirstName,
		LastName:    customer.LastName,
		Phone:       customer.Phone,
		DateOfBirth: customer.DateOfBirth,
	}

	resp, body := f.MakeRequest("POST", "/api/v1/customers/register", registrationReq, nil)
	apiResp := f.AssertSuccessResponse(resp, body, http.StatusCreated)

	// Parse customer data
	customerDataBytes, err := json.Marshal(apiResp.Data)
	require.NoError(f.t, err, "Failed to marshal customer data")

	var customerResp dto.CustomerResponse
	err = json.Unmarshal(customerDataBytes, &customerResp)
	require.NoError(f.t, err, "Failed to unmarshal customer response")

	// Update customer ID from response
	customer.ID = customerResp.ID

	// For testing purposes, we'll generate a mock JWT token
	// In a real implementation, this would come from authentication
	f.authTokens[customer.ID.String()] = f.generateMockJWTToken(customer.ID.String())

	return &customerResp
}

// GenerateTestAddress generates test address data
func (f *Phase4TestFramework) GenerateTestAddress(customerID uuid.UUID, suffix string) *TestAddress {
	addressLine2 := "Apartment " + suffix

	address := &TestAddress{
		ID:            uuid.New(),
		CustomerID:    customerID,
		Label:         "Address " + suffix,
		RecipientName: "John Doe " + suffix,
		Phone:         "+62812345" + fmt.Sprintf("%04d", len(f.testAddresses)+2000),
		AddressLine1:  fmt.Sprintf("Jl. Test Street No. %s", suffix),
		AddressLine2:  &addressLine2,
		City:          "Jakarta",
		State:         "DKI Jakarta",
		PostalCode:    "12190",
		Country:       "Indonesia",
		IsDefault:     len(f.testAddresses) == 0, // First address is default
	}

	f.testAddresses = append(f.testAddresses, address)
	return address
}

// GenerateTestStorefront generates test storefront data
func (f *Phase4TestFramework) GenerateTestStorefront(customerID uuid.UUID, suffix string) *TestStorefront {
	description := "Test store description " + suffix

	storefront := &TestStorefront{
		ID:          uuid.New(),
		CustomerID:  customerID,
		Name:        "Test Store " + suffix,
		Slug:        "test-store-" + suffix,
		Domain:      "teststore" + suffix + ".smartseller.com",
		Description: &description,
		Theme:       "modern",
		Status:      "active",
	}

	f.testStorefronts = append(f.testStorefronts, storefront)
	return storefront
}

// generateMockJWTToken generates a mock JWT token for testing
// In production, this would be handled by the authentication service
func (f *Phase4TestFramework) generateMockJWTToken(customerID string) string {
	// This is a mock token for testing purposes
	// In a real implementation, you'd use a proper JWT library
	return fmt.Sprintf("mock-jwt-token-%s-%d", customerID, time.Now().Unix())
}

// WaitForAsyncOperations waits for async operations to complete
func (f *Phase4TestFramework) WaitForAsyncOperations() {
	time.Sleep(100 * time.Millisecond) // Small delay for async operations
}

// AssertCustomerResponse validates customer response structure
func (f *Phase4TestFramework) AssertCustomerResponse(data interface{}, expectedEmail string) {
	customerData, ok := data.(map[string]interface{})
	require.True(f.t, ok, "Customer data should be an object")

	assert.NotEmpty(f.t, customerData["id"], "Customer ID should not be empty")
	assert.Equal(f.t, expectedEmail, customerData["email"], "Customer email should match")
	assert.NotEmpty(f.t, customerData["first_name"], "First name should not be empty")
	assert.NotEmpty(f.t, customerData["last_name"], "Last name should not be empty")
	assert.Equal(f.t, "active", customerData["status"], "Customer should be active")
	assert.NotEmpty(f.t, customerData["created_at"], "Created timestamp should not be empty")
}

// AssertAddressResponse validates address response structure
func (f *Phase4TestFramework) AssertAddressResponse(data interface{}, expectedCustomerID uuid.UUID) {
	addressData, ok := data.(map[string]interface{})
	require.True(f.t, ok, "Address data should be an object")

	assert.NotEmpty(f.t, addressData["id"], "Address ID should not be empty")
	assert.Equal(f.t, expectedCustomerID.String(), addressData["customer_id"], "Customer ID should match")
	assert.NotEmpty(f.t, addressData["label"], "Address label should not be empty")
	assert.NotEmpty(f.t, addressData["recipient_name"], "Recipient name should not be empty")
	assert.NotEmpty(f.t, addressData["address_line_1"], "Address line 1 should not be empty")
	assert.NotEmpty(f.t, addressData["city"], "City should not be empty")
	assert.NotEmpty(f.t, addressData["created_at"], "Created timestamp should not be empty")
}

// AssertStorefrontResponse validates storefront response structure
func (f *Phase4TestFramework) AssertStorefrontResponse(data interface{}, expectedCustomerID uuid.UUID) {
	storefrontData, ok := data.(map[string]interface{})
	require.True(f.t, ok, "Storefront data should be an object")

	assert.NotEmpty(f.t, storefrontData["id"], "Storefront ID should not be empty")
	assert.Equal(f.t, expectedCustomerID.String(), storefrontData["customer_id"], "Customer ID should match")
	assert.NotEmpty(f.t, storefrontData["name"], "Storefront name should not be empty")
	assert.NotEmpty(f.t, storefrontData["slug"], "Storefront slug should not be empty")
	assert.NotEmpty(f.t, storefrontData["domain"], "Storefront domain should not be empty")
	assert.Equal(f.t, "active", storefrontData["status"], "Storefront should be active")
	assert.NotEmpty(f.t, storefrontData["created_at"], "Created timestamp should not be empty")
}

// AssertPaginatedResponse validates paginated response structure
func (f *Phase4TestFramework) AssertPaginatedResponse(data interface{}, expectedMinItems int) {
	paginatedData, ok := data.(map[string]interface{})
	require.True(f.t, ok, "Paginated data should be an object")

	// Check pagination metadata
	pagination, ok := paginatedData["pagination"].(map[string]interface{})
	require.True(f.t, ok, "Pagination metadata should exist")

	assert.NotNil(f.t, pagination["total"], "Total should not be nil")
	assert.NotNil(f.t, pagination["per_page"], "Per page should not be nil")
	assert.NotNil(f.t, pagination["current_page"], "Current page should not be nil")
	assert.NotNil(f.t, pagination["last_page"], "Last page should not be nil")

	// Check data array
	items, ok := paginatedData["data"].([]interface{})
	require.True(f.t, ok, "Data should be an array")

	if expectedMinItems > 0 {
		assert.GreaterOrEqual(f.t, len(items), expectedMinItems, "Should have minimum expected items")
	}
}

// GetTestStorefrontID returns the test storefront ID
func (f *Phase4TestFramework) GetTestStorefrontID() uuid.UUID {
	return f.testStorefrontID
}

// GetTestCustomers returns all test customers
func (f *Phase4TestFramework) GetTestCustomers() []*TestCustomer {
	return f.testCustomers
}

// GetTestAddresses returns all test addresses
func (f *Phase4TestFramework) GetTestAddresses() []*TestAddress {
	return f.testAddresses
}

// GetTestStorefronts returns all test storefronts
func (f *Phase4TestFramework) GetTestStorefronts() []*TestStorefront {
	return f.testStorefronts
}
