package model

// Address represents a physical address
type Address struct {
	ID          string  `json:"id,omitempty"`
	UserID      string  `json:"user_id"`
	Name        string  `json:"name"`
	Phone       string  `json:"phone"`
	Email       string  `json:"email,omitempty"` // Added for SiCepat integration
	City        string  `json:"city"`
	Province    string  `json:"province"`
	District    string  `json:"district"`
	PostalCode  string  `json:"postal_code"`
	Country     string  `json:"country,omitempty"`
	Address     string  `json:"address"`
	AddressType string  `json:"address_type,omitempty"`
	IsPrimary   bool    `json:"is_primary,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}
