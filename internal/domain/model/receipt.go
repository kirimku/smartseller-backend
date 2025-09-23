package model

import (
	"time"
)

// Receipt represents a shipping receipt/label
type Receipt struct {
	ReceiptNumber  string `json:"receipt_number"`
	TrackingNumber string `json:"tracking_number"`
	CourierID      string `json:"courier_id"`
	ServiceType    string `json:"service_type"`

	// Sender information
	SenderName    string `json:"sender_name"`
	SenderPhone   string `json:"sender_phone"`
	SenderAddress string `json:"sender_address"`

	// Recipient information
	RecipientName    string `json:"recipient_name"`
	RecipientPhone   string `json:"recipient_phone"`
	RecipientAddress string `json:"recipient_address"`

	// Package information
	Weight      float64 `json:"weight"`
	Dimensions  string  `json:"dimensions"`
	PackageType string  `json:"package_type"`

	// Cost information
	ShippingCost  float64 `json:"shipping_cost"`
	InsuranceCost float64 `json:"insurance_cost"`
	TotalCost     float64 `json:"total_cost"`

	// Receipt details
	GeneratedAt time.Time  `json:"generated_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	BarcodeURL  string     `json:"barcode_url"`
	LabelURL    string     `json:"label_url"`

	// Items in the package
	Items []Item `json:"items"`

	// Additional metadata
	Notes               string `json:"notes"`
	SpecialInstructions string `json:"special_instructions"`
}

// Item represents an item in a shipping package
type Item struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"weight"`
	Value       float64 `json:"value"`
	SKU         string  `json:"sku"`

	// Dimensions
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	// Additional metadata
	IsFragile   bool   `json:"is_fragile"`
	IsLiquid    bool   `json:"is_liquid"`
	IsHazardous bool   `json:"is_hazardous"`
	Notes       string `json:"notes"`
}

// ReceiptRequest represents a request to generate a receipt
type ReceiptRequest struct {
	TransactionID  int                    `json:"transaction_id"`
	BookingData    map[string]interface{} `json:"booking_data"` // Generic booking response data
	IncludeBarcode bool                   `json:"include_barcode"`
	IncludeLabel   bool                   `json:"include_label"`
	Format         string                 `json:"format"` // pdf, png, jpg
}

// ReceiptResponse represents the response from receipt generation
type ReceiptResponse struct {
	Success      bool     `json:"success"`
	Receipt      *Receipt `json:"receipt,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty"`
	ErrorCode    string   `json:"error_code,omitempty"`
}
