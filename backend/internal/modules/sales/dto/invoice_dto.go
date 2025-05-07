// Package dto - DTOs para o módulo de faturas
// Este arquivo contém os DTOs específicos para operações de faturas,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// InvoiceCreate DTO para criar novas faturas.
type InvoiceCreate struct {
	ContactID    int                 `json:"contact_id" validate:"required"`                 // ID do cliente (obrigatório)
	SalesOrderID int                 `json:"sales_order_id,omitempty"`                       // ID do pedido de venda relacionado (opcional)
	IssueDate    time.Time           `json:"issue_date" validate:"required"`                 // Data de emissão (obrigatória)
	DueDate      time.Time           `json:"due_date" validate:"required,gtfield=IssueDate"` // Data de vencimento (obrigatória)
	Items        []InvoiceItemCreate `json:"items" validate:"required,dive,min=1"`           // Itens da fatura (obrigatório)
	PaymentTerms string              `json:"payment_terms,omitempty"`                        // Condições de pagamento
	Notes        string              `json:"notes,omitempty"`                                // Observações
}

// InvoiceUpdate DTO para atualizar faturas existentes.
// Campos omitempty permitem atualizações parciais.
type InvoiceUpdate struct {
	DueDate      time.Time           `json:"due_date" validate:"omitempty,gtfield=time.Now"` // Nova data de vencimento
	Items        []InvoiceItemCreate `json:"items" validate:"omitempty,dive,min=1"`          // Novos itens (opcional)
	PaymentTerms string              `json:"payment_terms,omitempty"`                        // Novas condições de pagamento
	Notes        string              `json:"notes,omitempty"`                                // Novas observações
}

// InvoiceStatusUpdateRequest DTO para atualização de status de fatura.
// Define validação específica para os status possíveis.
type InvoiceStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=draft sent partial paid overdue cancelled"` // Novo status
	Reason string `json:"reason,omitempty"`                                                           // Motivo (opcional)
}

// InvoiceResponse DTO para retornar dados completos de faturas.
// Inclui informações calculadas que são somente leitura.
type InvoiceResponse struct {
	ID            int                      `json:"id"`                       // ID da fatura
	InvoiceNo     string                   `json:"invoice_no"`               // Número da fatura (somente leitura)
	ContactID     int                      `json:"contact_id"`               // ID do cliente
	Contact       ContactResponse          `json:"contact"`                  // Dados do cliente
	SalesOrderID  int                      `json:"sales_order_id,omitempty"` // ID do pedido de venda relacionado
	SalesOrder    *SalesOrderShortResponse `json:"sales_order,omitempty"`    // Dados resumidos do pedido de venda
	Status        string                   `json:"status"`                   // Status atual
	CreatedAt     time.Time                `json:"created_at"`               // Data de criação (somente leitura)
	UpdatedAt     time.Time                `json:"updated_at"`               // Data de atualização (somente leitura)
	IssueDate     time.Time                `json:"issue_date"`               // Data de emissão
	DueDate       time.Time                `json:"due_date"`                 // Data de vencimento
	SubTotal      float64                  `json:"subtotal"`                 // Subtotal (somente leitura)
	TaxTotal      float64                  `json:"tax_total"`                // Total de impostos (somente leitura)
	DiscountTotal float64                  `json:"discount_total"`           // Total de descontos (somente leitura)
	GrandTotal    float64                  `json:"grand_total"`              // Total geral (somente leitura)
	AmountPaid    float64                  `json:"amount_paid"`              // Valor pago (somente leitura)
	Balance       float64                  `json:"balance"`                  // Saldo a pagar (somente leitura)
	Items         []InvoiceItemResponse    `json:"items"`                    // Itens da fatura
	Payments      []PaymentResponse        `json:"payments,omitempty"`       // Pagamentos relacionados
	PaymentTerms  string                   `json:"payment_terms,omitempty"`  // Condições de pagamento
	Notes         string                   `json:"notes,omitempty"`          // Observações
}

// InvoiceShortResponse DTO para respostas resumidas de faturas.
// Usado em listagens e como relacionamento em outros objetos.
type InvoiceShortResponse struct {
	ID         int                  `json:"id"`          // ID da fatura
	InvoiceNo  string               `json:"invoice_no"`  // Número da fatura
	ContactID  int                  `json:"contact_id"`  // ID do cliente
	Contact    ContactShortResponse `json:"contact"`     // Dados resumidos do cliente
	Status     string               `json:"status"`      // Status atual
	IssueDate  time.Time            `json:"issue_date"`  // Data de emissão
	DueDate    time.Time            `json:"due_date"`    // Data de vencimento
	GrandTotal float64              `json:"grand_total"` // Total geral
	AmountPaid float64              `json:"amount_paid"` // Valor pago
	Balance    float64              `json:"balance"`     // Saldo a pagar
	ItemsCount int                  `json:"items_count"` // Quantidade de itens
	IsOverdue  bool                 `json:"is_overdue"`  // Indica se está vencida
}

// PaginatedInvoiceResponse DTO para respostas paginadas.
// Usa a estrutura base Pagination.
type PaginatedInvoiceResponse struct {
	Items      []InvoiceResponse `json:"items"` // Lista de faturas
	Pagination                   // Campos de paginação herdados da estrutura base
}

// PaginatedInvoiceShortResponse DTO para respostas paginadas resumidas.
// Usa a estrutura base Pagination.
type PaginatedInvoiceShortResponse struct {
	Items      []InvoiceShortResponse `json:"items"` // Lista de faturas resumidas
	Pagination                        // Campos de paginação herdados da estrutura base
}

// InvoiceFilter estende o DocumentFilter para filtros específicos de faturas.
type InvoiceFilter struct {
	DocumentFilter           // Filtros comuns a documentos
	IssueStartDate time.Time `json:"issue_start_date,omitempty" form:"issueStartDate"` // Filtro por data de emissão (início)
	IssueEndDate   time.Time `json:"issue_end_date,omitempty" form:"issueEndDate"`     // Filtro por data de emissão (fim)
	DueStartDate   time.Time `json:"due_start_date,omitempty" form:"dueStartDate"`     // Filtro por data de vencimento (início)
	DueEndDate     time.Time `json:"due_end_date,omitempty" form:"dueEndDate"`         // Filtro por data de vencimento (fim)
	SalesOrderID   int       `json:"sales_order_id,omitempty" form:"salesOrderId"`     // Filtro por ID de pedido de venda
	IsOverdue      *bool     `json:"is_overdue,omitempty" form:"isOverdue"`            // Filtro por vencidas/não vencidas
	HasPayment     *bool     `json:"has_payment,omitempty" form:"hasPayment"`          // Filtro por faturas com/sem pagamento
	OnlyUnpaid     *bool     `json:"only_unpaid,omitempty" form:"onlyUnpaid"`          // Filtro somente não pagas
	MinBalance     float64   `json:"min_balance,omitempty" form:"minBalance"`          // Saldo mínimo a pagar
}
