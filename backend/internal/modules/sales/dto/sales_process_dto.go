// Package dto - DTOs para o módulo de processo de vendas
// Este arquivo contém os DTOs específicos para operações de processos de vendas,
// que representam o fluxo completo de vendas desde a cotação até o pagamento.
package dto

import "time"

// SalesProcessCreate DTO para criar um novo processo de vendas.
type SalesProcessCreate struct {
	ContactID int    `json:"contact_id" validate:"required"` // ID do contato (obrigatório)
	Status    string `json:"status" validate:"required"`     // Status inicial (obrigatório)
	Notes     string `json:"notes,omitempty"`                // Observações opcionais
}

// SalesProcessUpdate DTO para atualizar processos de vendas existentes.
// Campos omitempty permitem atualizações parciais.
type SalesProcessUpdate struct {
	Status     string  `json:"status,omitempty"`      // Novo status
	Notes      string  `json:"notes,omitempty"`       // Novas observações
	TotalValue float64 `json:"total_value,omitempty"` // Novo valor total
	Profit     float64 `json:"profit,omitempty"`      // Novo valor de lucro
}

// SalesProcessStatusUpdateRequest DTO para atualização de status de processo de vendas.
// Estende a estrutura base StatusUpdateRequest.
type SalesProcessStatusUpdateRequest struct {
	StatusUpdateRequest // Campos comuns herdados de StatusUpdateRequest
}

// SalesProcessResponse DTO para retornar dados completos de processos de vendas.
type SalesProcessResponse struct {
	ID            int                    `json:"id"`                       // ID do processo
	ContactID     int                    `json:"contact_id"`               // ID do contato
	Contact       ContactResponse        `json:"contact"`                  // Dados do contato
	Status        string                 `json:"status"`                   // Status atual
	CreatedAt     time.Time              `json:"created_at"`               // Data de criação (somente leitura)
	UpdatedAt     time.Time              `json:"updated_at"`               // Data de atualização (somente leitura)
	TotalValue    float64                `json:"total_value"`              // Valor total
	Profit        float64                `json:"profit"`                   // Lucro
	Notes         string                 `json:"notes,omitempty"`          // Observações
	Quotation     *QuotationResponse     `json:"quotation,omitempty"`      // Cotação relacionada
	SalesOrder    *SalesOrderResponse    `json:"sales_order,omitempty"`    // Pedido de venda relacionado
	PurchaseOrder *PurchaseOrderResponse `json:"purchase_order,omitempty"` // Pedido de compra relacionado
	Deliveries    []DeliveryResponse     `json:"deliveries,omitempty"`     // Entregas relacionadas
	Invoices      []InvoiceResponse      `json:"invoices,omitempty"`       // Faturas relacionadas
}

// SalesProcessShortResponse DTO para respostas resumidas de processos de vendas.
// Usado em listagens e como relacionamento em outros objetos.
type SalesProcessShortResponse struct {
	ID         int                  `json:"id"`          // ID do processo
	ContactID  int                  `json:"contact_id"`  // ID do contato
	Contact    ContactShortResponse `json:"contact"`     // Dados resumidos do contato
	Status     string               `json:"status"`      // Status atual
	CreatedAt  time.Time            `json:"created_at"`  // Data de criação
	TotalValue float64              `json:"total_value"` // Valor total
	Profit     float64              `json:"profit"`      // Lucro
}

// SalesProcessWithDocumentsResponse DTO para processos de vendas com contagens de documentos.
// Usado para visões gerais sem necessidade de carregar todos os documentos relacionados.
type SalesProcessWithDocumentsResponse struct {
	ID               int                  `json:"id"`                 // ID do processo
	ContactID        int                  `json:"contact_id"`         // ID do contato
	Contact          ContactShortResponse `json:"contact"`            // Dados resumidos do contato
	Status           string               `json:"status"`             // Status atual
	CreatedAt        time.Time            `json:"created_at"`         // Data de criação
	UpdatedAt        time.Time            `json:"updated_at"`         // Data de atualização
	TotalValue       float64              `json:"total_value"`        // Valor total
	Profit           float64              `json:"profit"`             // Lucro
	Notes            string               `json:"notes,omitempty"`    // Observações
	HasQuotation     bool                 `json:"has_quotation"`      // Indica se possui cotação
	HasSalesOrder    bool                 `json:"has_sales_order"`    // Indica se possui pedido de venda
	HasPurchaseOrder bool                 `json:"has_purchase_order"` // Indica se possui pedido de compra
	DeliveriesCount  int                  `json:"deliveries_count"`   // Quantidade de entregas
	InvoicesCount    int                  `json:"invoices_count"`     // Quantidade de faturas
	PaymentsCount    int                  `json:"payments_count"`     // Quantidade de pagamentos
}

// PaginatedSalesProcessResponse DTO para respostas paginadas de processos de vendas.
// Usa a estrutura base Pagination.
type PaginatedSalesProcessResponse struct {
	Items      []SalesProcessResponse `json:"items"` // Lista de processos de vendas
	Pagination                        // Campos de paginação herdados da estrutura base
}

// PaginatedSalesProcessShortResponse DTO para respostas paginadas resumidas.
// Usa a estrutura base Pagination.
type PaginatedSalesProcessShortResponse struct {
	Items      []SalesProcessShortResponse `json:"items"` // Lista de processos de vendas resumidos
	Pagination                             // Campos de paginação herdados da estrutura base
}

// SalesProcessFilter estende o DocumentFilter para filtros específicos de processos de vendas.
type SalesProcessFilter struct {
	DocumentFilter         // Filtros comuns a documentos
	MinProfit      float64 `json:"min_profit,omitempty" form:"minProfit"`          // Lucro mínimo
	MaxProfit      float64 `json:"max_profit,omitempty" form:"maxProfit"`          // Lucro máximo
	HasQuotation   *bool   `json:"has_quotation,omitempty" form:"hasQuotation"`    // Filtrar processos com/sem cotação
	HasSalesOrder  *bool   `json:"has_sales_order,omitempty" form:"hasSalesOrder"` // Filtrar processos com/sem pedido de venda
	HasInvoice     *bool   `json:"has_invoice,omitempty" form:"hasInvoice"`        // Filtrar processos com/sem fatura
}
