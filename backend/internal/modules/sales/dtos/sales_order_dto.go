package dtos

import "time"

// SalesOrderCreateDTO representa os dados para criar um sales order
type SalesOrderCreateDTO struct {
	QuotationID     int               `json:"quotation_id,omitempty"`
	ContactID       int               `json:"contact_id" validate:"required"`
	ExpectedDate    time.Time         `json:"expected_date" validate:"required"`
	PaymentTerms    string            `json:"payment_terms,omitempty"`
	ShippingAddress string            `json:"shipping_address,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	Items           []SOItemCreateDTO `json:"items" validate:"required,min=1,dive"`
}

// SalesOrderUpdateDTO representa os dados para atualizar um sales order
type SalesOrderUpdateDTO struct {
	ExpectedDate    *time.Time `json:"expected_date,omitempty"`
	PaymentTerms    *string    `json:"payment_terms,omitempty"`
	ShippingAddress *string    `json:"shipping_address,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

// SalesOrderResponseDTO representa os dados retornados de um sales order
type SalesOrderResponseDTO struct {
	ID              int                 `json:"id"`
	SONo            string              `json:"so_no"`
	QuotationID     int                 `json:"quotation_id,omitempty"`
	ContactID       int                 `json:"contact_id"`
	Contact         *ContactBasicInfo   `json:"contact,omitempty"`
	Status          string              `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
	ExpectedDate    time.Time           `json:"expected_date"`
	SubTotal        float64             `json:"subtotal"`
	TaxTotal        float64             `json:"tax_total"`
	DiscountTotal   float64             `json:"discount_total"`
	GrandTotal      float64             `json:"grand_total"`
	Notes           string              `json:"notes,omitempty"`
	PaymentTerms    string              `json:"payment_terms,omitempty"`
	ShippingAddress string              `json:"shipping_address,omitempty"`
	Items           []SOItemResponseDTO `json:"items,omitempty"`
	InvoiceCount    int                 `json:"invoice_count"`
	POCount         int                 `json:"po_count"`
	DeliveryCount   int                 `json:"delivery_count"`
	FulfillmentRate float64             `json:"fulfillment_rate"`
}

// SalesOrderListItemDTO representa uma versão resumida para listagens
type SalesOrderListItemDTO struct {
	ID              int               `json:"id"`
	SONo            string            `json:"so_no"`
	ContactID       int               `json:"contact_id"`
	Contact         *ContactBasicInfo `json:"contact,omitempty"`
	Status          string            `json:"status"`
	CreatedAt       time.Time         `json:"created_at"`
	ExpectedDate    time.Time         `json:"expected_date"`
	GrandTotal      float64           `json:"grand_total"`
	ItemCount       int               `json:"item_count"`
	InvoiceCount    int               `json:"invoice_count"`
	DeliveryCount   int               `json:"delivery_count"`
	FulfillmentRate float64           `json:"fulfillment_rate"`
}

// SOItemCreateDTO representa os dados para criar um item de SO
type SOItemCreateDTO struct {
	ProductID   int     `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name,omitempty"`
	ProductCode string  `json:"product_code,omitempty"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" validate:"min=0,max=100"`
	Tax         float64 `json:"tax" validate:"min=0"`
}

// SOItemUpdateDTO representa os dados para atualizar um item
type SOItemUpdateDTO struct {
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice   *float64 `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	Discount    *float64 `json:"discount,omitempty" validate:"omitempty,min=0,max=100"`
	Tax         *float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	Description *string  `json:"description,omitempty"`
}

// SOItemResponseDTO representa os dados retornados de um item
type SOItemResponseDTO struct {
	ID           int     `json:"id"`
	SalesOrderID int     `json:"sales_order_id"`
	ProductID    int     `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductCode  string  `json:"product_code"`
	Description  string  `json:"description,omitempty"`
	Quantity     int     `json:"quantity"`
	UnitPrice    float64 `json:"unit_price"`
	Discount     float64 `json:"discount"`
	Tax          float64 `json:"tax"`
	Total        float64 `json:"total"`
	DeliveredQty int     `json:"delivered_qty"`
	InvoicedQty  int     `json:"invoiced_qty"`
	PendingQty   int     `json:"pending_qty"`
}

// SOStatusUpdateDTO representa dados para atualizar status
type SOStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=draft confirmed processing completed cancelled"`
	Notes  string `json:"notes,omitempty"`
}

// CreateFromQuotationDTO representa dados para criar de uma quotation
type CreateFromQuotationDTO struct {
	QuotationID     int       `json:"quotation_id" validate:"required"`
	ExpectedDate    time.Time `json:"expected_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	IncludeAllItems bool      `json:"include_all_items"`
}

// SalesOrderConfirmDTO representa dados para confirmar SO
type SalesOrderConfirmDTO struct {
	SendConfirmation bool     `json:"send_confirmation"`
	EmailTo          []string `json:"email_to,omitempty" validate:"omitempty,dive,email"`
	EmailSubject     string   `json:"email_subject,omitempty"`
	EmailBody        string   `json:"email_body,omitempty"`
	Notes            string   `json:"notes,omitempty"`
}

// SOFulfillmentDTO representa dados de fulfillment
type SOFulfillmentDTO struct {
	SalesOrderID    int                    `json:"sales_order_id"`
	Items           []SOItemFulfillmentDTO `json:"items"`
	TotalItems      int                    `json:"total_items"`
	FulfilledItems  int                    `json:"fulfilled_items"`
	PendingItems    int                    `json:"pending_items"`
	FulfillmentRate float64                `json:"fulfillment_rate"`
}

// SOItemFulfillmentDTO representa fulfillment de item
type SOItemFulfillmentDTO struct {
	ItemID       int     `json:"item_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
	DeliveredQty int     `json:"delivered_qty"`
	InvoicedQty  int     `json:"invoiced_qty"`
	PendingQty   int     `json:"pending_qty"`
	Status       string  `json:"status"` // pending, partial, complete
	Percentage   float64 `json:"percentage"`
}

// SalesOrderCloneDTO representa dados para clonar SO
type SalesOrderCloneDTO struct {
	ContactID       int       `json:"contact_id,omitempty"`
	ExpectedDate    time.Time `json:"expected_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

// BackOrderDTO representa dados de back order
type BackOrderDTO struct {
	OriginalSOID int                `json:"original_so_id" validate:"required"`
	ExpectedDate time.Time          `json:"expected_date" validate:"required"`
	Notes        string             `json:"notes,omitempty"`
	Items        []BackOrderItemDTO `json:"items" validate:"required,min=1"`
}

// BackOrderItemDTO representa item de back order
type BackOrderItemDTO struct {
	OriginalItemID int `json:"original_item_id" validate:"required"`
	Quantity       int `json:"quantity" validate:"required,gt=0"`
}
