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

// QuotationRepository define a interface para operações de repositório de cotações
type QuotationRepository interface {
	// Operações CRUD básicas
	CreateQuotation(quotation *models.Quotation) error
	GetQuotationByID(id int) (*models.Quotation, error)
	GetAllQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateQuotation(id int, quotation *models.Quotation) error
	DeleteQuotation(id int) error

	// Métodos adicionais específicos
	GetQuotationsByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetExpiredQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// gormQuotationRepository é a implementação concreta usando GORM
type gormQuotationRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	quotationRepoInstance *gormQuotationRepository
	quotationRepoOnce     sync.Once
)

// NewQuotationRepository retorna uma instância do repositório de cotações
func NewQuotationRepository() (QuotationRepository, error) {
	var initErr error

	quotationRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("QuotationRepository")

		quotationRepoInstance = &gormQuotationRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return quotationRepoInstance, nil
}

// NewQuotationRepositoryWithDB creates a repository with a provided DB connection (for testing)
// func NewQuotationRepositoryWithDB(db *gorm.DB, log *zap.Logger) QuotationRepository {
// 	return &gormQuotationRepository{
// 		db:  db,
// 		log: log,
// 	}
// }

// CreateQuotation cria uma nova cotação no banco de dados
func (r *gormQuotationRepository) CreateQuotation(quotation *models.Quotation) error {
	r.log.Info("Iniciando criação de cotação",
		zap.Int("contact_id", quotation.ContactID),
		zap.String("operation", "CreateQuotation"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Define valores padrão se não fornecidos
	if quotation.Status == "" {
		quotation.Status = models.QuotationStatusDraft
	}

	// Preservar os itens em uma variável temporária
	items := quotation.Items

	// Remover os itens antes de criar a cotação
	quotation.Items = nil

	// Criar a cotação sem os itens
	if err := tx.Create(quotation).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar cotação", zap.Error(err))
		return fmt.Errorf("falha ao criar cotação: %w", err)
	}

	// Agora criar cada item separadamente, definindo o ID da cotação
	for i, item := range items {
		newItem := item
		newItem.ID = 0
		newItem.QuotationID = quotation.ID

		if err := tx.Create(&newItem).Error; err != nil {
			tx.Rollback()
			r.log.Error("Falha ao criar item da cotação",
				zap.Int("quotation_id", quotation.ID),
				zap.Int("item_index", i),
				zap.Error(err),
			)
			return fmt.Errorf("falha ao criar item da cotação: %w", err)
		}
	}

	// Restaurar os itens para a cotação
	if err := tx.Where("quotation_id = ?", quotation.ID).Find(&quotation.Items).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao recuperar itens da cotação", zap.Error(err))
		return fmt.Errorf("falha ao recuperar itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Cotação criada com sucesso",
		zap.Int("quotation_id", quotation.ID),
		zap.String("status", quotation.Status),
	)

	return nil
}

// GetQuotationByID recupera uma cotação pelo seu ID
func (r *gormQuotationRepository) GetQuotationByID(id int) (*models.Quotation, error) {
	r.log.Info("Buscando cotação por ID",
		zap.Int("quotation_id", id),
		zap.String("operation", "GetQuotationByID"),
	)

	var quotation models.Quotation
	if err := r.db.First(&quotation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Cotação não encontrada", zap.Int("quotation_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrQuotationNotFound, id)
		}
		r.log.Error("Erro ao buscar cotação", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar cotação: %w", err)
	}

	// Carrega os itens
	if err := r.db.Model(&quotation).Association("Items").Find(&quotation.Items); err != nil {
		r.log.Error("Erro ao carregar itens da cotação", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar itens: %w", err)
	}

	// Carrega informações do contato
	if err := r.db.Model(&quotation).Association("Contact").Find(&quotation.Contact); err != nil {
		r.log.Error("Erro ao carregar contato da cotação", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar contato: %w", err)
	}

	r.log.Info("Cotação recuperada com sucesso", zap.Int("quotation_id", id))
	return &quotation, nil
}

// GetAllQuotations recupera todas as cotações do banco de dados com paginação
func (r *gormQuotationRepository) GetAllQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando cotações com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllQuotations"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Quotation{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de cotações", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar cotações: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var quotations []models.Quotation
	if err := r.db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&quotations).Error; err != nil {
		r.log.Error("Erro ao buscar cotações paginadas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar cotações: %w", err)
	}

	// Carrega os relacionamentos para cada cotação
	for i := range quotations {
		if err := r.db.Model(&quotations[i]).Association("Items").Find(&quotations[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&quotations[i]).Association("Contact").Find(&quotations[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, quotations)

	r.log.Info("Cotações recuperadas com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdateQuotation atualiza uma cotação existente
func (r *gormQuotationRepository) UpdateQuotation(id int, quotation *models.Quotation) error {
	r.log.Info("Iniciando atualização de cotação",
		zap.Int("quotation_id", id),
		zap.String("operation", "UpdateQuotation"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se a cotação existe
	var existing models.Quotation
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Cotação não encontrada para atualização", zap.Int("quotation_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrQuotationNotFound, id)
		}
		r.log.Error("Erro ao verificar existência da cotação", zap.Error(err))
		return fmt.Errorf("erro ao verificar cotação: %w", err)
	}

	// Atualiza a cotação
	quotation.ID = id
	if err := tx.Model(&existing).Updates(quotation).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar dados básicos da cotação", zap.Error(err))
		return fmt.Errorf("falha ao atualizar cotação: %w", err)
	}

	// Deleta os itens existentes
	if err := tx.Where("quotation_id = ?", id).Delete(&models.QuotationItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao remover itens existentes", zap.Error(err))
		return fmt.Errorf("falha ao remover itens: %w", err)
	}

	// Define o ID da cotação para cada item
	for i := range quotation.Items {
		quotation.Items[i].QuotationID = id
		quotation.Items[i].ID = 0 // Redefine o ID para criar novos itens
	}

	// Cria os novos itens
	if err := tx.CreateInBatches(quotation.Items, 100).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar novos itens", zap.Error(err))
		return fmt.Errorf("falha ao criar novos itens: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Cotação atualizada com sucesso", zap.Int("quotation_id", id))
	return nil
}

// DeleteQuotation exclui uma cotação pelo seu ID
func (r *gormQuotationRepository) DeleteQuotation(id int) error {
	r.log.Info("Iniciando exclusão de cotação",
		zap.Int("quotation_id", id),
		zap.String("operation", "DeleteQuotation"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se existem pedidos de venda associados
	var count int64
	if err := tx.Model(&models.SalesOrder{}).Where("quotation_id = ?", id).Count(&count).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao verificar pedidos de venda associados", zap.Error(err))
		return fmt.Errorf("falha ao verificar pedidos de venda: %w", err)
	}

	if count > 0 {
		tx.Rollback()
		r.log.Warn("Cotação possui pedidos de venda associados", zap.Int("quotation_id", id))
		return fmt.Errorf("%w: cotação possui %d pedidos de venda associados", errors.ErrRelatedRecordsExist, count)
	}

	// Exclui os itens primeiro
	if err := tx.Where("quotation_id = ?", id).Delete(&models.QuotationItem{}).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir itens da cotação", zap.Error(err))
		return fmt.Errorf("falha ao excluir itens: %w", err)
	}

	// Exclui a cotação
	result := tx.Delete(&models.Quotation{}, id)
	if result.Error != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir cotação", zap.Error(result.Error))
		return fmt.Errorf("falha ao excluir cotação: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		r.log.Warn("Cotação não encontrada para exclusão", zap.Int("quotation_id", id))
		return fmt.Errorf("%w: ID %d", errors.ErrQuotationNotFound, id)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Cotação excluída com sucesso", zap.Int("quotation_id", id))
	return nil
}

// GetQuotationsByStatus recupera cotações por status com paginação
func (r *gormQuotationRepository) GetQuotationsByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando cotações por status",
		zap.String("status", status),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetQuotationsByStatus"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Quotation{}).Where("status = ?", status).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar cotações por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar cotações: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var quotations []models.Quotation
	if err := r.db.Where("status = ?", status).
		Order("expiry_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&quotations).Error; err != nil {
		r.log.Error("Erro ao buscar cotações por status", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar cotações: %w", err)
	}

	// Carrega os relacionamentos para cada cotação
	for i := range quotations {
		if err := r.db.Model(&quotations[i]).Association("Items").Find(&quotations[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&quotations[i]).Association("Contact").Find(&quotations[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, quotations)

	r.log.Info("Cotações por status recuperadas com sucesso",
		zap.String("status", status),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetQuotationsByContact recupera cotações por ID de contato com paginação
func (r *gormQuotationRepository) GetQuotationsByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando cotações por contato",
		zap.Int("contact_id", contactID),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetQuotationsByContact"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Quotation{}).Where("contact_id = ?", contactID).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar cotações por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar cotações: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var quotations []models.Quotation
	if err := r.db.Where("contact_id = ?", contactID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&quotations).Error; err != nil {
		r.log.Error("Erro ao buscar cotações por contato", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar cotações: %w", err)
	}

	// Carrega os relacionamentos para cada cotação
	for i := range quotations {
		if err := r.db.Model(&quotations[i]).Association("Items").Find(&quotations[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&quotations[i]).Association("Contact").Find(&quotations[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, quotations)

	r.log.Info("Cotações por contato recuperadas com sucesso",
		zap.Int("contact_id", contactID),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// GetExpiredQuotations recupera cotações expiradas com paginação
func (r *gormQuotationRepository) GetExpiredQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

	r.log.Info("Buscando cotações expiradas",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetExpiredQuotations"),
	)

	today := time.Now()

	var totalItems int64
	if err := r.db.Model(&models.Quotation{}).
		Where("expiry_date < ? AND status NOT IN (?, ?)",
			today,
			models.QuotationStatusAccepted,
			models.QuotationStatusRejected).
		Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar cotações expiradas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar cotações expiradas: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var quotations []models.Quotation
	if err := r.db.Where("expiry_date < ? AND status NOT IN (?, ?)",
		today,
		models.QuotationStatusAccepted,
		models.QuotationStatusRejected).
		Order("expiry_date ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&quotations).Error; err != nil {
		r.log.Error("Erro ao buscar cotações expiradas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar cotações expiradas: %w", err)
	}

	// Carrega os relacionamentos para cada cotação
	for i := range quotations {
		if err := r.db.Model(&quotations[i]).Association("Items").Find(&quotations[i].Items); err != nil {
			r.log.Error("Erro ao carregar itens das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar itens: %w", err)
		}

		if err := r.db.Model(&quotations[i]).Association("Contact").Find(&quotations[i].Contact); err != nil {
			r.log.Error("Erro ao carregar contatos das cotações", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar contatos: %w", err)
		}
	}

	// Atualiza o status das cotações expiradas para "expired" se ainda não estiverem
	for i := range quotations {
		if quotations[i].Status != models.QuotationStatusExpired {
			if err := r.db.Model(&quotations[i]).Update("status", models.QuotationStatusExpired).Error; err != nil {
				r.log.Error("Erro ao atualizar status da cotação para expirada",
					zap.Int("quotation_id", quotations[i].ID),
					zap.Error(err),
				)
				// Não interrompe o processo, apenas loga o erro
			} else {
				quotations[i].Status = models.QuotationStatusExpired
			}
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, quotations)

	r.log.Info("Cotações expiradas recuperadas com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}
