package dtos

import "time"

// DeliveryCreateDTO representa os dados para criar uma delivery
type DeliveryCreateDTO struct {
	PurchaseOrderID int                     `json:"purchase_order_id,omitempty"`
	SalesOrderID    int                     `json:"sales_order_id,omitempty"`
	DeliveryDate    time.Time               `json:"delivery_date" validate:"required"`
	ShippingMethod  string                  `json:"shipping_method,omitempty"`
	ShippingAddress string                  `json:"shipping_address" validate:"required"`
	Notes           string                  `json:"notes,omitempty"`
	Items           []DeliveryItemCreateDTO `json:"items" validate:"required,min=1,dive"`
}

// DeliveryUpdateDTO representa os dados para atualizar uma delivery
type DeliveryUpdateDTO struct {
	DeliveryDate    *time.Time `json:"delivery_date,omitempty"`
	ReceivedDate    *time.Time `json:"received_date,omitempty"`
	ShippingMethod  *string    `json:"shipping_method,omitempty"`
	TrackingNumber  *string    `json:"tracking_number,omitempty"`
	ShippingAddress *string    `json:"shipping_address,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

// DeliveryResponseDTO representa os dados retornados de uma delivery
type DeliveryResponseDTO struct {
	ID              int                       `json:"id"`
	DeliveryNo      string                    `json:"delivery_no"`
	PurchaseOrderID int                       `json:"purchase_order_id,omitempty"`
	PONo            string                    `json:"po_no,omitempty"`
	SalesOrderID    int                       `json:"sales_order_id,omitempty"`
	SONo            string                    `json:"so_no,omitempty"`
	Status          string                    `json:"status"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
	DeliveryDate    time.Time                 `json:"delivery_date"`
	ReceivedDate    *time.Time                `json:"received_date,omitempty"`
	ShippingMethod  string                    `json:"shipping_method,omitempty"`
	TrackingNumber  string                    `json:"tracking_number,omitempty"`
	ShippingAddress string                    `json:"shipping_address"`
	Notes           string                    `json:"notes,omitempty"`
	Items           []DeliveryItemResponseDTO `json:"items,omitempty"`
	Contact         *ContactBasicInfo         `json:"contact,omitempty"`
}

// DeliveryListItemDTO representa uma versão resumida para listagens
type DeliveryListItemDTO struct {
	ID             int               `json:"id"`
	DeliveryNo     string            `json:"delivery_no"`
	Status         string            `json:"status"`
	DeliveryDate   time.Time         `json:"delivery_date"`
	ReceivedDate   *time.Time        `json:"received_date,omitempty"`
	Contact        *ContactBasicInfo `json:"contact,omitempty"`
	ItemCount      int               `json:"item_count"`
	TrackingNumber string            `json:"tracking_number,omitempty"`
	DeliveryType   string            `json:"delivery_type"` // incoming ou outgoing
}

// DeliveryItemCreateDTO representa os dados para criar um item de delivery
type DeliveryItemCreateDTO struct {
	ProductID   int    `json:"product_id" validate:"required"`
	ProductName string `json:"product_name,omitempty"`
	ProductCode string `json:"product_code,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity" validate:"required,gt=0"`
	Notes       string `json:"notes,omitempty"`
}

// DeliveryItemUpdateDTO representa os dados para atualizar um item
type DeliveryItemUpdateDTO struct {
	ReceivedQty *int    `json:"received_qty,omitempty" validate:"omitempty,min=0"`
	Notes       *string `json:"notes,omitempty"`
}

// DeliveryItemResponseDTO representa os dados retornados de um item
type DeliveryItemResponseDTO struct {
	ID          int    `json:"id"`
	DeliveryID  int    `json:"delivery_id"`
	ProductID   int    `json:"product_id"`
	ProductName string `json:"product_name"`
	ProductCode string `json:"product_code"`
	Description string `json:"description,omitempty"`
	Quantity    int    `json:"quantity"`
	ReceivedQty int    `json:"received_qty"`
	Notes       string `json:"notes,omitempty"`
	Status      string `json:"status"` // pending, partial, complete
}

// DeliveryStatusUpdateDTO representa dados para atualizar status
type DeliveryStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=pending shipped delivered returned"`
	Notes  string `json:"notes,omitempty"`
}

// MarkAsShippedDTO representa dados para marcar como enviado
type MarkAsShippedDTO struct {
	TrackingNumber string `json:"tracking_number" validate:"required"`
	ShippingMethod string `json:"shipping_method,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

// MarkAsDeliveredDTO representa dados para marcar como entregue
type MarkAsDeliveredDTO struct {
	ReceivedDate time.Time `json:"received_date,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// MarkAsReturnedDTO representa dados para marcar como devolvido
type MarkAsReturnedDTO struct {
	Reason     string    `json:"reason" validate:"required"`
	ReturnDate time.Time `json:"return_date,omitempty"`
	Notes      string    `json:"notes,omitempty"`
}

// DeliveryBulkUpdateDTO representa dados para atualização em massa
type DeliveryBulkUpdateDTO struct {
	DeliveryIDs    []int  `json:"delivery_ids" validate:"required,min=1"`
	Status         string `json:"status,omitempty"`
	ShippingMethod string `json:"shipping_method,omitempty"`
	Notes          string `json:"notes,omitempty"`
}

// CreateDeliveryFromPODTO representa dados para criar delivery de PO
type CreateDeliveryFromPODTO struct {
	PurchaseOrderID int       `json:"purchase_order_id" validate:"required"`
	DeliveryDate    time.Time `json:"delivery_date" validate:"required"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	IncludeAllItems bool      `json:"include_all_items"`
}

// CreateDeliveryFromSODTO representa dados para criar delivery de SO
type CreateDeliveryFromSODTO struct {
	SalesOrderID    int       `json:"sales_order_id" validate:"required"`
	DeliveryDate    time.Time `json:"delivery_date" validate:"required"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	IncludeAllItems bool      `json:"include_all_items"`
}
