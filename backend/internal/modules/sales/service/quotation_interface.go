package service

import (
	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"context"
)

// QuotationService define métodos para operações de negócios relacionadas a cotações
type QuotationService interface {
	// Operações CRUD básicas
	Create(ctx context.Context, quotationDTO *dto.QuotationCreate) (*dto.QuotationResponse, error)
	GetByID(ctx context.Context, id int) (*dto.QuotationResponse, error)
	GetShortByID(ctx context.Context, id int) (*dto.QuotationShortResponse, error)
	Update(ctx context.Context, id int, quotationDTO *dto.QuotationUpdate) (*dto.QuotationResponse, error)
	Delete(ctx context.Context, id int) error

	// Listagem e filtragem
	Find(ctx context.Context, filter *dto.QuotationFilter, params *pagination.PaginationParams) (*dto.PaginatedQuotationResponse, error)
	FindShort(ctx context.Context, filter *dto.QuotationFilter, params *pagination.PaginationParams) (*dto.PaginatedQuotationShortResponse, error)
	Search(ctx context.Context, query string, params *pagination.PaginationParams) (*dto.PaginatedQuotationShortResponse, error)

	// Gestão de status
	UpdateStatus(ctx context.Context, id int, req *dto.QuotationStatusUpdateRequest) (*dto.QuotationResponse, error)

	// Automação
	ProcessExpirations(ctx context.Context) (int, error)
	NotifyExpiringQuotations(ctx context.Context, daysBeforeExpiry int) (int, error)

	// Gestão de itens
	AddItem(ctx context.Context, id int, item *dto.QuotationItemCreate) (*dto.QuotationResponse, error)
	UpdateItem(ctx context.Context, quotationID int, itemID int, item *dto.QuotationItemCreate) (*dto.QuotationResponse, error)
	RemoveItem(ctx context.Context, quotationID int, itemID int) (*dto.QuotationResponse, error)

	// Integração com outros documentos
	ConvertToSalesOrder(ctx context.Context, id int, data *dto.SalesOrderFromQuotationCreate) (*dto.SalesOrderResponse, error)
	Clone(ctx context.Context, id int, options *dto.QuotationCloneOptions) (*dto.QuotationResponse, error)

	// Análise e estatísticas
	GetStats(ctx context.Context, dateRange *dto.DateRange) (*dto.QuotationStats, error)
	GetConversionStats(ctx context.Context, dateRange *dto.DateRange) (*dto.ConversionRateStats, error)

	// Documentação e exportação
	GeneratePDF(ctx context.Context, id int) ([]byte, error)
	ExportToCSV(ctx context.Context, ids []int) ([]byte, error)
	SendByEmail(ctx context.Context, id int, options *dto.EmailOptions) error
}
