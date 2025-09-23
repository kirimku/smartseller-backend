package model

import "time"

// TrackingInfo represents tracking information for a package
type TrackingInfo struct {
	TrackingNumber    string     `json:"tracking_number"`
	Status            string     `json:"status"`
	StatusText        string     `json:"status_text"`
	Courier           string     `json:"courier"`
	Location          string     `json:"location"`
	LastUpdate        time.Time  `json:"last_update"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	ActualDelivery    *time.Time `json:"actual_delivery,omitempty"`

	// Actual shipping data from partner (for discrepancy detection)
	ActualWeight      *float64 `json:"actual_weight,omitempty"`
	ActualShippingFee *float64 `json:"actual_shipping_fee,omitempty"`
	ShipperAddress    *string  `json:"shipper_address,omitempty"`
	ReceiverAddress   *string  `json:"receiver_address,omitempty"`

	History []TrackingStep `json:"history"`
}

// TrackingStep represents a single step in a tracking history
type TrackingStep struct {
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
}
