package models

import (
	"time"
)

// Payment represents a payment made against an invoice
type Payment struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	InvoiceID     int       `json:"invoice_id" gorm:"index"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	PaymentDate   time.Time `json:"payment_date" gorm:"autoCreateTime"`
	PaymentMethod string    `json:"payment_method"`
	Reference     string    `json:"reference"`
	Notes         string    `json:"notes"`

	// Relationships
	Invoice *Invoice `json:"-" gorm:"foreignKey:InvoiceID"`
}
