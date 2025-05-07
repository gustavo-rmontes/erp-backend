// Package dto - DTOs para o módulo de pedidos de compra
// Este arquivo contém os DTOs específicos para operações de pedidos de compra,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// PurchaseOrderCreate DTO para criar novos pedidos de compra.
type PurchaseOrderCreate struct {
	ContactID       int                       `json:"contact_id" validate:"required"`                     // ID do fornecedor (obrigatório)
	SalesOrderID    int                       `json:"sales_order_id,omitempty"`                           // ID do pedido de venda relacionado (opcional)
	ExpectedDate    time.Time                 `json:"expected_date" validate:"required,gtfield=time.Now"` // Data prevista (obrigatória)
	Items           []PurchaseOrderItemCreate `json:"items" validate:"required,dive,min=1"`               // Itens do pedido (obrigatório)
	PaymentTerms    string                    `json:"payment_terms,omitempty"`                            // Condições de pagamento
	ShippingAddress string                    `json:"shipping_address" validate:"required"`               // Endereço de entrega (obrigatório)
	Notes           string                    `json:"notes,omitempty"`                                    // Observações
}

// PurchaseOrderUpdate DTO para atualizar pedidos de compra existentes.
// Campos omitempty permitem atualizações parciais.
type PurchaseOrderUpdate struct {
	ExpectedDate    time.Time                 `json:"expected_date" validate:"omitempty,gtfield=time.Now"` // Nova data prevista
	Items           []PurchaseOrderItemCreate `json:"items" validate:"omitempty,dive,min=1"`               // Novos itens (opcional)
	PaymentTerms    string                    `json:"payment_terms,omitempty"`                             // Novas condições de pagamento
	ShippingAddress string                    `json:"shipping_address,omitempty"`                          // Novo endereço de entrega
	Notes           string                    `json:"notes,omitempty"`                                     // Novas observações
}

// PurchaseOrderStatusUpdateRequest DTO para atualização de status de pedido de compra.
// Estende a estrutura base StatusUpdateRequest e adiciona validação específica.
type PurchaseOrderStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=draft sent confirmed received cancelled"` // Novo status
	Reason string `json:"reason,omitempty"`                                                         // Motivo (opcional)
}

// PurchaseOrderResponse DTO para retornar dados completos de pedidos de compra.
// Inclui informações calculadas que são somente leitura.
type PurchaseOrderResponse struct {
	ID              int                         `json:"id"`                       // ID do pedido
	PONo            string                      `json:"po_no"`                    // Número do pedido (somente leitura)
	ContactID       int                         `json:"contact_id"`               // ID do fornecedor
	Contact         ContactResponse             `json:"contact"`                  // Dados do fornecedor
	SalesOrderID    int                         `json:"sales_order_id,omitempty"` // ID do pedido de venda relacionado
	SalesOrder      *SalesOrderShortResponse    `json:"sales_order,omitempty"`    // Dados resumidos do pedido de venda
	Status          string                      `json:"status"`                   // Status atual
	CreatedAt       time.Time                   `json:"created_at"`               // Data de criação (somente leitura)
	UpdatedAt       time.Time                   `json:"updated_at"`               // Data de atualização (somente leitura)
	ExpectedDate    time.Time                   `json:"expected_date"`            // Data prevista para recebimento
	SubTotal        float64                     `json:"subtotal"`                 // Subtotal (somente leitura)
	TaxTotal        float64                     `json:"tax_total"`                // Total de impostos (somente leitura)
	DiscountTotal   float64                     `json:"discount_total"`           // Total de descontos (somente leitura)
	GrandTotal      float64                     `json:"grand_total"`              // Total geral (somente leitura)
	Items           []PurchaseOrderItemResponse `json:"items"`                    // Itens do pedido
	PaymentTerms    string                      `json:"payment_terms,omitempty"`  // Condições de pagamento
	ShippingAddress string                      `json:"shipping_address"`         // Endereço de entrega
	Notes           string                      `json:"notes,omitempty"`          // Observações
}

// PurchaseOrderShortResponse DTO para respostas resumidas de pedidos de compra.
// Usado em listagens e como relacionamento em outros objetos.
type PurchaseOrderShortResponse struct {
	ID           int                  `json:"id"`            // ID do pedido
	PONo         string               `json:"po_no"`         // Número do pedido
	ContactID    int                  `json:"contact_id"`    // ID do fornecedor
	Contact      ContactShortResponse `json:"contact"`       // Dados resumidos do fornecedor
	Status       string               `json:"status"`        // Status atual
	ExpectedDate time.Time            `json:"expected_date"` // Data prevista
	GrandTotal   float64              `json:"grand_total"`   // Total geral
	ItemsCount   int                  `json:"items_count"`   // Quantidade de itens
}

// PaginatedPurchaseOrderResponse DTO para respostas paginadas.
// Usa a estrutura base Pagination.
type PaginatedPurchaseOrderResponse struct {
	Items      []PurchaseOrderResponse `json:"items"` // Lista de pedidos de compra
	Pagination                         // Campos de paginação herdados da estrutura base
}

// PaginatedPurchaseOrderShortResponse DTO para respostas paginadas resumidas.
// Usa a estrutura base Pagination.
type PaginatedPurchaseOrderShortResponse struct {
	Items      []PurchaseOrderShortResponse `json:"items"` // Lista de pedidos de compra resumidos
	Pagination                              // Campos de paginação herdados da estrutura base
}

// PurchaseOrderFilter estende o DocumentFilter para filtros específicos de pedidos de compra.
type PurchaseOrderFilter struct {
	DocumentFilter              // Filtros comuns a documentos
	ExpectedStartDate time.Time `json:"expected_start_date,omitempty" form:"expectedStartDate"` // Filtro por data prevista (início)
	ExpectedEndDate   time.Time `json:"expected_end_date,omitempty" form:"expectedEndDate"`     // Filtro por data prevista (fim)
	SalesOrderID      int       `json:"sales_order_id,omitempty" form:"salesOrderId"`           // Filtro por ID de pedido de venda
	HasDelivery       *bool     `json:"has_delivery,omitempty" form:"hasDelivery"`              // Filtro por pedidos com/sem entrega
}
