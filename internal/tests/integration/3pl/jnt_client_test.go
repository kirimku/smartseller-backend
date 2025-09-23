package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/kirimku/smartseller-backend/internal/config"
    "github.com/kirimku/smartseller-backend/internal/domain/model"
    "github.com/kirimku/smartseller-backend/internal/infrastructure/external"
    "github.com/kirimku/smartseller-backend/internal/tests/mockserver"
)

func TestJNTClient(t *testing.T) {
    // Start the mock server
    mockServerURL := mock_server.StartMockServer(8081)
    defer mock_server.StopMockServer()

    // Load config
    err := config.LoadConfig()
    assert.NoError(t, err)

    // Override the JNT API URL to point to our mock server
    originalURL := config.AppConfig.JNTConfig.APIURL
    config.AppConfig.JNTConfig.APIURL = mockServerURL
    defer func() {
        // Restore the original URL after the test
        config.AppConfig.JNTConfig.APIURL = originalURL
    }()

    // Create JNT client
    jntClient := external.NewJNTClient(&config.AppConfig)

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

    // Call the client
    rate, err := jntClient.CalculateShippingFee(context.Background(), fromAddr, toAddr, 1000)
    assert.NoError(t, err)
    assert.NotNil(t, rate)

    // Verify the response
    assert.Equal(t, "jnt", rate.CourierID)
    assert.Equal(t, "J&T Express", rate.CourierName)
    assert.Greater(t, rate.Price, 0)

    // Check that the request was recorded
    requests := mock_server.GetRecordedJNTRequests()
    assert.GreaterOrEqual(t, len(requests), 1)
    assert.Equal(t, "/jandt_track/inquiry.action", requests[0].Path)
}