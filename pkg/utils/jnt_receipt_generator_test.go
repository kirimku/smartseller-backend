package utils

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kirimku/smartseller-backend/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTransactionRepository is a mock for testing purposes
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByID(ctx context.Context, id int) (*entity.Transaction, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByUniqueID(ctx context.Context, uniqueID string) (*entity.Transaction, error) {
	args := m.Called(ctx, uniqueID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByTrackingNumber(ctx context.Context, trackingNumber string) (*entity.Transaction, error) {
	args := m.Called(ctx, trackingNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) FindByBookingCode(ctx context.Context, bookingCode string) (*entity.Transaction, error) {
	args := m.Called(ctx, bookingCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) CreateWithWalletPayment(ctx context.Context, tx *sqlx.Tx, transaction *entity.Transaction) error {
	args := m.Called(ctx, tx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindByUserID(ctx context.Context, userID string, page, limit int, states []string, couriers []string, cashbackStates []string, bookingCodes []string) ([]*entity.Transaction, int, error) {
	args := m.Called(ctx, userID, page, limit, states, couriers, cashbackStates, bookingCodes)
	return args.Get(0).([]*entity.Transaction), args.Int(1), args.Error(2)
}

func (m *MockTransactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	args := m.Called(ctx, transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) UpdateState(ctx context.Context, id int, state string) error {
	args := m.Called(ctx, id, state)
	return args.Error(0)
}

func (m *MockTransactionRepository) FindPaidTransactionsNeedingProcessing(ctx context.Context, limit int) ([]*entity.Transaction, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*entity.Transaction), args.Error(1)
}

func (m *MockTransactionRepository) UpdatePostPaymentProcessing(ctx context.Context, transactionID int, processed bool, error string, attempts int) error {
	args := m.Called(ctx, transactionID, processed, error, attempts)
	return args.Error(0)
}

func TestGenerateJNTReceiptNumber(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockTransactionRepository{}

	tests := []struct {
		name          string
		transactionID int
		wantPrefix    string
		wantLength    int
	}{
		{
			name:          "Generate receipt for transaction ID 1",
			transactionID: 1,
			wantPrefix:    "JB",
			wantLength:    12,
		},
		{
			name:          "Generate receipt for transaction ID 123",
			transactionID: 123,
			wantPrefix:    "JB",
			wantLength:    12,
		},
		{
			name:          "Generate receipt for transaction ID 999999",
			transactionID: 999999,
			wantPrefix:    "JB",
			wantLength:    12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate receipt number
			receiptNumber, err := GenerateJNTReceiptNumber(ctx, mockRepo, tt.transactionID)

			// Assertions
			assert.NoError(t, err)
			assert.NotEmpty(t, receiptNumber)
			assert.Equal(t, tt.wantLength, len(receiptNumber))
			assert.Equal(t, tt.wantPrefix, receiptNumber[:2])

			// Verify it's a valid receipt number
			assert.True(t, ValidateJNTReceiptNumber(receiptNumber))

			// Verify the numeric part contains only digits
			for i := 2; i < len(receiptNumber); i++ {
				assert.True(t, receiptNumber[i] >= '0' && receiptNumber[i] <= '9',
					"Position %d should be a digit, got: %c", i, receiptNumber[i])
			}
		})
	}
}

func TestGenerateJNTReceiptNumber_Uniqueness(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockTransactionRepository{}

	// Generate multiple receipt numbers for the same transaction ID
	// They should be unique due to timestamp differences
	receiptNumbers := make(map[string]bool)
	transactionID := 1

	for i := 0; i < 10; i++ {
		receiptNumber, err := GenerateJNTReceiptNumber(ctx, mockRepo, transactionID)
		assert.NoError(t, err)

		// Check that this receipt number hasn't been generated before
		assert.False(t, receiptNumbers[receiptNumber],
			"Receipt number %s was generated twice", receiptNumber)

		receiptNumbers[receiptNumber] = true
	}

	// Verify we generated 10 unique receipt numbers
	assert.Equal(t, 10, len(receiptNumbers))
}

func TestValidateJNTReceiptNumber(t *testing.T) {
	tests := []struct {
		name          string
		receiptNumber string
		want          bool
	}{
		{
			name:          "Valid receipt number",
			receiptNumber: "JB0000000001",
			want:          true,
		},
		{
			name:          "Valid receipt number with larger number",
			receiptNumber: "JB1234567890",
			want:          true,
		},
		{
			name:          "Invalid - wrong prefix",
			receiptNumber: "AB0000000001",
			want:          false,
		},
		{
			name:          "Invalid - too short",
			receiptNumber: "JB00000001",
			want:          false,
		},
		{
			name:          "Invalid - too long",
			receiptNumber: "JB00000000001",
			want:          false,
		},
		{
			name:          "Invalid - contains letters in numeric part",
			receiptNumber: "JB000000000A",
			want:          false,
		},
		{
			name:          "Invalid - empty string",
			receiptNumber: "",
			want:          false,
		},
		{
			name:          "Invalid - only prefix",
			receiptNumber: "JB",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateJNTReceiptNumber(tt.receiptNumber)
			assert.Equal(t, tt.want, result)
		})
	}
}
