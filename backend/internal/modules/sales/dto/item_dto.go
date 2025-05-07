// Package dto - módulo de DTOs para itens
// Este arquivo contém DTOs específicos para os diferentes tipos de itens no sistema,
// utilizando a estrutura base definida em base_dto.go para reduzir duplicação.
package dto

// QuotationItemCreate DTO para criar itens de cotação.
// Estende a estrutura BaseItem.
type QuotationItemCreate struct {
	BaseItem // Campos comuns herdados de BaseItem
}

// SalesOrderItemCreate DTO para criar itens de pedido de venda.
// Estende a estrutura BaseItem.
type SalesOrderItemCreate struct {
	BaseItem // Campos comuns herdados de BaseItem
}

// PurchaseOrderItemCreate DTO para criar itens de pedido de compra.
// Estende a estrutura BaseItem.
type PurchaseOrderItemCreate struct {
	BaseItem // Campos comuns herdados de BaseItem
}

// InvoiceItemCreate DTO para criar itens de fatura.
// Estende a estrutura BaseItem.
type InvoiceItemCreate struct {
	BaseItem // Campos comuns herdados de BaseItem
}

// DeliveryItemCreate DTO para criar itens de entrega.
// Não estende BaseItem pois tem estrutura diferente.
type DeliveryItemCreate struct {
	ProductID   int    `json:"product_id" validate:"required"`    // ID do produto
	ProductName string `json:"product_name" validate:"required"`  // Nome do produto
	Quantity    int    `json:"quantity" validate:"required,gt=0"` // Quantidade a entregar
	Notes       string `json:"notes,omitempty"`                   // Observações sobre o item
}

// QuotationItemResponse DTO para resposta de itens de cotação.
// Estende a estrutura BaseItemResponse.
type QuotationItemResponse struct {
	BaseItemResponse // Campos comuns herdados de BaseItemResponse
}

// SalesOrderItemResponse DTO para resposta de itens de pedido de venda.
// Estende a estrutura BaseItemResponse.
type SalesOrderItemResponse struct {
	BaseItemResponse // Campos comuns herdados de BaseItemResponse
}

// PurchaseOrderItemResponse DTO para resposta de itens de pedido de compra.
// Estende a estrutura BaseItemResponse.
type PurchaseOrderItemResponse struct {
	BaseItemResponse // Campos comuns herdados de BaseItemResponse
}

// InvoiceItemResponse DTO para resposta de itens de fatura.
// Estende a estrutura BaseItemResponse.
type InvoiceItemResponse struct {
	BaseItemResponse // Campos comuns herdados de BaseItemResponse
}

// DeliveryItemResponse DTO para resposta de itens de entrega.
// Não estende BaseItemResponse pois tem estrutura diferente.
type DeliveryItemResponse struct {
	ID          int    `json:"id"`                     // ID do item
	ProductID   int    `json:"product_id"`             // ID do produto
	ProductName string `json:"product_name"`           // Nome do produto
	ProductCode string `json:"product_code,omitempty"` // Código do produto
	Description string `json:"description,omitempty"`  // Descrição
	Quantity    int    `json:"quantity"`               // Quantidade a entregar
	ReceivedQty int    `json:"received_qty"`           // Quantidade recebida
	Notes       string `json:"notes,omitempty"`        // Observações
}

// DeliveryItemReceiptUpdate DTO para atualizar quantidades recebidas em entregas.
type DeliveryItemReceiptUpdate struct {
	DeliveryItemID int    `json:"delivery_item_id" validate:"required"`   // ID do item de entrega
	ReceivedQty    int    `json:"received_qty" validate:"required,gte=0"` // Quantidade recebida
	Notes          string `json:"notes,omitempty"`                        // Observações
}
