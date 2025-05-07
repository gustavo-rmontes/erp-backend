// Package dto - DTOs para o módulo de pagamentos
// Este arquivo contém os DTOs específicos para operações de pagamentos,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// PaymentCreate DTO para criar novos pagamentos.
type PaymentCreate struct {
	InvoiceID     int       `json:"invoice_id" validate:"required"`     // ID da fatura (obrigatório)
	Amount        float64   `json:"amount" validate:"required,gt=0"`    // Valor do pagamento (obrigatório)
	PaymentDate   time.Time `json:"payment_date" validate:"required"`   // Data do pagamento (obrigatória)
	PaymentMethod string    `json:"payment_method" validate:"required"` // Método de pagamento (obrigatório)
	Reference     string    `json:"reference,omitempty"`                // Referência/comprovante
	Notes         string    `json:"notes,omitempty"`                    // Observações
}

// PaymentUpdate DTO para atualizar pagamentos existentes.
// Campos omitempty permitem atualizações parciais.
type PaymentUpdate struct {
	Amount        float64   `json:"amount" validate:"omitempty,gt=0"` // Novo valor
	PaymentDate   time.Time `json:"payment_date,omitempty"`           // Nova data
	PaymentMethod string    `json:"payment_method,omitempty"`         // Novo método
	Reference     string    `json:"reference,omitempty"`              // Nova referência
	Notes         string    `json:"notes,omitempty"`                  // Novas observações
}

// PaymentResponse DTO para retornar dados completos de pagamentos.
type PaymentResponse struct {
	ID            int                   `json:"id"`                  // ID do pagamento
	InvoiceID     int                   `json:"invoice_id"`          // ID da fatura
	Invoice       *InvoiceShortResponse `json:"invoice,omitempty"`   // Dados resumidos da fatura
	Amount        float64               `json:"amount"`              // Valor do pagamento
	PaymentDate   time.Time             `json:"payment_date"`        // Data do pagamento
	PaymentMethod string                `json:"payment_method"`      // Método de pagamento
	Reference     string                `json:"reference,omitempty"` // Referência/comprovante
	Notes         string                `json:"notes,omitempty"`     // Observações
}

// PaginatedPaymentResponse DTO para respostas paginadas.
// Usa a estrutura base Pagination.
type PaginatedPaymentResponse struct {
	Items      []PaymentResponse `json:"items"` // Lista de pagamentos
	Pagination                   // Campos de paginação herdados da estrutura base
}

// BulkPaymentCreate DTO para criar múltiplos pagamentos de uma vez.
type BulkPaymentCreate struct {
	InvoiceIDs    []int     `json:"invoice_ids" validate:"required,min=1"`                  // IDs das faturas (obrigatório)
	PaymentMethod string    `json:"payment_method" validate:"required"`                     // Método de pagamento (obrigatório)
	PaymentDate   time.Time `json:"payment_date" validate:"required"`                       // Data do pagamento (obrigatória)
	IsFullPayment bool      `json:"is_full_payment"`                                        // Pagar valor total (se true)
	Amount        float64   `json:"amount" validate:"required_if=IsFullPayment false,gt=0"` // Valor por fatura (se IsFullPayment=false)
	Reference     string    `json:"reference,omitempty"`                                    // Referência/comprovante
	Notes         string    `json:"notes,omitempty"`                                        // Observações
}

// PaymentReceiptRequest DTO para solicitação de recibos de pagamento.
type PaymentReceiptRequest struct {
	PaymentIDs    []int         `json:"payment_ids" validate:"required,min=1"`                              // IDs dos pagamentos (obrigatório)
	ReceiptFormat string        `json:"receipt_format" validate:"required,oneof=pdf html email"`            // Formato do recibo (obrigatório)
	EmailOptions  *EmailOptions `json:"email_options,omitempty" validate:"required_if=ReceiptFormat email"` // Opções de e-mail (obrigatório se format=email)
}

// PaymentFilter estende o DocumentFilter para filtros específicos de pagamentos.
type PaymentFilter struct {
	StartDate     time.Time `json:"start_date,omitempty" form:"startDate"`         // Data inicial
	EndDate       time.Time `json:"end_date,omitempty" form:"endDate"`             // Data final
	InvoiceID     int       `json:"invoice_id,omitempty" form:"invoiceId"`         // Filtro por ID de fatura
	ContactID     int       `json:"contact_id,omitempty" form:"contactId"`         // Filtro por ID de contato
	MinValue      float64   `json:"min_value,omitempty" form:"minValue"`           // Valor mínimo
	MaxValue      float64   `json:"max_value,omitempty" form:"maxValue"`           // Valor máximo
	PaymentMethod string    `json:"payment_method,omitempty" form:"paymentMethod"` // Método de pagamento
	Reference     string    `json:"reference,omitempty" form:"reference"`          // Referência
}
