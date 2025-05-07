package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DeliveryRepository define a interface para operações de repositório de entregas
type DeliveryRepository interface {
	// Operações CRUD básicas
	CreateDelivery(delivery *models.Delivery) error
	GetDeliveryByID(id int) (*models.Delivery, error)
	GetAllDeliveries(pagination *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateDelivery(id int, delivery *models.Delivery) error
	DeleteDelivery(id int) error

	// Métodos adicionais específicos
	GetDeliveriesByStatus(status string, pagination *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesBySalesOrder(salesOrderID int) ([]models.Delivery, error)
	GetDeliveriesByPurchaseOrder(purchaseOrderID int) ([]models.Delivery, error)
	GetPendingDeliveries(pagination *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// gormDeliveryRepository é a implementação concreta usando GORM
type gormDeliveryRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	deliveryRepoInstance *gormDeliveryRepository
	deliveryRepoOnce     sync.Once
)

// NewDeliveryRepository retorna uma instância do repositório de entregas
func NewDeliveryRepository() (DeliveryRepository, error) {
	var initErr error

	deliveryRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usa o logger centralizado com o módulo específico
		log := logger.WithModule("DeliveryRepository")

		deliveryRepoInstance = &gormDeliveryRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return deliveryRepoInstance, nil
}

// CreateDelivery cria uma nova entrega no banco de dados
func (r *gormDeliveryRepository) CreateDelivery(delivery *models.Delivery) error {
	r.log.Info("Iniciando criação de entrega",
		zap.Int("purchase_order_id", delivery.PurchaseOrderID),
		zap.Int("sales_order_id", delivery.SalesOrderID),
		zap.String("operation", "CreateDelivery"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Define valores padrão se não fornecidos
	if delivery.Status == "" {
		delivery.Status = models.DeliveryStatusPending
	}

	// Preservar os itens em uma variável temporária
	items := delivery.Items

	// Remover os itens antes de criar a entrega
	delivery.Items = nil

	// Criar a entrega sem os itens
	if err := tx.Create(delivery).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar entrega", zap.Error(err))
		return fmt.Errorf("falha ao criar entrega: %w", err)
	}

	// Agora criar cada item separadamente, definindo o ID da entrega
	for i, item := range items {
		newItem := item
		newItem.ID = 0
		newItem.DeliveryID = delivery.ID

		if err := tx.Create(&newItem).Error; err != nil {
			tx.Rollback()
			r.log.Error("Falha ao criar item da entrega",
				zap.Int("delivery_id", delivery.ID),
				zap.Int("item_index", i),
				zap.Error(err),
			)
			return fmt.Errorf("falha ao criar item da entrega: %w", err)
		}
	}

	// Restaurar os itens para a entrega
	if err := tx.Where("delivery_id = ?", delivery.ID).Find(&delivery.Items).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao recuperar itens da entrega", zap.Error(err))
		return fmt.Errorf("falha ao recuperar itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Entrega criada com sucesso",
		zap.Int("delivery_id", delivery.ID),
		zap.String("status", delivery.Status),
	)

	return nil
}

// GetDeliveryByID recupera uma entrega pelo seu ID
func (r *gormDeliveryRepository) GetDeliveryByID(id int) (*models.Delivery, error) {
	r.log.Info("Buscando entrega por ID",
		zap.Int("delivery_id", id),
		zap.String("operation", "GetDeliveryByID"),
	)

	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Entrega não encontrada", zap.Int("delivery_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrDeliveryNotFound, id)
		}
		r.log.Error("Erro ao buscar entrega", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar entrega: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&delivery).Association("Items").Find(&delivery.Items); err != nil {
		r.log.Error("Erro ao carregar itens da entrega", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	// Carrega o pedido de compra se existir
	if delivery.PurchaseOrderID > 0 {
		if err := r.db.Model(&delivery).Association("PurchaseOrder").Find(&delivery.PurchaseOrder); err != nil {
			r.log.Error("Erro ao carregar pedido de compra da entrega", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pedido de compra: %w", err)
		}
	}

	// Carrega o pedido de venda se existir
	if delivery.SalesOrderID > 0 {
		if err := r.db.Model(&delivery).Association("SalesOrder").Find(&delivery.SalesOrder); err != nil {
			r.log.Error("Erro ao carregar pedido de venda da entrega", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pedido de venda: %w", err)
		}
	}

	r.log.Info("Entrega recuperada com sucesso", zap.Int("delivery_id", id))
	return &delivery, nil
}

// GetAllDeliveries recupera todas as entregas do banco de dados com paginação
func (r *gormDeliveryRepository) GetAllDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Valor padrão para paginação
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	if params != nil {
		if !params.Validate() {
			return nil, errors.ErrInvalidPagination
		}
		page = params.Page
		pageSize = params.PageSize
	}

	r.log.Info("Buscando entregas com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllDeliveries"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Delivery{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de entregas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar entregas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var deliveries []models.Delivery
	if err := r.db.Offset(offset).Limit(pageSize).Find(&deliveries).Error; err != nil {
		r.log.Error("Erro ao buscar entregas paginadas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar entregas: %w", err)
	}

	// Carrega os relacionamentos para cada entrega
	for i := range deliveries {
		if err := r.db.Model(&deliveries[i]).Association("Items").Find(&deliveries[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das entregas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if deliveries[i].PurchaseOrderID > 0 {
			if err := r.db.Model(&deliveries[i]).Association("PurchaseOrder").Find(&deliveries[i].PurchaseOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de compra das entregas", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de compra: %w", err)
			}
		}

		if deliveries[i].SalesOrderID > 0 {
			if err := r.db.Model(&deliveries[i]).Association("SalesOrder").Find(&deliveries[i].SalesOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de venda das entregas", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de venda: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, deliveries)

	r.log.Info("Entregas recuperadas com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdateDelivery atualiza uma entrega existente
func (r *gormDeliveryRepository) UpdateDelivery(id int, delivery *models.Delivery) error {
	r.log.Info("Iniciando atualização de entrega",
		zap.Int("delivery_id", id),
		zap.String("operation", "UpdateDelivery"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se a entrega existe
	var existing models.Delivery
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Entrega não encontrada para atualização", zap.Int("delivery_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrDeliveryNotFound, id)
		}
		r.log.Error("Erro ao verificar existência da entrega", zap.Error(err))
		return fmt.Errorf("erro ao verificar entrega: %w", err)
	}

	// Atualiza a entrega
	delivery.ID = id
	if err := tx.Model(&existing).Updates(delivery).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados básicos da entrega", zap.Error(err))
		return fmt.Errorf("falha ao atualizar entrega: %w", err)
	}

	// Deleta os itens existentes
	if err := tx.Where("delivery_id = ?", id).Delete(&models.DeliveryItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao remover itens existentes", zap.Error(err))
		return fmt.Errorf("falha ao remover itens: %w", err)
	}

	// Define o ID da entrega para cada item
	for i := range delivery.Items {
		delivery.Items[i].DeliveryID = id
		delivery.Items[i].ID = 0 // Redefine o ID para criar novos itens
	}

	// Cria os novos itens
	if err := tx.CreateInBatches(delivery.Items, 100).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar novos itens", zap.Error(err))
		return fmt.Errorf("falha ao criar novos itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Entrega atualizada com sucesso", zap.Int("delivery_id", id))
	return nil
}

// DeleteDelivery exclui uma entrega pelo seu ID
func (r *gormDeliveryRepository) DeleteDelivery(id int) error {
	r.log.Info("Iniciando exclusão de entrega",
		zap.Int("delivery_id", id),
		zap.String("operation", "DeleteDelivery"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Exclui os itens primeiro
	if err := tx.Where("delivery_id = ?", id).Delete(&models.DeliveryItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir itens da entrega", zap.Error(err))
		return fmt.Errorf("falha ao excluir itens: %w", err)
	}

	// Exclui a entrega
	result := tx.Delete(&models.Delivery{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir entrega", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir entrega: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Entrega não encontrada para exclusão", zap.Int("delivery_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrDeliveryNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Entrega excluída com sucesso", zap.Int("delivery_id", id))
	return nil
}

// GetDeliveriesByStatus recupera entregas por status com paginação
func (r *gormDeliveryRepository) GetDeliveriesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Valor padrão para paginação
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	if params != nil {
		if !params.Validate() {
			return nil, errors.ErrInvalidPagination
		}
		page = params.Page
		pageSize = params.PageSize
	}

	r.log.Info("Buscando entregas por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetDeliveriesByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Delivery{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar entregas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar entregas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var deliveries []models.Delivery
	if err := r.db.Where("status = ?", status).Offset(offset).Limit(pageSize).Find(&deliveries).Error; err != nil {
		r.log.Error("Erro ao buscar entregas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar entregas: %w", err)
	}

	// Carrega os relacionamentos para cada entrega
	for i := range deliveries {
		if err := r.db.Model(&deliveries[i]).Association("Items").Find(&deliveries[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das entregas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if deliveries[i].PurchaseOrderID > 0 {
			if err := r.db.Model(&deliveries[i]).Association("PurchaseOrder").Find(&deliveries[i].PurchaseOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de compra das entregas", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de compra: %w", err)
			}
		}

		if deliveries[i].SalesOrderID > 0 {
			if err := r.db.Model(&deliveries[i]).Association("SalesOrder").Find(&deliveries[i].SalesOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de venda das entregas", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de venda: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, deliveries)

	r.log.Info("Entregas por status recuperadas com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetDeliveriesBySalesOrder recupera entregas por ID de pedido de venda
func (r *gormDeliveryRepository) GetDeliveriesBySalesOrder(salesOrderID int) ([]models.Delivery, error) {
	r.log.Info("Buscando entregas por pedido de venda",
		zap.Int("sales_order_id", salesOrderID),
		zap.String("operation", "GetDeliveriesBySalesOrder"),
	)

	var deliveries []models.Delivery
	if err := r.db.Where("sales_order_id = ?", salesOrderID).Find(&deliveries).Error; err != nil {
		r.log.Error("Erro ao buscar entregas por pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar entregas por pedido de venda: %w", err)
	}

	// Carrega os relacionamentos para cada entrega
	for i := range deliveries {
		if err := r.db.Model(&deliveries[i]).Association("Items").Find(&deliveries[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das entregas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}
	}

	r.log.Info("Entregas por pedido de venda recuperadas com sucesso",
		zap.Int("sales_order_id", salesOrderID),
		zap.Int("count", len(deliveries)),
	)

	return deliveries, nil
}

// GetDeliveriesByPurchaseOrder recupera entregas por ID de pedido de compra
func (r *gormDeliveryRepository) GetDeliveriesByPurchaseOrder(purchaseOrderID int) ([]models.Delivery, error) {
	r.log.Info("Buscando entregas por pedido de compra",
		zap.Int("purchase_order_id", purchaseOrderID),
		zap.String("operation", "GetDeliveriesByPurchaseOrder"),
	)

	var deliveries []models.Delivery
	if err := r.db.Where("purchase_order_id = ?", purchaseOrderID).Find(&deliveries).Error; err != nil {
		r.log.Error("Erro ao buscar entregas por pedido de compra", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar entregas por pedido de compra: %w", err)
	}

	// Carrega os relacionamentos para cada entrega
	for i := range deliveries {
		if err := r.db.Model(&deliveries[i]).Association("Items").Find(&deliveries[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das entregas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}
	}

	r.log.Info("Entregas por pedido de compra recuperadas com sucesso",
		zap.Int("purchase_order_id", purchaseOrderID),
		zap.Int("count", len(deliveries)),
	)

	return deliveries, nil
}

// GetPendingDeliveries recupera entregas pendentes com paginação
func (r *gormDeliveryRepository) GetPendingDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	r.log.Info("Buscando entregas pendentes",
		zap.String("operation", "GetPendingDeliveries"),
	)

	// Reutilizando o método GetDeliveriesByStatus com o status "pending"
	return r.GetDeliveriesByStatus(models.DeliveryStatusPending, params)
}
