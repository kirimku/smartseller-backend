package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kirimku/smartseller-backend/tests/integration/setup"
)

// Global test setup instance
var testSetup *setup.TestSetup

// TestMain sets up and tears down the test suite
func TestMain(m *testing.M) {
	var err error

	// Setup test suite
	testSetup, err = setup.TestSuiteSetup()
	if err != nil {
		fmt.Printf("Failed to setup test suite: %v\n", err)
		return
	}

	// Run tests
	code := m.Run()

	// Cleanup test suite
	if err := setup.TestSuiteTeardown(testSetup); err != nil {
		fmt.Printf("Failed to teardown test suite: %v\n", err)
	}

	// Exit with test result code
	os.Exit(code)
}

// TestProductCRUD_CreateProduct_Success demonstrates product creation test
func TestProductCRUD_CreateProduct_Success(t *testing.T) {
	// Setup individual test
	testSetup.SetupTest(t)
	defer testSetup.TeardownTest(t)

	// Get test category ID
	categoryID, err := testSetup.GetTestProductCategory("electronics")
	require.NoError(t, err, "Failed to get test category")

	// Create product request
	createReq := map[string]interface{}{
		"name":        "Test Smartphone",
		"description": "A test smartphone for integration testing",
		"price":       299.99,
		"category_id": categoryID,
		"weight":      150.0,
		"dimensions": map[string]float64{
			"length": 14.2,
			"width":  7.1,
			"height": 0.8,
		},
	}

	// Create authenticated request using shared auth helper
	req, err := testSetup.Auth.AuthenticatedTestRequest("POST", "/api/v1/products", createReq)
	require.NoError(t, err, "Failed to create authenticated request")

	// Execute request (this would be done through your HTTP handler in actual test)
	// For demonstration, we're showing the structure
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err, "Request failed")
	defer resp.Body.Close()

	// Verify response
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdProduct map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdProduct)
	require.NoError(t, err, "Failed to decode response")

	// Verify product data
	assert.Equal(t, "Test Smartphone", createdProduct["name"])
	assert.Equal(t, 299.99, createdProduct["price"])
	assert.Equal(t, "active", createdProduct["status"]) // Default status
	assert.NotEmpty(t, createdProduct["id"])
	assert.NotEmpty(t, createdProduct["sku"]) // Auto-generated

	// Verify the product was created in database
	productID := createdProduct["id"].(string)
	var dbProductName string
	err = testSetup.DB.QueryRow("SELECT name FROM products WHERE id = $1", productID).Scan(&dbProductName)
	require.NoError(t, err, "Failed to query created product")
	assert.Equal(t, "Test Smartphone", dbProductName)
}

// TestProductCRUD_CreateProduct_ValidationErrors demonstrates validation error handling
func TestProductCRUD_CreateProduct_ValidationErrors(t *testing.T) {
	testSetup.SetupTest(t)
	defer testSetup.TeardownTest(t)

	testCases := []struct {
		name           string
		request        map[string]interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Missing required name field",
			request:        map[string]interface{}{"price": 99.99},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "name is required",
		},
		{
			name: "Invalid price (negative)",
			request: map[string]interface{}{
				"name":  "Test Product",
				"price": -10.0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "price must be positive",
		},
		{
			name: "Invalid category ID",
			request: map[string]interface{}{
				"name":        "Test Product",
				"price":       99.99,
				"category_id": "invalid-uuid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid category_id format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create authenticated request using shared auth helper
			req, err := testSetup.Auth.AuthenticatedTestRequest("POST", "/api/v1/products", tc.request)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)

			var errorResp map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&errorResp)
			require.NoError(t, err)

			assert.Contains(t, errorResp["message"].(string), tc.expectedError)
		})
	}
}

// TestProductCRUD_GetProduct_Success demonstrates product retrieval
func TestProductCRUD_GetProduct_Success(t *testing.T) {
	testSetup.SetupTest(t)
	defer testSetup.TeardownTest(t)

	// Create a test product first
	productID, err := testSetup.CreateTestProduct("Test Product", "TEST-001", 99.99, "electronics")
	require.NoError(t, err, "Failed to create test product")

	// Get the product using authenticated request
	req, err := testSetup.Auth.AuthenticatedTestRequest("GET", "/api/v1/products/"+productID, nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var product map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&product)
	require.NoError(t, err)

	assert.Equal(t, productID, product["id"])
	assert.Equal(t, "Test Product", product["name"])
	assert.Equal(t, "TEST-001", product["sku"])
	assert.Equal(t, 99.99, product["price"])
}

// TestProductCRUD_ListProducts_WithFilters demonstrates product listing with filters
func TestProductCRUD_ListProducts_WithFilters(t *testing.T) {
	testSetup.SetupTest(t)
	defer testSetup.TeardownTest(t)

	// Create multiple test products
	electronicsCategory, err := testSetup.GetTestProductCategory("electronics")
	require.NoError(t, err)

	clothingCategory, err := testSetup.GetTestProductCategory("clothing")
	require.NoError(t, err)

	// Create products in different categories with different prices
	products := []struct {
		name       string
		sku        string
		price      float64
		categoryID string
		status     string
	}{
		{"Laptop", "LAP-001", 999.99, electronicsCategory, "active"},
		{"Phone", "PHN-001", 599.99, electronicsCategory, "active"},
		{"Shirt", "SHT-001", 29.99, clothingCategory, "active"},
		{"Jeans", "JNS-001", 79.99, clothingCategory, "inactive"},
	}

	for _, p := range products {
		// Create products with specific status
		_, err = testSetup.DB.Exec(`
			INSERT INTO products (id, name, sku, price, category_id, status, created_at, updated_at)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, NOW(), NOW())
		`, p.name, p.sku, p.price, p.categoryID, p.status)
		require.NoError(t, err)
	}

	testCases := []struct {
		name             string
		queryParams      string
		expectedCount    int
		expectedProducts []string
	}{
		{
			name:             "All active products",
			queryParams:      "?status=active",
			expectedCount:    3,
			expectedProducts: []string{"Laptop", "Phone", "Shirt"},
		},
		{
			name:             "Electronics category only",
			queryParams:      "?category_id=" + electronicsCategory,
			expectedCount:    2,
			expectedProducts: []string{"Laptop", "Phone"},
		},
		{
			name:             "Price range filter",
			queryParams:      "?min_price=50&max_price=700",
			expectedCount:    2,
			expectedProducts: []string{"Phone", "Jeans"},
		},
		{
			name:             "Search by name",
			queryParams:      "?q=phone",
			expectedCount:    1,
			expectedProducts: []string{"Phone"},
		},
		{
			name:          "Pagination",
			queryParams:   "?page=1&limit=2",
			expectedCount: 2, // Should return first 2 products
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := testSetup.Auth.AuthenticatedTestRequest("GET", "/api/v1/products"+tc.queryParams, nil)
			require.NoError(t, err)

			client := &http.Client{}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)

			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			data := response["data"].([]interface{})
			assert.Equal(t, tc.expectedCount, len(data))

			if len(tc.expectedProducts) > 0 {
				var actualNames []string
				for _, item := range data {
					product := item.(map[string]interface{})
					actualNames = append(actualNames, product["name"].(string))
				}

				for _, expectedName := range tc.expectedProducts {
					assert.Contains(t, actualNames, expectedName)
				}
			}

			// Verify pagination metadata
			if meta, exists := response["meta"]; exists {
				metaData := meta.(map[string]interface{})
				assert.NotNil(t, metaData["total"])
				assert.NotNil(t, metaData["page"])
				assert.NotNil(t, metaData["limit"])
			}
		})
	}
}

// TestAuthentication_InvalidToken demonstrates authentication error handling
func TestAuthentication_InvalidToken(t *testing.T) {
	testSetup.SetupTest(t)
	defer testSetup.TeardownTest(t)

	// Create request with invalid token
	req, err := http.NewRequest("GET", testSetup.Config.BaseURL+"/api/v1/products", nil)
	require.NoError(t, err)

	req.Header.Set("Authorization", "Bearer invalid-token")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var errorResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp["message"].(string), "invalid token")
}

// TestAuthentication_MissingToken demonstrates missing authentication error
func TestAuthentication_MissingToken(t *testing.T) {
	// Create request without authentication header
	req, err := http.NewRequest("GET", testSetup.Config.BaseURL+"/api/v1/products", nil)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
