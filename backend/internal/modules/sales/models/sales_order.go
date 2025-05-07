package models

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	product "ERP-ONSMART/backend/internal/modules/products/models"
	"time"
)

// SalesOrder represents a sales order from a client
type SalesOrder struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	SONo            string    `json:"so_no" validate:"required" gorm:"uniqueIndex"`
	QuotationID     int       `json:"quotation_id" gorm:"index"`
	ContactID       int       `json:"contact_id" validate:"required" gorm:"index"`
	Status          string    `json:"status" validate:"required" gorm:"default:draft"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	ExpectedDate    time.Time `json:"expected_date"`
	SubTotal        float64   `json:"subtotal" gorm:"column:subtotal"`
	TaxTotal        float64   `json:"tax_total" gorm:"column:tax_total"`
	DiscountTotal   float64   `json:"discount_total" gorm:"column:discount_total"`
	GrandTotal      float64   `json:"grand_total" gorm:"column:grand_total"`
	Notes           string    `json:"notes"`
	PaymentTerms    string    `json:"payment_terms"`
	ShippingAddress string    `json:"shipping_address"`

	// Relationships
	Contact   *contact.Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Quotation *Quotation       `json:"quotation,omitempty" gorm:"foreignKey:QuotationID"`
	Items     []SOItem         `json:"items,omitempty" gorm:"foreignKey:SalesOrderID"`
}

// SOItem represents items in a sales order
type SOItem struct {
	ID           int     `json:"id" gorm:"primaryKey"`
	SalesOrderID int     `json:"sales_order_id" gorm:"index"`
	ProductID    int     `json:"product_id" validate:"required" gorm:"index"`
	ProductName  string  `json:"product_name"`
	ProductCode  string  `json:"product_code"`
	Description  string  `json:"description"`
	Quantity     int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice    float64 `json:"unit_price" validate:"required,gt=0"`
	Discount     float64 `json:"discount" gorm:"default:0"`
	Tax          float64 `json:"tax" gorm:"default:0"`
	Total        float64 `json:"total"`

	// Relationships
	Product    *product.Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	SalesOrder *SalesOrder      `json:"-" gorm:"foreignKey:SalesOrderID"`
}

// TableName define o nome da tabela para o modelo SOItem
func (SOItem) TableName() string {
	return "sales_order_items"
}
