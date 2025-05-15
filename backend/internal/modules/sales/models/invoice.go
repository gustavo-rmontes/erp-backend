package models

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	product "ERP-ONSMART/backend/internal/modules/products/models"
	"time"
)

// Invoice represents an invoice to a client
type Invoice struct {
	ID            int       `json:"id" gorm:"primaryKey"`
	InvoiceNo     string    `json:"invoice_no" validate:"required" gorm:"uniqueIndex"`
	SalesOrderID  int       `json:"sales_order_id" gorm:"index"`
	SONo          string    `json:"so_no"`
	ContactID     int       `json:"contact_id" validate:"required" gorm:"index"`
	Status        string    `json:"status" validate:"required" gorm:"default:draft"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	IssueDate     time.Time `json:"issue_date"`
	DueDate       time.Time `json:"due_date" validate:"required"`
	SubTotal      float64   `json:"subtotal" gorm:"column:subtotal"`
	TaxTotal      float64   `json:"tax_total" gorm:"column:tax_total"`
	DiscountTotal float64   `json:"discount_total" gorm:"column:discount_total"`
	GrandTotal    float64   `json:"grand_total" gorm:"column:grand_total"`
	AmountPaid    float64   `json:"amount_paid" gorm:"default:0"`
	PaymentTerms  string    `json:"payment_terms"`
	Notes         string    `json:"notes"`

	// Relationships
	Contact    *contact.Contact `json:"contact,omitempty" gorm:"foreignKey:ContactID"`
	SalesOrder *SalesOrder      `json:"sales_order,omitempty" gorm:"foreignKey:SalesOrderID"`
	Items      []InvoiceItem    `json:"items,omitempty" gorm:"foreignKey:InvoiceID"`
	Payments   []Payment        `json:"payments,omitempty" gorm:"foreignKey:InvoiceID"`
}

// InvoiceItem represents items in an invoice
type InvoiceItem struct {
	ID          int     `json:"id" gorm:"primaryKey"`
	InvoiceID   int     `json:"invoice_id" gorm:"index"`
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
	Product *product.Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Invoice *Invoice         `json:"-" gorm:"foreignKey:InvoiceID"`
}
