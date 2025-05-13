package dtos

import "time"

// PaymentCreateDTO representa os dados para criar um payment
type PaymentCreateDTO struct {
	InvoiceID     int       `json:"invoice_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	PaymentDate   time.Time `json:"payment_date" validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	Reference     string    `json:"reference,omitempty"`
	Notes         string    `json:"notes,omitempty"`
}

// PaymentUpdateDTO representa os dados para atualizar um payment
type PaymentUpdateDTO struct {
	Amount        *float64   `json:"amount,omitempty" validate:"omitempty,gt=0"`
	PaymentDate   *time.Time `json:"payment_date,omitempty"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	Reference     *string    `json:"reference,omitempty"`
	Notes         *string    `json:"notes,omitempty"`
}

// PaymentResponseDTO representa os dados retornados de um payment
type PaymentResponseDTO struct {
	ID            int               `json:"id"`
	InvoiceID     int               `json:"invoice_id"`
	InvoiceNo     string            `json:"invoice_no,omitempty"`
	Amount        float64           `json:"amount"`
	PaymentDate   time.Time         `json:"payment_date"`
	PaymentMethod string            `json:"payment_method"`
	Reference     string            `json:"reference,omitempty"`
	Notes         string            `json:"notes,omitempty"`
	Contact       *ContactBasicInfo `json:"contact,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// PaymentListItemDTO representa uma versão resumida para listagens
type PaymentListItemDTO struct {
	ID            int               `json:"id"`
	InvoiceID     int               `json:"invoice_id"`
	InvoiceNo     string            `json:"invoice_no"`
	Contact       *ContactBasicInfo `json:"contact,omitempty"`
	Amount        float64           `json:"amount"`
	PaymentDate   time.Time         `json:"payment_date"`
	PaymentMethod string            `json:"payment_method"`
	Reference     string            `json:"reference,omitempty"`
}

// PaymentReconciliationDTO representa dados para reconciliação
type PaymentReconciliationDTO struct {
	PaymentID     int       `json:"payment_id" validate:"required"`
	BankReference string    `json:"bank_reference" validate:"required"`
	ReconcileDate time.Time `json:"reconcile_date" validate:"required"`
	Notes         string    `json:"notes,omitempty"`
}

// ProcessInvoicePaymentDTO representa dados para processar pagamento
type ProcessInvoicePaymentDTO struct {
	InvoiceID     int       `json:"invoice_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	Reference     string    `json:"reference,omitempty"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
	SendReceipt   bool      `json:"send_receipt"`
	Notes         string    `json:"notes,omitempty"`
}

// PaymentMethodSummaryDTO representa resumo por método de pagamento
type PaymentMethodSummaryDTO struct {
	PaymentMethod string    `json:"payment_method"`
	Count         int       `json:"count"`
	TotalAmount   float64   `json:"total_amount"`
	AverageAmount float64   `json:"average_amount"`
	Percentage    float64   `json:"percentage"`
	LastUsed      time.Time `json:"last_used"`
}

// BulkPaymentDTO representa dados para pagamento em massa
type BulkPaymentDTO struct {
	InvoiceIDs    []int     `json:"invoice_ids" validate:"required,min=1"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	PaymentDate   time.Time `json:"payment_date" validate:"required"`
	Reference     string    `json:"reference,omitempty"`
	Notes         string    `json:"notes,omitempty"`
}

// PaymentScheduleDTO representa agendamento de pagamento
type PaymentScheduleDTO struct {
	InvoiceID     int       `json:"invoice_id" validate:"required"`
	Amount        float64   `json:"amount" validate:"required,gt=0"`
	ScheduleDate  time.Time `json:"schedule_date" validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"required"`
	IsRecurring   bool      `json:"is_recurring"`
	Frequency     string    `json:"frequency,omitempty" validate:"omitempty,oneof=weekly monthly quarterly"`
	Notes         string    `json:"notes,omitempty"`
}

// PaymentReceiptDTO representa dados para recibo de pagamento
type PaymentReceiptDTO struct {
	PaymentID    int      `json:"payment_id" validate:"required"`
	EmailTo      []string `json:"email_to" validate:"required,min=1,dive,email"`
	EmailCC      []string `json:"email_cc,omitempty" validate:"omitempty,dive,email"`
	EmailSubject string   `json:"email_subject,omitempty"`
	EmailBody    string   `json:"email_body,omitempty"`
	AttachPDF    bool     `json:"attach_pdf"`
}

// RefundDTO representa dados para reembolso
type RefundDTO struct {
	PaymentID    int       `json:"payment_id" validate:"required"`
	Amount       float64   `json:"amount" validate:"required,gt=0"`
	RefundDate   time.Time `json:"refund_date" validate:"required"`
	Reason       string    `json:"reason" validate:"required"`
	RefundMethod string    `json:"refund_method,omitempty"`
	Reference    string    `json:"reference,omitempty"`
	Notes        string    `json:"notes,omitempty"`
}

// PaymentHistoryDTO representa histórico de pagamentos
type PaymentHistoryDTO struct {
	InvoiceID       int                  `json:"invoice_id"`
	InvoiceNo       string               `json:"invoice_no"`
	GrandTotal      float64              `json:"grand_total"`
	AmountPaid      float64              `json:"amount_paid"`
	BalanceDue      float64              `json:"balance_due"`
	Payments        []PaymentResponseDTO `json:"payments"`
	LastPaymentDate *time.Time           `json:"last_payment_date,omitempty"`
	Status          string               `json:"status"`
}
