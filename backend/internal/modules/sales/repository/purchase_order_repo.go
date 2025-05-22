package repository

import (
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PurchaseOrderRepository define as operações do repositório de purchase orders
type PurchaseOrderRepository interface {
	// CRUD básico
	CreatePurchaseOrder(ctx context.Context, purchaseOrder *models.PurchaseOrder) error
	GetPurchaseOrderByID(ctx context.Context, id int) (*models.PurchaseOrder, error)
	UpdatePurchaseOrder(ctx context.Context, id int, purchaseOrder *models.PurchaseOrder) error
	DeletePurchaseOrder(ctx context.Context, id int) error

	// Consultas com paginação
	GetAllPurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersBySalesOrder(ctx context.Context, salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByPeriod(sctx context.Context, tartDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByExpectedDateRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContactType(ctx context.Context, contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPendingPurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverduePurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada (opcional, considere mover para serviço se contiver muita lógica de negócio)
	SearchPurchaseOrders(ctx context.Context, filter PurchaseOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

type purchaseOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPurchaseOrderRepository cria uma nova instância do repositório
func NewPurchaseOrderRepository(db *gorm.DB, logger *zap.Logger) PurchaseOrderRepository {
	return &purchaseOrderRepository{
		db:     db,
		logger: logger.With(zap.String("module", "purchase_order_repository")),
	}
}

// CreatePurchaseOrder cria um novo purchase order no banco
func (r *purchaseOrderRepository) CreatePurchaseOrder(ctx context.Context, purchaseOrder *models.PurchaseOrder) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "CreatePurchaseOrder"))
			return errors.WrapError(ctx.Err(), "timeout ao criar purchase order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "CreatePurchaseOrder"))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao criar purchase order")
		}
	}

	// Preparação do purchase order
	if purchaseOrder.PONo == "" {
		purchaseOrder.PONo = r.generatePurchaseOrderNumber()
	}

	if purchaseOrder.Status == "" {
		purchaseOrder.Status = models.POStatusDraft
	}

	// Inicia transação com contexto
	tx := r.db.WithContext(ctx).Begin()

	// Verifica novamente o contexto após iniciar transação
	if ctx.Err() != nil {
		tx.Rollback()
		return errors.WrapError(ctx.Err(), "contexto expirou após iniciar transação")
	}

	// Cria o purchase order, omitindo sales_order_id se for 0 (para permitir NULL)
	var err error
	if purchaseOrder.SalesOrderID == 0 {
		err = tx.Omit("sales_order_id").Create(purchaseOrder).Error
	} else {
		err = tx.Create(purchaseOrder).Error
	}

	if err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar purchase order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar purchase order")
	}

	// Se houver itens, cria os itens
	if len(purchaseOrder.Items) > 0 {
		for i := range purchaseOrder.Items {
			// Verifica contexto durante loop para operações longas
			if ctx.Err() != nil {
				tx.Rollback()
				return errors.WrapError(ctx.Err(), "contexto expirou durante criação de itens")
			}

			purchaseOrder.Items[i].PurchaseOrderID = purchaseOrder.ID
			if err := tx.Create(&purchaseOrder.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item do purchase order",
					zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d do purchase order", i))
			}
		}
	}

	// Verificação final do contexto antes do commit
	if ctx.Err() != nil {
		tx.Rollback()
		return errors.WrapError(ctx.Err(), "contexto expirou antes do commit")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("purchase order criado com sucesso",
		zap.Int("id", purchaseOrder.ID),
		zap.String("po_no", purchaseOrder.PONo))
	return nil
}

// GetPurchaseOrderByID busca um purchase order pelo ID
func (r *purchaseOrderRepository) GetPurchaseOrderByID(ctx context.Context, id int) (*models.PurchaseOrder, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "GetPurchaseOrderByID"), zap.Int("id", id))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar purchase order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "GetPurchaseOrderByID"), zap.Int("id", id))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar purchase order")
		}
	}

	var purchaseOrder models.PurchaseOrder

	query := r.db.WithContext(ctx).Preload("Contact").
		Preload("SalesOrder").
		Preload("Items").
		Preload("Items.Product")

	if err := query.First(&purchaseOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPurchaseOrderNotFound
		}
		r.logger.Error("erro ao buscar purchase order por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar purchase order")
	}

	return &purchaseOrder, nil
}

// UpdatePurchaseOrder atualiza um purchase order existente
func (r *purchaseOrderRepository) UpdatePurchaseOrder(ctx context.Context, id int, purchaseOrder *models.PurchaseOrder) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "UpdatePurchaseOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "timeout ao atualizar purchase order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "UpdatePurchaseOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao atualizar purchase order")
		}
	}

	// Verifica se o purchase order existe
	var existing models.PurchaseOrder
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPurchaseOrderNotFound
		}
		return errors.WrapError(err, "falha ao verificar purchase order existente")
	}

	// Verifica contexto antes da operação de update
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou antes do update")
	}

	// Atualiza os campos
	purchaseOrder.ID = id

	// Trata SalesOrderID = 0 como omissão (para manter NULL no banco)
	var err error
	if purchaseOrder.SalesOrderID == 0 {
		err = r.db.WithContext(ctx).Omit("sales_order_id").Save(purchaseOrder).Error
	} else {
		err = r.db.WithContext(ctx).Save(purchaseOrder).Error
	}

	if err != nil {
		r.logger.Error("erro ao atualizar purchase order", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar purchase order")
	}

	r.logger.Info("purchase order atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeletePurchaseOrder remove um purchase order
func (r *purchaseOrderRepository) DeletePurchaseOrder(ctx context.Context, id int) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "DeletePurchaseOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "timeout ao deletar purchase order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "DeletePurchaseOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao deletar purchase order")
		}
	}

	// Verifica se existem deliveries relacionadas
	var deliveryCount int64
	if err := r.db.WithContext(ctx).Model(&models.Delivery{}).Where("purchase_order_id = ?", id).Count(&deliveryCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar deliveries relacionadas")
	}

	if deliveryCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Verifica contexto entre as operações de verificação
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou durante verificações de integridade")
	}

	// Verificação final do contexto antes da operação de delete
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou antes do delete")
	}

	// Remove o purchase order (cascade removerá os itens)
	result := r.db.WithContext(ctx).Delete(&models.PurchaseOrder{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar purchase order", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar purchase order")
	}

	if result.RowsAffected == 0 {
		return errors.ErrPurchaseOrderNotFound
	}

	r.logger.Info("purchase order deletado com sucesso", zap.Int("id", id))
	return nil
}

// generatePurchaseOrderNumber gera um número único para o purchase order
func (r *purchaseOrderRepository) generatePurchaseOrderNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastPurchaseOrder models.PurchaseOrder

	r.db.Order("id DESC").First(&lastPurchaseOrder)

	year := time.Now().Year()
	sequence := lastPurchaseOrder.ID + 1

	return fmt.Sprintf("PO-%d-%06d", year, sequence)
}
