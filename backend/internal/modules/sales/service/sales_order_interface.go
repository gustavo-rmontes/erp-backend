package service

import (
	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"context"
)

// SalesOrderService define métodos para operações de negócios relacionadas a pedidos de venda
type SalesOrderService interface {
	// Operações CRUD básicas
	Create(ctx context.Context, orderDTO *dto.SalesOrderCreate) (*dto.SalesOrderResponse, error)
	GetByID(ctx context.Context, id int) (*dto.SalesOrderResponse, error)
	GetShortByID(ctx context.Context, id int) (*dto.SalesOrderShortResponse, error)
	Update(ctx context.Context, id int, orderDTO *dto.SalesOrderUpdate) (*dto.SalesOrderResponse, error)
	Delete(ctx context.Context, id int) error

	// Listagem e filtragem
	Find(ctx context.Context, filter *dto.SalesOrderFilter, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderResponse, error)
	FindShort(ctx context.Context, filter *dto.SalesOrderFilter, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderShortResponse, error)
	Search(ctx context.Context, query string, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderShortResponse, error)

	// Gestão de status
	UpdateStatus(ctx context.Context, id int, req *dto.SalesOrderStatusUpdateRequest) (*dto.SalesOrderResponse, error)

	// Gestão de itens
	AddItem(ctx context.Context, id int, item *dto.SalesOrderItemCreate) (*dto.SalesOrderResponse, error)
	UpdateItem(ctx context.Context, orderID int, itemID int, item *dto.SalesOrderItemCreate) (*dto.SalesOrderResponse, error)
	RemoveItem(ctx context.Context, orderID int, itemID int) (*dto.SalesOrderResponse, error)

	// Integração com outros documentos
	CreateFromQuotation(ctx context.Context, quotationID int, data *dto.SalesOrderFromQuotationCreate) (*dto.SalesOrderResponse, error)
	CreatePurchaseOrder(ctx context.Context, id int, data *dto.PurchaseOrderFromSOCreate) (*dto.PurchaseOrderResponse, error)
	CreateInvoice(ctx context.Context, id int, data *dto.InvoiceFromSOCreate) (*dto.InvoiceResponse, error)
	CreateDelivery(ctx context.Context, id int, data *dto.DeliveryFromSOCreate) (*dto.DeliveryResponse, error)
	Clone(ctx context.Context, id int, options *dto.SalesOrderCloneOptions) (*dto.SalesOrderResponse, error)

	// Análise e estatísticas
	GetStats(ctx context.Context, dateRange *dto.DateRange) (*dto.SalesOrderStats, error)

	// Documentação e exportação
	GeneratePDF(ctx context.Context, id int) ([]byte, error)
	ExportToCSV(ctx context.Context, ids []int) ([]byte, error)
}
