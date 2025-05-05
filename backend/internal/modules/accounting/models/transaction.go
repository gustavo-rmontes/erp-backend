package models

type Transaction struct {
	ID          int     `json:"id,omitempty"`
	Description string  `json:"description" validate:"required"`
	Amount      float64 `json:"amount" validate:"required"`
	Date        string  `json:"date" validate:"required,datetime=02/01/2006"`
}
