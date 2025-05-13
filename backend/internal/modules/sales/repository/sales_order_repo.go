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
	CreateSalesOrder(salesOrder *models.SalesOrder) error
	GetSalesOrderByID(id int) (*models.SalesOrder, error)
	GetAllSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateSalesOrder(id int, salesOrder *models.SalesOrder) error
	DeleteSalesOrder(id int) error
	GetSalesOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByQuotation(quotationID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrdersByExpectedDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchSalesOrders(filter SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesOrderStats(filter SalesOrderFilter) (*SalesOrderStats, error)
	GetContactSalesOrdersSummary(contactID int) (*ContactSalesOrdersSummary, error)
	GetSalesOrdersByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	CreateInvoiceFromSalesOrder(salesOrderID int) error
	CreatePurchaseOrderFromSalesOrder(salesOrderID int) error
	GetPendingSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
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

	query := r.db.Model(&models.SalesOrder{})

	// Aplica os filtros
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	// Filtro por tipo de contato ou pessoa
	if filter.ContactType != "" || filter.PersonType != "" {
		contactQuery := r.db.Model(&contact.Contact{})
		if filter.ContactType != "" {
			contactQuery = contactQuery.Where("type = ?", filter.ContactType)
		}
		if filter.PersonType != "" {
			contactQuery = contactQuery.Where("person_type = ?", filter.PersonType)
		}
		var contactIDs []int
		contactQuery.Pluck("id", &contactIDs)
		if len(contactIDs) > 0 {
			query = query.Where("contact_id IN ?", contactIDs)
		}
	}

	// Filtros de data
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	if !filter.ExpectedDateStart.IsZero() && !filter.ExpectedDateEnd.IsZero() {
		query = query.Where("expected_date >= ? AND expected_date <= ?", filter.ExpectedDateStart, filter.ExpectedDateEnd)
	}

	// Filtros de valor
	if filter.MinAmount > 0 {
		query = query.Where("grand_total >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("grand_total <= ?", filter.MaxAmount)
	}

	// Filtro de invoice
	if filter.HasInvoice != nil {
		if *filter.HasInvoice {
			var soIDs []int
			r.db.Model(&models.Invoice{}).Distinct("sales_order_id").Where("sales_order_id IS NOT NULL").Pluck("sales_order_id", &soIDs)
			if len(soIDs) > 0 {
				query = query.Where("id IN ?", soIDs)
			}
		} else {
			var soIDs []int
			r.db.Model(&models.Invoice{}).Distinct("sales_order_id").Where("sales_order_id IS NOT NULL").Pluck("sales_order_id", &soIDs)
			if len(soIDs) > 0 {
				query = query.Where("id NOT IN ?", soIDs)
			}
		}
	}

	// Filtro de purchase order
	if filter.HasPurchaseOrder != nil {
		if *filter.HasPurchaseOrder {
			var soIDs []int
			r.db.Model(&models.PurchaseOrder{}).Distinct("sales_order_id").Where("sales_order_id IS NOT NULL").Pluck("sales_order_id", &soIDs)
			if len(soIDs) > 0 {
				query = query.Where("id IN ?", soIDs)
			}
		} else {
			var soIDs []int
			r.db.Model(&models.PurchaseOrder{}).Distinct("sales_order_id").Where("sales_order_id IS NOT NULL").Pluck("sales_order_id", &soIDs)
			if len(soIDs) > 0 {
				query = query.Where("id NOT IN ?", soIDs)
			}
		}
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Joins("LEFT JOIN contacts ON contacts.id = sales_orders.contact_id").
			Where("sales_orders.so_no LIKE ? OR sales_orders.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
	}

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

// GetSalesOrderStats retorna estatísticas de sales orders
func (r *salesOrderRepository) GetSalesOrderStats(filter SalesOrderFilter) (*SalesOrderStats, error) {
	stats := &SalesOrderStats{
		CountByStatus: make(map[string]int),
	}

	query := r.db.Model(&models.SalesOrder{})

	// Aplica filtros básicos
	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Contagem total e valores
	var result struct {
		Count      int
		TotalValue float64
	}

	if err := query.Select("COUNT(*) as count, SUM(grand_total) as total_value").
		Scan(&result).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas")
	}

	stats.TotalOrders = result.Count
	stats.TotalValue = result.TotalValue

	// Valores por status específicos
	statusQueries := map[string]string{
		"confirmed":  models.SOStatusConfirmed,
		"processing": models.SOStatusProcessing,
		"completed":  models.SOStatusCompleted,
		"cancelled":  models.SOStatusCancelled,
	}

	for key, status := range statusQueries {
		var value float64
		if err := query.Where("status = ?", status).
			Select("SUM(grand_total)").
			Scan(&value).Error; err != nil {
			r.logger.Warn("erro ao calcular valor para status", zap.String("status", status), zap.Error(err))
		}

		switch key {
		case "confirmed":
			stats.TotalConfirmed = value
		case "processing":
			stats.TotalProcessing = value
		case "completed":
			stats.TotalCompleted = value
		case "cancelled":
			stats.TotalCancelled = value
		}
	}

	// Contagem por status
	rows, err := query.Select("status, COUNT(*) as count").
		Group("status").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao contar por status")
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats.CountByStatus[status] = count
	}

	// Calcula taxa de cumprimento
	completedCount := stats.CountByStatus[models.SOStatusCompleted]
	totalCount := stats.TotalOrders
	if totalCount > 0 {
		stats.FulfillmentRate = float64(completedCount) / float64(totalCount) * 100
	}

	return stats, nil
}

// GetContactSalesOrdersSummary retorna um resumo dos sales orders de um contato
func (r *salesOrderRepository) GetContactSalesOrdersSummary(contactID int) (*ContactSalesOrdersSummary, error) {
	summary := &ContactSalesOrdersSummary{
		ContactID: contactID,
	}

	// Busca informações do contato
	var contact contact.Contact
	if err := r.db.First(&contact, contactID).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao buscar contato")
	}

	summary.ContactName = contact.Name
	if contact.CompanyName != "" {
		summary.ContactName = contact.CompanyName
	}
	summary.ContactType = contact.Type

	// Estatísticas dos sales orders
	var stats struct {
		Count      int
		TotalValue float64
	}

	if err := r.db.Model(&models.SalesOrder{}).
		Where("contact_id = ?", contactID).
		Select("COUNT(*) as count, SUM(grand_total) as total_value").
		Scan(&stats).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas do contato")
	}

	summary.TotalOrders = stats.Count
	summary.TotalValue = stats.TotalValue

	// Valores por status
	statusQueries := map[string]string{
		"completed": models.SOStatusCompleted,
		"cancelled": models.SOStatusCancelled,
	}

	for key, status := range statusQueries {
		var value float64
		if err := r.db.Model(&models.SalesOrder{}).
			Where("contact_id = ? AND status = ?", contactID, status).
			Select("SUM(grand_total)").
			Scan(&value).Error; err != nil {
			r.logger.Warn("erro ao calcular valor para status", zap.String("status", status), zap.Error(err))
		}

		switch key {
		case "completed":
			summary.TotalCompleted = value
		case "cancelled":
			summary.TotalCancelled = value
		}
	}

	// Sales orders pendentes
	var pendingStats struct {
		Count int
		Value float64
	}

	if err := r.db.Model(&models.SalesOrder{}).
		Where("contact_id = ? AND status IN ?", contactID, []string{models.SOStatusDraft, models.SOStatusConfirmed, models.SOStatusProcessing}).
		Select("COUNT(*) as count, SUM(grand_total) as value").
		Scan(&pendingStats).Error; err != nil {
		r.logger.Warn("erro ao calcular sales orders pendentes do contato", zap.Error(err))
	}

	summary.PendingCount = pendingStats.Count
	summary.PendingValue = pendingStats.Value

	// Calcula taxa de cumprimento
	var completedCount int64
	if err := r.db.Model(&models.SalesOrder{}).
		Where("contact_id = ? AND status = ?", contactID, models.SOStatusCompleted).
		Count(&completedCount).Error; err != nil {
		r.logger.Warn("erro ao contar sales orders completados", zap.Error(err))
	}

	if summary.TotalOrders > 0 {
		summary.FulfillmentRate = float64(completedCount) / float64(summary.TotalOrders) * 100
	}

	// Último sales order
	var lastOrder models.SalesOrder
	if err := r.db.Model(&models.SalesOrder{}).
		Where("contact_id = ?", contactID).
		Order("created_at DESC").
		First(&lastOrder).Error; err == nil {
		summary.LastOrderDate = lastOrder.CreatedAt
	}

	return summary, nil
}

// GetSalesOrdersByContactType busca sales orders por tipo de contato
func (r *salesOrderRepository) GetSalesOrdersByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	// Primeiro, busca os IDs dos contatos do tipo especificado
	var contactIDs []int
	if err := r.db.Model(&contact.Contact{}).
		Where("type = ?", contactType).
		Pluck("id", &contactIDs).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao buscar contatos por tipo")
	}

	if len(contactIDs) == 0 {
		// Retorna resultado vazio se não houver contatos do tipo especificado
		return pagination.NewPaginatedResult(0, params.Page, params.PageSize, []models.SalesOrder{}), nil
	}

	// Busca os sales orders dos contatos encontrados
	query := r.db.Model(&models.SalesOrder{}).Where("contact_id IN ?", contactIDs)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao contar sales orders por tipo de contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao buscar sales orders por tipo de contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
}

// CreateInvoiceFromSalesOrder cria uma invoice a partir de um sales order
func (r *salesOrderRepository) CreateInvoiceFromSalesOrder(salesOrderID int) error {
	// Busca o sales order
	salesOrder, err := r.GetSalesOrderByID(salesOrderID)
	if err != nil {
		return err
	}

	// Verifica se o sales order está confirmado
	if salesOrder.Status != models.SOStatusConfirmed && salesOrder.Status != models.SOStatusCompleted {
		return errors.WrapError(gorm.ErrInvalidData, "sales order não está confirmado")
	}

	// TODO: Implementar a criação da invoice
	// Isso seria feito em conjunto com o InvoiceRepository
	// Por enquanto, apenas registramos a operação
	r.logger.Info("sales order pronto para criação de invoice", zap.Int("sales_order_id", salesOrderID))

	return nil
}

// CreatePurchaseOrderFromSalesOrder cria um purchase order a partir de um sales order
func (r *salesOrderRepository) CreatePurchaseOrderFromSalesOrder(salesOrderID int) error {
	// Busca o sales order
	salesOrder, err := r.GetSalesOrderByID(salesOrderID)
	if err != nil {
		return err
	}

	// Verifica se o sales order está confirmado
	if salesOrder.Status != models.SOStatusConfirmed {
		return errors.WrapError(gorm.ErrInvalidData, "sales order não está confirmado")
	}

	// TODO: Implementar a criação do purchase order
	// Isso seria feito em conjunto com o PurchaseOrderRepository
	// Por enquanto, apenas registramos a operação
	r.logger.Info("sales order pronto para criação de purchase order", zap.Int("sales_order_id", salesOrderID))

	return nil
}

// GetPendingSalesOrders busca sales orders pendentes
func (r *salesOrderRepository) GetPendingSalesOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesOrders []models.SalesOrder
	var total int64

	pendingStatuses := []string{models.SOStatusDraft, models.SOStatusConfirmed, models.SOStatusProcessing}
	query := r.db.Model(&models.SalesOrder{}).Where("status IN ?", pendingStatuses)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders pendentes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders pendentes")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesOrders).Error; err != nil {
		r.logger.Error("erro ao buscar sales orders pendentes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales orders pendentes")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesOrders)
	return result, nil
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
