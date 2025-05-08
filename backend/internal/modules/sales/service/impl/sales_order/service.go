package sales_order

import (
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/modules/sales/service"

	"go.uber.org/zap"
)

// Service implementa a interface SalesOrderService
type Service struct {
	salesOrderRepo    repository.SalesOrderRepository
	quotationRepo     repository.QuotationRepository
	purchaseOrderRepo repository.PurchaseOrderRepository
	invoiceRepo       repository.InvoiceRepository
	deliveryRepo      repository.DeliveryRepository
	salesProcessRepo  repository.SalesProcessRepository
	logger            *zap.Logger
}

// New cria uma nova instância do serviço de pedidos de venda
func New(
	salesOrderRepo repository.SalesOrderRepository,
	quotationRepo repository.QuotationRepository,
	purchaseOrderRepo repository.PurchaseOrderRepository,
	invoiceRepo repository.InvoiceRepository,
	deliveryRepo repository.DeliveryRepository,
	salesProcessRepo repository.SalesProcessRepository,
	logger *zap.Logger,
) service.SalesOrderService {
	return &Service{
		salesOrderRepo:    salesOrderRepo,
		quotationRepo:     quotationRepo,
		purchaseOrderRepo: purchaseOrderRepo,
		invoiceRepo:       invoiceRepo,
		deliveryRepo:      deliveryRepo,
		salesProcessRepo:  salesProcessRepo,
		logger:            logger.With(zap.String("service", "SalesOrderService")),
	}
}
