package models

import (
	product "ERP-ONSMART/backend/internal/modules/products/models"
	"time"
)

// Delivery represents a delivery of items
type Delivery struct {
	ID              int       `json:"id" gorm:"primaryKey"`
	DeliveryNo      string    `json:"delivery_no" validate:"required" gorm:"uniqueIndex"`
	PurchaseOrderID int       `json:"purchase_order_id" gorm:"index"`
	PONo            string    `json:"po_no"`
	SalesOrderID    int       `json:"sales_order_id" gorm:"index"`
	SONo            string    `json:"so_no"`
	Status          string    `json:"status" validate:"required" gorm:"default:pending"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	DeliveryDate    time.Time `json:"delivery_date"`
	ReceivedDate    time.Time `json:"received_date"`
	ShippingMethod  string    `json:"shipping_method"`
	TrackingNumber  string    `json:"tracking_number"`
	ShippingAddress string    `json:"shipping_address"`
	Notes           string    `json:"notes"`

	// Relationships
	PurchaseOrder *PurchaseOrder `json:"purchase_order,omitempty" gorm:"foreignKey:PurchaseOrderID"`
	SalesOrder    *SalesOrder    `json:"sales_order,omitempty" gorm:"foreignKey:SalesOrderID"`
	Items         []DeliveryItem `json:"items,omitempty" gorm:"foreignKey:DeliveryID"`
}

// DeliveryItem represents an item in a delivery
type DeliveryItem struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	DeliveryID  int    `json:"delivery_id" gorm:"index"`
	ProductID   int    `json:"product_id" validate:"required" gorm:"index"`
	ProductName string `json:"product_name"`
	ProductCode string `json:"product_code"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	ReceivedQty int    `json:"received_qty" gorm:"default:0"`
	Notes       string `json:"notes"`

	// Relationships
	Product  *product.Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
	Delivery *Delivery        `json:"-" gorm:"foreignKey:DeliveryID"`
}
