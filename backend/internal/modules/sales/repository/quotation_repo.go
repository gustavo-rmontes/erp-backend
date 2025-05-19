package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	stderrors "errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// QuotationRepository define as operações do repositório de quotations
type QuotationRepository interface {
	CreateQuotation(quotation *models.Quotation) error
	GetQuotationByID(id int) (*models.Quotation, error)
	GetAllQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateQuotation(id int, quotation *models.Quotation) error
	DeleteQuotation(id int) error
	GetQuotationsByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetExpiredQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByExpiryDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchQuotations(filter QuotationFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationStats(filter QuotationFilter) (*QuotationStats, error)
	GetContactQuotationsSummary(contactID int) (*ContactQuotationsSummary, error)
	GetQuotationsByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	ConvertToSalesOrder(quotationID int) error
	GetExpiringQuotations(days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SetCreatedAtForTesting(quotationID int, createdAt time.Time) error
}

// QuotationFilter define os filtros para busca avançada
type QuotationFilter struct {
	Status         []string
	ContactID      int
	ContactType    string // cliente, fornecedor, lead
	PersonType     string // pf, pj
	DateRangeStart time.Time
	DateRangeEnd   time.Time
	ExpiryStart    time.Time
	ExpiryEnd      time.Time
	MinAmount      float64
	MaxAmount      float64
	IsExpired      *bool
	SearchQuery    string
}

// QuotationStats representa estatísticas de quotations
type QuotationStats struct {
	TotalQuotations int            `json:"total_quotations"`
	TotalValue      float64        `json:"total_value"`
	TotalAccepted   float64        `json:"total_accepted"`
	TotalRejected   float64        `json:"total_rejected"`
	TotalPending    float64        `json:"total_pending"`
	TotalExpired    float64        `json:"total_expired"`
	CountByStatus   map[string]int `json:"count_by_status"`
	ConversionRate  float64        `json:"conversion_rate"`
}

// ContactQuotationsSummary representa um resumo das quotations de um contato
type ContactQuotationsSummary struct {
	ContactID         int       `json:"contact_id"`
	ContactName       string    `json:"contact_name"`
	ContactType       string    `json:"contact_type"`
	TotalQuotations   int       `json:"total_quotations"`
	TotalValue        float64   `json:"total_value"`
	TotalAccepted     float64   `json:"total_accepted"`
	TotalRejected     float64   `json:"total_rejected"`
	PendingCount      int       `json:"pending_count"`
	PendingValue      float64   `json:"pending_value"`
	ConversionRate    float64   `json:"conversion_rate"`
	LastQuotationDate time.Time `json:"last_quotation_date"`
}

type quotationRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewQuotationRepository cria uma nova instância do repositório
func NewQuotationRepository() (QuotationRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &quotationRepository{
		db:     db,
		logger: logger.WithModule("quotation_repository"),
	}, nil
}

// CreateQuotation cria uma nova quotation no banco
func (r *quotationRepository) CreateQuotation(quotation *models.Quotation) error {
	// Gera o número da quotation se não foi fornecido
	if quotation.QuotationNo == "" {
		quotation.QuotationNo = r.generateQuotationNumber()
	}

	// Define status padrão se não foi fornecido
	if quotation.Status == "" {
		quotation.Status = models.QuotationStatusDraft
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria a quotation
	if err := tx.Create(quotation).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar quotation", zap.Error(err))
		return errors.WrapError(err, "falha ao criar quotation")
	}

	// Se houver itens, cria os itens
	if len(quotation.Items) > 0 {
		for i := range quotation.Items {
			quotation.Items[i].QuotationID = quotation.ID
			if err := tx.Create(&quotation.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item da quotation", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d da quotation", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("quotation criada com sucesso", zap.Int("id", quotation.ID), zap.String("quotation_no", quotation.QuotationNo))
	return nil
}

// GetQuotationByID busca uma quotation pelo ID
func (r *quotationRepository) GetQuotationByID(id int) (*models.Quotation, error) {
	var quotation models.Quotation

	query := r.db.Preload("Contact").
		Preload("Items").
		Preload("Items.Product")

	if err := query.First(&quotation, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrQuotationNotFound
		}
		r.logger.Error("erro ao buscar quotation por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar quotation")
	}

	return &quotation, nil
}

// GetAllQuotations retorna todas as quotations com paginação
func (r *quotationRepository) GetAllQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	// Query base
	query := r.db.Model(&models.Quotation{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// UpdateQuotation atualiza uma quotation existente
func (r *quotationRepository) UpdateQuotation(id int, quotation *models.Quotation) error {
	// Verifica se a quotation existe
	var existing models.Quotation
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrQuotationNotFound
		}
		return errors.WrapError(err, "falha ao verificar quotation existente")
	}

	// Atualiza os campos
	quotation.ID = id
	if err := r.db.Save(quotation).Error; err != nil {
		r.logger.Error("erro ao atualizar quotation", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar quotation")
	}

	r.logger.Info("quotation atualizada com sucesso", zap.Int("id", id))
	return nil
}

// DeleteQuotation remove uma quotation
func (r *quotationRepository) DeleteQuotation(id int) error {
	// Verifica se existem sales orders relacionadas
	var salesOrderCount int64
	if err := r.db.Model(&models.SalesOrder{}).Where("quotation_id = ?", id).Count(&salesOrderCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar pedidos de venda relacionados")
	}

	if salesOrderCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove a quotation (cascade removerá os itens)
	result := r.db.Delete(&models.Quotation{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar quotation", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar quotation")
	}

	if result.RowsAffected == 0 {
		return errors.ErrQuotationNotFound
	}

	r.logger.Info("quotation deletada com sucesso", zap.Int("id", id))
	return nil
}

// GetQuotationsByStatus busca quotations por status
func (r *quotationRepository) GetQuotationsByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.Model(&models.Quotation{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar quotations por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar quotations por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// GetQuotationsByContact busca quotations por contato
func (r *quotationRepository) GetQuotationsByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.Model(&models.Quotation{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar quotations por contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao buscar quotations por contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// GetExpiredQuotations busca quotations expiradas
func (r *quotationRepository) GetExpiredQuotations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	now := time.Now()
	query := r.db.Model(&models.Quotation{}).
		Where("expiry_date < ? AND status NOT IN ?", now, []string{models.QuotationStatusAccepted, models.QuotationStatusRejected, models.QuotationStatusCancelled})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations expiradas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations expiradas")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("expiry_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations expiradas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations expiradas")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// GetQuotationsByPeriod busca quotations por período (usando created_at)
func (r *quotationRepository) GetQuotationsByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.Model(&models.Quotation{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// GetQuotationsByExpiryDateRange busca quotations por período de expiração
func (r *quotationRepository) GetQuotationsByExpiryDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.Model(&models.Quotation{}).
		Where("expiry_date >= ? AND expiry_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations por período de expiração", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations por período de expiração")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("expiry_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations por período de expiração", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations por período de expiração")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// SearchQuotations busca quotations com filtros combinados
func (r *quotationRepository) SearchQuotations(filter QuotationFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.Model(&models.Quotation{})

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

	if !filter.ExpiryStart.IsZero() && !filter.ExpiryEnd.IsZero() {
		query = query.Where("expiry_date >= ? AND expiry_date <= ?", filter.ExpiryStart, filter.ExpiryEnd)
	}

	// Filtros de valor
	if filter.MinAmount > 0 {
		query = query.Where("grand_total >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("grand_total <= ?", filter.MaxAmount)
	}

	// Filtro de expiração
	if filter.IsExpired != nil && *filter.IsExpired {
		now := time.Now()
		query = query.Where("expiry_date < ? AND status NOT IN ?", now, []string{models.QuotationStatusAccepted, models.QuotationStatusRejected, models.QuotationStatusCancelled})
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Joins("LEFT JOIN contacts ON contacts.id = quotations.contact_id").
			Where("quotations.quotation_no LIKE ? OR quotations.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// GetQuotationStats retorna estatísticas de quotations
func (r *quotationRepository) GetQuotationStats(filter QuotationFilter) (*QuotationStats, error) {
	stats := &QuotationStats{
		CountByStatus: make(map[string]int),
	}

	query := r.db.Model(&models.Quotation{})

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

	// Usando COALESCE para tratar NULL na soma total
	if err := query.Select("COUNT(*) as count, COALESCE(SUM(grand_total), 0) as total_value").
		Scan(&result).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas")
	}

	stats.TotalQuotations = result.Count
	stats.TotalValue = result.TotalValue

	// Valores por status específicos
	statusQueries := map[string]string{
		"accepted": models.QuotationStatusAccepted,
		"rejected": models.QuotationStatusRejected,
		"pending":  models.QuotationStatusSent,
		"expired":  models.QuotationStatusExpired,
	}

	for key, status := range statusQueries {
		var value float64
		statusQuery := r.db.Model(&models.Quotation{})

		// Aplicar os mesmos filtros básicos
		if filter.ContactID > 0 {
			statusQuery = statusQuery.Where("contact_id = ?", filter.ContactID)
		}

		if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
			statusQuery = statusQuery.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
		}

		if key == "expired" {
			now := time.Now()
			statusQuery = statusQuery.Where("expiry_date < ? AND status NOT IN ?", now, []string{models.QuotationStatusAccepted, models.QuotationStatusRejected, models.QuotationStatusCancelled})
		} else if key == "pending" {
			statusQuery = statusQuery.Where("status IN ?", []string{models.QuotationStatusDraft, models.QuotationStatusSent})
		} else {
			statusQuery = statusQuery.Where("status = ?", status)
		}

		// Usando COALESCE para tratar NULL nas somas por status
		if err := statusQuery.Select("COALESCE(SUM(grand_total), 0) as total").Scan(&value).Error; err != nil {
			r.logger.Warn("erro ao calcular valor para status", zap.String("status", status), zap.Error(err))
			continue
		}

		switch key {
		case "accepted":
			stats.TotalAccepted = value
		case "rejected":
			stats.TotalRejected = value
		case "pending":
			stats.TotalPending = value
		case "expired":
			stats.TotalExpired = value
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

	// Calcula taxa de conversão
	acceptedCount := stats.CountByStatus[models.QuotationStatusAccepted]
	totalCount := stats.TotalQuotations
	if totalCount > 0 {
		stats.ConversionRate = float64(acceptedCount) / float64(totalCount) * 100
	}

	return stats, nil
}

// GetContactQuotationsSummary retorna um resumo das quotations de um contato
func (r *quotationRepository) GetContactQuotationsSummary(contactID int) (*ContactQuotationsSummary, error) {
	summary := &ContactQuotationsSummary{
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

	// Estatísticas das quotations
	var stats struct {
		Count      int
		TotalValue float64
	}

	if err := r.db.Model(&models.Quotation{}).
		Where("contact_id = ?", contactID).
		Select("COUNT(*) as count, SUM(grand_total) as total_value").
		Scan(&stats).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas do contato")
	}

	summary.TotalQuotations = stats.Count
	summary.TotalValue = stats.TotalValue

	// Valores por status
	statusQueries := map[string]string{
		"accepted": models.QuotationStatusAccepted,
		"rejected": models.QuotationStatusRejected,
	}

	for key, status := range statusQueries {
		var value float64
		if err := r.db.Model(&models.Quotation{}).
			Where("contact_id = ? AND status = ?", contactID, status).
			Select("SUM(grand_total)").
			Scan(&value).Error; err != nil {
			r.logger.Warn("erro ao calcular valor para status", zap.String("status", status), zap.Error(err))
		}

		switch key {
		case "accepted":
			summary.TotalAccepted = value
		case "rejected":
			summary.TotalRejected = value
		}
	}

	// Quotations pendentes
	var pendingStats struct {
		Count int
		Value float64
	}

	if err := r.db.Model(&models.Quotation{}).
		Where("contact_id = ? AND status IN ?", contactID, []string{models.QuotationStatusDraft, models.QuotationStatusSent}).
		Select("COUNT(*) as count, SUM(grand_total) as value").
		Scan(&pendingStats).Error; err != nil {
		r.logger.Warn("erro ao calcular quotations pendentes do contato", zap.Error(err))
	}

	summary.PendingCount = pendingStats.Count
	summary.PendingValue = pendingStats.Value

	// Calcula taxa de conversão
	var acceptedCount int64
	if err := r.db.Model(&models.Quotation{}).
		Where("contact_id = ? AND status = ?", contactID, models.QuotationStatusAccepted).
		Count(&acceptedCount).Error; err != nil {
		r.logger.Warn("erro ao contar quotations aceitas", zap.Error(err))
	}

	if summary.TotalQuotations > 0 {
		summary.ConversionRate = float64(acceptedCount) / float64(summary.TotalQuotations) * 100
	}

	// Última quotation
	var lastQuotation models.Quotation
	if err := r.db.Model(&models.Quotation{}).
		Where("contact_id = ?", contactID).
		Order("created_at DESC").
		First(&lastQuotation).Error; err == nil {
		summary.LastQuotationDate = lastQuotation.CreatedAt
	}

	return summary, nil
}

// GetQuotationsByContactType busca quotations por tipo de contato
func (r *quotationRepository) GetQuotationsByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
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
		return pagination.NewPaginatedResult(0, params.Page, params.PageSize, []models.Quotation{}), nil
	}

	// Busca as quotations dos contatos encontrados
	query := r.db.Model(&models.Quotation{}).Where("contact_id IN ?", contactIDs)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao contar quotations por tipo de contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao buscar quotations por tipo de contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// ConvertToSalesOrder converte uma quotation aceita em pedido de venda
func (r *quotationRepository) ConvertToSalesOrder(quotationID int) error {
	// Busca a quotation
	quotation, err := r.GetQuotationByID(quotationID)
	if err != nil {
		return err
	}

	// Verifica se a quotation está aceita
	if quotation.Status != models.QuotationStatusAccepted {
		return errors.WrapError(stderrors.New("quotation não está aceita"), "status inválido para conversão")
	}

	// Inicia transação
	tx := r.db.Begin()

	// Gera número do pedido de venda
	soNumber := r.generateSalesOrderNumber(tx)

	// Cria o Sales Order
	salesOrder := &models.SalesOrder{
		SONo:            soNumber,
		QuotationID:     quotation.ID,
		ContactID:       quotation.ContactID,
		Status:          models.SOStatusConfirmed,     // Já começa como confirmado, já que a cotação já foi aceita
		ExpectedDate:    time.Now().AddDate(0, 0, 15), // Data estimada de 15 dias por padrão
		SubTotal:        quotation.SubTotal,
		TaxTotal:        quotation.TaxTotal,
		DiscountTotal:   quotation.DiscountTotal,
		GrandTotal:      quotation.GrandTotal,
		Notes:           quotation.Notes,
		PaymentTerms:    quotation.Terms,
		ShippingAddress: "", // Precisaria ser preenchido com informações do contato
	}

	// Cria o sales order
	if err := tx.Create(salesOrder).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar sales order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar sales order")
	}

	// Copia os itens da cotação para o pedido de venda
	for _, item := range quotation.Items {
		orderItem := &models.SOItem{
			SalesOrderID: salesOrder.ID,
			ProductID:    item.ProductID,
			ProductName:  item.ProductName,
			ProductCode:  item.ProductCode,
			Description:  item.Description,
			Quantity:     item.Quantity,
			UnitPrice:    item.UnitPrice,
			Discount:     item.Discount,
			Tax:          item.Tax,
			Total:        item.Total,
		}

		if err := tx.Create(orderItem).Error; err != nil {
			tx.Rollback()
			r.logger.Error("erro ao criar item do sales order", zap.Error(err))
			return errors.WrapError(err, fmt.Sprintf("falha ao criar item do sales order para o produto %s", item.ProductName))
		}
	}

	// Opcionalmente, podemos atualizar o status da cotação para indicar que foi convertida
	// Poderia ser criado um status específico como "converted" no enums.go
	// Por enquanto, mantém o status atual

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao confirmar transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("quotation convertida em pedido de venda com sucesso",
		zap.Int("quotation_id", quotationID),
		zap.Int("sales_order_id", salesOrder.ID),
		zap.String("sales_order_no", salesOrder.SONo))

	return nil
}

// GetExpiringQuotations busca quotations que expirarão em X dias
func (r *quotationRepository) GetExpiringQuotations(days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	now := time.Now()
	expiryLimit := now.AddDate(0, 0, days)

	query := r.db.Model(&models.Quotation{}).
		Where("expiry_date >= ? AND expiry_date <= ?", now, expiryLimit).
		Where("status IN ?", []string{models.QuotationStatusDraft, models.QuotationStatusSent})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar quotations expirando", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar quotations expirando")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("expiry_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&quotations).Error; err != nil {
		r.logger.Error("erro ao buscar quotations expirando", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar quotations expirando")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, quotations)
	return result, nil
}

// generateQuotationNumber gera um número único para a quotation
func (r *quotationRepository) generateQuotationNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastQuotation models.Quotation

	r.db.Order("id DESC").First(&lastQuotation)

	year := time.Now().Year()
	sequence := lastQuotation.ID + 1

	return fmt.Sprintf("QT-%d-%06d", year, sequence)
}

func (r *quotationRepository) generateSalesOrderNumber(tx *gorm.DB) string {
	var lastOrder models.SalesOrder

	// Se tx for nil, usa r.db
	db := r.db
	if tx != nil {
		db = tx
	}

	db.Order("id DESC").First(&lastOrder)

	year := time.Now().Year()
	sequence := lastOrder.ID + 1
	if sequence == 0 {
		sequence = 1 // Evita problemas com o primeiro registro
	}

	return fmt.Sprintf("SO-%d-%06d", year, sequence)
}

// Apenas para uso em testes
func (r *quotationRepository) SetCreatedAtForTesting(quotationID int, createdAt time.Time) error {
	return r.db.Exec("UPDATE quotations SET created_at = ? WHERE id = ?", createdAt, quotationID).Error
}
