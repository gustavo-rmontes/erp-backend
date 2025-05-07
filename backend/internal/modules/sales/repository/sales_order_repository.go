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

// SalesOrderRepository define a interface para operações de repositório de pedidos de venda
type SalesOrderRepository interface {
	// Operações CRUD básicas
	CreateSalesOrder(order *models.SalesOrder) error
	GetSalesOrderByID(id int) (*models.SalesOrder, error)
	GetAllSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateSalesOrder(id int, order *models.SalesOrder) error
	DeleteSalesOrder(id int) error

	// Métodos adicionais específicos
	GetSalesOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByQuotation(quotationID int) (*models.SalesOrder, error)
}

// gormSalesOrderRepository é a implementação concreta usando GORM
type gormSalesOrderRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	salesOrderRepoInstance *gormSalesOrderRepository
	salesOrderRepoOnce     sync.Once
)

// NewSalesOrderRepository retorna uma instância do repositório de pedidos de venda
func NewSalesOrderRepository() (SalesOrderRepository, error) {
	var initErr error

	salesOrderRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("SalesOrderRepository")

		salesOrderRepoInstance = &gormSalesOrderRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return salesOrderRepoInstance, nil
}

// CreateSalesOrder cria um novo pedido de venda no banco de dados
func (r *gormSalesOrderRepository) CreateSalesOrder(order *models.SalesOrder) error {
	r.log.Info("Iniciando criação de pedido de venda",
		zap.Int("contact_id", order.ContactID),
		zap.Int("quotation_id", order.QuotationID),
		zap.String("operation", "CreateSalesOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Define valores padrão se não fornecidos
	if order.Status == "" {
		order.Status = models.SOStatusDraft
	}

	// Preservar os itens em uma variável temporária
	items := order.Items

	// Remover os itens antes de criar o pedido
	order.Items = nil

	// Criar o pedido sem os itens
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar pedido de venda", zap.Error(err))
		return fmt.Errorf("falha ao criar pedido de venda: %w", err)
	}

	// Agora criar cada item separadamente, definindo o ID do pedido
	for i, item := range items {
		newItem := item
		newItem.ID = 0
		newItem.SalesOrderID = order.ID

		if err := tx.Create(&newItem).Error; err != nil {
			tx.Rollback()
			r.log.Error("Falha ao criar item do pedido de venda",
				zap.Int("sales_order_id", order.ID),
				zap.Int("item_index", i),
				zap.Error(err),
			)
			return fmt.Errorf("falha ao criar item do pedido de venda: %w", err)
		}
	}

	// Restaurar os itens para o pedido
	if err := tx.Where("sales_order_id = ?", order.ID).Find(&order.Items).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao recuperar itens do pedido de venda", zap.Error(err))
		return fmt.Errorf("falha ao recuperar itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pedido de venda criado com sucesso",
		zap.Int("sales_order_id", order.ID),
		zap.String("status", order.Status),
	)

	return nil
}

// GetSalesOrderByID recupera um pedido de venda pelo seu ID
func (r *gormSalesOrderRepository) GetSalesOrderByID(id int) (*models.SalesOrder, error) {
	r.log.Info("Buscando pedido de venda por ID",
		zap.Int("sales_order_id", id),
		zap.String("operation", "GetSalesOrderByID"),
	)

	var order models.SalesOrder
	if err := r.db.First(&order, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pedido de venda não encontrado", zap.Int("sales_order_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrSalesOrderNotFound, id)
		}
		r.log.Error("Erro ao buscar pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedido de venda: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&order).Association("Items").Find(&order.Items); err != nil {
		r.log.Error("Erro ao carregar itens do pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	// Carrega informações do contato
	if err := r.db.Model(&order).Association("Contact").Find(&order.Contact); err != nil {
		r.log.Error("Erro ao carregar contato do pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar contato: %w", err)
	}

	// Carrega informações da cotação se existir
	if order.QuotationID > 0 {
		if err := r.db.Model(&order).Association("Quotation").Find(&order.Quotation); err != nil {
			r.log.Error("Erro ao carregar cotação do pedido de venda", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar cotação: %w", err)
		}
	}

	r.log.Info("Pedido de venda recuperado com sucesso", zap.Int("sales_order_id", id))
	return &order, nil
}

// GetAllSalesOrders recupera todos os pedidos de venda do banco de dados com paginação
func (r *gormSalesOrderRepository) GetAllSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando pedidos de venda com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllSalesOrders"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesOrder{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de pedidos de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pedidos de venda: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var orders []models.SalesOrder
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de venda paginados", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de venda: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if orders[i].QuotationID > 0 {
			if err := r.db.Model(&orders[i]).Association("Quotation").Find(&orders[i].Quotation); err != nil {
				r.log.Error("Erro ao carregar cotações dos pedidos", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar cotações: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, orders)

	r.log.Info("Pedidos de venda recuperados com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdateSalesOrder atualiza um pedido de venda existente
func (r *gormSalesOrderRepository) UpdateSalesOrder(id int, order *models.SalesOrder) error {
	r.log.Info("Iniciando atualização de pedido de venda",
		zap.Int("sales_order_id", id),
		zap.String("operation", "UpdateSalesOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se o pedido existe
	var existing models.SalesOrder
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pedido de venda não encontrado para atualização", zap.Int("sales_order_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesOrderNotFound, id)
		}
		r.log.Error("Erro ao verificar existência do pedido de venda", zap.Error(err))
		return fmt.Errorf("erro ao verificar pedido de venda: %w", err)
	}

	// Atualiza o pedido
	order.ID = id
	if err := tx.Model(&existing).Updates(order).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados básicos do pedido de venda", zap.Error(err))
		return fmt.Errorf("falha ao atualizar pedido de venda: %w", err)
	}

	// Deleta os itens existentes
	if err := tx.Where("sales_order_id = ?", id).Delete(&models.SOItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao remover itens existentes", zap.Error(err))
		return fmt.Errorf("falha ao remover itens: %w", err)
	}

	// Define o ID do pedido para cada item
	for i := range order.Items {
		order.Items[i].SalesOrderID = id
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

	r.log.Info("Pedido de venda atualizado com sucesso", zap.Int("sales_order_id", id))
	return nil
}

// DeleteSalesOrder exclui um pedido de venda pelo seu ID
func (r *gormSalesOrderRepository) DeleteSalesOrder(id int) error {
	r.log.Info("Iniciando exclusão de pedido de venda",
		zap.Int("sales_order_id", id),
		zap.String("operation", "DeleteSalesOrder"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se há POs vinculados a este SO
	var count int64
	if err := tx.Model(&models.PurchaseOrder{}).Where("sales_order_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao verificar pedidos de compra vinculados", zap.Error(err))
		return fmt.Errorf("falha ao verificar pedidos de compra vinculados: %w", err)
	}

	if count > 0 {
		tx.Rollback()
		r.log.Warn("Pedido de venda possui pedidos de compra vinculados", zap.Int("sales_order_id", id))
		return fmt.Errorf("%w: pedido possui %d pedidos de compra vinculados", errors.ErrRelatedRecordsExist, count)
	}

	// Verifica se há faturas vinculadas a este SO
	if err := tx.Model(&models.Invoice{}).Where("sales_order_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao verificar faturas vinculadas", zap.Error(err))
		return fmt.Errorf("falha ao verificar faturas vinculadas: %w", err)
	}

	if count > 0 {
		tx.Rollback()
		r.log.Warn("Pedido de venda possui faturas vinculadas", zap.Int("sales_order_id", id))
		return fmt.Errorf("%w: pedido possui %d faturas vinculadas", errors.ErrRelatedRecordsExist, count)
	}

	// Exclui os itens primeiro
	if err := tx.Where("sales_order_id = ?", id).Delete(&models.SOItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir itens do pedido de venda", zap.Error(err))
		return fmt.Errorf("falha ao excluir itens: %w", err)
	}

	// Exclui o pedido
	result := tx.Delete(&models.SalesOrder{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir pedido de venda", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir pedido de venda: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Pedido de venda não encontrado para exclusão", zap.Int("sales_order_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrSalesOrderNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pedido de venda excluído com sucesso", zap.Int("sales_order_id", id))
	return nil
}

// GetSalesOrdersByStatus recupera pedidos de venda por status com paginação
func (r *gormSalesOrderRepository) GetSalesOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando pedidos de venda por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetSalesOrdersByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesOrder{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar pedidos de venda por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pedidos de venda: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var orders []models.SalesOrder
	if err := r.db.Where("status = ?", status).
		Order("delivery_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de venda por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de venda: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if orders[i].QuotationID > 0 {
			if err := r.db.Model(&orders[i]).Association("Quotation").Find(&orders[i].Quotation); err != nil {
				r.log.Error("Erro ao carregar cotações dos pedidos", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar cotações: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, orders)

	r.log.Info("Pedidos de venda por status recuperados com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetSalesOrdersByContact recupera pedidos de venda por ID de contato com paginação
func (r *gormSalesOrderRepository) GetSalesOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando pedidos de venda por contato",
		zap.Int("contact_id", contactID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetSalesOrdersByContact"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesOrder{}).Where("contact_id = ?", contactID).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar pedidos de venda por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pedidos de venda: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var orders []models.SalesOrder
	if err := r.db.Where("contact_id = ?", contactID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&orders).Error; err != nil {
		r.log.Error("Erro ao buscar pedidos de venda por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedidos de venda: %w", err)
	}

	// Carrega os relacionamentos para cada pedido
	for i := range orders {
		if err := r.db.Model(&orders[i]).Association("Items").Find(&orders[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&orders[i]).Association("Contact").Find(&orders[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos pedidos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if orders[i].QuotationID > 0 {
			if err := r.db.Model(&orders[i]).Association("Quotation").Find(&orders[i].Quotation); err != nil {
				r.log.Error("Erro ao carregar cotações dos pedidos", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar cotações: %w", err)
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, orders)

	r.log.Info("Pedidos de venda por contato recuperados com sucesso",
		zap.Int("contact_id", contactID),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetSalesOrdersByQuotation recupera um pedido de venda pelo ID da cotação
func (r *gormSalesOrderRepository) GetSalesOrdersByQuotation(quotationID int) (*models.SalesOrder, error) {
	r.log.Info("Buscando pedido de venda por cotação",
		zap.Int("quotation_id", quotationID),
		zap.String("operation", "GetSalesOrdersByQuotation"),
	)

	var order models.SalesOrder
	if err := r.db.Where("quotation_id = ?", quotationID).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pedido de venda não encontrado para a cotação", zap.Int("quotation_id", quotationID))
			return nil, nil // Retorna nil sem erro, pois pode não existir um pedido para esta cotação
		}
		r.log.Error("Erro ao buscar pedido de venda por cotação", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pedido de venda por cotação: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&order).Association("Items").Find(&order.Items); err != nil {
		r.log.Error("Erro ao carregar itens do pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	r.log.Info("Pedido de venda por cotação recuperado com sucesso",
		zap.Int("quotation_id", quotationID),
		zap.Int("sales_order_id", order.ID),
	)

	return &order, nil
}
