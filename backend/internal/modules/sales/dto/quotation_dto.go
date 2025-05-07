// Package dto - DTOs para o módulo de cotações
// Este arquivo contém os DTOs específicos para operações de cotações,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// QuotationCreate DTO para criar novas cotações.
type QuotationCreate struct {
	ContactID  int                   `json:"contact_id" validate:"required"`                   // ID do contato (obrigatório)
	ExpiryDate time.Time             `json:"expiry_date" validate:"required,gtfield=time.Now"` // Data de expiração (obrigatória)
	Items      []QuotationItemCreate `json:"items" validate:"required,dive,min=1"`             // Itens da cotação (obrigatório)
	Notes      string                `json:"notes,omitempty"`                                  // Observações opcionais
	Terms      string                `json:"terms,omitempty"`                                  // Termos e condições opcionais
}

// QuotationUpdate DTO para atualizar cotações existentes.
// Campos omitempty permitem atualizações parciais.
type QuotationUpdate struct {
	ExpiryDate time.Time             `json:"expiry_date" validate:"omitempty,gtfield=time.Now"` // Nova data de expiração
	Items      []QuotationItemCreate `json:"items" validate:"omitempty,dive,min=1"`             // Novos itens (opcional)
	Notes      string                `json:"notes,omitempty"`                                   // Novas observações
	Terms      string                `json:"terms,omitempty"`                                   // Novos termos e condições
}

// QuotationResponse DTO para retornar dados completos de cotações.
// Inclui informações calculadas como subtotal e impostos que são somente leitura.
type QuotationResponse struct {
	ID            int                     `json:"id"`              // ID da cotação
	QuotationNo   string                  `json:"quotation_no"`    // Número da cotação (somente leitura)
	ContactID     int                     `json:"contact_id"`      // ID do contato
	Contact       ContactResponse         `json:"contact"`         // Dados do contato
	Status        string                  `json:"status"`          // Status atual
	CreatedAt     time.Time               `json:"created_at"`      // Data de criação (somente leitura)
	UpdatedAt     time.Time               `json:"updated_at"`      // Data de atualização (somente leitura)
	ExpiryDate    time.Time               `json:"expiry_date"`     // Data de expiração
	SubTotal      float64                 `json:"subtotal"`        // Subtotal (somente leitura)
	TaxTotal      float64                 `json:"tax_total"`       // Total de impostos (somente leitura)
	DiscountTotal float64                 `json:"discount_total"`  // Total de descontos (somente leitura)
	GrandTotal    float64                 `json:"grand_total"`     // Total geral (somente leitura)
	Items         []QuotationItemResponse `json:"items"`           // Itens da cotação
	Notes         string                  `json:"notes,omitempty"` // Observações
	Terms         string                  `json:"terms,omitempty"` // Termos e condições
}

// QuotationShortResponse DTO para respostas resumidas de cotações.
// Usado em listagens e como relacionamento em outros objetos.
type QuotationShortResponse struct {
	ID          int                  `json:"id"`           // ID da cotação
	QuotationNo string               `json:"quotation_no"` // Número da cotação
	ContactID   int                  `json:"contact_id"`   // ID do contato
	Contact     ContactShortResponse `json:"contact"`      // Dados resumidos do contato
	Status      string               `json:"status"`       // Status atual
	ExpiryDate  time.Time            `json:"expiry_date"`  // Data de expiração
	GrandTotal  float64              `json:"grand_total"`  // Total geral
	ItemsCount  int                  `json:"items_count"`  // Quantidade de itens
}

// QuotationStatusUpdateRequest DTO para solicitações de atualização de status.
type QuotationStatusUpdateRequest struct {
	Status string `json:"status" validate:"required,oneof=draft sent accepted rejected expired cancelled"` // Novo status
	Reason string `json:"reason,omitempty"`                                                                // Motivo (opcional)
}

// Usando a estrutura genérica de paginação para respostas paginadas
// PaginatedQuotationResponse equivale a PaginatedResponse<QuotationResponse>
type PaginatedQuotationResponse struct {
	Items      []QuotationResponse `json:"items"` // Lista de cotações
	Pagination                     // Campos de paginação herdados da estrutura base
}

// PaginatedQuotationShortResponse equivale a PaginatedResponse<QuotationShortResponse>
type PaginatedQuotationShortResponse struct {
	Items      []QuotationShortResponse `json:"items"` // Lista de cotações resumidas
	Pagination                          // Campos de paginação herdados da estrutura base
}

// QuotationFilter estende o DocumentFilter para filtros específicos de cotações.
type QuotationFilter struct {
	DocumentFilter            // Filtros comuns a documentos
	ExpiryStartDate time.Time `json:"expiry_start_date,omitempty"` // Filtro por data de expiração (início)
	ExpiryEndDate   time.Time `json:"expiry_end_date,omitempty"`   // Filtro por data de expiração (fim)
	IsExpired       *bool     `json:"is_expired,omitempty"`        // Filtro por expirado/não expirado
}
