// Package dto - módulo de DTOs de integração
// Este arquivo contém todos os DTOs utilizados para integração entre diferentes módulos do sistema,
// como conversão de cotação para pedido de venda, pedido de venda para pedido de compra, etc.
package dto

import "time"

// SalesOrderFromQuotationCreate DTO para criar um pedido de venda a partir de uma cotação.
// Contém apenas os campos adicionais necessários, já que outras informações virão da cotação.
type SalesOrderFromQuotationCreate struct {
	ExpectedDate    time.Time `json:"expected_date" validate:"required,gtfield=time.Now"` // Data prevista para entrega
	PaymentTerms    string    `json:"payment_terms,omitempty"`                            // Condições de pagamento
	ShippingAddress string    `json:"shipping_address" validate:"required"`               // Endereço de entrega
	Notes           string    `json:"notes,omitempty"`                                    // Observações adicionais
	CopyAllItems    bool      `json:"copy_all_items"`                                     // Copiar todos os itens (true) ou selecionar (false)
	ItemIDs         []int     `json:"item_ids,omitempty" validate:"omitempty,min=1"`      // IDs dos itens da cotação a serem copiados
}

// PurchaseOrderFromSOCreate DTO para criar um pedido de compra a partir de um pedido de venda.
type PurchaseOrderFromSOCreate struct {
	ContactID       int       `json:"contact_id" validate:"required"`                     // ID do fornecedor
	ExpectedDate    time.Time `json:"expected_date" validate:"required,gtfield=time.Now"` // Data prevista de recebimento
	PaymentTerms    string    `json:"payment_terms,omitempty"`                            // Condições de pagamento
	ShippingAddress string    `json:"shipping_address" validate:"required"`               // Endereço de entrega
	Notes           string    `json:"notes,omitempty"`                                    // Observações adicionais
	CopyAllItems    bool      `json:"copy_all_items"`                                     // Copiar todos os itens (true) ou selecionar (false)
	ItemIDs         []int     `json:"item_ids,omitempty" validate:"omitempty,min=1"`      // IDs dos itens do pedido de venda a serem copiados
}

// InvoiceFromSOCreate DTO para criar uma fatura a partir de um pedido de venda.
type InvoiceFromSOCreate struct {
	IssueDate    time.Time `json:"issue_date" validate:"required"`                 // Data de emissão
	DueDate      time.Time `json:"due_date" validate:"required,gtfield=IssueDate"` // Data de vencimento
	PaymentTerms string    `json:"payment_terms,omitempty"`                        // Condições de pagamento
	Notes        string    `json:"notes,omitempty"`                                // Observações adicionais
	CopyAllItems bool      `json:"copy_all_items"`                                 // Copiar todos os itens (true) ou selecionar (false)
	ItemIDs      []int     `json:"item_ids,omitempty" validate:"omitempty,min=1"`  // IDs dos itens do pedido de venda a serem copiados
}

// DeliveryFromSOCreate DTO para criar uma entrega a partir de um pedido de venda.
type DeliveryFromSOCreate struct {
	DeliveryDate    time.Time   `json:"delivery_date" validate:"required"`             // Data prevista de entrega
	ShippingMethod  string      `json:"shipping_method" validate:"required"`           // Método de envio
	TrackingNumber  string      `json:"tracking_number,omitempty"`                     // Número de rastreamento
	ShippingAddress string      `json:"shipping_address" validate:"required"`          // Endereço de entrega
	Notes           string      `json:"notes,omitempty"`                               // Observações adicionais
	CopyAllItems    bool        `json:"copy_all_items"`                                // Copiar todos os itens (true) ou selecionar (false)
	ItemIDs         []int       `json:"item_ids,omitempty" validate:"omitempty,min=1"` // IDs dos itens do pedido de venda a serem copiados
	Quantities      map[int]int `json:"quantities,omitempty"`                          // Mapa de ID do item -> quantidade a ser entregue
}

// DeliveryFromPOCreate DTO para criar uma entrega a partir de um pedido de compra.
type DeliveryFromPOCreate struct {
	DeliveryDate   time.Time   `json:"delivery_date" validate:"required"`             // Data de entrega
	ReceivedDate   time.Time   `json:"received_date,omitempty"`                       // Data de recebimento
	ShippingMethod string      `json:"shipping_method" validate:"required"`           // Método de envio
	TrackingNumber string      `json:"tracking_number,omitempty"`                     // Número de rastreamento
	Notes          string      `json:"notes,omitempty"`                               // Observações adicionais
	CopyAllItems   bool        `json:"copy_all_items"`                                // Copiar todos os itens (true) ou selecionar (false)
	ItemIDs        []int       `json:"item_ids,omitempty" validate:"omitempty,min=1"` // IDs dos itens do pedido de compra a serem copiados
	Quantities     map[int]int `json:"quantities,omitempty"`                          // Mapa de ID do item -> quantidade a ser entregue
}

// QuotationCloneOptions define as opções para clonagem de cotação.
type QuotationCloneOptions struct {
	ContactID       int       `json:"contact_id,omitempty"`                                        // Se diferente, usar novo contato
	ExpiryDate      time.Time `json:"expiry_date,omitempty" validate:"omitempty,gtfield=time.Now"` // Nova data de expiração
	AdjustPrices    bool      `json:"adjust_prices"`                                               // Indica se deve atualizar preços dos produtos
	PriceAdjustment float64   `json:"price_adjustment,omitempty" validate:"omitempty"`             // Percentual de ajuste
	CopyNotes       bool      `json:"copy_notes"`                                                  // Copiar notas ou deixar em branco
	CopyTerms       bool      `json:"copy_terms"`                                                  // Copiar termos ou deixar em branco
}

// SalesOrderCloneOptions define as opções para clonagem de pedido de venda.
type SalesOrderCloneOptions struct {
	ContactID       int       `json:"contact_id,omitempty"`                                          // Se diferente, usar novo contato
	ExpectedDate    time.Time `json:"expected_date,omitempty" validate:"omitempty,gtfield=time.Now"` // Nova data prevista
	AdjustPrices    bool      `json:"adjust_prices"`                                                 // Indica se deve atualizar preços dos produtos
	PriceAdjustment float64   `json:"price_adjustment,omitempty" validate:"omitempty"`               // Percentual de ajuste
	CopyNotes       bool      `json:"copy_notes"`                                                    // Copiar notas ou deixar em branco
	CopyItems       bool      `json:"copy_items"`                                                    // Copiar todos os itens
	ItemIDs         []int     `json:"item_ids,omitempty" validate:"omitempty,min=1"`                 // IDs dos itens a serem copiados se CopyItems=false
}

// SalesProcessDocumentsLinkCreate DTO para vincular documentos a um processo de vendas.
type SalesProcessDocumentsLinkCreate struct {
	QuotationID     int   `json:"quotation_id,omitempty"`      // ID da cotação
	SalesOrderID    int   `json:"sales_order_id,omitempty"`    // ID do pedido de venda
	PurchaseOrderID int   `json:"purchase_order_id,omitempty"` // ID do pedido de compra
	InvoiceIDs      []int `json:"invoice_ids,omitempty"`       // IDs das faturas
	DeliveryIDs     []int `json:"delivery_ids,omitempty"`      // IDs das entregas
}
