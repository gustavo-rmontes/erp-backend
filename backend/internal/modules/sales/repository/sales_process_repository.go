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

// SalesProcessRepository define a interface para operações de repositório de processos de vendas
type SalesProcessRepository interface {
	// Operações CRUD básicas
	CreateSalesProcess(process *models.SalesProcess) error
	GetSalesProcessByID(id int) (*models.SalesProcess, error)
	GetAllSalesProcesses(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateSalesProcess(id int, process *models.SalesProcess) error
	DeleteSalesProcess(id int) error

	// Métodos auxiliares específicos para o SalesProcess
	LinkQuotationToProcess(processID int, quotationID int) error
	LinkSalesOrderToProcess(processID int, salesOrderID int) error
	LinkPurchaseOrderToProcess(processID int, purchaseOrderID int) error
	LinkDeliveryToProcess(processID int, deliveryID int) error
	LinkInvoiceToProcess(processID int, invoiceID int) error

	// Métodos adicionais de busca
	GetSalesProcessByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesProcessByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// gormSalesProcessRepository é a implementação concreta usando GORM
type gormSalesProcessRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	salesProcessRepoInstance *gormSalesProcessRepository
	salesProcessRepoOnce     sync.Once
)

// NewSalesProcessRepository retorna uma instância do repositório de processos de vendas
func NewSalesProcessRepository() (SalesProcessRepository, error) {
	var initErr error

	salesProcessRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("SalesProcessRepository")

		salesProcessRepoInstance = &gormSalesProcessRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return salesProcessRepoInstance, nil
}

// NewTestSalesProcessRepository creates a repository instance for testing
func NewTestSalesProcessRepository(db *gorm.DB) SalesProcessRepository {
	return &gormSalesProcessRepository{
		db:  db,
		log: logger.WithModule("TestSalesProcessRepository"),
	}
}

// CreateSalesProcess cria um novo processo de vendas no banco de dados
func (r *gormSalesProcessRepository) CreateSalesProcess(process *models.SalesProcess) error {
	r.log.Info("Iniciando criação de processo de vendas",
		zap.Int("contact_id", process.ContactID),
		zap.String("operation", "CreateSalesProcess"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Criar o processo de vendas
	if err := tx.Create(process).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar processo de vendas", zap.Error(err))
		return fmt.Errorf("falha ao criar processo de vendas: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Processo de vendas criado com sucesso",
		zap.Int("process_id", process.ID),
		zap.String("status", process.Status),
	)

	return nil
}

// GetSalesProcessByID recupera um processo de vendas pelo seu ID
func (r *gormSalesProcessRepository) GetSalesProcessByID(id int) (*models.SalesProcess, error) {
	r.log.Info("Buscando processo de vendas por ID",
		zap.Int("process_id", id),
		zap.String("operation", "GetSalesProcessByID"),
	)

	var process models.SalesProcess
	if err := r.db.First(&process, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, id)
		}
		r.log.Error("Erro ao buscar processo de vendas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar processo de vendas: %w", err)
	}

	// Carrega informações do contato
	if err := r.db.Model(&process).Association("Contact").Find(&process.Contact); err != nil {
		r.log.Error("Erro ao carregar contato do processo de vendas", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar contato: %w", err)
	}

	// Carrega documentos relacionados
	if err := r.loadSalesProcessDocuments(&process); err != nil {
		r.log.Error("Erro ao carregar documentos relacionados ao processo de vendas", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar documentos relacionados: %w", err)
	}

	r.log.Info("Processo de vendas recuperado com sucesso", zap.Int("process_id", id))
	return &process, nil
}

// loadSalesProcessDocuments carrega todos os documentos relacionados a um processo de vendas
func (r *gormSalesProcessRepository) loadSalesProcessDocuments(process *models.SalesProcess) error {
	// Carrega a cotação associada ao processo
	var quotationIDs []int
	if err := r.db.Table("process_quotations").Where("process_id = ?", process.ID).Pluck("quotation_id", &quotationIDs).Error; err != nil {
		return err
	}

	if len(quotationIDs) > 0 {
		quotationRepo, err := NewQuotationRepository()
		if err != nil {
			return fmt.Errorf("erro ao inicializar repositório de cotações: %w", err)
		}

		quotation, err := quotationRepo.GetQuotationByID(quotationIDs[0])
		if err == nil {
			process.Quotation = quotation
		}
	}

	// Carrega o pedido de venda associado ao processo
	var salesOrderIDs []int
	if err := r.db.Table("process_sales_orders").Where("process_id = ?", process.ID).Pluck("sales_order_id", &salesOrderIDs).Error; err != nil {
		return err
	}

	if len(salesOrderIDs) > 0 {
		salesOrderRepo, err := NewSalesOrderRepository()
		if err != nil {
			return fmt.Errorf("erro ao inicializar repositório de pedidos de venda: %w", err)
		}

		salesOrder, err := salesOrderRepo.GetSalesOrderByID(salesOrderIDs[0])
		if err == nil {
			process.SalesOrder = salesOrder
		}
	}

	// Carrega o pedido de compra associado ao processo
	var purchaseOrderIDs []int
	if err := r.db.Table("process_purchase_orders").Where("process_id = ?", process.ID).Pluck("purchase_order_id", &purchaseOrderIDs).Error; err != nil {
		return err
	}

	if len(purchaseOrderIDs) > 0 {
		purchaseOrderRepo, err := NewPurchaseOrderRepository()
		if err != nil {
			return fmt.Errorf("erro ao inicializar repositório de pedidos de compra: %w", err)
		}

		purchaseOrder, err := purchaseOrderRepo.GetPurchaseOrderByID(purchaseOrderIDs[0])
		if err == nil {
			process.PurchaseOrder = purchaseOrder
		}
	}

	// Carrega as entregas associadas ao processo
	var deliveryIDs []int
	if err := r.db.Table("process_deliveries").Where("process_id = ?", process.ID).Pluck("delivery_id", &deliveryIDs).Error; err != nil {
		return err
	}

	if len(deliveryIDs) > 0 {
		deliveryRepo, err := NewDeliveryRepository()
		if err != nil {
			return fmt.Errorf("erro ao inicializar repositório de entregas: %w", err)
		}

		process.Deliveries = []models.Delivery{}
		for _, deliveryID := range deliveryIDs {
			delivery, err := deliveryRepo.GetDeliveryByID(deliveryID)
			if err == nil {
				process.Deliveries = append(process.Deliveries, *delivery)
			}
		}
	}

	// Carrega as faturas associadas ao processo
	var invoiceIDs []int
	if err := r.db.Table("process_invoices").Where("process_id = ?", process.ID).Pluck("invoice_id", &invoiceIDs).Error; err != nil {
		return err
	}

	if len(invoiceIDs) > 0 {
		invoiceRepo, err := NewInvoiceRepository()
		if err != nil {
			return fmt.Errorf("erro ao inicializar repositório de faturas: %w", err)
		}

		process.Invoices = []models.Invoice{}
		for _, invoiceID := range invoiceIDs {
			invoice, err := invoiceRepo.GetInvoiceByID(invoiceID)
			if err == nil {
				process.Invoices = append(process.Invoices, *invoice)
			}
		}
	}

	return nil
}

// GetAllSalesProcesses recupera todos os processos de vendas do banco de dados com paginação
func (r *gormSalesProcessRepository) GetAllSalesProcesses(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando processos de vendas com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllSalesProcesses"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesProcess{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de processos de vendas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar processos de vendas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var processes []models.SalesProcess
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&processes).Error; err != nil {
		r.log.Error("Erro ao buscar processos de vendas paginados", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar processos de vendas: %w", err)
	}

	// Carrega informações do contato para cada processo
	for i := range processes {
		if err := r.db.Model(&processes[i]).Association("Contact").Find(&processes[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos processos de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		// Carrega documentos relacionados
		if err := r.loadSalesProcessDocuments(&processes[i]); err != nil {
			r.log.Error("Erro ao carregar documentos relacionados ao processo de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar documentos relacionados: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, processes)

	r.log.Info("Processos de vendas recuperados com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdateSalesProcess atualiza um processo de vendas existente
func (r *gormSalesProcessRepository) UpdateSalesProcess(id int, process *models.SalesProcess) error {
	r.log.Info("Iniciando atualização de processo de vendas",
		zap.Int("process_id", id),
		zap.String("operation", "UpdateSalesProcess"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se o processo existe
	var existing models.SalesProcess
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado para atualização", zap.Int("process_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, id)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Atualiza o processo
	process.ID = id
	if err := tx.Model(&existing).Updates(process).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados do processo de vendas", zap.Error(err))
		return fmt.Errorf("falha ao atualizar processo de vendas: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Processo de vendas atualizado com sucesso", zap.Int("process_id", id))
	return nil
}

// DeleteSalesProcess exclui um processo de vendas pelo seu ID
func (r *gormSalesProcessRepository) DeleteSalesProcess(id int) error {
	r.log.Info("Iniciando exclusão de processo de vendas",
		zap.Int("process_id", id),
		zap.String("operation", "DeleteSalesProcess"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Exclui primeiro as relações nas tabelas de ligação
	if err := tx.Table("process_quotations").Where("process_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir relações com cotações", zap.Error(err))
		return fmt.Errorf("falha ao excluir relações com cotações: %w", err)
	}

	if err := tx.Table("process_sales_orders").Where("process_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir relações com pedidos de venda", zap.Error(err))
		return fmt.Errorf("falha ao excluir relações com pedidos de venda: %w", err)
	}

	if err := tx.Table("process_purchase_orders").Where("process_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir relações com pedidos de compra", zap.Error(err))
		return fmt.Errorf("falha ao excluir relações com pedidos de compra: %w", err)
	}

	if err := tx.Table("process_deliveries").Where("process_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir relações com entregas", zap.Error(err))
		return fmt.Errorf("falha ao excluir relações com entregas: %w", err)
	}

	if err := tx.Table("process_invoices").Where("process_id = ?", id).Delete(nil).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir relações com faturas", zap.Error(err))
		return fmt.Errorf("falha ao excluir relações com faturas: %w", err)
	}

	// Exclui o processo
	result := tx.Delete(&models.SalesProcess{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir processo de vendas", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir processo de vendas: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Processo de vendas não encontrado para exclusão", zap.Int("process_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Processo de vendas excluído com sucesso", zap.Int("process_id", id))
	return nil
}

// LinkQuotationToProcess vincula uma cotação a um processo de vendas
func (r *gormSalesProcessRepository) LinkQuotationToProcess(processID int, quotationID int) error {
	r.log.Info("Vinculando cotação ao processo de vendas",
		zap.Int("process_id", processID),
		zap.Int("quotation_id", quotationID),
		zap.String("operation", "LinkQuotationToProcess"),
	)

	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", processID))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, processID)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Verifica se a cotação existe
	quotationRepo, err := NewQuotationRepository()
	if err != nil {
		return fmt.Errorf("erro ao inicializar repositório de cotações: %w", err)
	}

	_, err = quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		r.log.Error("Erro ao verificar existência da cotação", zap.Error(err))
		return fmt.Errorf("erro ao verificar cotação: %w", err)
	}

	// Verifica se já existe um vínculo
	var count int64
	if err := r.db.Table("process_quotations").Where("process_id = ? AND quotation_id = ?", processID, quotationID).Count(&count).Error; err != nil {
		r.log.Error("Erro ao verificar vínculo existente", zap.Error(err))
		return fmt.Errorf("erro ao verificar vínculo existente: %w", err)
	}

	if count > 0 {
		r.log.Info("Vínculo já existe, nenhuma ação necessária")
		return nil
	}

	// Cria o vínculo
	query := `INSERT INTO process_quotations (process_id, quotation_id) VALUES (?, ?)`
	if err := r.db.Exec(query, processID, quotationID).Error; err != nil {
		r.log.Error("Erro ao criar vínculo entre processo e cotação", zap.Error(err))
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	r.log.Info("Cotação vinculada ao processo com sucesso",
		zap.Int("process_id", processID),
		zap.Int("quotation_id", quotationID),
	)

	return nil
}

// LinkSalesOrderToProcess vincula um pedido de venda a um processo de vendas
func (r *gormSalesProcessRepository) LinkSalesOrderToProcess(processID int, salesOrderID int) error {
	r.log.Info("Vinculando pedido de venda ao processo de vendas",
		zap.Int("process_id", processID),
		zap.Int("sales_order_id", salesOrderID),
		zap.String("operation", "LinkSalesOrderToProcess"),
	)

	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", processID))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, processID)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Verifica se o pedido de venda existe
	salesOrderRepo, err := NewSalesOrderRepository()
	if err != nil {
		return fmt.Errorf("erro ao inicializar repositório de pedidos de venda: %w", err)
	}

	_, err = salesOrderRepo.GetSalesOrderByID(salesOrderID)
	if err != nil {
		r.log.Error("Erro ao verificar existência do pedido de venda", zap.Error(err))
		return fmt.Errorf("erro ao verificar pedido de venda: %w", err)
	}

	// Verifica se já existe um vínculo
	var count int64
	if err := r.db.Table("process_sales_orders").Where("process_id = ? AND sales_order_id = ?", processID, salesOrderID).Count(&count).Error; err != nil {
		r.log.Error("Erro ao verificar vínculo existente", zap.Error(err))
		return fmt.Errorf("erro ao verificar vínculo existente: %w", err)
	}

	if count > 0 {
		r.log.Info("Vínculo já existe, nenhuma ação necessária")
		return nil
	}

	// Cria o vínculo
	query := `INSERT INTO process_sales_orders (process_id, sales_order_id) VALUES (?, ?)`
	if err := r.db.Exec(query, processID, salesOrderID).Error; err != nil {
		r.log.Error("Erro ao criar vínculo entre processo e pedido de venda", zap.Error(err))
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	r.log.Info("Pedido de venda vinculado ao processo com sucesso",
		zap.Int("process_id", processID),
		zap.Int("sales_order_id", salesOrderID),
	)

	return nil
}

// LinkPurchaseOrderToProcess vincula um pedido de compra a um processo de vendas
func (r *gormSalesProcessRepository) LinkPurchaseOrderToProcess(processID int, purchaseOrderID int) error {
	r.log.Info("Vinculando pedido de compra ao processo de vendas",
		zap.Int("process_id", processID),
		zap.Int("purchase_order_id", purchaseOrderID),
		zap.String("operation", "LinkPurchaseOrderToProcess"),
	)

	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", processID))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, processID)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Verifica se o pedido de compra existe
	purchaseOrderRepo, err := NewPurchaseOrderRepository()
	if err != nil {
		return fmt.Errorf("erro ao inicializar repositório de pedidos de compra: %w", err)
	}

	_, err = purchaseOrderRepo.GetPurchaseOrderByID(purchaseOrderID)
	if err != nil {
		r.log.Error("Erro ao verificar existência do pedido de compra", zap.Error(err))
		return fmt.Errorf("erro ao verificar pedido de compra: %w", err)
	}

	// Verifica se já existe um vínculo
	var count int64
	if err := r.db.Table("process_purchase_orders").Where("process_id = ? AND purchase_order_id = ?", processID, purchaseOrderID).Count(&count).Error; err != nil {
		r.log.Error("Erro ao verificar vínculo existente", zap.Error(err))
		return fmt.Errorf("erro ao verificar vínculo existente: %w", err)
	}

	if count > 0 {
		r.log.Info("Vínculo já existe, nenhuma ação necessária")
		return nil
	}

	// Cria o vínculo
	query := `INSERT INTO process_purchase_orders (process_id, purchase_order_id) VALUES (?, ?)`
	if err := r.db.Exec(query, processID, purchaseOrderID).Error; err != nil {
		r.log.Error("Erro ao criar vínculo entre processo e pedido de compra", zap.Error(err))
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	r.log.Info("Pedido de compra vinculado ao processo com sucesso",
		zap.Int("process_id", processID),
		zap.Int("purchase_order_id", purchaseOrderID),
	)

	return nil
}

// LinkDeliveryToProcess vincula uma entrega a um processo de vendas
func (r *gormSalesProcessRepository) LinkDeliveryToProcess(processID int, deliveryID int) error {
	r.log.Info("Vinculando entrega ao processo de vendas",
		zap.Int("process_id", processID),
		zap.Int("delivery_id", deliveryID),
		zap.String("operation", "LinkDeliveryToProcess"),
	)

	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", processID))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, processID)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Verifica se a entrega existe
	deliveryRepo, err := NewDeliveryRepository()
	if err != nil {
		return fmt.Errorf("erro ao inicializar repositório de entregas: %w", err)
	}

	_, err = deliveryRepo.GetDeliveryByID(deliveryID)
	if err != nil {
		r.log.Error("Erro ao verificar existência da entrega", zap.Error(err))
		return fmt.Errorf("erro ao verificar entrega: %w", err)
	}

	// Verifica se já existe um vínculo
	var count int64
	if err := r.db.Table("process_deliveries").Where("process_id = ? AND delivery_id = ?", processID, deliveryID).Count(&count).Error; err != nil {
		r.log.Error("Erro ao verificar vínculo existente", zap.Error(err))
		return fmt.Errorf("erro ao verificar vínculo existente: %w", err)
	}

	if count > 0 {
		r.log.Info("Vínculo já existe, nenhuma ação necessária")
		return nil
	}

	// Cria o vínculo
	query := `INSERT INTO process_deliveries (process_id, delivery_id) VALUES (?, ?)`
	if err := r.db.Exec(query, processID, deliveryID).Error; err != nil {
		r.log.Error("Erro ao criar vínculo entre processo e entrega", zap.Error(err))
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	r.log.Info("Entrega vinculada ao processo com sucesso",
		zap.Int("process_id", processID),
		zap.Int("delivery_id", deliveryID),
	)

	return nil
}

// LinkInvoiceToProcess vincula uma fatura a um processo de vendas
func (r *gormSalesProcessRepository) LinkInvoiceToProcess(processID int, invoiceID int) error {
	r.log.Info("Vinculando fatura ao processo de vendas",
		zap.Int("process_id", processID),
		zap.Int("invoice_id", invoiceID),
		zap.String("operation", "LinkInvoiceToProcess"),
	)

	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Processo de vendas não encontrado", zap.Int("process_id", processID))
			return fmt.Errorf("%w: ID %d", errors.ErrSalesProcessNotFound, processID)
		}
		r.log.Error("Erro ao verificar existência do processo de vendas", zap.Error(err))
		return fmt.Errorf("erro ao verificar processo de vendas: %w", err)
	}

	// Verifica se a fatura existe
	invoiceRepo, err := NewInvoiceRepository()
	if err != nil {
		return fmt.Errorf("erro ao inicializar repositório de faturas: %w", err)
	}

	_, err = invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		r.log.Error("Erro ao verificar existência da fatura", zap.Error(err))
		return fmt.Errorf("erro ao verificar fatura: %w", err)
	}

	// Verifica se já existe um vínculo
	var count int64
	if err := r.db.Table("process_invoices").Where("process_id = ? AND invoice_id = ?", processID, invoiceID).Count(&count).Error; err != nil {
		r.log.Error("Erro ao verificar vínculo existente", zap.Error(err))
		return fmt.Errorf("erro ao verificar vínculo existente: %w", err)
	}

	if count > 0 {
		r.log.Info("Vínculo já existe, nenhuma ação necessária")
		return nil
	}

	// Cria o vínculo
	query := `INSERT INTO process_invoices (process_id, invoice_id) VALUES (?, ?)`
	if err := r.db.Exec(query, processID, invoiceID).Error; err != nil {
		r.log.Error("Erro ao criar vínculo entre processo e fatura", zap.Error(err))
		return fmt.Errorf("erro ao criar vínculo: %w", err)
	}

	r.log.Info("Fatura vinculada ao processo com sucesso",
		zap.Int("process_id", processID),
		zap.Int("invoice_id", invoiceID),
	)

	return nil
}

// GetSalesProcessByContact recupera processos de vendas por ID de contato com paginação
func (r *gormSalesProcessRepository) GetSalesProcessByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando processos de vendas por contato",
		zap.Int("contact_id", contactID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetSalesProcessByContact"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesProcess{}).Where("contact_id = ?", contactID).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar processos de vendas por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar processos de vendas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var processes []models.SalesProcess
	if err := r.db.Where("contact_id = ?", contactID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&processes).Error; err != nil {
		r.log.Error("Erro ao buscar processos de vendas por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar processos de vendas: %w", err)
	}

	// Carrega os relacionamentos para cada processo
	for i := range processes {
		if err := r.db.Model(&processes[i]).Association("Contact").Find(&processes[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos processos de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		// Carrega documentos relacionados
		if err := r.loadSalesProcessDocuments(&processes[i]); err != nil {
			r.log.Error("Erro ao carregar documentos relacionados ao processo de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar documentos relacionados: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, processes)

	r.log.Info("Processos de vendas por contato recuperados com sucesso",
		zap.Int("contact_id", contactID),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetSalesProcessByStatus recupera processos de vendas por status com paginação
func (r *gormSalesProcessRepository) GetSalesProcessByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando processos de vendas por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetSalesProcessByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.SalesProcess{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar processos de vendas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar processos de vendas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var processes []models.SalesProcess
	if err := r.db.Where("status = ?", status).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&processes).Error; err != nil {
		r.log.Error("Erro ao buscar processos de vendas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar processos de vendas: %w", err)
	}

	// Carrega os relacionamentos para cada processo
	for i := range processes {
		if err := r.db.Model(&processes[i]).Association("Contact").Find(&processes[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos dos processos de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		// Carrega documentos relacionados
		if err := r.loadSalesProcessDocuments(&processes[i]); err != nil {
			r.log.Error("Erro ao carregar documentos relacionados ao processo de vendas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar documentos relacionados: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, processes)

	r.log.Info("Processos de vendas por status recuperados com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}
