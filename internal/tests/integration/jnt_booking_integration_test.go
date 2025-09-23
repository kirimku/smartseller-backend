package integration

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/kirimku/smartseller-backend/internal/application/service"
	"github.com/kirimku/smartseller-backend/internal/config"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/kirimku/smartseller-backend/internal/domain/model"
	domainservice "github.com/kirimku/smartseller-backend/internal/domain/service"
	"github.com/kirimku/smartseller-backend/internal/infrastructure/external"
	"github.com/kirimku/smartseller-backend/pkg/logger"
)

// JNTBookingIntegrationTestSuite tests the complete JNT booking integration
type JNTBookingIntegrationTestSuite struct {
	suite.Suite
	logisticBookingService *service.LogisticBookingServiceImpl
	jntClient              *external.JNTClient
	transactionRepo        *persistence.MockTransactionRepository
}

// SetupSuite sets up the test suite
func (s *JNTBookingIntegrationTestSuite) SetupSuite() {
	// Initialize logger
	logger.InitLogger("test")

	// Set environment variables for testing with mockserver
	os.Setenv("JNT_API_URL", "http://localhost:3001")
	os.Setenv("JNT_API_KEY", "test_api_key")
	os.Setenv("JNT_ECOM_ID", "test_ecom_id")
	os.Setenv("JNT_WHITELABEL_ID", "test_whitelabel_id")
	os.Setenv("JNT_BUKASEND_ID", "test_bukasend_id")

	// Load config
	cfg := &config.Config{}

	// Create JNT client
	jntClient, err := external.NewJNTClient(cfg)
	require.NoError(s.T(), err)
	s.jntClient = jntClient

	// Create mock transaction repository
	s.transactionRepo = &persistence.MockTransactionRepository{}

	// Create LogisticBookingService
	packageCategoryService := domainservice.NewPackageCategoryService()
	s.logisticBookingService = service.NewLogisticBookingService(
		s.transactionRepo,
		nil, // User repository not needed for this test
		cfg,
		packageCategoryService,
		jntClient,
		nil, // JNE client not needed for this test
		nil, // SiCepat client not needed for this test
		nil, // SiCepat booking service not needed for this test
		nil, // SAPX client not needed for this test
		nil, // ResiGenerator not needed for this test
	).(*service.LogisticBookingServiceImpl)
}

// TestJNTBookingEndToEnd tests the complete JNT booking flow
func (s *JNTBookingIntegrationTestSuite) TestJNTBookingEndToEnd() {
	ctx := context.Background()

	// Create a test transaction
	transaction := &entity.Transaction{
		ID:              123,
		TotalAmount:     50000,
		Weight:          1000,
		InsuranceAmount: 0,
		Courier:         "jnt",
		COD:             false,
		CODValue:        0,
		FromAddress: entity.TransactionAddress{
			Name:     "Test Sender",
			Phone:    "081234567890",
			Address:  "Test Pickup Address",
			Area:     "Menteng",
			City:     "Jakarta",
			Province: "DKI Jakarta",
			PostCode: "10350",
		},
		ToAddress: entity.TransactionAddress{
			Name:     "Test Receiver",
			Phone:    "081234567891",
			Address:  "Test Receiver Address",
			Area:     "Cicendo",
			City:     "Bandung",
			Province: "Jawa Barat",
			PostCode: "40175",
		},
	}

	// Mock the repository Update method
	s.transactionRepo.On("Update", ctx, transaction).Return(nil)

	// Test booking creation
	bookingResult, err := s.logisticBookingService.CreateBooking(ctx, transaction)

	// Assertions
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), bookingResult)
	assert.Equal(s.T(), "jnt", bookingResult.CourierID)
	assert.NotEmpty(s.T(), bookingResult.BookingCode)
	assert.Contains(s.T(), bookingResult.BookingCode, "JNT")

	// Verify transaction was updated with booking code
	assert.Equal(s.T(), bookingResult.BookingCode, transaction.BookingCode)
	assert.Equal(s.T(), bookingResult.CourierID, transaction.Courier)

	// Verify repository was called
	s.transactionRepo.AssertExpectations(s.T())

	s.T().Logf("✅ JNT booking created successfully:")
	s.T().Logf("   - Booking Code: %s", bookingResult.BookingCode)
	s.T().Logf("   - Courier ID: %s", bookingResult.CourierID)
	s.T().Logf("   - Tracking URL: %s", bookingResult.TrackingURL)
}

// TestJNTBookingValidation tests validation logic
func (s *JNTBookingIntegrationTestSuite) TestJNTBookingValidation() {
	ctx := context.Background()

	// Test with missing sender address
	transaction := &entity.Transaction{
		ID:          124,
		TotalAmount: 50000,
		Weight:      1000,
		Courier:     "jnt",
		FromAddress: entity.TransactionAddress{
			// Missing Name and Phone
			Address:  "Test Address",
			City:     "Jakarta",
			Province: "DKI Jakarta",
		},
		ToAddress: entity.TransactionAddress{
			Name:     "Test Receiver",
			Phone:    "081234567891",
			Address:  "Test Address",
			City:     "Bandung",
			Province: "Jawa Barat",
		},
	}

	// Test booking creation - should fail validation
	bookingResult, err := s.logisticBookingService.CreateBooking(ctx, transaction)

	// Assertions
	assert.Error(s.T(), err)
	assert.Nil(s.T(), bookingResult)
	assert.Contains(s.T(), err.Error(), "sender address is required")

	s.T().Logf("✅ Validation working correctly: %s", err.Error())
}

// TestJNTBookingExistingBookingCode tests behavior when booking code already exists
func (s *JNTBookingIntegrationTestSuite) TestJNTBookingExistingBookingCode() {
	ctx := context.Background()

	// Create a test transaction with existing booking code
	transaction := &entity.Transaction{
		ID:          125,
		TotalAmount: 50000,
		Weight:      1000,
		Courier:     "jnt",
		BookingCode: "EXISTING_JNT123", // Already has booking code
		FromAddress: entity.TransactionAddress{
			Name:     "Test Sender",
			Phone:    "081234567890",
			Address:  "Test Address",
			City:     "Jakarta",
			Province: "DKI Jakarta",
		},
		ToAddress: entity.TransactionAddress{
			Name:     "Test Receiver",
			Phone:    "081234567891",
			Address:  "Test Address",
			City:     "Bandung",
			Province: "Jawa Barat",
		},
	}

	// Test booking creation - should return existing booking
	bookingResult, err := s.logisticBookingService.CreateBooking(ctx, transaction)

	// Assertions
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), bookingResult)
	assert.Equal(s.T(), "EXISTING_JNT123", bookingResult.BookingCode)
	assert.Equal(s.T(), "jnt", bookingResult.CourierID)

	s.T().Logf("✅ Existing booking code handled correctly: %s", bookingResult.BookingCode)
}

// TestJNTServiceTypeSelection tests service type determination logic
func (s *JNTBookingIntegrationTestSuite) TestJNTServiceTypeSelection() {
	ctx := context.Background()

	// Test high-value transaction (should use pickup)
	highValueTransaction := &entity.Transaction{
		ID:                 126,
		TotalAmount:        1500000, // > 1M IDR
		Weight:             1000,
		Courier:            "jnt",
		CourierServiceType: "", // Not specified, should default based on amount
		FromAddress: entity.TransactionAddress{
			Name:     "Test Sender",
			Phone:    "081234567890",
			Address:  "Test Address",
			City:     "Jakarta",
			Province: "DKI Jakarta",
		},
		ToAddress: entity.TransactionAddress{
			Name:     "Test Receiver",
			Phone:    "081234567891",
			Address:  "Test Address",
			City:     "Bandung",
			Province: "Jawa Barat",
		},
	}

	// Mock the repository Update method
	s.transactionRepo.On("Update", ctx, highValueTransaction).Return(nil)

	// Test booking creation
	bookingResult, err := s.logisticBookingService.CreateBooking(ctx, highValueTransaction)

	// Assertions
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), bookingResult)
	assert.NotEmpty(s.T(), bookingResult.BookingCode)

	s.T().Logf("✅ High-value transaction booking successful: %s", bookingResult.BookingCode)

	// Test explicit service type
	explicitServiceTransaction := &entity.Transaction{
		ID:                 127,
		TotalAmount:        50000,
		Weight:             1000,
		Courier:            "jnt",
		CourierServiceType: "pickup", // Explicitly set to pickup
		FromAddress: entity.TransactionAddress{
			Name:     "Test Sender",
			Phone:    "081234567890",
			Address:  "Test Address",
			City:     "Jakarta",
			Province: "DKI Jakarta",
		},
		ToAddress: entity.TransactionAddress{
			Name:     "Test Receiver",
			Phone:    "081234567891",
			Address:  "Test Address",
			City:     "Bandung",
			Province: "Jawa Barat",
		},
	}

	// Mock the repository Update method
	s.transactionRepo.On("Update", ctx, explicitServiceTransaction).Return(nil)

	// Test booking creation
	bookingResult2, err := s.logisticBookingService.CreateBooking(ctx, explicitServiceTransaction)

	// Assertions
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), bookingResult2)
	assert.NotEmpty(s.T(), bookingResult2.BookingCode)

	s.T().Logf("✅ Explicit pickup service booking successful: %s", bookingResult2.BookingCode)

	// Verify both repository calls
	s.transactionRepo.AssertExpectations(s.T())
}

// TestJNTClientDirectly tests the JNT client directly
func (s *JNTBookingIntegrationTestSuite) TestJNTClientDirectly() {
	ctx := context.Background()

	// Test shipping fee calculation
	fromAddr := &entity.Address{
		City:     "Jakarta",
		Province: "DKI Jakarta",
	}

	toAddr := &entity.Address{
		City:     "Bandung",
		Province: "Jawa Barat",
	}

	// Convert to model.Address for JNT client
	fromModel := &model.Address{
		City:     fromAddr.City,
		Province: fromAddr.Province,
		Country:  "Indonesia",
	}

	toModel := &model.Address{
		City:     toAddr.City,
		Province: toAddr.Province,
		Country:  "Indonesia",
	}

	rate, err := s.jntClient.CalculateShippingFee(ctx, fromModel, toModel, 1000, "REG")

	// Assertions
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), rate)
	assert.Equal(s.T(), "jnt", rate.CourierID)
	assert.Greater(s.T(), rate.Price, float64(0))

	s.T().Logf("✅ JNT shipping fee calculation successful:")
	s.T().Logf("   - Courier: %s", rate.CourierName)
	s.T().Logf("   - Service: %s", rate.ServiceID)
	s.T().Logf("   - Price: %.0f", rate.Price)
}

// TestSuite runs all the tests
func TestJNTBookingIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(JNTBookingIntegrationTestSuite))
}
