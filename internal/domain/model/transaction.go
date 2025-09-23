package model

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeSend    TransactionType = "send"
	TransactionTypeReceive TransactionType = "receive"
	TransactionTypeReturn  TransactionType = "return"
)

// Transaction represents a shipping transaction
type Transaction struct {
	ID              string
	UniqueID        string // Added to match entity.Transaction
	TransactionType TransactionType
	Amount          float64
	OrderPrice      float64 // Order price from database, different from total amount
	Weight          int
	WeightInGrams   int // Added for SiCepat integration
	InsuranceCost   float64
	ServiceType     string // Added for SiCepat integration
	PackageCategory string // Package category for shipping classification
	IsWhiteLabel    bool   // Added for SiCepat integration
	Items           []TransactionItem
}

// TransactionItem represents an item in a transaction
type TransactionItem struct {
	Name     string
	Price    float64
	Quantity int
}
