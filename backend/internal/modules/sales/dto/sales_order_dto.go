// Package dto - DTOs para o módulo de pedidos de venda
// Este arquivo contém os DTOs específicos para operações de pedidos de venda,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// SalesOrderCreate DTO para criar novos pedidos de venda.
type SalesOrderCreate struct {
	ContactID       int                    `json:"contact_id" validate:"required"`                     // ID do contato (obrigatório)
	QuotationID     int                    `json:"quotation_id,omitempty"`                             // ID da cotação (opcional)
	ExpectedDate    time.Time              `json:"expected_date" validate:"required,gtfield=time.Now"` // Data prevista (obrigatória)
	Items           []SalesOrderItemCreate `json:"items" validate:"required,dive,min=1"`               // Itens do pedido (obrigatório)
	PaymentTerms    string                 `json:"payment_terms,omitempty"`                            // Condições de pagamento
	ShippingAddress string                 `json:"shipping_address" validate:"required"`               // Endereço de entrega (obrigatório)
	Notes           string                 `json:"notes,omitempty"`                                    // Observações
}

// SalesOrderUpdate DTO para atualizar pedidos de venda existentes.
// Campos omitempty permitem atualizações parciais.
type SalesOrderUpdate struct {
	ExpectedDate    time.Time              `json:"expected_date" validate:"omitempty,gtfield=time.Now"` // Nova data prevista
	Items           []SalesOrderItemCreate `json:"items" validate:"omitempty,dive,min=1"`               // Novos itens (opcional)
	PaymentTerms    string                 `json:"payment_terms,omitempty"`                             // Novas condições de pagamento
	ShippingAddress string                 `json:"shipping_address,omitempty"`                          // Novo endereço de entrega
	Notes           string                 `json:"notes,omitempty"`                                     // Novas observações
}

// SalesOrderResponse DTO para retornar dados completos de pedidos de venda.
// Inclui informações calculadas que são somente leitura.
type SalesOrderResponse struct {
	ID              int                      `json:"id"`                      // ID do pedido
	SONo            string                   `json:"so_no"`                   // Número do pedido (somente leitura)
	ContactID       int                      `json:"contact_id"`              // ID do contato
	Contact         ContactResponse          `json:"contact"`                 // Dados do contato
	QuotationID     int                      `json:"quotation_id,omitempty"`  // ID da cotação de origem, se houver
	Quotation       *QuotationShortResponse  `json:"quotation,omitempty"`     // Dados resumidos da cotação
	Status          string                   `json:"status"`                  // Status atual
	CreatedAt       time.Time                `json:"created_at"`              // Data de criação (somente leitura)
	UpdatedAt       time.Time                `json:"updated_at"`              // Data de atualização (somente leitura)
	ExpectedDate    time.Time                `json:"expected_date"`           // Data prevista para entrega
	SubTotal        float64                  `json:"subtotal"`                // Subtotal (somente leitura)
	TaxTotal        float64                  `json:"tax_total"`               // Total de impostos (somente leitura)
	DiscountTotal   float64                  `json:"discount_total"`          // Total de descontos (somente leitura)
	GrandTotal      float64                  `json:"grand_total"`             // Total geral (somente leitura)
	Items           []SalesOrderItemResponse `json:"items"`                   // Itens do pedido
	PaymentTerms    string                   `json:"payment_terms,omitempty"` // Condições de pagamento
	ShippingAddress string                   `json:"shipping_address"`        // Endereço de entrega
	Notes           string                   `json:"notes,omitempty"`         // Observações
}

// SalesOrderShortResponse DTO para respostas resumidas de pedidos de venda.
// Usado em listagens e como relacionamento em outros objetos.
type SalesOrderShortResponse struct {
	ID           int                  `json:"id"`            // ID do pedido
	SONo         string               `json:"so_no"`         // Número do pedido
	ContactID    int                  `json:"contact_id"`    // ID do contato
	Contact      ContactShortResponse `json:"contact"`       // Dados resumidos do contato
	Status       string               `json:"status"`        // Status atual
	ExpectedDate time.Time            `json:"expected_date"` // Data prevista
	GrandTotal   float64              `json:"grand_total"`   // Total geral
	ItemsCount   int                  `json:"items_count"`   // Quantidade de itens
}

// SalesOrderStatusUpdateRequest DTO para solicitações de atualização de status.
type SalesOrderStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=draft confirmed processing completed cancelled"` // Novo status
	Reason string `json:"reason,omitempty"`                                                                // Motivo (opcional)
}

// Usando a estrutura genérica de paginação para respostas paginadas
// PaginatedSalesOrderResponse equivale a PaginatedResponse<SalesOrderResponse>
type PaginatedSalesOrderResponse struct {
	Items      []SalesOrderResponse `json:"items"` // Lista de pedidos
	Pagination                      // Campos de paginação herdados da estrutura base
}

// PaginatedSalesOrderShortResponse equivale a PaginatedResponse<SalesOrderShortResponse>
type PaginatedSalesOrderShortResponse struct {
	Items      []SalesOrderShortResponse `json:"items"` // Lista de pedidos resumidos
	Pagination                           // Campos de paginação herdados da estrutura base
}

// SalesOrderFilter estende o DocumentFilter para filtros específicos de pedidos de venda.
type SalesOrderFilter struct {
	DocumentFilter              // Filtros comuns a documentos
	ExpectedStartDate time.Time `json:"expected_start_date,omitempty"` // Filtro por data prevista (início)
	ExpectedEndDate   time.Time `json:"expected_end_date,omitempty"`   // Filtro por data prevista (fim)
	QuotationID       int       `json:"quotation_id,omitempty"`        // Filtro por cotação de origem
	HasDelivery       *bool     `json:"has_delivery,omitempty"`        // Filtro por pedidos com/sem entrega
	HasInvoice        *bool     `json:"has_invoice,omitempty"`         // Filtro por pedidos com/sem fatura
}
