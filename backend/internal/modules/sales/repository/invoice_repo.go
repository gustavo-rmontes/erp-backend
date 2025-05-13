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

// InvoiceRepository define as operações do repositório de invoices
type InvoiceRepository interface {
	CreateInvoice(invoice *models.Invoice) error
	GetInvoiceByID(id int) (*models.Invoice, error)
	GetAllInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateInvoice(id int, invoice *models.Invoice) error
	DeleteInvoice(id int) error
	GetInvoicesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoicesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverdueInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoicesBySalesOrder(salesOrderID int) ([]models.Invoice, error)
	GetInvoicesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoicesByDueDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoicesByIssueDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchInvoices(filter InvoiceFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetInvoiceStats(filter InvoiceFilter) (*InvoiceStats, error)
	GetContactInvoicesSummary(contactID int) (*ContactInvoicesSummary, error)
	GetInvoicesByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// InvoiceFilter define os filtros para busca avançada
type InvoiceFilter struct {
	Status         []string
	ContactID      int
	ContactType    string // cliente, fornecedor, lead
	PersonType     string // pf, pj
	DateRangeStart time.Time
	DateRangeEnd   time.Time
	DueDateStart   time.Time
	DueDateEnd     time.Time
	MinAmount      float64
	MaxAmount      float64
	HasPayment     *bool
	IsOverdue      *bool
	SearchQuery    string
}

// InvoiceStats representa estatísticas de invoices
type InvoiceStats struct {
	TotalInvoices int            `json:"total_invoices"`
	TotalValue    float64        `json:"total_value"`
	TotalPaid     float64        `json:"total_paid"`
	TotalPending  float64        `json:"total_pending"`
	TotalOverdue  float64        `json:"total_overdue"`
	CountByStatus map[string]int `json:"count_by_status"`
}

// ContactInvoicesSummary representa um resumo das invoices de um contato
type ContactInvoicesSummary struct {
	ContactID     int     `json:"contact_id"`
	ContactName   string  `json:"contact_name"`
	ContactType   string  `json:"contact_type"`
	TotalInvoices int     `json:"total_invoices"`
	TotalValue    float64 `json:"total_value"`
	TotalPaid     float64 `json:"total_paid"`
	TotalPending  float64 `json:"total_pending"`
	OverdueCount  int     `json:"overdue_count"`
	OverdueValue  float64 `json:"overdue_value"`
}

type invoiceRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewInvoiceRepository cria uma nova instância do repositório
func NewInvoiceRepository() (InvoiceRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &invoiceRepository{
		db:     db,
		logger: logger.WithModule("invoice_repository"),
	}, nil
}

// CreateInvoice cria uma nova invoice no banco
func (r *invoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	// Gera o número da invoice se não foi fornecido
	if invoice.InvoiceNo == "" {
		invoice.InvoiceNo = r.generateInvoiceNumber()
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria a invoice
	if err := tx.Create(invoice).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar invoice", zap.Error(err))
		return errors.WrapError(err, "falha ao criar invoice")
	}

	// Se houver itens, cria os itens
	if len(invoice.Items) > 0 {
		for i := range invoice.Items {
			invoice.Items[i].InvoiceID = invoice.ID
			if err := tx.Create(&invoice.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item da invoice", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d da invoice", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("invoice criada com sucesso", zap.Int("id", invoice.ID), zap.String("invoice_no", invoice.InvoiceNo))
	return nil
}

// GetInvoiceByID busca uma invoice pelo ID
func (r *invoiceRepository) GetInvoiceByID(id int) (*models.Invoice, error) {
	var invoice models.Invoice

	query := r.db.Preload("Contact").
		Preload("SalesOrder").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payments")

	if err := query.First(&invoice, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrInvoiceNotFound
		}
		r.logger.Error("erro ao buscar invoice por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar invoice")
	}

	return &invoice, nil
}

// GetAllInvoices retorna todas as invoices com paginação
func (r *invoiceRepository) GetAllInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	// Query base
	query := r.db.Model(&models.Invoice{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// UpdateInvoice atualiza uma invoice existente
func (r *invoiceRepository) UpdateInvoice(id int, invoice *models.Invoice) error {
	// Verifica se a invoice existe
	var existing models.Invoice
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrInvoiceNotFound
		}
		return errors.WrapError(err, "falha ao verificar invoice existente")
	}

	// Atualiza os campos
	invoice.ID = id
	if err := r.db.Save(invoice).Error; err != nil {
		r.logger.Error("erro ao atualizar invoice", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar invoice")
	}

	r.logger.Info("invoice atualizada com sucesso", zap.Int("id", id))
	return nil
}

// DeleteInvoice remove uma invoice
func (r *invoiceRepository) DeleteInvoice(id int) error {
	// Verifica se existem pagamentos relacionados
	var paymentCount int64
	if err := r.db.Model(&models.Payment{}).Where("invoice_id = ?", id).Count(&paymentCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar pagamentos relacionados")
	}

	if paymentCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove a invoice (cascade removerá os itens)
	result := r.db.Delete(&models.Invoice{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar invoice", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar invoice")
	}

	if result.RowsAffected == 0 {
		return errors.ErrInvoiceNotFound
	}

	r.logger.Info("invoice deletada com sucesso", zap.Int("id", id))
	return nil
}

// GetInvoicesByStatus busca invoices por status
func (r *invoiceRepository) GetInvoicesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar invoices por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar invoices por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetInvoicesByContact busca invoices por contato
func (r *invoiceRepository) GetInvoicesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar invoices por contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao buscar invoices por contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetOverdueInvoices busca invoices vencidas
func (r *invoiceRepository) GetOverdueInvoices(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	now := time.Now()
	query := r.db.Model(&models.Invoice{}).
		Where("due_date < ? AND status != ?", now, models.InvoiceStatusPaid).
		Where("status != ?", models.InvoiceStatusCancelled)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices vencidas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices vencidas")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("due_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices vencidas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices vencidas")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetInvoicesBySalesOrder busca invoices por pedido de venda
func (r *invoiceRepository) GetInvoicesBySalesOrder(salesOrderID int) ([]models.Invoice, error) {
	var invoices []models.Invoice

	if err := r.db.Where("sales_order_id = ?", salesOrderID).
		Preload("Contact").
		Preload("Items").
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por pedido de venda", zap.Error(err), zap.Int("sales_order_id", salesOrderID))
		return nil, errors.WrapError(err, "falha ao buscar invoices por pedido de venda")
	}

	return invoices, nil
}

// GetInvoicesByPeriod busca invoices por período (usando created_at)
func (r *invoiceRepository) GetInvoicesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetInvoicesByDueDateRange busca invoices por período de vencimento
func (r *invoiceRepository) GetInvoicesByDueDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).
		Where("due_date >= ? AND due_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por período de vencimento", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices por período de vencimento")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("due_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por período de vencimento", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices por período de vencimento")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetInvoicesByIssueDateRange busca invoices por período de emissão
func (r *invoiceRepository) GetInvoicesByIssueDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{}).
		Where("issue_date >= ? AND issue_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por período de emissão", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices por período de emissão")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("issue_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por período de emissão", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices por período de emissão")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// SearchInvoices busca invoices com filtros combinados
func (r *invoiceRepository) SearchInvoices(filter InvoiceFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
	var total int64

	query := r.db.Model(&models.Invoice{})

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

	if !filter.DueDateStart.IsZero() && !filter.DueDateEnd.IsZero() {
		query = query.Where("due_date >= ? AND due_date <= ?", filter.DueDateStart, filter.DueDateEnd)
	}

	// Filtros de valor
	if filter.MinAmount > 0 {
		query = query.Where("grand_total >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("grand_total <= ?", filter.MaxAmount)
	}

	// Filtro de vencimento
	if filter.IsOverdue != nil && *filter.IsOverdue {
		now := time.Now()
		query = query.Where("due_date < ? AND status != ?", now, models.InvoiceStatusPaid).
			Where("status != ?", models.InvoiceStatusCancelled)
	}

	// Filtro de pagamento
	if filter.HasPayment != nil {
		if *filter.HasPayment {
			query = query.Where("amount_paid > 0")
		} else {
			query = query.Where("amount_paid = 0")
		}
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Joins("LEFT JOIN contacts ON contacts.id = invoices.contact_id").
			Where("invoices.invoice_no LIKE ? OR invoices.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar invoices na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar invoices")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// GetInvoiceStats retorna estatísticas de invoices
func (r *invoiceRepository) GetInvoiceStats(filter InvoiceFilter) (*InvoiceStats, error) {
	stats := &InvoiceStats{
		CountByStatus: make(map[string]int),
	}

	query := r.db.Model(&models.Invoice{})

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
		TotalPaid  float64
	}

	if err := query.Select("COUNT(*) as count, SUM(grand_total) as total_value, SUM(amount_paid) as total_paid").
		Scan(&result).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas")
	}

	stats.TotalInvoices = result.Count
	stats.TotalValue = result.TotalValue
	stats.TotalPaid = result.TotalPaid
	stats.TotalPending = stats.TotalValue - stats.TotalPaid

	// Valor vencido
	now := time.Now()
	var overdueValue float64
	if err := query.Where("due_date < ? AND status != ?", now, models.InvoiceStatusPaid).
		Where("status != ?", models.InvoiceStatusCancelled).
		Select("SUM(grand_total - amount_paid)").
		Scan(&overdueValue).Error; err != nil {
		r.logger.Warn("erro ao calcular valor vencido", zap.Error(err))
	}
	stats.TotalOverdue = overdueValue

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

	return stats, nil
}

// GetContactInvoicesSummary retorna um resumo das invoices de um contato
func (r *invoiceRepository) GetContactInvoicesSummary(contactID int) (*ContactInvoicesSummary, error) {
	summary := &ContactInvoicesSummary{
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

	// Estatísticas das invoices
	var stats struct {
		Count      int
		TotalValue float64
		TotalPaid  float64
	}

	if err := r.db.Model(&models.Invoice{}).
		Where("contact_id = ?", contactID).
		Select("COUNT(*) as count, SUM(grand_total) as total_value, SUM(amount_paid) as total_paid").
		Scan(&stats).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas do contato")
	}

	summary.TotalInvoices = stats.Count
	summary.TotalValue = stats.TotalValue
	summary.TotalPaid = stats.TotalPaid
	summary.TotalPending = stats.TotalValue - stats.TotalPaid

	// Invoices vencidas
	now := time.Now()
	var overdueStats struct {
		Count int
		Value float64
	}

	if err := r.db.Model(&models.Invoice{}).
		Where("contact_id = ? AND due_date < ? AND status != ?", contactID, now, models.InvoiceStatusPaid).
		Where("status != ?", models.InvoiceStatusCancelled).
		Select("COUNT(*) as count, SUM(grand_total - amount_paid) as value").
		Scan(&overdueStats).Error; err != nil {
		r.logger.Warn("erro ao calcular invoices vencidas do contato", zap.Error(err))
	}

	summary.OverdueCount = overdueStats.Count
	summary.OverdueValue = overdueStats.Value

	return summary, nil
}

// GetInvoicesByContactType busca invoices por tipo de contato
func (r *invoiceRepository) GetInvoicesByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var invoices []models.Invoice
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
		return pagination.NewPaginatedResult(0, params.Page, params.PageSize, []models.Invoice{}), nil
	}

	// Busca as invoices dos contatos encontrados
	query := r.db.Model(&models.Invoice{}).Where("contact_id IN ?", contactIDs)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar invoices por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao contar invoices por tipo de contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&invoices).Error; err != nil {
		r.logger.Error("erro ao buscar invoices por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao buscar invoices por tipo de contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, invoices)
	return result, nil
}

// generateInvoiceNumber gera um número único para a invoice
func (r *invoiceRepository) generateInvoiceNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastInvoice models.Invoice

	r.db.Order("id DESC").First(&lastInvoice)

	year := time.Now().Year()
	sequence := lastInvoice.ID + 1

	return fmt.Sprintf("INV-%d-%06d", year, sequence)
}
