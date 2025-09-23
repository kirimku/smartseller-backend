package model

// Booking represents a courier booking response
type Booking struct {
	CourierID   string
	BookingCode string
	Status      string
	Reference   map[string]interface{}
}