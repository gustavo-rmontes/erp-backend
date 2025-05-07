// Package dto - DTOs para o módulo de entregas
// Este arquivo contém os DTOs específicos para operações de entregas,
// seguindo o padrão CRUD (Create/Read/Update/Delete).
package dto

import "time"

// DeliveryCreate DTO para criar novas entregas.
type DeliveryCreate struct {
	SalesOrderID    int                  `json:"sales_order_id,omitempty"`             // ID do pedido de venda (opcional)
	PurchaseOrderID int                  `json:"purchase_order_id,omitempty"`          // ID do pedido de compra (opcional)
	DeliveryDate    time.Time            `json:"delivery_date" validate:"required"`    // Data prevista (obrigatória)
	ReceivedDate    time.Time            `json:"received_date,omitempty"`              // Data de recebimento (opcional)
	ShippingMethod  string               `json:"shipping_method" validate:"required"`  // Método de envio (obrigatório)
	TrackingNumber  string               `json:"tracking_number,omitempty"`            // Número de rastreamento
	ShippingAddress string               `json:"shipping_address" validate:"required"` // Endereço de entrega (obrigatório)
	Items           []DeliveryItemCreate `json:"items" validate:"required,dive,min=1"` // Itens da entrega (obrigatório)
	Notes           string               `json:"notes,omitempty"`                      // Observações
}

// DeliveryUpdate DTO para atualizar entregas existentes.
// Campos omitempty permitem atualizações parciais.
type DeliveryUpdate struct {
	DeliveryDate    time.Time            `json:"delivery_date,omitempty"`                         // Nova data prevista
	ReceivedDate    time.Time            `json:"received_date,omitempty"`                         // Nova data de recebimento
	ShippingMethod  string               `json:"shipping_method,omitempty"`                       // Novo método de envio
	TrackingNumber  string               `json:"tracking_number,omitempty"`                       // Novo número de rastreamento
	ShippingAddress string               `json:"shipping_address,omitempty"`                      // Novo endereço de entrega
	Items           []DeliveryItemCreate `json:"items,omitempty" validate:"omitempty,dive,min=1"` // Novos itens (opcional)
	Notes           string               `json:"notes,omitempty"`                                 // Novas observações
}

// DeliveryStatusUpdateRequest DTO para atualização de status de entrega.
// Define validação específica para os status possíveis.
type DeliveryStatusUpdateRequest struct {
	Status       string    `json:"status" validate:"required,oneof=pending shipped delivered returned"` // Novo status
	ReceivedDate time.Time `json:"received_date,omitempty" validate:"required_if=Status delivered"`     // Data de recebimento (obrigatória se status=delivered)
	Reason       string    `json:"reason,omitempty" validate:"required_if=Status returned"`             // Motivo (obrigatório se status=returned)
}

// DeliveryResponse DTO para retornar dados completos de entregas.
type DeliveryResponse struct {
	ID              int                         `json:"id"`                          // ID da entrega
	DeliveryNo      string                      `json:"delivery_no"`                 // Número da entrega (somente leitura)
	SalesOrderID    int                         `json:"sales_order_id,omitempty"`    // ID do pedido de venda
	SalesOrder      *SalesOrderShortResponse    `json:"sales_order,omitempty"`       // Dados resumidos do pedido de venda
	PurchaseOrderID int                         `json:"purchase_order_id,omitempty"` // ID do pedido de compra
	PurchaseOrder   *PurchaseOrderShortResponse `json:"purchase_order,omitempty"`    // Dados resumidos do pedido de compra
	Status          string                      `json:"status"`                      // Status atual
	CreatedAt       time.Time                   `json:"created_at"`                  // Data de criação (somente leitura)
	UpdatedAt       time.Time                   `json:"updated_at"`                  // Data de atualização (somente leitura)
	DeliveryDate    time.Time                   `json:"delivery_date"`               // Data prevista
	ReceivedDate    time.Time                   `json:"received_date,omitempty"`     // Data de recebimento
	ShippingMethod  string                      `json:"shipping_method"`             // Método de envio
	TrackingNumber  string                      `json:"tracking_number,omitempty"`   // Número de rastreamento
	ShippingAddress string                      `json:"shipping_address"`            // Endereço de entrega
	Items           []DeliveryItemResponse      `json:"items"`                       // Itens da entrega
	Notes           string                      `json:"notes,omitempty"`             // Observações
}

// DeliveryShortResponse DTO para respostas resumidas de entregas.
// Usado em listagens e como relacionamento em outros objetos.
type DeliveryShortResponse struct {
	ID             int       `json:"id"`                        // ID da entrega
	DeliveryNo     string    `json:"delivery_no"`               // Número da entrega
	Status         string    `json:"status"`                    // Status atual
	DeliveryDate   time.Time `json:"delivery_date"`             // Data prevista
	ReceivedDate   time.Time `json:"received_date,omitempty"`   // Data de recebimento
	ShippingMethod string    `json:"shipping_method"`           // Método de envio
	TrackingNumber string    `json:"tracking_number,omitempty"` // Número de rastreamento
	ItemsCount     int       `json:"items_count"`               // Quantidade de itens
	TotalQuantity  int       `json:"total_quantity"`            // Quantidade total
}

// DeliveryReceiptUpdate DTO para atualizar o recebimento de uma entrega.
type DeliveryReceiptUpdate struct {
	ReceivedDate time.Time                   `json:"received_date" validate:"required"`    // Data de recebimento (obrigatória)
	Items        []DeliveryItemReceiptUpdate `json:"items" validate:"required,dive,min=1"` // Itens recebidos (obrigatório)
	Notes        string                      `json:"notes,omitempty"`                      // Observações
}

// DeliveryScheduleResponse DTO para respostas de agenda de entregas.
type DeliveryScheduleResponse struct {
	DueToday          []DeliveryShortResponse `json:"due_today"`          // Entregas para hoje
	Upcoming          []DeliveryShortResponse `json:"upcoming"`           // Próximas entregas
	Overdue           []DeliveryShortResponse `json:"overdue"`            // Entregas atrasadas
	RecentlyDelivered []DeliveryShortResponse `json:"recently_delivered"` // Entregas recentes
}

// DeliveryTrackingUpdateRequest DTO para atualizar informações de rastreamento.
type DeliveryTrackingUpdateRequest struct {
	TrackingNumber    string    `json:"tracking_number" validate:"required"` // Número de rastreamento (obrigatório)
	ShippingMethod    string    `json:"shipping_method,omitempty"`           // Método de envio
	TrackingURL       string    `json:"tracking_url,omitempty"`              // URL de rastreamento
	EstimatedDelivery time.Time `json:"estimated_delivery,omitempty"`        // Data estimada de entrega
}

// PaginatedDeliveryResponse DTO para respostas paginadas.
// Usa a estrutura base Pagination.
type PaginatedDeliveryResponse struct {
	Items      []DeliveryResponse `json:"items"` // Lista de entregas
	Pagination                    // Campos de paginação herdados da estrutura base
}

// PaginatedDeliveryShortResponse DTO para respostas paginadas resumidas.
// Usa a estrutura base Pagination.
type PaginatedDeliveryShortResponse struct {
	Items      []DeliveryShortResponse `json:"items"` // Lista de entregas resumidas
	Pagination                         // Campos de paginação herdados da estrutura base
}

// DeliveryFilter estende o DocumentFilter para filtros específicos de entregas.
type DeliveryFilter struct {
	DocumentFilter              // Filtros comuns a documentos
	DeliveryStartDate time.Time `json:"delivery_start_date,omitempty" form:"deliveryStartDate"` // Filtro por data prevista (início)
	DeliveryEndDate   time.Time `json:"delivery_end_date,omitempty" form:"deliveryEndDate"`     // Filtro por data prevista (fim)
	ReceivedStartDate time.Time `json:"received_start_date,omitempty" form:"receivedStartDate"` // Filtro por data de recebimento (início)
	ReceivedEndDate   time.Time `json:"received_end_date,omitempty" form:"receivedEndDate"`     // Filtro por data de recebimento (fim)
	SalesOrderID      int       `json:"sales_order_id,omitempty" form:"salesOrderId"`           // Filtro por ID de pedido de venda
	PurchaseOrderID   int       `json:"purchase_order_id,omitempty" form:"purchaseOrderId"`     // Filtro por ID de pedido de compra
	ShippingMethod    string    `json:"shipping_method,omitempty" form:"shippingMethod"`        // Filtro por método de envio
	TrackingNumber    string    `json:"tracking_number,omitempty" form:"trackingNumber"`        // Filtro por número de rastreamento
	IsReceived        *bool     `json:"is_received,omitempty" form:"isReceived"`                // Filtro por recebido/não recebido
}
