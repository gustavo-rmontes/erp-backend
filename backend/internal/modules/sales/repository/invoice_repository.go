package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// InvoiceRepository define a interface para operações de repositório de faturas
type InvoiceRepository interface {
	// Operações CRUD básicas
	CreateInvoice(invoice *models.Invoice) error
	GetInvoiceByID(id int) (*models.Invoice, error)
	GetAllInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateInvoice(id int, invoice *models.Invoice) error
	DeleteInvoice(id int) error

	// Métodos adicionais específicos
	GetInvoicesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoicesBySalesOrder(salesOrderID int) ([]models.Invoice, error)
	GetInvoicesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverdueInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// gormInvoiceRepository é a implementação concreta usando GORM
type gormInvoiceRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	invoiceRepoInstance *gormInvoiceRepository
	invoiceRepoOnce     sync.Once
)

// NewInvoiceRepository retorna uma instância do repositório de faturas
func NewInvoiceRepository() (InvoiceRepository, error) {
	var initErr error

	invoiceRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("InvoiceRepository")

		invoiceRepoInstance = &gormInvoiceRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return invoiceRepoInstance, nil
}

// CreateInvoice cria uma nova fatura no banco de dados
func (r *gormInvoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	r.log.Info("Iniciando criação de fatura",
		zap.Int("contact_id", invoice.ContactID),
		zap.Int("sales_order_id", invoice.SalesOrderID),
		zap.String("operation", "CreateInvoice"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Define valores padrão se não fornecidos
	if invoice.Status == "" {
		invoice.Status = models.InvoiceStatusDraft
	}

	// Preservar os itens em uma variável temporária
	items := invoice.Items

	// Remover os itens antes de criar a fatura
	invoice.Items = nil

	// Criar a fatura sem os itens
	if err := tx.Create(invoice).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar fatura", zap.Error(err))
		return fmt.Errorf("falha ao criar fatura: %w", err)
	}

	// Agora criar cada item separadamente, definindo o ID da fatura
	for i, item := range items {
		newItem := item
		newItem.ID = 0
		newItem.InvoiceID = invoice.ID

		if err := tx.Create(&newItem).Error; err != nil {
			tx.Rollback()
			r.log.Error("Falha ao criar item da fatura",
				zap.Int("invoice_id", invoice.ID),
				zap.Int("item_index", i),
				zap.Error(err),
			)
			return fmt.Errorf("falha ao criar item da fatura: %w", err)
		}
	}

	// Restaurar os itens para a fatura
	if err := tx.Where("invoice_id = ?", invoice.ID).Find(&invoice.Items).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao recuperar itens da fatura", zap.Error(err))
		return fmt.Errorf("falha ao recuperar itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Fatura criada com sucesso",
		zap.Int("invoice_id", invoice.ID),
		zap.String("status", invoice.Status),
	)

	return nil
}

// GetInvoiceByID recupera uma fatura pelo seu ID
func (r *gormInvoiceRepository) GetInvoiceByID(id int) (*models.Invoice, error) {
	r.log.Info("Buscando fatura por ID",
		zap.Int("invoice_id", id),
		zap.String("operation", "GetInvoiceByID"),
	)

	var invoice models.Invoice
	if err := r.db.First(&invoice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Fatura não encontrada", zap.Int("invoice_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrInvoiceNotFound, id)
		}
		r.log.Error("Erro ao buscar fatura", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar fatura: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&invoice).Association("Items").Find(&invoice.Items); err != nil {
		r.log.Error("Erro ao carregar itens da fatura", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	// Carrega informações do contato
	if err := r.db.Model(&invoice).Association("Contact").Find(&invoice.Contact); err != nil {
		r.log.Error("Erro ao carregar contato da fatura", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar contato: %w", err)
	}

	// Carrega informações do pedido de venda se existir
	if invoice.SalesOrderID > 0 {
		if err := r.db.Model(&invoice).Association("SalesOrder").Find(&invoice.SalesOrder); err != nil {
			r.log.Error("Erro ao carregar pedido de venda da fatura", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pedido de venda: %w", err)
		}
	}

	// Carrega pagamentos
	if err := r.db.Model(&invoice).Association("Payments").Find(&invoice.Payments); err != nil {
		r.log.Error("Erro ao carregar pagamentos da fatura", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
	}

	r.log.Info("Fatura recuperada com sucesso", zap.Int("invoice_id", id))
	return &invoice, nil
}

// GetAllInvoices recupera todas as faturas do banco de dados com paginação
func (r *gormInvoiceRepository) GetAllInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando faturas com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllInvoices"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Invoice{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de faturas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar faturas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var invoices []models.Invoice
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&invoices).Error; err != nil {
		r.log.Error("Erro ao buscar faturas paginadas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar faturas: %w", err)
	}

	// Carrega os relacionamentos para cada fatura
	for i := range invoices {
		if err := r.db.Model(&invoices[i]).Association("Items").Find(&invoices[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Contact").Find(&invoices[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if invoices[i].SalesOrderID > 0 {
			if err := r.db.Model(&invoices[i]).Association("SalesOrder").Find(&invoices[i].SalesOrder); err != nil {
				r.log.Error("Erro ao carregar pedidos de venda das faturas", zap.Error(err))
				return nil, fmt.Errorf("erro ao carregar pedidos de venda: %w", err)
			}
		}

		if err := r.db.Model(&invoices[i]).Association("Payments").Find(&invoices[i].Payments); err != nil {
			r.log.Error("Erro ao carregar pagamentos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, invoices)

	r.log.Info("Faturas recuperadas com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdateInvoice atualiza uma fatura existente
func (r *gormInvoiceRepository) UpdateInvoice(id int, invoice *models.Invoice) error {
	r.log.Info("Iniciando atualização de fatura",
		zap.Int("invoice_id", id),
		zap.String("operation", "UpdateInvoice"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se a fatura existe
	var existing models.Invoice
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Fatura não encontrada para atualização", zap.Int("invoice_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrInvoiceNotFound, id)
		}
		r.log.Error("Erro ao verificar existência da fatura", zap.Error(err))
		return fmt.Errorf("erro ao verificar fatura: %w", err)
	}

	// Preserva o valor pago se não for fornecido
	if invoice.AmountPaid == 0 {
		invoice.AmountPaid = existing.AmountPaid
	}

	// Atualiza a fatura
	invoice.ID = id
	if err := tx.Model(&existing).Updates(invoice).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados básicos da fatura", zap.Error(err))
		return fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	// Deleta os itens existentes
	if err := tx.Where("invoice_id = ?", id).Delete(&models.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao remover itens existentes", zap.Error(err))
		return fmt.Errorf("falha ao remover itens: %w", err)
	}

	// Define o ID da fatura para cada item
	for i := range invoice.Items {
		invoice.Items[i].InvoiceID = id
		invoice.Items[i].ID = 0 // Redefine o ID para criar novos itens
	}

	// Cria os novos itens
	if err := tx.CreateInBatches(invoice.Items, 100).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar novos itens", zap.Error(err))
		return fmt.Errorf("falha ao criar novos itens: %w", err)
	}

	// Atualiza o status da fatura com base no valor pago
	// Somente se o status não foi explicitamente definido
	if invoice.Status == existing.Status {
		if invoice.AmountPaid >= invoice.GrandTotal {
			invoice.Status = models.InvoiceStatusPaid
			if err := tx.Model(&existing).Update("status", invoice.Status).Error; err != nil {
				tx.Rollback()
				r.log.Error("Falha ao atualizar status da fatura", zap.Error(err))
				return fmt.Errorf("falha ao atualizar status: %w", err)
			}
		} else if invoice.AmountPaid > 0 {
			invoice.Status = models.InvoiceStatusPartial
			if err := tx.Model(&existing).Update("status", invoice.Status).Error; err != nil {
				tx.Rollback()
				r.log.Error("Falha ao atualizar status da fatura", zap.Error(err))
				return fmt.Errorf("falha ao atualizar status: %w", err)
			}
		}
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Fatura atualizada com sucesso", zap.Int("invoice_id", id))
	return nil
}

// DeleteInvoice exclui uma fatura pelo seu ID
func (r *gormInvoiceRepository) DeleteInvoice(id int) error {
	r.log.Info("Iniciando exclusão de fatura",
		zap.Int("invoice_id", id),
		zap.String("operation", "DeleteInvoice"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se existem pagamentos associados
	var paymentsCount int64
	if err := tx.Model(&models.Payment{}).Where("invoice_id = ?", id).Count(&paymentsCount).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao verificar pagamentos associados", zap.Error(err))
		return fmt.Errorf("falha ao verificar pagamentos: %w", err)
	}

	if paymentsCount > 0 {
		tx.Rollback()
		r.log.Warn("Não é possível excluir fatura com pagamentos associados",
			zap.Int("invoice_id", id),
			zap.Int64("payments_count", paymentsCount),
		)
		return fmt.Errorf("%w: fatura possui %d pagamentos associados", errors.ErrRelatedRecordsExist, paymentsCount)
	}

	// Exclui os itens primeiro
	if err := tx.Where("invoice_id = ?", id).Delete(&models.InvoiceItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir itens da fatura", zap.Error(err))
		return fmt.Errorf("falha ao excluir itens: %w", err)
	}

	// Exclui a fatura
	result := tx.Delete(&models.Invoice{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir fatura", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir fatura: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Fatura não encontrada para exclusão", zap.Int("invoice_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrInvoiceNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Fatura excluída com sucesso", zap.Int("invoice_id", id))
	return nil
}

// GetInvoicesByStatus recupera faturas por status com paginação
func (r *gormInvoiceRepository) GetInvoicesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando faturas por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetInvoicesByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Invoice{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar faturas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar faturas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var invoices []models.Invoice
	if err := r.db.Where("status = ?", status).
		Order("due_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&invoices).Error; err != nil {
		r.log.Error("Erro ao buscar faturas por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar faturas: %w", err)
	}

	// Carrega os relacionamentos para cada fatura
	for i := range invoices {
		if err := r.db.Model(&invoices[i]).Association("Items").Find(&invoices[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Contact").Find(&invoices[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Payments").Find(&invoices[i].Payments); err != nil {
			r.log.Error("Erro ao carregar pagamentos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, invoices)

	r.log.Info("Faturas por status recuperadas com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetInvoicesBySalesOrder recupera faturas por ID de pedido de venda
func (r *gormInvoiceRepository) GetInvoicesBySalesOrder(salesOrderID int) ([]models.Invoice, error) {
	r.log.Info("Buscando faturas por pedido de venda",
		zap.Int("sales_order_id", salesOrderID),
		zap.String("operation", "GetInvoicesBySalesOrder"),
	)

	var invoices []models.Invoice
	if err := r.db.Where("sales_order_id = ?", salesOrderID).
		Order("created_at DESC").
		Find(&invoices).Error; err != nil {
		r.log.Error("Erro ao buscar faturas por pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar faturas por pedido de venda: %w", err)
	}

	// Carrega os relacionamentos para cada fatura
	for i := range invoices {
		if err := r.db.Model(&invoices[i]).Association("Items").Find(&invoices[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Payments").Find(&invoices[i].Payments); err != nil {
			r.log.Error("Erro ao carregar pagamentos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
		}
	}

	r.log.Info("Faturas por pedido de venda recuperadas com sucesso",
		zap.Int("sales_order_id", salesOrderID),
		zap.Int("count", len(invoices)),
	)

	return invoices, nil
}

// GetInvoicesByContact recupera faturas por ID de contato com paginação
func (r *gormInvoiceRepository) GetInvoicesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando faturas por contato",
		zap.Int("contact_id", contactID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetInvoicesByContact"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Invoice{}).Where("contact_id = ?", contactID).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar faturas por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar faturas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var invoices []models.Invoice
	if err := r.db.Where("contact_id = ?", contactID).
		Order("due_date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&invoices).Error; err != nil {
		r.log.Error("Erro ao buscar faturas por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar faturas: %w", err)
	}

	// Carrega os relacionamentos para cada fatura
	for i := range invoices {
		if err := r.db.Model(&invoices[i]).Association("Items").Find(&invoices[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Contact").Find(&invoices[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Payments").Find(&invoices[i].Payments); err != nil {
			r.log.Error("Erro ao carregar pagamentos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, invoices)

	r.log.Info("Faturas por contato recuperadas com sucesso",
		zap.Int("contact_id", contactID),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetOverdueInvoices recupera faturas vencidas com paginação
func (r *gormInvoiceRepository) GetOverdueInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando faturas vencidas",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetOverdueInvoices"),
	)

	today := time.Now()

	var totalItems int64
	if err := r.db.Model(&models.Invoice{}).
		Where("due_date < ? AND status NOT IN (?, ?)",
			today,
			models.InvoiceStatusPaid,
			models.InvoiceStatusCancelled).
		Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar faturas vencidas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar faturas vencidas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var invoices []models.Invoice
	if err := r.db.Where("due_date < ? AND status NOT IN (?, ?)",
		today,
		models.InvoiceStatusPaid,
		models.InvoiceStatusCancelled).
		Order("due_date ASC"). // Ordena pelo vencimento mais antigo primeiro
		Offset(offset).
		Limit(pageSize).
		Find(&invoices).Error; err != nil {
		r.log.Error("Erro ao buscar faturas vencidas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar faturas vencidas: %w", err)
	}

	// Carrega os relacionamentos para cada fatura
	for i := range invoices {
		if err := r.db.Model(&invoices[i]).Association("Items").Find(&invoices[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Contact").Find(&invoices[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}

		if err := r.db.Model(&invoices[i]).Association("Payments").Find(&invoices[i].Payments); err != nil {
			r.log.Error("Erro ao carregar pagamentos das faturas", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar pagamentos: %w", err)
		}
	}

	// Atualiza o status das faturas vencidas para "overdue" se ainda não estiverem
	for i := range invoices {
		if invoices[i].Status != models.InvoiceStatusOverdue {
			if err := r.db.Model(&invoices[i]).Update("status", models.InvoiceStatusOverdue).Error; err != nil {
				r.log.Error("Erro ao atualizar status da fatura para vencida",
					zap.Int("invoice_id", invoices[i].ID),
					zap.Error(err),
				)
				// Não interrompe o processo, apenas loga o erro
			} else {
				invoices[i].Status = models.InvoiceStatusOverdue
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, invoices)

	r.log.Info("Faturas vencidas recuperadas com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}
