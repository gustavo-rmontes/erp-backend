package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PurchaseOrderRepository define as operações do repositório de purchase orders
type PurchaseOrderRepository interface {
	// CRUD básico
	CreatePurchaseOrder(purchaseOrder *models.PurchaseOrder) error
	GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error)
	UpdatePurchaseOrder(id int, purchaseOrder *models.PurchaseOrder) error
	DeletePurchaseOrder(id int) error

	// Consultas com paginação
	GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersBySalesOrder(salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByExpectedDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPendingPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverduePurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada (opcional, considere mover para serviço se contiver muita lógica de negócio)
	SearchPurchaseOrders(filter PurchaseOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

type purchaseOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPurchaseOrderRepository cria uma nova instância do repositório
func NewPurchaseOrderRepository() (PurchaseOrderRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &purchaseOrderRepository{
		db:     db,
		logger: logger.WithModule("purchase_order_repository"),
	}, nil
}

// CreatePurchaseOrder cria um novo purchase order no banco
func (r *purchaseOrderRepository) CreatePurchaseOrder(purchaseOrder *models.PurchaseOrder) error {
	// Gera o número do purchase order se não foi fornecido
	if purchaseOrder.PONo == "" {
		purchaseOrder.PONo = r.generatePurchaseOrderNumber()
	}

	// Define status padrão se não foi fornecido
	if purchaseOrder.Status == "" {
		purchaseOrder.Status = models.POStatusDraft
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria o purchase order
	if err := tx.Create(purchaseOrder).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar purchase order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar purchase order")
	}

	// Se houver itens, cria os itens
	if len(purchaseOrder.Items) > 0 {
		for i := range purchaseOrder.Items {
			purchaseOrder.Items[i].PurchaseOrderID = purchaseOrder.ID
			if err := tx.Create(&purchaseOrder.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item do purchase order", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d do purchase order", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("purchase order criado com sucesso", zap.Int("id", purchaseOrder.ID), zap.String("po_no", purchaseOrder.PONo))
	return nil
}

// GetPurchaseOrderByID busca um purchase order pelo ID
func (r *purchaseOrderRepository) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	var purchaseOrder models.PurchaseOrder

	query := r.db.Preload("Contact").
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

// GetAllPurchaseOrders retorna todos os purchase orders com paginação
func (r *purchaseOrderRepository) GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	// Query base
	query := r.db.Model(&models.PurchaseOrder{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// UpdatePurchaseOrder atualiza um purchase order existente
func (r *purchaseOrderRepository) UpdatePurchaseOrder(id int, purchaseOrder *models.PurchaseOrder) error {
	// Verifica se o purchase order existe
	var existing models.PurchaseOrder
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPurchaseOrderNotFound
		}
		return errors.WrapError(err, "falha ao verificar purchase order existente")
	}

	// Atualiza os campos
	purchaseOrder.ID = id
	if err := r.db.Save(purchaseOrder).Error; err != nil {
		r.logger.Error("erro ao atualizar purchase order", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar purchase order")
	}

	r.logger.Info("purchase order atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeletePurchaseOrder remove um purchase order
func (r *purchaseOrderRepository) DeletePurchaseOrder(id int) error {
	// Verifica se existem deliveries relacionadas
	var deliveryCount int64
	if err := r.db.Model(&models.Delivery{}).Where("purchase_order_id = ?", id).Count(&deliveryCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar deliveries relacionadas")
	}

	if deliveryCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove o purchase order (cascade removerá os itens)
	result := r.db.Delete(&models.PurchaseOrder{}, id)
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
