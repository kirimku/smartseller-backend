package model

// JNEReceipt represents a receipt from JNE
type JNEReceipt struct {
	ID            string
	TransactionID string
	ReceiptNumber string
	ServiceType   string // JNE service type (REG, YES, JTR)
}
