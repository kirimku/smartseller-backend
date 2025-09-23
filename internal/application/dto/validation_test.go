package dto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateWalletTransactionCreateRequest(t *testing.T) {
	t.Run("Valid Request", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:  "Test Order",
				OrderPrice: 100000,
				Weight:     1000,
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateWalletTransactionCreateRequest(req)
		assert.Nil(t, errors)
	})

	t.Run("Invalid Wallet ID", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:  "Test Order",
				OrderPrice: 100000,
				Weight:     1000,
			},
			WalletID: "invalid-uuid",
		}

		errors := ValidateWalletTransactionCreateRequest(req)
		assert.NotNil(t, errors)
		assert.Len(t, errors.Errors, 1)
		assert.Equal(t, "wallet_id", errors.Errors[0].Field)
	})

	t.Run("Invalid Order Price", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:  "Test Order",
				OrderPrice: -100,
				Weight:     1000,
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateWalletTransactionCreateRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "Order price must be greater than 0")
	})

	t.Run("Order Price Exceeds Maximum", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:  "Test Order",
				OrderPrice: 200000000, // 200M IDR
				Weight:     1000,
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateWalletTransactionCreateRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "exceeds maximum allowed amount")
	})
}

func TestValidateWalletRefundRequest(t *testing.T) {
	t.Run("Valid Request", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        "Customer requested refund",
		}

		errors := ValidateWalletRefundRequest(req)
		assert.Nil(t, errors)
	})

	t.Run("Invalid Transaction ID", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 0,
			RefundAmount:  50000,
			Reason:        "Customer requested refund",
		}

		errors := ValidateWalletRefundRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "Transaction ID must be greater than 0")
	})

	t.Run("Invalid Refund Amount", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  0,
			Reason:        "Customer requested refund",
		}

		errors := ValidateWalletRefundRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "Refund amount must be greater than 0")
	})

	t.Run("Reason Too Short", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        "Bad",
		}

		errors := ValidateWalletRefundRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "at least 5 characters")
	})

	t.Run("Reason Too Long", func(t *testing.T) {
		longReason := ""
		for i := 0; i < 600; i++ {
			longReason += "a"
		}

		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        longReason,
		}

		errors := ValidateWalletRefundRequest(req)
		assert.NotNil(t, errors)
		assert.Contains(t, errors.Error(), "must not exceed 500 characters")
	})
}

func TestSanitizeString(t *testing.T) {
	t.Run("Remove Control Characters", func(t *testing.T) {
		input := "Normal text\x00with\x01control\x02characters"
		expected := "Normal textwithcontrolcharacters"
		result := SanitizeString(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Preserve Newlines and Tabs", func(t *testing.T) {
		input := "Text with\nnewline and\ttab"
		expected := "Text with\nnewline and\ttab"
		result := SanitizeString(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Trim Whitespace", func(t *testing.T) {
		input := "  text with spaces  "
		expected := "text with spaces"
		result := SanitizeString(input)
		assert.Equal(t, expected, result)
	})
}

func TestSanitizeFilename(t *testing.T) {
	t.Run("Remove Dangerous Characters", func(t *testing.T) {
		input := "file<name>with:danger\"ous/characters"
		expected := "file_name_with_danger_ous_characters"
		result := SanitizeFilename(input)
		assert.Equal(t, expected, result)
	})

	t.Run("Trim Dots and Spaces", func(t *testing.T) {
		input := " .filename. "
		expected := "filename"
		result := SanitizeFilename(input)
		assert.Equal(t, expected, result)
	})
}

func TestIsValidUUID(t *testing.T) {
	t.Run("Valid UUID", func(t *testing.T) {
		validUUID := "550e8400-e29b-41d4-a716-446655440000"
		assert.True(t, isValidUUID(validUUID))
	})

	t.Run("Invalid UUID", func(t *testing.T) {
		invalidUUID := "invalid-uuid-format"
		assert.False(t, isValidUUID(invalidUUID))
	})

	t.Run("Empty String", func(t *testing.T) {
		assert.False(t, isValidUUID(""))
	})
}

func TestValidateRefundAmount(t *testing.T) {
	t.Run("Valid Refund Amount", func(t *testing.T) {
		err := ValidateRefundAmount(50000, 100000, 0)
		assert.Nil(t, err)
	})

	t.Run("Refund Amount Too High", func(t *testing.T) {
		err := ValidateRefundAmount(150000, 100000, 0)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "cannot exceed original transaction amount")
	})

	t.Run("Total Refund Exceeds Original", func(t *testing.T) {
		err := ValidateRefundAmount(60000, 100000, 50000)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "total refund amount cannot exceed")
	})

	t.Run("Zero Refund Amount", func(t *testing.T) {
		err := ValidateRefundAmount(0, 100000, 0)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})
}

func TestValidateWalletBalance(t *testing.T) {
	t.Run("Sufficient Balance", func(t *testing.T) {
		err := ValidateWalletBalance(50000, 100000)
		assert.Nil(t, err)
	})

	t.Run("Insufficient Balance", func(t *testing.T) {
		err := ValidateWalletBalance(150000, 100000)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "insufficient wallet balance")
	})

	t.Run("Zero Requested Amount", func(t *testing.T) {
		err := ValidateWalletBalance(0, 100000)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "must be greater than 0")
	})
}

func TestValidateCODFields(t *testing.T) {
	t.Run("Valid COD Enabled", func(t *testing.T) {
		err := ValidateCODFields(true, 100000, "buyer")
		assert.Nil(t, err)
	})

	t.Run("Valid COD Disabled", func(t *testing.T) {
		err := ValidateCODFields(false, 0, "")
		assert.Nil(t, err)
	})

	t.Run("COD Enabled But No Value", func(t *testing.T) {
		err := ValidateCODFields(true, 0, "buyer")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "COD value must be greater than 0")
	})

	t.Run("COD Disabled But Has Value", func(t *testing.T) {
		err := ValidateCODFields(false, 100000, "")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "COD value must be 0 when COD is disabled")
	})

	t.Run("Invalid COD Admin Fee Paid By", func(t *testing.T) {
		err := ValidateCODFields(true, 100000, "invalid")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "must be either 'seller' or 'buyer'")
	})
}

func TestValidateTransactionConsistency(t *testing.T) {
	t.Run("Valid COD Configuration", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:          "Test Order",
				OrderPrice:         100000,
				Weight:             1000,
				COD:                true,
				CODValue:           100000,
				CODAdminFeePaidBy:  "buyer",
				Courier:            "jne",
				CourierServiceType: "reg",
				From: TransactionAddressDTO{
					Name:     "Sender",
					Phone:    "081234567890",
					Province: "DKI Jakarta",
					City:     "Jakarta",
					Area:     "Kemang",
					Address:  "Jl. Test No. 123",
					PostCode: "12345",
				},
				To: TransactionAddressDTO{
					Name:     "Receiver",
					Phone:    "081234567891",
					Province: "Jawa Barat",
					City:     "Bandung",
					Area:     "Dago",
					Address:  "Jl. Dago No. 456",
					PostCode: "67890",
				},
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateTransactionConsistency(req)
		assert.Nil(t, errors)
	})

	t.Run("Invalid COD Configuration", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:          "Test Order",
				OrderPrice:         100000,
				Weight:             1000,
				COD:                true,
				CODValue:           0, // Invalid: COD enabled but value is 0
				CODAdminFeePaidBy:  "buyer",
				Courier:            "jne",
				CourierServiceType: "reg", // Add this to avoid courier service validation error
				From: TransactionAddressDTO{
					Name:     "Sender",
					Phone:    "081234567890",
					Province: "DKI Jakarta",
					City:     "Jakarta",
					Area:     "Kemang",
					Address:  "Jl. Test No. 123",
					PostCode: "12345",
				},
				To: TransactionAddressDTO{
					Name:     "Receiver",
					Phone:    "081234567891",
					Province: "Jawa Barat",
					City:     "Bandung",
					Area:     "Dago",
					Address:  "Jl. Dago No. 456",
					PostCode: "67890",
				},
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateTransactionConsistency(req)
		assert.NotNil(t, errors)
		assert.Len(t, errors.Errors, 1)
		assert.Equal(t, "cod_fields", errors.Errors[0].Field)
	})

	t.Run("Invalid Postal Code", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:          "Test Order",
				OrderPrice:         100000,
				Weight:             1000,
				Courier:            "jne",
				CourierServiceType: "reg", // Add this to avoid courier service validation error
				From: TransactionAddressDTO{
					Name:     "Sender",
					Phone:    "081234567890",
					Province: "DKI Jakarta",
					City:     "Jakarta",
					Area:     "Kemang",
					Address:  "Jl. Test No. 123",
					PostCode: "123", // Invalid: too short
				},
				To: TransactionAddressDTO{
					Name:     "Receiver",
					Phone:    "081234567891",
					Province: "Jawa Barat",
					City:     "Bandung",
					Area:     "Dago",
					Address:  "Jl. Dago No. 456",
					PostCode: "67890",
				},
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateTransactionConsistency(req)
		assert.NotNil(t, errors)
		assert.Len(t, errors.Errors, 1)
		assert.Equal(t, "from.post_code", errors.Errors[0].Field)
	})
}

func TestValidateRefundBusinessRules(t *testing.T) {
	t.Run("Valid Refund Request", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        "customer_request - Customer changed mind",
			AdminNote:     "Approved by admin",
		}

		errors := ValidateRefundBusinessRules(req, 100000, 0, "completed")
		assert.Nil(t, errors)
	})

	t.Run("Invalid Payment Status", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        "customer_request - Customer changed mind",
		}

		errors := ValidateRefundBusinessRules(req, 100000, 0, "pending")
		assert.NotNil(t, errors)
		assert.Len(t, errors.Errors, 1)
		assert.Equal(t, "transaction_status", errors.Errors[0].Field)
	})

	t.Run("Refund Amount Exceeds Original", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  150000, // Exceeds original amount
			Reason:        "customer_request - Customer changed mind",
		}

		errors := ValidateRefundBusinessRules(req, 100000, 0, "completed")
		assert.NotNil(t, errors)
		assert.True(t, len(errors.Errors) > 0)
		// Should have errors for both exceeding original amount and business rule validation
	})

	t.Run("Partial Refund Exceeds Remaining", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  40000, // Would exceed remaining amount (30000)
			Reason:        "customer_request - Customer changed mind",
		}

		errors := ValidateRefundBusinessRules(req, 100000, 70000, "completed")
		assert.NotNil(t, errors)
		assert.True(t, len(errors.Errors) > 0)
		// Check for remaining amount error
		found := false
		for _, err := range errors.Errors {
			if err.Field == "refund_amount" && strings.Contains(err.Message, "remaining refundable amount") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("Below Minimum Refund Amount", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  500, // Below 1000 IDR minimum
			Reason:        "customer_request - Customer changed mind",
		}

		errors := ValidateRefundBusinessRules(req, 100000, 0, "completed")
		assert.NotNil(t, errors)
		found := false
		for _, err := range errors.Errors {
			if err.Field == "refund_amount" && strings.Contains(err.Message, "Minimum refund amount") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("Invalid Refund Reason Category", func(t *testing.T) {
		req := &WalletRefundRequest{
			TransactionID: 123,
			RefundAmount:  50000,
			Reason:        "just because", // Invalid reason
		}

		errors := ValidateRefundBusinessRules(req, 100000, 0, "completed")
		assert.NotNil(t, errors)
		found := false
		for _, err := range errors.Errors {
			if err.Field == "reason" && strings.Contains(err.Message, "valid categories") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestSanitizeAndValidateAmount(t *testing.T) {
	t.Run("Valid Amount", func(t *testing.T) {
		err := SanitizeAndValidateAmount(50000, "test_amount", 1000, 100000)
		assert.Nil(t, err)
	})

	t.Run("Amount Below Minimum", func(t *testing.T) {
		err := SanitizeAndValidateAmount(500, "test_amount", 1000, 100000)
		assert.NotNil(t, err)
		assert.Equal(t, "test_amount", err.Field)
		assert.Contains(t, err.Message, "at least")
	})

	t.Run("Amount Above Maximum", func(t *testing.T) {
		err := SanitizeAndValidateAmount(150000, "test_amount", 1000, 100000)
		assert.NotNil(t, err)
		assert.Equal(t, "test_amount", err.Field)
		assert.Contains(t, err.Message, "not exceed")
	})
}

func TestSanitizeAndValidateAddress(t *testing.T) {
	t.Run("Valid Address", func(t *testing.T) {
		address := &TransactionAddressDTO{
			Name:     "John Doe",
			Phone:    "081234567890",
			Email:    "john@example.com",
			Province: "DKI Jakarta",
			City:     "Jakarta",
			Area:     "Kemang",
			Address:  "Jl. Test No. 123",
			PostCode: "12345",
		}

		errors := SanitizeAndValidateAddress(address, "test_address")
		assert.Empty(t, errors)
	})

	t.Run("Invalid Postal Code", func(t *testing.T) {
		address := &TransactionAddressDTO{
			Name:     "John Doe",
			Phone:    "081234567890",
			Province: "DKI Jakarta",
			City:     "Jakarta",
			Area:     "Kemang",
			Address:  "Jl. Test No. 123",
			PostCode: "123", // Invalid: too short
		}

		errors := SanitizeAndValidateAddress(address, "test_address")
		assert.NotEmpty(t, errors)
		assert.Equal(t, "test_address.post_code", errors[0].Field)
	})

	t.Run("Invalid Phone Number", func(t *testing.T) {
		address := &TransactionAddressDTO{
			Name:     "John Doe",
			Phone:    "123", // Invalid phone number
			Province: "DKI Jakarta",
			City:     "Jakarta",
			Area:     "Kemang",
			Address:  "Jl. Test No. 123",
			PostCode: "12345",
		}

		errors := SanitizeAndValidateAddress(address, "test_address")
		assert.NotEmpty(t, errors)
		assert.Equal(t, "test_address.phone", errors[0].Field)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		address := &TransactionAddressDTO{
			Name:     "John Doe",
			Phone:    "081234567890",
			Email:    "invalid-email", // Invalid email
			Province: "DKI Jakarta",
			City:     "Jakarta",
			Area:     "Kemang",
			Address:  "Jl. Test No. 123",
			PostCode: "12345",
		}

		errors := SanitizeAndValidateAddress(address, "test_address")
		assert.NotEmpty(t, errors)
		assert.Equal(t, "test_address.email", errors[0].Field)
	})
}

func TestValidateAgainstXSS(t *testing.T) {
	t.Run("Clean Input", func(t *testing.T) {
		err := ValidateAgainstXSS("This is a clean input", "test_field")
		assert.Nil(t, err)
	})

	t.Run("XSS Script Tag", func(t *testing.T) {
		err := ValidateAgainstXSS("<script>alert('xss')</script>", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
		assert.Contains(t, err.Message, "dangerous content")
	})

	t.Run("XSS JavaScript", func(t *testing.T) {
		err := ValidateAgainstXSS("javascript:alert('xss')", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
	})

	t.Run("XSS Event Handler", func(t *testing.T) {
		err := ValidateAgainstXSS("onload=alert('xss')", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
	})
}

func TestValidateAgainstSQLInjection(t *testing.T) {
	t.Run("Clean Input", func(t *testing.T) {
		err := ValidateAgainstSQLInjection("This is a clean input", "test_field")
		assert.Nil(t, err)
	})

	t.Run("SQL Union Select", func(t *testing.T) {
		err := ValidateAgainstSQLInjection("' UNION SELECT * FROM users --", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
		assert.Contains(t, err.Message, "SQL patterns")
	})

	t.Run("SQL Drop Table", func(t *testing.T) {
		err := ValidateAgainstSQLInjection("'; DROP TABLE users; --", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
	})

	t.Run("SQL Comment", func(t *testing.T) {
		err := ValidateAgainstSQLInjection("test -- comment", "test_field")
		assert.NotNil(t, err)
		assert.Equal(t, "test_field", err.Field)
	})
}

func TestComprehensiveInputValidation(t *testing.T) {
	t.Run("Clean Input", func(t *testing.T) {
		errors := ComprehensiveInputValidation("This is a clean input", "test_field")
		assert.Empty(t, errors)
	})

	t.Run("Multiple Security Issues", func(t *testing.T) {
		errors := ComprehensiveInputValidation("<script>alert('xss')</script>' UNION SELECT * FROM users --", "test_field")
		assert.NotEmpty(t, errors)
		assert.True(t, len(errors) >= 2) // Should catch both XSS and SQL injection
	})

	t.Run("Null Bytes", func(t *testing.T) {
		errors := ComprehensiveInputValidation("test\x00input", "test_field")
		assert.NotEmpty(t, errors)
		assert.Equal(t, "test_field", errors[0].Field)
		assert.Contains(t, errors[0].Message, "null bytes")
	})
}

func TestValidateWalletTransactionCreateRequestAdvanced(t *testing.T) {
	t.Run("Valid Advanced Request", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:          "Clean Order Name",
				OrderPrice:         50000,
				Weight:             1000,
				COD:                true,
				CODValue:           50000,
				CODAdminFeePaidBy:  "buyer",
				Courier:            "jne",
				CourierServiceType: "reg",
				From: TransactionAddressDTO{
					Name:     "Sender Name",
					Phone:    "081234567890",
					Email:    "sender@example.com",
					Province: "DKI Jakarta",
					City:     "Jakarta",
					Area:     "Kemang",
					Address:  "Jl. Kemang Raya No. 123",
					PostCode: "12345",
				},
				To: TransactionAddressDTO{
					Name:     "Receiver Name",
					Phone:    "081234567891",
					Email:    "receiver@example.com",
					Province: "Jawa Barat",
					City:     "Bandung",
					Area:     "Dago",
					Address:  "Jl. Dago No. 456",
					PostCode: "67890",
				},
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateWalletTransactionCreateRequestAdvanced(req)
		assert.Nil(t, errors)
	})

	t.Run("XSS in Order Name", func(t *testing.T) {
		req := &WalletTransactionCreateRequest{
			TransactionCreateRequest: TransactionCreateRequest{
				OrderName:  "<script>alert('xss')</script>",
				OrderPrice: 50000,
				Weight:     1000,
				Courier:    "jne",
				From: TransactionAddressDTO{
					Name:     "Sender",
					Phone:    "081234567890",
					Province: "DKI Jakarta",
					City:     "Jakarta",
					Area:     "Kemang",
					Address:  "Jl. Test No. 123",
					PostCode: "12345",
				},
				To: TransactionAddressDTO{
					Name:     "Receiver",
					Phone:    "081234567891",
					Province: "Jawa Barat",
					City:     "Bandung",
					Area:     "Dago",
					Address:  "Jl. Dago No. 456",
					PostCode: "67890",
				},
			},
			WalletID: "550e8400-e29b-41d4-a716-446655440000",
		}

		errors := ValidateWalletTransactionCreateRequestAdvanced(req)
		assert.NotNil(t, errors)
		// Should have XSS error
		found := false
		for _, err := range errors.Errors {
			if err.Field == "order_name" && strings.Contains(err.Message, "dangerous content") {
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}
