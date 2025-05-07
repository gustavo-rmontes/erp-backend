// sales_process.go
package models

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	product "ERP-ONSMART/backend/internal/modules/products/models"
	"time"
)

// SalesItem represents an item in a quotation, SO, or PO
type SalesItem struct {
	ID          int     `json:"id" gorm:"primaryKey"`
	ProductID   int     `json:"product_id" validate:"required" gorm:"index"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" gorm:"default:0"`
	Tax         float64 `json:"tax" gorm:"default:0"`
	Total       float64 `json:"total"`

	// Relationships (not stored in DB)
	Product *product.Product `json:"product,omitempty" gorm:"-"`
}

// SalesProcess represents the full sales process linking all documents
type SalesProcess struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	ContactID  int       `json:"contact_id" validate:"required" gorm:"index"`
	Status     string    `json:"status" validate:"required"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	TotalValue float64   `json:"total_value"`
	Profit     float64   `json:"profit"`
	Notes      string    `json:"notes"`

	// Relationships
	Contact       *contact.Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	Quotation     *Quotation       `json:"quotation,omitempty" gorm:"-"`
	SalesOrder    *SalesOrder      `json:"sales_order,omitempty" gorm:"-"`
	PurchaseOrder *PurchaseOrder   `json:"purchase_order,omitempty" gorm:"-"`
	Deliveries    []Delivery       `json:"deliveries,omitempty" gorm:"-"`
	Invoices      []Invoice        `json:"invoices,omitempty" gorm:"-"`
}
