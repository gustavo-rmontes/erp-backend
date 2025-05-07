package models

type Sale struct {
	ID       int     `json:"id,omitempty"`
	Product  string  `json:"product" validate:"required"`
	Quantity int     `json:"quantity" validate:"required,gt=0"`
	Price    float64 `json:"price" validate:"required,gt=0"`
	Customer string  `json:"customer" validate:"required,email"`
}
