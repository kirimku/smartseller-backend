package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	parentpkg "github.com/kirimku/smartseller-backend/internal/tests/integration"
)

// TestAddressIntegration tests the complete address management flow
func TestAddressIntegration(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping address integration test in short mode")
	}

	// Get auth token
	token := parentpkg.GetAuthToken(t)
	require.NotEmpty(t, token, "Authentication token should not be empty")

	// Define HTTP client with timeout
	client := parentpkg.CreateHTTPClient()

	// Define base URL
	baseURL := parentpkg.GetBaseURL()

	// Test variables to store state between tests
	var (
		provinces []string
		cities    []string
		districts []string
		postcodes []string
		addressID int64
	)

	// 1. Get Provinces
	t.Run("Get Provinces", func(t *testing.T) {
		url := fmt.Sprintf("%s/provinces", baseURL)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "Failed to create provinces request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send provinces request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for provinces")

		var response struct {
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Data    []string `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode provinces response")

		require.True(t, response.Success, "Expected success:true for provinces")
		require.NotEmpty(t, response.Data, "Expected non-empty provinces list")

		t.Logf("Got %d provinces", len(response.Data))
		provinces = response.Data
	})

	// 2. Get Cities for first province
	t.Run("Get Cities", func(t *testing.T) {
		if len(provinces) == 0 {
			t.Skip("No provinces available to get cities")
		}

		province := provinces[0]
		t.Logf("Getting cities for province: %s", province)

		// Build URL with query parameters
		u, err := url.Parse(fmt.Sprintf("%s/cities", baseURL))
		require.NoError(t, err, "Failed to parse cities URL")

		q := u.Query()
		q.Add("province", province)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		require.NoError(t, err, "Failed to create cities request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send cities request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for cities")

		var response struct {
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Data    []string `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode cities response")

		require.True(t, response.Success, "Expected success:true for cities")
		require.NotEmpty(t, response.Data, "Expected non-empty cities list")

		t.Logf("Got %d cities for province %s", len(response.Data), province)
		cities = response.Data
	})

	// 3. Get Districts for first province and city
	t.Run("Get Districts", func(t *testing.T) {
		if len(provinces) == 0 || len(cities) == 0 {
			t.Skip("No provinces or cities available to get districts")
		}

		province := provinces[0]
		city := cities[0]
		t.Logf("Getting districts for province: %s, city: %s", province, city)

		// Build URL with query parameters
		u, err := url.Parse(fmt.Sprintf("%s/districts", baseURL))
		require.NoError(t, err, "Failed to parse districts URL")

		q := u.Query()
		q.Add("province", province)
		q.Add("city", city)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		require.NoError(t, err, "Failed to create districts request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send districts request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for districts")

		var response struct {
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Data    []string `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode districts response")

		require.True(t, response.Success, "Expected success:true for districts")
		require.NotEmpty(t, response.Data, "Expected non-empty districts list")

		t.Logf("Got %d districts for province %s and city %s", len(response.Data), province, city)
		districts = response.Data
	})

	// 4. Get Postcodes for first province, city, and district
	t.Run("Get Postcodes", func(t *testing.T) {
		if len(provinces) == 0 || len(cities) == 0 || len(districts) == 0 {
			t.Skip("No provinces, cities or districts available to get postcodes")
		}

		province := provinces[0]
		city := cities[0]
		district := districts[0]
		t.Logf("Getting postcodes for province: %s, city: %s, district: %s", province, city, district)

		// Build URL with query parameters
		u, err := url.Parse(fmt.Sprintf("%s/postcodes", baseURL))
		require.NoError(t, err, "Failed to parse postcodes URL")

		q := u.Query()
		q.Add("province", province)
		q.Add("city", city)
		q.Add("district", district)
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(http.MethodGet, u.String(), nil)
		require.NoError(t, err, "Failed to create postcodes request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send postcodes request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for postcodes")

		var response struct {
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Data    []string `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode postcodes response")

		require.True(t, response.Success, "Expected success:true for postcodes")
		require.NotEmpty(t, response.Data, "Expected non-empty postcodes list")

		t.Logf("Got %d postcodes for province %s, city %s, and district %s",
			len(response.Data), province, city, district)
		postcodes = response.Data
	})

	// 5. Create Address
	t.Run("Create Address", func(t *testing.T) {
		if len(provinces) == 0 || len(cities) == 0 || len(districts) == 0 || len(postcodes) == 0 {
			t.Skip("Missing location data for creating address")
		}

		addressRequest := map[string]interface{}{
			"title":                   "Home",
			"name":                    "Test User",
			"phone":                   "081234567890",
			"email":                   "test@example.com",
			"main_address":            true,
			"geolocation_set_by_user": true,
			"address":                 "Jl. Test No. 123",
			"area":                    districts[0],
			"city":                    cities[0],
			"province":                provinces[0],
			"post_code":               postcodes[0],
			"latitude":                -6.1753924,
			"longitude":               106.8271528,
		}

		jsonPayload, err := json.Marshal(addressRequest)
		require.NoError(t, err, "Failed to marshal address request")

		url := fmt.Sprintf("%s/addresses", baseURL)
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		require.NoError(t, err, "Failed to create address creation request")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send address creation request")
		defer resp.Body.Close()

		t.Logf("Create address response status: %d", resp.StatusCode)

		// If unauthorized, the test token might not be valid
		if resp.StatusCode == http.StatusUnauthorized {
			t.Fatalf("Unauthorized: Invalid or expired token")
		}

		require.Equal(t, http.StatusCreated, resp.StatusCode, "Expected status Created for address creation")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID int64 `json:"id"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode address creation response")

		require.True(t, response.Success, "Expected success:true for address creation")
		require.NotZero(t, response.Data.ID, "Expected non-zero address ID")

		// Store address ID for later tests
		addressID = response.Data.ID
		t.Logf("Created address with ID: %d", addressID)
	})

	// 6. Get Address
	t.Run("Get Address", func(t *testing.T) {
		if addressID == 0 {
			t.Skip("No address ID available to retrieve")
		}

		url := fmt.Sprintf("%s/addresses/%d", baseURL, addressID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "Failed to create get address request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send get address request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for getting address")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID        int64   `json:"id"`
				Title     string  `json:"title"`
				Name      string  `json:"name"`
				Phone     string  `json:"phone"`
				Email     string  `json:"email"`
				Address   string  `json:"address"`
				Area      string  `json:"area"`
				City      string  `json:"city"`
				Province  string  `json:"province"`
				PostCode  string  `json:"post_code"`
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode get address response")

		require.True(t, response.Success, "Expected success:true for getting address")
		require.Equal(t, addressID, response.Data.ID, "Expected matching address ID")
		require.Equal(t, "Test User", response.Data.Name, "Expected correct name")
		require.Equal(t, "Jl. Test No. 123", response.Data.Address, "Expected correct address")

		t.Logf("Successfully retrieved address: %s at %s", response.Data.Name, response.Data.Address)
	})

	// 7. List Addresses
	t.Run("List Addresses", func(t *testing.T) {
		url := fmt.Sprintf("%s/addresses", baseURL)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		require.NoError(t, err, "Failed to create list addresses request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send list addresses request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for listing addresses")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    []struct {
				ID          int64  `json:"id"`
				Name        string `json:"name"`
				MainAddress bool   `json:"main_address"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode list addresses response")

		require.True(t, response.Success, "Expected success:true for listing addresses")
		require.NotEmpty(t, response.Data, "Expected non-empty addresses list")

		// Check if our created address is in the list
		var found bool
		for _, addr := range response.Data {
			if addr.ID == addressID {
				found = true
				break
			}
		}
		require.True(t, found, "Expected to find the created address in the list")

		t.Logf("Successfully listed %d addresses", len(response.Data))
	})

	// 8. Update Address
	t.Run("Update Address", func(t *testing.T) {
		if addressID == 0 {
			t.Skip("No address ID available to update")
		}

		// Prepare update payload
		updateRequest := map[string]interface{}{
			"title":                   "Office",
			"name":                    "Updated Test User",
			"phone":                   "087654321098",
			"email":                   "updated_test@example.com",
			"main_address":            true,
			"geolocation_set_by_user": true,
			"address":                 "Jl. Updated Test No. 456",
			"area":                    districts[0],
			"city":                    cities[0],
			"province":                provinces[0],
			"post_code":               postcodes[0],
			"latitude":                -6.2753924,
			"longitude":               106.7271528,
		}

		jsonPayload, err := json.Marshal(updateRequest)
		require.NoError(t, err, "Failed to marshal address update request")

		url := fmt.Sprintf("%s/addresses/%d", baseURL, addressID)
		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonPayload))
		require.NoError(t, err, "Failed to create address update request")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send address update request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for address update")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Data    struct {
				ID      int64  `json:"id"`
				Name    string `json:"name"`
				Title   string `json:"title"`
				Address string `json:"address"`
			} `json:"data"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode address update response")

		require.True(t, response.Success, "Expected success:true for address update")
		require.Equal(t, addressID, response.Data.ID, "Expected matching address ID")
		assert.Equal(t, "Updated Test User", response.Data.Name, "Expected updated name")
		assert.Equal(t, "Office", response.Data.Title, "Expected updated title")
		assert.Equal(t, "Jl. Updated Test No. 456", response.Data.Address, "Expected updated address")

		t.Logf("Successfully updated address to: %s at %s", response.Data.Name, response.Data.Address)
	})

	// 9. Set Main Address
	t.Run("Set Main Address", func(t *testing.T) {
		if addressID == 0 {
			t.Skip("No address ID available to set as main")
		}

		url := fmt.Sprintf("%s/addresses/%d/main", baseURL, addressID)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		require.NoError(t, err, "Failed to create set main address request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send set main address request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for setting main address")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode set main address response")

		require.True(t, response.Success, "Expected success:true for setting main address")

		t.Logf("Successfully set address %d as main address", addressID)
	})

	// 10. Delete Address
	t.Run("Delete Address", func(t *testing.T) {
		if addressID == 0 {
			t.Skip("No address ID available to delete")
		}

		url := fmt.Sprintf("%s/addresses/%d", baseURL, addressID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err, "Failed to create delete address request")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		resp, err := client.Do(req)
		require.NoError(t, err, "Failed to send delete address request")
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK for deleting address")

		var response struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "Failed to decode delete address response")

		require.True(t, response.Success, "Expected success:true for deleting address")

		t.Logf("Successfully deleted address %d", addressID)

		// Verify address is deleted by trying to get it
		getReq, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/addresses/%d", baseURL, addressID), nil)
		require.NoError(t, err, "Failed to create verification request")
		getReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

		getResp, err := client.Do(getReq)
		require.NoError(t, err, "Failed to send verification request")
		defer getResp.Body.Close()

		// We expect either a 404 Not Found or a 200 OK with a "deleted" flag
		if getResp.StatusCode == http.StatusNotFound {
			t.Log("Address confirmed deleted: received 404 Not Found response")
		} else {
			var getResponse map[string]interface{}
			err = json.NewDecoder(getResp.Body).Decode(&getResponse)
			require.NoError(t, err, "Failed to decode verification response")
			t.Logf("Get deleted address response: %v", getResponse)
		}
	})
}
