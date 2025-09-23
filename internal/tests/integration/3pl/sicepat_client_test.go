package integration

import (
    "context"
    "encoding/json"
    "net/http"
    "os"
    "strings"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/kirimku/smartseller-backend/internal/config"
    "github.com/kirimku/smartseller-backend/internal/domain/model"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/external"
    "github.com/kirimku/smartseller-backend/internal/tests/mockserver"
)

func TestSiCepatClient(t *testing.T) {
    // Set environment variables for testing
    oldAPIURL := os.Getenv("SICEPAT_API_URL")
    oldAPIKey := os.Getenv("SICEPAT_API_KEY")
    
    // Start the mock server
    mockServerURL := mock_server.StartMockServer(8082)
    defer mock_server.StopMockServer()
    
    // Override environment variables to point to our mock server
    os.Setenv("SICEPAT_API_URL", mockServerURL)
    os.Setenv("SICEPAT_API_KEY", "test_api_key")
    
    // Restore environment variables after the test
    defer func() {
        os.Setenv("SICEPAT_API_URL", oldAPIURL)
        os.Setenv("SICEPAT_API_KEY", oldAPIKey)
    }()
    
    // Setup mock responses
    setupMockResponses(t)
    
    // Load config
    err := config.LoadConfig()
    assert.NoError(t, err)
    
    // Initialize location data for tests
    setupLocationData(t, &config.AppConfig)
    
    // Create SiCepat client
    scClient := external.NewSiCepatClient(&config.AppConfig)
    
    // Test addresses
    fromAddr := &model.Address{
        City:     "Jakarta Selatan", // Changed to match mapping in the client
        District: "Kebayoran Baru",  // Changed to match mapping in the client
        Province: "DKI Jakarta",
        Country:  "Indonesia",
    }
    
    toAddr := &model.Address{
        City:     "Bandung",
        District: "Bandung Kota",
        Province: "Jawa Barat",
        Country:  "Indonesia",
    }
    
    // Test different services
    t.Run("REG Service", func(t *testing.T) {
        // Calculate shipping fee
        rate, err := scClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, external.SCServiceREG)
        
        // Assert that there were no errors
        require.NoError(t, err)
        require.NotNil(t, rate)
        
        // Check response details
        assert.Equal(t, "sicepat", rate.CourierID)
        assert.Equal(t, "SiCepat", rate.CourierName)
        assert.Equal(t, external.SCServiceREG, rate.ServiceID)
        assert.Equal(t, "SiCepat REG", rate.ServiceName)
        assert.Equal(t, "regular", rate.ServiceType)
        assert.Greater(t, rate.Price, 0)
    })
    
    t.Run("BEST Service", func(t *testing.T) {
        // Calculate shipping fee
        rate, err := scClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, external.SCServiceBEST)
        
        // Assert that there were no errors
        require.NoError(t, err)
        require.NotNil(t, rate)
        
        // Check response details
        assert.Equal(t, "sicepat", rate.CourierID)
        assert.Equal(t, "SiCepat", rate.CourierName)
        assert.Equal(t, external.SCServiceBEST, rate.ServiceID)
        assert.Equal(t, "SiCepat BEST", rate.ServiceName)
        assert.Equal(t, "express", rate.ServiceType)
        assert.Greater(t, rate.Price, 0)
    })
    
    t.Run("GOKIL Service", func(t *testing.T) {
        // Calculate shipping fee
        rate, err := scClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, external.SCServiceGOKIL)
        
        // Assert that there were no errors
        require.NoError(t, err)
        require.NotNil(t, rate)
        
        // Check response details
        assert.Equal(t, "sicepat", rate.CourierID)
        assert.Equal(t, "SiCepat", rate.CourierName)
        assert.Equal(t, external.SCServiceGOKIL, rate.ServiceID)
        assert.Equal(t, "SiCepat GOKIL", rate.ServiceName)
        assert.Equal(t, "same_day", rate.ServiceType)
        assert.Greater(t, rate.Price, 0)
    })
    
    t.Run("HALU Service", func(t *testing.T) {
        // Calculate shipping fee
        rate, err := scClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, external.SCServiceHALU)
        
        // Assert that there were no errors
        require.NoError(t, err)
        require.NotNil(t, rate)
        
        // Check response details
        assert.Equal(t, "sicepat", rate.CourierID)
        assert.Equal(t, "SiCepat", rate.CourierName)
        assert.Equal(t, external.SCServiceHALU, rate.ServiceID)
        assert.Equal(t, "SiCepat HALU", rate.ServiceName)
        assert.Equal(t, "regular", rate.ServiceType)
        assert.Greater(t, rate.Price, 0)
    })
    
    t.Run("Invalid Service", func(t *testing.T) {
        // Calculate shipping fee with invalid service
        _, err := scClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, "invalid_service")
        
        // Assert that there was an error
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid SiCepat service")
    })
    
    t.Run("Unknown Area", func(t *testing.T) {
        // Create test address with an unknown area
        unknownAddr := &model.Address{
            City:     "Unknown City",
            District: "Unknown District",
            Province: "Unknown Province",
            Country:  "Indonesia",
        }
        
        // Calculate shipping fee - should fail
        _, err := scClient.CalculateShippingFee(context.Background(), fromAddr, unknownAddr, 1000, external.SCServiceREG)
        
        // Assert that there was an error
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid destination address")
    })
    
    t.Run("International Shipment", func(t *testing.T) {
        // Create test address with an international destination
        internationalAddr := &model.Address{
            City:     "Singapore",
            District: "Central",
            Province: "Singapore",
            Country:  "Singapore",
        }
        
        // Calculate shipping fee - should fail for international
        _, err := scClient.CalculateShippingFee(context.Background(), fromAddr, internationalAddr, 1000, external.SCServiceREG)
        
        // Assert that there was an error
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "domestic shipping")
    })
    
    t.Run("Missing Required Fields", func(t *testing.T) {
        // Test with missing sender address
        _, err := scClient.CalculateShippingFee(context.Background(), nil, toAddr, 1000, external.SCServiceREG)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid sender address")
        
        // Test with missing recipient address
        _, err = scClient.CalculateShippingFee(context.Background(), fromAddr, nil, 1000, external.SCServiceREG)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid recipient address")
        
        // Test with missing district
        incompleteAddr := &model.Address{
            City:     "Bandung",
            Province: "Jawa Barat",
            Country:  "Indonesia",
        }
        _, err = scClient.CalculateShippingFee(context.Background(), fromAddr, incompleteAddr, 1000, external.SCServiceREG)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "requires destination district")
    })
    
    // Check that at least one request was recorded
    requests := mock_server.GetRecordedRequests()
    var foundSiCepatRequest bool
    for _, req := range requests {
        if strings.Contains(req.Path, "/customer/tariff") {
            foundSiCepatRequest = true
            break
        }
    }
    assert.True(t, foundSiCepatRequest, "Should have made at least one SiCepat tariff request")
}

// setupMockResponses registers the necessary mock responses for the tests
func setupMockResponses(t *testing.T) {
    // Register a handler for SiCepat tariff endpoint
    mock_server.RegisterHandler("/customer/tariff", func(w http.ResponseWriter, r *http.Request) {
        // Get query parameters
        q := r.URL.Query()
        apiKey := q.Get("api-key")
        origin := q.Get("origin")
        destination := q.Get("destination")
        
        // Validate API key
        if apiKey != "test_api_key" {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "sicepat": map[string]interface{}{
                    "status": map[string]interface{}{
                        "code":        "401",
                        "description": "Invalid API key",
                    },
                },
            })
            return
        }
        
        // Check for unknown areas
        if origin == "" || destination == "" || strings.Contains(destination, "Unknown") {
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(map[string]interface{}{
                "sicepat": map[string]interface{}{
                    "status": map[string]interface{}{
                        "code":        "400",
                        "description": "Can't get tariff from database",
                    },
                },
            })
            return
        }
        
        // Return successful response with tariffs for different services
        w.WriteHeader(http.StatusOK)
        w.Header().Set("Content-Type", "application/json")
        
        // Generate mock tariffs
        tariffs := []map[string]interface{}{
            {
                "service": "REG",
                "tariff":  15000,
                "etd":     "2-3",
            },
            {
                "service": "BEST",
                "tariff":  20000,
                "etd":     "1-2",
            },
            {
                "service": "GOKIL",
                "tariff":  30000,
                "etd":     "0-1",
            },
            {
                "service": "HALU",
                "tariff":  12000,
                "etd":     "3-4",
            },
        }
        
        json.NewEncoder(w).Encode(map[string]interface{}{
            "sicepat": map[string]interface{}{
                "status": map[string]interface{}{
                    "code":        "200",
                    "description": "Success",
                },
                "results": tariffs,
            },
        })
    })
}

// setupLocationData initializes the location data for testing
func setupLocationData(t *testing.T, config *config.Config) {
    // Initialize PROVINCES map if it doesn't exist
    if config.LocationData.PROVINCES == nil {
        config.LocationData.PROVINCES = make(map[string]string)
    }
    
    // Add city to province mappings
    config.LocationData.PROVINCES["Jakarta Selatan"] = "DKI Jakarta"
    config.LocationData.PROVINCES["Jakarta Pusat"] = "DKI Jakarta"
    config.LocationData.PROVINCES["Bandung"] = "Jawa Barat"
    config.LocationData.PROVINCES["Bogor"] = "Jawa Barat"
    
    // Initialize DISTRICTS map if it doesn't exist
    if config.LocationData.DISTRICTS == nil {
        config.LocationData.DISTRICTS = make(map[string]map[string][]string)
    }
    
    // Add district mappings for DKI Jakarta
    jakartaDistricts := make(map[string][]string)
    jakartaDistricts["Jakarta Selatan"] = []string{"Kebayoran Baru", "Pancoran", "Tebet"}
    jakartaDistricts["Jakarta Pusat"] = []string{"Menteng", "Tanah Abang"}
    config.LocationData.DISTRICTS["DKI Jakarta"] = jakartaDistricts
    
    // Add district mappings for Jawa Barat
    jabarDistricts := make(map[string][]string)
    jabarDistricts["Bandung"] = []string{"Bandung Kota", "Cicendo"}
    jabarDistricts["Bogor"] = []string{"Bogor Kota", "Bogor Utara"}
    config.LocationData.DISTRICTS["Jawa Barat"] = jabarDistricts
}