package models

type Campaign struct {
	ID          int     `json:"id,omitempty"`
	Title       string  `json:"title" validate:"required"`
	Description string  `json:"description" validate:"required"`
	Budget      float64 `json:"budget" validate:"required,gt=0"`
	StartDate   string  `json:"start_date" validate:"required,datetime=02/01/2006"`
	EndDate     string  `json:"end_date" validate:"required,datetime=02/01/2006"`
}
