// quotation.go
package models

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	product "ERP-ONSMART/backend/internal/modules/products/models"
	"time"
)

// Quotation represents a sales quotation sent to a client
type Quotation struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	QuotationNo   string    `json:"quotation_no" validate:"required" gorm:"uniqueIndex"`
	ContactID     int       `json:"contact_id" validate:"required" gorm:"index"`
	Status        string    `json:"status" validate:"required" gorm:"default:draft"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiryDate    time.Time `json:"expiry_date" validate:"required"`
	SubTotal      float64   `json:"subtotal" gorm:"column:subtotal"`
	TaxTotal      float64   `json:"tax_total" gorm:"column:tax_total"`
	DiscountTotal float64   `json:"discount_total" gorm:"column:discount_total"`
	GrandTotal    float64   `json:"grand_total" gorm:"column:grand_total"`
	Notes         string    `json:"notes"`
	Terms         string    `json:"terms"`

	// Relationships
	Contact *contact.Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Items   []QuotationItem  `json:"items,omitempty" gorm:"foreignKey:QuotationID"`
}

// QuotationItem represents items in a quotation
type QuotationItem struct {
	ID          int     `json:"id" gorm:"primaryKey"`
	QuotationID int     `json:"quotation_id" gorm:"index"`
	ProductID   int     `json:"product_id" validate:"required" gorm:"index"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" gorm:"default:0"`
	Tax         float64 `json:"tax" gorm:"default:0"`
	Total       float64 `json:"total"`

	// Relationships
	Product   *product.Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Quotation *Quotation       `json:"-" gorm:"foreignKey:QuotationID"`
}
