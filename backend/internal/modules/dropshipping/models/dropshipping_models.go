package models

// Dropshipping representa uma transação de revenda de hardware/software/garantia.
type Dropshipping struct {
	ID         int     `json:"id"`
	ProductID  int     `json:"product"`
	WarrantyID int     `json:"warranty"`
	Cliente    string  `json:"contact"`
	Price      float64 `json:"price" binding:"required,gt=0"`
	Quantity   int     `json:"quantity" binding:"required,gte=0"`
	TotalPrice float64 `json:"total_price"`
	StartDate  string  `json:"start_date" binding:"required"`
	UpdatedAt  string  `json:"updated_at"`
}
