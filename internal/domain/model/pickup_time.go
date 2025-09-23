package model

import "time"

// PickupTime represents an available pickup time slot from a courier
type PickupTime struct {
	PickupRequestTime time.Time `json:"pickup_request_time"`
	Date              string    `json:"date"`
	StartTime         string    `json:"start_time"`
	EndTime           string    `json:"end_time"`
	IsAvailable       bool      `json:"is_available"`
	CourierID         string    `json:"courier_id"`
	ServiceType       string    `json:"service_type,omitempty"`
}
