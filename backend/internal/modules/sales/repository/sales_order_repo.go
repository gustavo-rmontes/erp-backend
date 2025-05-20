package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SalesOrderRepository define as operações do repositório de sales orders
type SalesOrderRepository interface {
	// CRUD básico
	CreateSalesOrder(salesOrder *models.SalesOrder) error
	GetSalesOrderByID(id int) (*models.SalesOrder, error)
	UpdateSalesOrder(id int, salesOrder *models.SalesOrder) error
	DeleteSalesOrder(id int) error

	// Consultas com paginação
	GetAllSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByQuotation(quotationID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByExpectedDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada (opcional, considere mover para serviço se contiver muita lógica de negócio)
	SearchSalesOrders(filter SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// SalesOrderFilter define os filtros para busca avançada
type SalesOrderFilter struct {
	Status            []string
	ContactID         int
	ContactType       string // cliente, fornecedor, lead
	PersonType        string // pf, pj
	DateRangeStart    time.Time
	DateRangeEnd      time.Time
	ExpectedDateStart time.Time
	ExpectedDateEnd   time.Time
	MinAmount         float64
	MaxAmount         float64
	HasInvoice        *bool
	HasPurchaseOrder  *bool
	SearchQuery       string
}

// SalesOrderStats representa estatísticas de sales orders
type SalesOrderStats struct {
	TotalOrders     int            `json:"total_orders"`
	TotalValue      float64        `json:"total_value"`
	TotalConfirmed  float64        `json:"total_confirmed"`
	TotalProcessing float64        `json:"total_processing"`
	TotalCompleted  float64        `json:"total_completed"`
	TotalCancelled  float64        `json:"total_cancelled"`
	CountByStatus   map[string]int `json:"count_by_status"`
	FulfillmentRate float64        `json:"fulfillment_rate"`
}

// ContactSalesOrdersSummary representa um resumo dos sales orders de um contato
type ContactSalesOrdersSummary struct {
	ContactID       int       `json:"contact_id"`
	ContactName     string    `json:"contact_name"`
	ContactType     string    `json:"contact_type"`
	TotalOrders     int       `json:"total_orders"`
	TotalValue      float64   `json:"total_value"`
	TotalCompleted  float64   `json:"total_completed"`
	TotalCancelled  float64   `json:"total_cancelled"`
	PendingCount    int       `json:"pending_count"`
	PendingValue    float64   `json:"pending_value"`
	FulfillmentRate float64   `json:"fulfillment_rate"`
	LastOrderDate   time.Time `json:"last_order_date"`
}

type salesOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSalesOrderRepository cria uma nova instância do repositório
func NewSalesOrderRepository() (SalesOrderRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &salesOrderRepository{
		db:     db,
		logger: logger.WithModule("sales_order_repository"),
	}, nil
}

// CreateSalesOrder cria um novo sales order no banco
func (r *salesOrderRepository) CreateSalesOrder(salesOrder *models.SalesOrder) error {
	// Gera o número do sales order se não foi fornecido
	if salesOrder.SONo == "" {
		salesOrder.SONo = r.generateSalesOrderNumber()
	}

	// Define status padrão se não foi fornecido
	if salesOrder.Status == "" {
		salesOrder.Status = models.SOStatusDraft
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria o sales order
	if err := tx.Create(salesOrder).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar sales order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar sales order")
	}

	// Se houver itens, cria os itens
	if len(salesOrder.Items) > 0 {
		for i := range salesOrder.Items {
			salesOrder.Items[i].SalesOrderID = salesOrder.ID
			if err := tx.Create(&salesOrder.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item do sales order", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d do sales order", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("sales order criado com sucesso", zap.Int("id", salesOrder.ID), zap.String("so_no", salesOrder.SONo))
	return nil
}

// GetSalesOrderByID busca um sales order pelo ID
func (r *salesOrderRepository) GetSalesOrderByID(id int) (*models.SalesOrder, error) {
	var salesOrder models.SalesOrder

	query := r.db.Preload("Contact").
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

// GetAllSalesOrders retorna todos os sales orders com paginação
func (r *salesOrderRepository) GetAllSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	// Query base
	query := r.db.Model(&models.SalesOrder{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales orders")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// UpdateSalesOrder atualiza um sales order existente
func (r *salesOrderRepository) UpdateSalesOrder(id int, salesOrder *models.SalesOrder) error {
	// Verifica se o sales order existe
	var existing models.SalesOrder
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesOrderNotFound
		}
		return errors.WrapError(err, "falha ao verificar sales order existente")
	}

	// Atualiza os campos
	salesOrder.ID = id
	if err := r.db.Save(salesOrder).Error; err != nil {
		r.logger.Error("erro ao atualizar sales order", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar sales order")
	}

	r.logger.Info("sales order atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeleteSalesOrder remove um sales order
func (r *salesOrderRepository) DeleteSalesOrder(id int) error {
	// Verifica se existem invoices ou purchase orders relacionados
	var invoiceCount int64
	if err := r.db.Model(&models.Invoice{}).Where("sales_order_id = ?", id).Count(&invoiceCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar invoices relacionadas")
	}

	if invoiceCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	var poCount int64
	if err := r.db.Model(&models.PurchaseOrder{}).Where("sales_order_id = ?", id).Count(&poCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar purchase orders relacionadas")
	}

	if poCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove o sales order (cascade removerá os itens)
	result := r.db.Delete(&models.SalesOrder{}, id)
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

// GetSalesOrdersByStatus busca sales orders por status
func (r *salesOrderRepository) GetSalesOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	query := r.db.Model(&models.SalesOrder{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar sales orders por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// GetSalesOrdersByContact busca sales orders por contato
func (r *salesOrderRepository) GetSalesOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	query := r.db.Model(&models.SalesOrder{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar sales orders por contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// GetSalesOrdersByQuotation busca sales orders por quotation
func (r *salesOrderRepository) GetSalesOrdersByQuotation(quotationID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	query := r.db.Model(&models.SalesOrder{}).Where("quotation_id = ?", quotationID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por quotation", zap.Error(err), zap.Int("quotation_id", quotationID))
		return nil, errors.WrapError(err, "falha ao contar sales orders por quotation")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Quotation").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por quotation", zap.Error(err), zap.Int("quotation_id", quotationID))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por quotation")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// GetSalesOrdersByPeriod busca sales orders por período (usando created_at)
func (r *salesOrderRepository) GetSalesOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	query := r.db.Model(&models.SalesOrder{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// GetSalesOrdersByExpectedDate busca sales orders por data esperada
func (r *salesOrderRepository) GetSalesOrdersByExpectedDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	query := r.db.Model(&models.SalesOrder{}).
		Where("expected_date >= ? AND expected_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por data esperada", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders por data esperada")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("expected_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por data esperada", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por data esperada")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// SearchSalesOrders busca sales orders com filtros combinados
func (r *salesOrderRepository) SearchSalesOrders(filter SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	// Inicia a query base
	query := r.db.Model(&models.SalesOrder{})

	// Aplica os diversos filtros usando métodos auxiliares
	query = r.applyStatusFilter(query, filter)
	query = r.applyContactFilter(query, filter)
	query = r.applyDateRangeFilters(query, filter)
	query = r.applyAmountFilters(query, filter)
	query = r.applyRelatedEntityFilters(query, filter)
	query = r.applyTextSearchFilter(query, filter)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales orders")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// Método auxiliar para filtrar por status
func (r *salesOrderRepository) applyStatusFilter(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	if len(filter.Status) > 0 {
		return query.Where("status IN ?", filter.Status)
	}
	return query
}

// Método auxiliar para filtrar por contacto e tipo de pessoa
func (r *salesOrderRepository) applyContactFilter(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	// Filtro simples de ID do contato
	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	// Filtro por tipo de contato ou pessoa
	if filter.ContactType != "" || filter.PersonType != "" {
		var contactIDs []int
		contactQuery := r.db.Model(&contact.Contact{})

		if filter.ContactType != "" {
			contactQuery = contactQuery.Where("type = ?", filter.ContactType)
		}
		if filter.PersonType != "" {
			contactQuery = contactQuery.Where("person_type = ?", filter.PersonType)
		}

		contactQuery.Pluck("id", &contactIDs)

		if len(contactIDs) > 0 {
			query = query.Where("contact_id IN ?", contactIDs)
		}
	}

	return query
}

// Método auxiliar para filtrar por datas
func (r *salesOrderRepository) applyDateRangeFilters(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	// Filtro de data de criação
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?",
			filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Filtro de data esperada
	if !filter.ExpectedDateStart.IsZero() && !filter.ExpectedDateEnd.IsZero() {
		query = query.Where("expected_date >= ? AND expected_date <= ?",
			filter.ExpectedDateStart, filter.ExpectedDateEnd)
	}

	return query
}

// Método auxiliar para filtrar por valores monetários
func (r *salesOrderRepository) applyAmountFilters(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	if filter.MinAmount > 0 {
		query = query.Where("grand_total >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("grand_total <= ?", filter.MaxAmount)
	}

	return query
}

// Método auxiliar genérico para aplicar filtros de entidades relacionadas
func (r *salesOrderRepository) getRelatedOrderIDs(entityType string, hasRelation bool) ([]int, error) {
	var tableName string

	switch entityType {
	case "invoice":
		tableName = "invoices"
	case "purchase_order":
		tableName = "purchase_orders"
	default:
		return nil, fmt.Errorf("tipo de entidade não suportado: %s", entityType)
	}

	var orderIDs []int
	query := r.db.Table(tableName).
		Distinct("sales_order_id").
		Where("sales_order_id IS NOT NULL")

	if err := query.Pluck("sales_order_id", &orderIDs).Error; err != nil {
		return nil, err
	}

	return orderIDs, nil
}

// Método auxiliar para filtrar por relações com outras entidades
func (r *salesOrderRepository) applyRelatedEntityFilters(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	// Filtro de invoice
	if filter.HasInvoice != nil {
		if orderIDs, err := r.getRelatedOrderIDs("invoice", true); err == nil && len(orderIDs) > 0 {
			if *filter.HasInvoice {
				query = query.Where("id IN ?", orderIDs)
			} else {
				query = query.Where("id NOT IN ?", orderIDs)
			}
		}
	}

	// Filtro de purchase order
	if filter.HasPurchaseOrder != nil {
		if orderIDs, err := r.getRelatedOrderIDs("purchase_order", true); err == nil && len(orderIDs) > 0 {
			if *filter.HasPurchaseOrder {
				query = query.Where("id IN ?", orderIDs)
			} else {
				query = query.Where("id NOT IN ?", orderIDs)
			}
		}
	}

	return query
}

// Método auxiliar para busca textual
func (r *salesOrderRepository) applyTextSearchFilter(query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"

		// Fazemos um join com contatos para buscar também nos campos de contato
		query = query.Joins("LEFT JOIN contacts ON contacts.id = sales_orders.contact_id").
			Where("sales_orders.so_no LIKE ? OR sales_orders.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
	}

	return query
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
