package model

// JNTReceipt represents a receipt from JNT
type JNTReceipt struct {
	ID            string
	TransactionID string
	ReceiptNumber string
	ServiceType   int
}