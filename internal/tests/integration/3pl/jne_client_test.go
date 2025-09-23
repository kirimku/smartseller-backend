package integration

import (
    "context"
    "os"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/kirimku/smartseller-backend/internal/config"
    "github.com/kirimku/smartseller-backend/internal/domain/model"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/external"
    "github.com/kirimku/smartseller-backend/internal/tests/mockserver"
)

func TestJNEClient(t *testing.T) {
    // Set environment variables for testing
    os.Setenv("JNE_API_URL", "")  // Will be overridden
    os.Setenv("JNE_USERNAME", "test_user")
    os.Setenv("JNE_API_KEY", "test_key")
    
    // Start the mock server
    mockServerURL := mock_server.StartMockServer(8081)
    defer mock_server.StopMockServer()
    
    // Override the JNE API URL to point to our mock server
    oldURL := os.Getenv("JNE_API_URL")
    os.Setenv("JNE_API_URL", mockServerURL)
    defer os.Setenv("JNE_API_URL", oldURL)

    // Create JNE client
    jneClient := external.NewJNEClient(&config.AppConfig)

    // Test addresses
    fromAddr := &model.Address{
        City:     "Jakarta",
        District: "Jakarta Selatan",
        Province: "DKI Jakarta",
        Country:  "Indonesia",
    }

    toAddr := &model.Address{
        City:     "Bandung",
        District: "Bandung Kota",
        Province: "Jawa Barat",
        Country:  "Indonesia",
    }

    // Call the client with different service types
    testServices := []string{
        external.JNEServiceREG,
        external.JNEServiceYES,
        external.JNEServiceTrucking,
    }

    for _, service := range testServices {
        t.Run(service, func(t *testing.T) {
            // Calculate shipping fee
            rate, err := jneClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000, service)
            
            // Assert that there were no errors
            assert.NoError(t, err)
            assert.NotNil(t, rate)
            
            // Check response details
            assert.Equal(t, "jne", rate.CourierID)
            assert.Equal(t, "JNE", rate.CourierName)
            assert.Contains(t, rate.ServiceID, "jne_")
            assert.Equal(t, service, rate.ServiceName)
            assert.Greater(t, rate.Price, 0)
        })
    }

    // Check that the request was recorded
    requests := mock_server.GetRecordedJNERequests()
    assert.GreaterOrEqual(t, len(requests), 1)
    assert.Equal(t, "/tracing/api/pricedev", requests[0].Path)
}