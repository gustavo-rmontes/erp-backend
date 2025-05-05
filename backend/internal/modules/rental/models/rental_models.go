package models

type Rental struct {
	ID          int     `json:"id"`
	ClientName  string  `json:"client_name" binding:"required"`
	Equipment   string  `json:"equipment" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     string  `json:"end_date" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	BillingType string  `json:"billing_type" binding:"required"` // mensal, anual, etc.
}
