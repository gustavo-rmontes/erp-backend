package dtos

// BaseFilter contém campos comuns de filtro
type BaseFilter struct {
	DateRange   *DateRange `json:"date_range,omitempty"`
	ContactID   int        `json:"contact_id,omitempty"`
	ContactType string     `json:"contact_type,omitempty"`
	PersonType  string     `json:"person_type,omitempty"`
	SearchQuery string     `json:"search_query,omitempty"`
	Status      []string   `json:"status,omitempty"`
	PaginationRequest
}

// DeliveryFilter define os filtros para busca avançada de deliveries
type DeliveryFilter struct {
	BaseFilter
	PurchaseOrderID   int        `json:"purchase_order_id,omitempty"`
	SalesOrderID      int        `json:"sales_order_id,omitempty"`
	DeliveryDateRange *DateRange `json:"delivery_date_range,omitempty"`
	ReceivedDateRange *DateRange `json:"received_date_range,omitempty"`
	ShippingMethod    string     `json:"shipping_method,omitempty"`
	HasTrackingNumber *bool      `json:"has_tracking_number,omitempty"`
	IsOverdue         *bool      `json:"is_overdue,omitempty"`
	DeliveryType      string     `json:"delivery_type,omitempty" validate:"omitempty,oneof=incoming outgoing"`
}

// InvoiceFilter define os filtros para busca avançada de invoices
type InvoiceFilter struct {
	BaseFilter
	SalesOrderID   int          `json:"sales_order_id,omitempty"`
	DueDateRange   *DateRange   `json:"due_date_range,omitempty"`
	IssueDateRange *DateRange   `json:"issue_date_range,omitempty"`
	AmountRange    *AmountRange `json:"amount_range,omitempty"`
	HasPayment     *bool        `json:"has_payment,omitempty"`
	IsOverdue      *bool        `json:"is_overdue,omitempty"`
}

// PaymentFilter define os filtros para busca avançada de payments
type PaymentFilter struct {
	InvoiceID     int          `json:"invoice_id,omitempty"`
	ContactID     int          `json:"contact_id,omitempty"`
	DateRange     *DateRange   `json:"date_range,omitempty"`
	AmountRange   *AmountRange `json:"amount_range,omitempty"`
	PaymentMethod []string     `json:"payment_method,omitempty"`
	HasReference  *bool        `json:"has_reference,omitempty"`
	SearchQuery   string       `json:"search_query,omitempty"`
	PaginationRequest
}

// PurchaseOrderFilter define os filtros para busca avançada de purchase orders
type PurchaseOrderFilter struct {
	BaseFilter
	SalesOrderID      int          `json:"sales_order_id,omitempty"`
	ExpectedDateRange *DateRange   `json:"expected_date_range,omitempty"`
	AmountRange       *AmountRange `json:"amount_range,omitempty"`
	HasDelivery       *bool        `json:"has_delivery,omitempty"`
	IsOverdue         *bool        `json:"is_overdue,omitempty"`
}

// QuotationFilter define os filtros para busca avançada de quotations
type QuotationFilter struct {
	BaseFilter
	ExpiryDateRange *DateRange   `json:"expiry_date_range,omitempty"`
	AmountRange     *AmountRange `json:"amount_range,omitempty"`
	IsExpired       *bool        `json:"is_expired,omitempty"`
}

// SalesOrderFilter define os filtros para busca avançada de sales orders
type SalesOrderFilter struct {
	BaseFilter
	QuotationID       int          `json:"quotation_id,omitempty"`
	ExpectedDateRange *DateRange   `json:"expected_date_range,omitempty"`
	AmountRange       *AmountRange `json:"amount_range,omitempty"`
	HasInvoice        *bool        `json:"has_invoice,omitempty"`
	HasPurchaseOrder  *bool        `json:"has_purchase_order,omitempty"`
}

// SalesProcessFilter define os filtros para busca avançada de sales processes
type SalesProcessFilter struct {
	BaseFilter
	AmountRange      *AmountRange `json:"amount_range,omitempty"`
	ProfitRange      *AmountRange `json:"profit_range,omitempty"`
	HasQuotation     *bool        `json:"has_quotation,omitempty"`
	HasSalesOrder    *bool        `json:"has_sales_order,omitempty"`
	HasPurchaseOrder *bool        `json:"has_purchase_order,omitempty"`
	HasInvoice       *bool        `json:"has_invoice,omitempty"`
	IsComplete       *bool        `json:"is_complete,omitempty"`
}
