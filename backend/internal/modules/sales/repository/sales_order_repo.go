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

// SalesOrderRepository define as operações do repositório de sales orders
type SalesOrderRepository interface {
	// CRUD básico
	CreateSalesOrder(ctx context.Context, salesOrder *models.SalesOrder) error
	GetSalesOrderByID(ctx context.Context, id int) (*models.SalesOrder, error)
	UpdateSalesOrder(ctx context.Context, id int, salesOrder *models.SalesOrder) error
	DeleteSalesOrder(ctx context.Context, id int) error

	// Consultas com paginação
	GetAllSalesOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByQuotation(ctx context.Context, quotationID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByPeriod(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByExpectedDate(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada (opcional, considere mover para serviço se contiver muita lógica de negócio)
	SearchSalesOrders(ctx context.Context, filter SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

type salesOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSalesOrderRepository cria uma nova instância do repositório
func NewSalesOrderRepository(db *gorm.DB, logger *zap.Logger) SalesOrderRepository {
	return &salesOrderRepository{
		db:     db,
		logger: logger.With(zap.String("module", "sales_order_repository")),
	}
}

// CreateSalesOrder cria um novo sales order no banco
func (r *salesOrderRepository) CreateSalesOrder(ctx context.Context, salesOrder *models.SalesOrder) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "CreateSalesOrder"))
			return errors.WrapError(ctx.Err(), "timeout ao criar sales order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "CreateSalesOrder"))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao criar sales order")
		}
	}

	// Preparação do sales order
	if salesOrder.SONo == "" {
		salesOrder.SONo = r.generateSalesOrderNumber()
	}

	if salesOrder.Status == "" {
		salesOrder.Status = models.SOStatusDraft
	}

	// Inicia transação com contexto
	tx := r.db.WithContext(ctx).Begin()

	// Verifica novamente o contexto após iniciar transação
	if ctx.Err() != nil {
		tx.Rollback()
		return errors.WrapError(ctx.Err(), "contexto expirou após iniciar transação")
	}

	// Cria o sales order
	if err := tx.Create(salesOrder).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar sales order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar sales order")
	}

	// Se houver itens, cria os itens
	if len(salesOrder.Items) > 0 {
		for i := range salesOrder.Items {
			// Verifica contexto durante loop para operações longas
			if ctx.Err() != nil {
				tx.Rollback()
				return errors.WrapError(ctx.Err(), "contexto expirou durante criação de itens")
			}

			salesOrder.Items[i].SalesOrderID = salesOrder.ID
			if err := tx.Create(&salesOrder.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item do sales order",
					zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d do sales order", i))
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

	r.logger.Info("sales order criado com sucesso",
		zap.Int("id", salesOrder.ID),
		zap.String("so_no", salesOrder.SONo))
	return nil
}

// GetSalesOrderByID busca um sales order pelo ID
func (r *salesOrderRepository) GetSalesOrderByID(ctx context.Context, id int) (*models.SalesOrder, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "GetSalesOrderByID"), zap.Int("id", id))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "GetSalesOrderByID"), zap.Int("id", id))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales order")
		}
	}

	var salesOrder models.SalesOrder

	query := r.db.WithContext(ctx).Preload("Contact").
		Preload("Quotation").
		Preload("Items").
		Preload("Items.Product")

	if err := query.First(&salesOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSalesOrderNotFound
		}
		r.logger.Error("erro ao buscar sales order por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar sales order")
	}

	return &salesOrder, nil
}

// UpdateSalesOrder atualiza um sales order existente
func (r *salesOrderRepository) UpdateSalesOrder(ctx context.Context, id int, salesOrder *models.SalesOrder) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "UpdateSalesOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "timeout ao atualizar sales order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "UpdateSalesOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao atualizar sales order")
		}
	}

	// Verifica se o sales order existe
	var existing models.SalesOrder
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesOrderNotFound
		}
		return errors.WrapError(err, "falha ao verificar sales order existente")
	}

	// Verifica contexto antes da operação de update
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou antes do update")
	}

	// Atualiza os campos
	salesOrder.ID = id
	if err := r.db.WithContext(ctx).Save(salesOrder).Error; err != nil {
		r.logger.Error("erro ao atualizar sales order", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar sales order")
	}

	r.logger.Info("sales order atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeleteSalesOrder remove um sales order
func (r *salesOrderRepository) DeleteSalesOrder(ctx context.Context, id int) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "DeleteSalesOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "timeout ao deletar sales order")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "DeleteSalesOrder"), zap.Int("id", id))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao deletar sales order")
		}
	}

	// Verifica se existem invoices ou purchase orders relacionados
	var invoiceCount int64
	if err := r.db.WithContext(ctx).Model(&models.Invoice{}).Where("sales_order_id = ?", id).Count(&invoiceCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar invoices relacionadas")
	}

	if invoiceCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Verifica contexto entre as operações de verificação
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou durante verificações de integridade")
	}

	var poCount int64
	if err := r.db.WithContext(ctx).Model(&models.PurchaseOrder{}).Where("sales_order_id = ?", id).Count(&poCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar purchase orders relacionadas")
	}

	if poCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Verificação final do contexto antes da operação de delete
	if ctx.Err() != nil {
		return errors.WrapError(ctx.Err(), "contexto expirou antes do delete")
	}

	// Remove o sales order (cascade removerá os itens)
	result := r.db.WithContext(ctx).Delete(&models.SalesOrder{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar sales order", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar sales order")
	}

	if result.RowsAffected == 0 {
		return errors.ErrSalesOrderNotFound
	}

	r.logger.Info("sales order deletado com sucesso", zap.Int("id", id))
	return nil
}

// generateSalesOrderNumber gera um número único para o sales order
func (r *salesOrderRepository) generateSalesOrderNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastSalesOrder models.SalesOrder

	r.db.Order("id DESC").First(&lastSalesOrder)

	year := time.Now().Year()
	sequence := lastSalesOrder.ID + 1

	return fmt.Sprintf("SO-%d-%06d", year, sequence)
}
