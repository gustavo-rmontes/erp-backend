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

// PurchaseOrderRepository define a interface para operações de repositório de pedidos de compra
type PurchaseOrderRepository interface {
	// Operações CRUD básicas
	CreatePurchaseOrder(order *models.PurchaseOrder) error
	GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error)
	GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdatePurchaseOrder(id int, order *models.PurchaseOrder) error
	DeletePurchaseOrder(id int) error

	// Métodos adicionais específicos
	GetPurchaseOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersBySalesOrder(salesOrderID int) ([]models.PurchaseOrder, error)
}

// gormPurchaseOrderRepository é a implementação concreta usando GORM
type gormPurchaseOrderRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	purchaseOrderRepoInstance *gormPurchaseOrderRepository
	purchaseOrderRepoOnce     sync.Once
)

// NewPurchaseOrderRepository retorna uma instância do repositório de pedidos de compra
func NewPurchaseOrderRepository() (PurchaseOrderRepository, error) {
	var initErr error

	purchaseOrderRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("PurchaseOrderRepository")

		purchaseOrderRepoInstance = &gormPurchaseOrderRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return purchaseOrderRepoInstance, nil
}

// NewPurchaseOrderRepositoryWithDB creates a repository with a provided DB connection (for testing)
// func NewPurchaseOrderRepositoryWithDB(db *gorm.DB, log *zap.Logger) PurchaseOrderRepository {
// 	return &gormPurchaseOrderRepository{
// 		db:  db,
// 		log: log,
// 	}
// }

// CreatePurchaseOrder cria um novo pedido de compra no banco de dados
func (r *gormPurchaseOrderRepository) CreatePurchaseOrder(order *models.PurchaseOrder) error {
	r.log.Info("Iniciando criação de pedido de compra",
		zap.Int("contact_id", order.ContactID),
		zap.Int("sales_order_id", order.SalesOrderID),
		zap.String("operation", "CreatePurchaseOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Define valores padrão se não fornecidos
	if order.Status == "" {
		order.Status = models.POStatusDraft
	}

	// Preservar os itens em uma variável temporária
	items := order.Items

	// Remover os itens antes de criar o pedido
	order.Items = nil

	// Criar o pedido sem os itens
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar pedido de compra", zap.Error(err))
		return fmt.Errorf("falha ao criar pedido de compra: %w", err)
	}

	// Agora criar cada item separadamente, definindo o ID do pedido
	for i, item := range items {
		newItem := item
		newItem.ID = 0
		newItem.PurchaseOrderID = order.ID

		if err := tx.Create(&newItem).Error; err != nil {
			tx.Rollback()
			r.log.Error("Falha ao criar item do pedido de compra",
				zap.Int("purchase_order_id", order.ID),
				zap.Int("item_index", i),
				zap.Error(err),
			)
			return fmt.Errorf("falha ao criar item do pedido de compra: %w", err)
		}
	}

	// Restaurar os itens para o pedido
	if err := tx.Where("purchase_order_id = ?", order.ID).Find(&order.Items).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao recuperar itens do pedido de compra", zap.Error(err))
		return fmt.Errorf("falha ao recuperar itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pedido de compra criado com sucesso",
		zap.Int("purchase_order_id", order.ID),
		zap.String("status", order.Status),
	)

	return nil
}

// GetPurchaseOrderByID recupera um pedido de compra pelo seu ID
func (r *gormPurchaseOrderRepository) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	r.log.Info("Buscando pedido de compra por ID",
		zap.Int("purchase_order_id", id),
		zap.String("operation", "GetPurchaseOrderByID"),
	)

	var order models.PurchaseOrder
	if err := r.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pedido de compra não encontrado", zap.Int("purchase_order_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrPurchaseOrderNotFound, id)
		}
		r.log.Error("Erro ao buscar pedido de compra", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedido de compra: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&order).Association("Items").Find(&order.Items); err != nil {
		r.log.Error("Erro ao carregar itens do pedido de compra", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	// Carrega informações do contato
	if err := r.db.Model(&order).Association("Contact").Find(&order.Contact); err != nil {
		r.log.Error("Erro ao carregar contato do pedido de compra", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar contato: %w", err)
	}

	// Carrega informações do pedido de venda se existir
	if order.SalesOrderID > 0 {
		if err := r.db.Model(&order).Association("SalesOrder").Find(&order.SalesOrder); err != nil {
			r.log.Error("Erro ao carregar pedido de venda do pedido de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pedido de venda: %w", err)
		}
	}

	r.log.Info("Pedido de compra recuperado com sucesso", zap.Int("purchase_order_id", id))
	return &order, nil
}

// GetAllPurchaseOrders recupera todos os pedidos de compra do banco de dados com paginação
func (r *gormPurchaseOrderRepository) GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando pedidos de compra com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllPurchaseOrders"),
	)

	var totalItems int64
	if err := r.db.Model(&models.PurchaseOrder{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de pedidos de compra", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pedidos de compra: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var orders []models.PurchaseOrder
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de compra paginados", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de compra: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if orders[i].SalesOrderID > 0 {
			if err := r.db.Model(&orders[i]).Association("SalesOrder").Find(&orders[i].SalesOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de venda dos pedidos de compra", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de venda: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, orders)

	r.log.Info("Pedidos de compra recuperados com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdatePurchaseOrder atualiza um pedido de compra existente
func (r *gormPurchaseOrderRepository) UpdatePurchaseOrder(id int, order *models.PurchaseOrder) error {
	r.log.Info("Iniciando atualização de pedido de compra",
		zap.Int("purchase_order_id", id),
		zap.String("operation", "UpdatePurchaseOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se o pedido existe
	var existing models.PurchaseOrder
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pedido de compra não encontrado para atualização", zap.Int("purchase_order_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrPurchaseOrderNotFound, id)
		}
		r.log.Error("Erro ao verificar existência do pedido de compra", zap.Error(err))
		return fmt.Errorf("erro ao verificar pedido de compra: %w", err)
	}

	// Atualiza o pedido
	order.ID = id
	if err := tx.Model(&existing).Updates(order).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados básicos do pedido de compra", zap.Error(err))
		return fmt.Errorf("falha ao atualizar pedido de compra: %w", err)
	}

	// Deleta os itens existentes
	if err := tx.Where("purchase_order_id = ?", id).Delete(&models.POItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao remover itens existentes", zap.Error(err))
		return fmt.Errorf("falha ao remover itens: %w", err)
	}

	// Define o ID do pedido para cada item
	for i := range order.Items {
		order.Items[i].PurchaseOrderID = id
		order.Items[i].ID = 0 // Redefine o ID para criar novos itens
	}

	// Cria os novos itens
	if err := tx.CreateInBatches(order.Items, 100).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar novos itens", zap.Error(err))
		return fmt.Errorf("falha ao criar novos itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pedido de compra atualizado com sucesso", zap.Int("purchase_order_id", id))
	return nil
}

// DeletePurchaseOrder exclui um pedido de compra pelo seu ID
func (r *gormPurchaseOrderRepository) DeletePurchaseOrder(id int) error {
	r.log.Info("Iniciando exclusão de pedido de compra",
		zap.Int("purchase_order_id", id),
		zap.String("operation", "DeletePurchaseOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se há entregas vinculadas a este PO
	var count int64
	if err := tx.Model(&models.Delivery{}).Where("purchase_order_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao verificar entregas vinculadas", zap.Error(err))
		return fmt.Errorf("falha ao verificar entregas vinculadas: %w", err)
	}

	if count > 0 {
		tx.Rollback()
		r.log.Warn("Pedido de compra possui entregas vinculadas", zap.Int("purchase_order_id", id))
		return fmt.Errorf("%w: pedido possui %d entregas vinculadas", errors.ErrRelatedRecordsExist, count)
	}

	// Exclui os itens primeiro
	if err := tx.Where("purchase_order_id = ?", id).Delete(&models.POItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir itens do pedido de compra", zap.Error(err))
		return fmt.Errorf("falha ao excluir itens: %w", err)
	}

	// Exclui o pedido
	result := tx.Delete(&models.PurchaseOrder{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir pedido de compra", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir pedido de compra: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Pedido de compra não encontrado para exclusão", zap.Int("purchase_order_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrPurchaseOrderNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pedido de compra excluído com sucesso", zap.Int("purchase_order_id", id))
	return nil
}

// GetPurchaseOrdersByStatus recupera pedidos de compra por status com paginação
func (r *gormPurchaseOrderRepository) GetPurchaseOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando pedidos de compra por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetPurchaseOrdersByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.PurchaseOrder{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar pedidos de compra por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pedidos de compra: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var orders []models.PurchaseOrder
	if err := r.db.Where("status = ?", status).
		Order("expected_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de compra por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de compra: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if orders[i].SalesOrderID > 0 {
			if err := r.db.Model(&orders[i]).Association("SalesOrder").Find(&orders[i].SalesOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de venda dos pedidos de compra", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de venda: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, orders)

	r.log.Info("Pedidos de compra por status recuperados com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetPurchaseOrdersBySalesOrder recupera pedidos de compra por ID de pedido de venda
func (r *gormPurchaseOrderRepository) GetPurchaseOrdersBySalesOrder(salesOrderID int) ([]models.PurchaseOrder, error) {
	r.log.Info("Buscando pedidos de compra por pedido de venda",
		zap.Int("sales_order_id", salesOrderID),
		zap.String("operation", "GetPurchaseOrdersBySalesOrder"),
	)

	var orders []models.PurchaseOrder
	if err := r.db.Where("sales_order_id = ?", salesOrderID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de compra por pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de compra por pedido de venda: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos de compra", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}
	}

	r.log.Info("Pedidos de compra por pedido de venda recuperados com sucesso",
		zap.Int("sales_order_id", salesOrderID),
		zap.Int("count", len(orders)),
	)

	return orders, nil
}
