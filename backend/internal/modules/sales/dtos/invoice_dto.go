package dtos

import "time"

// InvoiceCreateDTO representa os dados para criar uma invoice
type InvoiceCreateDTO struct {
	SalesOrderID int                    `json:"sales_order_id,omitempty"`
	ContactID    int                    `json:"contact_id" validate:"required"`
	IssueDate    time.Time              `json:"issue_date" validate:"required"`
	DueDate      time.Time              `json:"due_date" validate:"required"`
	PaymentTerms string                 `json:"payment_terms,omitempty"`
	Notes        string                 `json:"notes,omitempty"`
	Items        []InvoiceItemCreateDTO `json:"items" validate:"required,min=1,dive"`
}

// InvoiceUpdateDTO representa os dados para atualizar uma invoice
type InvoiceUpdateDTO struct {
	IssueDate    *time.Time `json:"issue_date,omitempty"`
	DueDate      *time.Time `json:"due_date,omitempty"`
	PaymentTerms *string    `json:"payment_terms,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
}

// InvoiceResponseDTO representa os dados retornados de uma invoice
type InvoiceResponseDTO struct {
	ID            int                      `json:"id"`
	InvoiceNo     string                   `json:"invoice_no"`
	SalesOrderID  int                      `json:"sales_order_id,omitempty"`
	SONo          string                   `json:"so_no,omitempty"`
	ContactID     int                      `json:"contact_id"`
	Contact       *ContactBasicInfo        `json:"contact,omitempty"`
	Status        string                   `json:"status"`
	CreatedAt     time.Time                `json:"created_at"`
	UpdatedAt     time.Time                `json:"updated_at"`
	IssueDate     time.Time                `json:"issue_date"`
	DueDate       time.Time                `json:"due_date"`
	SubTotal      float64                  `json:"subtotal"`
	TaxTotal      float64                  `json:"tax_total"`
	DiscountTotal float64                  `json:"discount_total"`
	GrandTotal    float64                  `json:"grand_total"`
	AmountPaid    float64                  `json:"amount_paid"`
	BalanceDue    float64                  `json:"balance_due"`
	PaymentTerms  string                   `json:"payment_terms,omitempty"`
	Notes         string                   `json:"notes,omitempty"`
	Items         []InvoiceItemResponseDTO `json:"items,omitempty"`
	Payments      []PaymentResponseDTO     `json:"payments,omitempty"`
	IsOverdue     bool                     `json:"is_overdue"`
	DaysOverdue   int                      `json:"days_overdue,omitempty"`
}

// InvoiceListItemDTO representa uma versão resumida para listagens
type InvoiceListItemDTO struct {
	ID          int               `json:"id"`
	InvoiceNo   string            `json:"invoice_no"`
	ContactID   int               `json:"contact_id"`
	Contact     *ContactBasicInfo `json:"contact,omitempty"`
	Status      string            `json:"status"`
	IssueDate   time.Time         `json:"issue_date"`
	DueDate     time.Time         `json:"due_date"`
	GrandTotal  float64           `json:"grand_total"`
	AmountPaid  float64           `json:"amount_paid"`
	BalanceDue  float64           `json:"balance_due"`
	IsOverdue   bool              `json:"is_overdue"`
	DaysOverdue int               `json:"days_overdue,omitempty"`
}

// InvoiceItemCreateDTO representa os dados para criar um item de invoice
type InvoiceItemCreateDTO struct {
	ProductID   int     `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name,omitempty"`
	ProductCode string  `json:"product_code,omitempty"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" validate:"min=0,max=100"`
	Tax         float64 `json:"tax" validate:"min=0"`
}

// InvoiceItemUpdateDTO representa os dados para atualizar um item
type InvoiceItemUpdateDTO struct {
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice   *float64 `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	Discount    *float64 `json:"discount,omitempty" validate:"omitempty,min=0,max=100"`
	Tax         *float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	Description *string  `json:"description,omitempty"`
}

// InvoiceItemResponseDTO representa os dados retornados de um item
type InvoiceItemResponseDTO struct {
	ID          int     `json:"id"`
	InvoiceID   int     `json:"invoice_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	Tax         float64 `json:"tax"`
	Total       float64 `json:"total"`
}

// InvoiceStatusUpdateDTO representa dados para atualizar status
type InvoiceStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=draft sent partial paid overdue cancelled"`
	Notes  string `json:"notes,omitempty"`
}

// InvoicePaymentSummaryDTO representa resumo de pagamentos
type InvoicePaymentSummaryDTO struct {
	InvoiceID       int        `json:"invoice_id"`
	InvoiceNo       string     `json:"invoice_no"`
	GrandTotal      float64    `json:"grand_total"`
	AmountPaid      float64    `json:"amount_paid"`
	BalanceDue      float64    `json:"balance_due"`
	LastPaymentDate *time.Time `json:"last_payment_date,omitempty"`
	PaymentCount    int        `json:"payment_count"`
	Status          string     `json:"status"`
}

// CreateInvoiceFromSODTO representa dados para criar invoice de SO
type CreateInvoiceFromSODTO struct {
	SalesOrderID    int       `json:"sales_order_id" validate:"required"`
	IssueDate       time.Time `json:"issue_date" validate:"required"`
	DueDate         time.Time `json:"due_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	IncludeAllItems bool      `json:"include_all_items"`
}

// InvoiceSendDTO representa dados para enviar invoice
type InvoiceSendDTO struct {
	EmailTo      []string `json:"email_to" validate:"required,min=1,dive,email"`
	EmailCC      []string `json:"email_cc,omitempty" validate:"omitempty,dive,email"`
	EmailSubject string   `json:"email_subject,omitempty"`
	EmailBody    string   `json:"email_body,omitempty"`
	AttachPDF    bool     `json:"attach_pdf"`
}

// InvoiceCloneDTO representa dados para clonar invoice
type InvoiceCloneDTO struct {
	ContactID int       `json:"contact_id,omitempty"`
	IssueDate time.Time `json:"issue_date" validate:"required"`
	DueDate   time.Time `json:"due_date" validate:"required"`
	Notes     string    `json:"notes,omitempty"`
}

// RecurringInvoiceDTO representa dados para invoice recorrente
type RecurringInvoiceDTO struct {
	BaseInvoiceID int        `json:"base_invoice_id" validate:"required"`
	Frequency     string     `json:"frequency" validate:"required,oneof=weekly monthly quarterly yearly"`
	StartDate     time.Time  `json:"start_date" validate:"required"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	NextDueDate   time.Time  `json:"next_due_date"`
	AutoSend      bool       `json:"auto_send"`
	IsActive      bool       `json:"is_active"`
}
