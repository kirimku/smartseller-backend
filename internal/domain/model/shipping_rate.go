package model

import "time"

// ShippingRate represents a calculated shipping rate from a courier
type ShippingRate struct {
	CourierID    string    `json:"courier_id"`
	CourierName  string    `json:"courier_name"`
	ServiceID    string    `json:"service_id"`
	ServiceName  string    `json:"service_name"`
	ServiceType  string    `json:"service_type"`
	Description  string    `json:"description"`
	Price        float64   `json:"price"`
	EstimatedETA string    `json:"estimated_eta"`
	CreatedAt    time.Time `json:"created_at"`
}
