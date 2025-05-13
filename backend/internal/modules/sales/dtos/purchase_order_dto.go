package dtos

import "time"

// PurchaseOrderCreateDTO representa os dados para criar um purchase order
type PurchaseOrderCreateDTO struct {
	SalesOrderID    int               `json:"sales_order_id,omitempty"`
	ContactID       int               `json:"contact_id" validate:"required"`
	ExpectedDate    time.Time         `json:"expected_date" validate:"required"`
	PaymentTerms    string            `json:"payment_terms,omitempty"`
	ShippingAddress string            `json:"shipping_address,omitempty"`
	Notes           string            `json:"notes,omitempty"`
	Items           []POItemCreateDTO `json:"items" validate:"required,min=1,dive"`
}

// PurchaseOrderUpdateDTO representa os dados para atualizar um purchase order
type PurchaseOrderUpdateDTO struct {
	ExpectedDate    *time.Time `json:"expected_date,omitempty"`
	PaymentTerms    *string    `json:"payment_terms,omitempty"`
	ShippingAddress *string    `json:"shipping_address,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
	Status          *string    `json:"status,omitempty" validate:"omitempty,oneof=draft sent confirmed received cancelled"`
}

// PurchaseOrderResponseDTO representa os dados retornados de um purchase order
type PurchaseOrderResponseDTO struct {
	ID              int                 `json:"id"`
	PONo            string              `json:"po_no"`
	SONo            string              `json:"so_no,omitempty"`
	SalesOrderID    int                 `json:"sales_order_id,omitempty"`
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
	Items           []POItemResponseDTO `json:"items,omitempty"`
	DeliveryCount   int                 `json:"delivery_count"`
	IsOverdue       bool                `json:"is_overdue"`
	DaysOverdue     int                 `json:"days_overdue,omitempty"`
}

// PurchaseOrderListItemDTO representa uma versão resumida para listagens
type PurchaseOrderListItemDTO struct {
	ID            int               `json:"id"`
	PONo          string            `json:"po_no"`
	SONo          string            `json:"so_no,omitempty"`
	ContactID     int               `json:"contact_id"`
	Contact       *ContactBasicInfo `json:"contact,omitempty"`
	Status        string            `json:"status"`
	CreatedAt     time.Time         `json:"created_at"`
	ExpectedDate  time.Time         `json:"expected_date"`
	GrandTotal    float64           `json:"grand_total"`
	ItemCount     int               `json:"item_count"`
	DeliveryCount int               `json:"delivery_count"`
	IsOverdue     bool              `json:"is_overdue"`
	DaysOverdue   int               `json:"days_overdue,omitempty"`
}

// POItemCreateDTO representa os dados para criar um item de PO
type POItemCreateDTO struct {
	ProductID   int     `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name,omitempty"`
	ProductCode string  `json:"product_code,omitempty"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" validate:"min=0,max=100"`
	Tax         float64 `json:"tax" validate:"min=0"`
}

// POItemUpdateDTO representa os dados para atualizar um item
type POItemUpdateDTO struct {
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice   *float64 `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	Discount    *float64 `json:"discount,omitempty" validate:"omitempty,min=0,max=100"`
	Tax         *float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	Description *string  `json:"description,omitempty"`
}

// POItemResponseDTO representa os dados retornados de um item
type POItemResponseDTO struct {
	ID              int     `json:"id"`
	PurchaseOrderID int     `json:"purchase_order_id"`
	ProductID       int     `json:"product_id"`
	ProductName     string  `json:"product_name"`
	ProductCode     string  `json:"product_code"`
	Description     string  `json:"description,omitempty"`
	Quantity        int     `json:"quantity"`
	UnitPrice       float64 `json:"unit_price"`
	Discount        float64 `json:"discount"`
	Tax             float64 `json:"tax"`
	Total           float64 `json:"total"`
	ReceivedQty     int     `json:"received_qty,omitempty"`
	PendingQty      int     `json:"pending_qty,omitempty"`
}

// POStatusUpdateDTO representa dados para atualizar status
type POStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=draft sent confirmed received cancelled"`
	Notes  string `json:"notes,omitempty"`
}

// CreatePOFromSODTO representa dados para criar PO de SO
type CreatePOFromSODTO struct {
	SalesOrderID    int              `json:"sales_order_id" validate:"required"`
	ContactID       int              `json:"contact_id" validate:"required"`
	ExpectedDate    time.Time        `json:"expected_date" validate:"required"`
	PaymentTerms    string           `json:"payment_terms,omitempty"`
	ShippingAddress string           `json:"shipping_address,omitempty"`
	Notes           string           `json:"notes,omitempty"`
	ItemMapping     []ItemMappingDTO `json:"item_mapping,omitempty"`
	IncludeAllItems bool             `json:"include_all_items"`
}

// ItemMappingDTO representa mapeamento de itens entre documentos
type ItemMappingDTO struct {
	FromItemID  int     `json:"from_item_id" validate:"required"`
	ToProductID int     `json:"to_product_id" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// PurchaseOrderSendDTO representa dados para enviar PO
type PurchaseOrderSendDTO struct {
	EmailTo      []string `json:"email_to" validate:"required,min=1,dive,email"`
	EmailCC      []string `json:"email_cc,omitempty" validate:"omitempty,dive,email"`
	EmailSubject string   `json:"email_subject,omitempty"`
	EmailBody    string   `json:"email_body,omitempty"`
	AttachPDF    bool     `json:"attach_pdf"`
}

// PurchaseOrderCloneDTO representa dados para clonar PO
type PurchaseOrderCloneDTO struct {
	ContactID       int       `json:"contact_id,omitempty"`
	ExpectedDate    time.Time `json:"expected_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	CloneItems      bool      `json:"clone_items"`
}

// SupplierPerformanceDTO representa performance do fornecedor
type SupplierPerformanceDTO struct {
	ContactID         int     `json:"contact_id"`
	ContactName       string  `json:"contact_name"`
	TotalOrders       int     `json:"total_orders"`
	CompletedOrders   int     `json:"completed_orders"`
	OnTimeDeliveries  int     `json:"on_time_deliveries"`
	DelayedDeliveries int     `json:"delayed_deliveries"`
	FulfillmentRate   float64 `json:"fulfillment_rate"`
	OnTimeRate        float64 `json:"on_time_rate"`
	AverageDelay      float64 `json:"average_delay_days"`
	TotalValue        float64 `json:"total_value"`
	AverageValue      float64 `json:"average_value"`
	QualityScore      float64 `json:"quality_score,omitempty"`
}

// POPriceComparisonDTO representa comparação de preços entre fornecedores
type POPriceComparisonDTO struct {
	ProductID           int                `json:"product_id"`
	ProductName         string             `json:"product_name"`
	ProductCode         string             `json:"product_code"`
	Suppliers           []SupplierPriceDTO `json:"suppliers"`
	AveragePrice        float64            `json:"average_price"`
	MinPrice            float64            `json:"min_price"`
	MaxPrice            float64            `json:"max_price"`
	RecommendedSupplier *SupplierPriceDTO  `json:"recommended_supplier,omitempty"`
}

// SupplierPriceDTO representa preço de fornecedor
type SupplierPriceDTO struct {
	ContactID      int       `json:"contact_id"`
	ContactName    string    `json:"contact_name"`
	UnitPrice      float64   `json:"unit_price"`
	MinQuantity    int       `json:"min_quantity,omitempty"`
	LeadTimeDays   int       `json:"lead_time_days,omitempty"`
	LastUpdateDate time.Time `json:"last_update_date"`
	Notes          string    `json:"notes,omitempty"`
}

// POApprovalDTO representa dados de aprovação de PO
type POApprovalDTO struct {
	PurchaseOrderID  int      `json:"purchase_order_id" validate:"required"`
	Action           string   `json:"action" validate:"required,oneof=approve reject request_changes"`
	Reason           string   `json:"reason,omitempty"`
	Notes            string   `json:"notes,omitempty"`
	ChangesRequested []string `json:"changes_requested,omitempty"`
}

// BulkPOCreateDTO representa criação de PO em massa
type BulkPOCreateDTO struct {
	TemplateID int                    `json:"template_id,omitempty"`
	ContactID  int                    `json:"contact_id" validate:"required"`
	Items      []BulkPOItemDTO        `json:"items" validate:"required,min=1"`
	SplitBy    string                 `json:"split_by,omitempty" validate:"omitempty,oneof=category supplier warehouse"`
	CommonData PurchaseOrderCommonDTO `json:"common_data"`
}

// BulkPOItemDTO representa item para PO em massa
type BulkPOItemDTO struct {
	ProductID  int     `json:"product_id" validate:"required"`
	Quantity   int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice  float64 `json:"unit_price" validate:"required,gt=0"`
	SupplierID int     `json:"supplier_id,omitempty"`
}

// PurchaseOrderCommonDTO representa dados comuns para PO
type PurchaseOrderCommonDTO struct {
	ExpectedDate    time.Time `json:"expected_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
}

// RecurringPODTO representa PO recorrente
type RecurringPODTO struct {
	BasePOID     int        `json:"base_po_id" validate:"required"`
	Frequency    string     `json:"frequency" validate:"required,oneof=weekly monthly quarterly"`
	StartDate    time.Time  `json:"start_date" validate:"required"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	NextPODate   time.Time  `json:"next_po_date"`
	AutoApprove  bool       `json:"auto_approve"`
	IsActive     bool       `json:"is_active"`
	NotifyBefore int        `json:"notify_before_days,omitempty"`
}
