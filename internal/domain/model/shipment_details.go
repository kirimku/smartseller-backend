package model

import "time"

// ShipmentDetails represents detailed shipment information extracted from logistics partners
// This struct contains actual data for discrepancy detection and comprehensive tracking
type ShipmentDetails struct {
	TrackingNumber  string     `json:"tracking_number"`
	ActualWeight    *float64   `json:"actual_weight,omitempty"`
	ActualFee       *float64   `json:"actual_fee,omitempty"`
	ServiceType     string     `json:"service_type"`
	Origin          string     `json:"origin"`
	Destination     string     `json:"destination"`
	ReceiverName    string     `json:"receiver_name"`
	ReceiverAddress string     `json:"receiver_address"`
	ShipperAddress  string     `json:"shipper_address"`
	DeliveryTime    *time.Time `json:"delivery_time,omitempty"`
	IsDelivered     bool       `json:"is_delivered"`
}
