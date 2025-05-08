package quotation

import (
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/modules/sales/service"

	"go.uber.org/zap"
)

// Service implementa a interface QuotationService
type Service struct {
	quotationRepo    repository.QuotationRepository
	salesOrderRepo   repository.SalesOrderRepository
	salesProcessRepo repository.SalesProcessRepository
	logger           *zap.Logger
}

// New cria uma nova instância do serviço de cotações
func New(
	quotationRepo repository.QuotationRepository,
	salesOrderRepo repository.SalesOrderRepository,
	salesProcessRepo repository.SalesProcessRepository,
	logger *zap.Logger,
) service.QuotationService {
	return &Service{
		quotationRepo:    quotationRepo,
		salesOrderRepo:   salesOrderRepo,
		salesProcessRepo: salesProcessRepo,
		logger:           logger.With(zap.String("service", "QuotationService")),
	}
}
