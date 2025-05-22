package repository

import (
	"ERP-ONSMART/backend/internal/errors"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"context"
	"time"

	"go.uber.org/zap"
)

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

// GetAllQuotations retorna todas as quotations com paginação
func (r *quotationRepository) GetAllQuotations(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	// Query base
	query := r.db.WithContext(ctx).Model(&models.Quotation{})

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

// GetQuotationsByStatus busca quotations por status
func (r *quotationRepository) GetQuotationsByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Quotation{}).Where("status = ?", status)

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
func (r *quotationRepository) GetQuotationsByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Quotation{}).Where("contact_id = ?", contactID)

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
func (r *quotationRepository) GetExpiredQuotations(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	now := time.Now()
	query := r.db.WithContext(ctx).Model(&models.Quotation{}).
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

// GetQuotationsByDateRange busca quotations por período (usando created_at)
func (r *quotationRepository) GetQuotationsByDateRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Quotation{}).
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

// GetQuotationsByExpiryRange busca quotations por período de expiração
func (r *quotationRepository) GetQuotationsByExpiryRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Quotation{}).
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
func (r *quotationRepository) SearchQuotations(ctx context.Context, filter QuotationFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&models.Quotation{})

	// Aplica os filtros
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	// Filtro por tipo de contato ou pessoa
	if filter.ContactType != "" || filter.PersonType != "" {
		contactQuery := r.db.WithContext(ctx).Model(&contact.Contact{})
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

// GetQuotationsByContactType busca quotations por tipo de contato
func (r *quotationRepository) GetQuotationsByContactType(ctx context.Context, contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	// Primeiro, busca os IDs dos contatos do tipo especificado
	var contactIDs []int
	if err := r.db.WithContext(ctx).Model(&contact.Contact{}).
		Where("type = ?", contactType).
		Pluck("id", &contactIDs).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao buscar contatos por tipo")
	}

	if len(contactIDs) == 0 {
		// Retorna resultado vazio se não houver contatos do tipo especificado
		return pagination.NewPaginatedResult(0, params.Page, params.PageSize, []models.Quotation{}), nil
	}

	// Busca as quotations dos contatos encontrados
	query := r.db.WithContext(ctx).Model(&models.Quotation{}).Where("contact_id IN ?", contactIDs)

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

// GetExpiringQuotations busca quotations que expirarão em X dias
func (r *quotationRepository) GetExpiringQuotations(ctx context.Context, days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var quotations []models.Quotation
	var total int64

	now := time.Now()
	expiryLimit := now.AddDate(0, 0, days)

	query := r.db.WithContext(ctx).Model(&models.Quotation{}).
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
