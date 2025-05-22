package repository

import (
	"ERP-ONSMART/backend/internal/errors"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

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

// GetAllSalesOrders retorna todos os sales orders com paginação
func (r *salesOrderRepository) GetAllSalesOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "GetAllSalesOrders"))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "GetAllSalesOrders"))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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

// GetSalesOrdersByStatus busca sales orders por status
func (r *salesOrderRepository) GetSalesOrdersByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "GetSalesOrdersByStatus"),
				zap.String("status", status))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders por status")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "GetSalesOrdersByStatus"),
				zap.String("status", status))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders por status")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto e filtro por status
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar sales orders por status")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) GetSalesOrdersByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "GetSalesOrdersByContact"),
				zap.Int("contact_id", contactID))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders por contato")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "GetSalesOrdersByContact"),
				zap.Int("contact_id", contactID))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders por contato")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto e filtro por contato
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar sales orders por contato")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) GetSalesOrdersByQuotation(ctx context.Context, quotationID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "GetSalesOrdersByQuotation"),
				zap.Int("quotation_id", quotationID))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders por quotation")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "GetSalesOrdersByQuotation"),
				zap.Int("quotation_id", quotationID))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders por quotation")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto e filtro por quotation
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{}).Where("quotation_id = ?", quotationID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por quotation", zap.Error(err), zap.Int("quotation_id", quotationID))
		return nil, errors.WrapError(err, "falha ao contar sales orders por quotation")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) GetSalesOrdersByPeriod(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "GetSalesOrdersByPeriod"),
				zap.Time("start_date", startDate),
				zap.Time("end_date", endDate))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders por período")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "GetSalesOrdersByPeriod"),
				zap.Time("start_date", startDate),
				zap.Time("end_date", endDate))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders por período")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto e filtro por período
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders por período")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) GetSalesOrdersByExpectedDate(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "GetSalesOrdersByExpectedDate"),
				zap.Time("start_date", startDate),
				zap.Time("end_date", endDate))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders por data esperada")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "GetSalesOrdersByExpectedDate"),
				zap.Time("start_date", startDate),
				zap.Time("end_date", endDate))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders por data esperada")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Query base com contexto e filtro por data esperada
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{}).
		Where("expected_date >= ? AND expected_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders por data esperada", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders por data esperada")
	}

	// Verifica contexto antes da operação principal de busca
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) SearchSalesOrders(ctx context.Context, filter SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação",
				zap.String("op", "SearchSalesOrders"),
				zap.String("search_query", filter.SearchQuery))
			return nil, errors.WrapError(ctx.Err(), "timeout ao buscar sales orders")
		case context.Canceled:
			r.logger.Info("operação cancelada",
				zap.String("op", "SearchSalesOrders"),
				zap.String("search_query", filter.SearchQuery))
			return nil, errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return nil, errors.WrapError(ctx.Err(), "erro de contexto ao buscar sales orders")
		}
	}

	var salesOrders []models.SalesOrder
	var total int64

	// Inicia a query base com contexto
	query := r.db.WithContext(ctx).Model(&models.SalesOrder{})

	// Aplica os diversos filtros usando métodos auxiliares
	query = r.applyStatusFilter(query, filter)
	query = r.applyContactFilter(ctx, query, filter)
	query = r.applyDateRangeFilters(query, filter)
	query = r.applyAmountFilters(query, filter)

	// Verifica contexto antes de aplicar filtros mais complexos
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou durante aplicação de filtros")
	}

	query = r.applyRelatedEntityFilters(ctx, query, filter)
	query = r.applyTextSearchFilter(query, filter)

	// Verifica contexto antes da contagem
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da contagem")
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales orders na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales orders na busca")
	}

	// Verifica contexto antes da busca principal
	if ctx.Err() != nil {
		return nil, errors.WrapError(ctx.Err(), "contexto expirou antes da busca principal")
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
func (r *salesOrderRepository) applyContactFilter(ctx context.Context, query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	// Filtro simples de ID do contato
	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	// Filtro por tipo de contato ou pessoa
	if filter.ContactType != "" || filter.PersonType != "" {
		var contactIDs []int
		contactQuery := r.db.WithContext(ctx).Model(&contact.Contact{})

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
func (r *salesOrderRepository) getRelatedOrderIDs(ctx context.Context, entityType string, hasRelation bool) ([]int, error) {
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
	query := r.db.WithContext(ctx).Table(tableName).
		Distinct("sales_order_id").
		Where("sales_order_id IS NOT NULL")

	if err := query.Pluck("sales_order_id", &orderIDs).Error; err != nil {
		return nil, err
	}

	return orderIDs, nil
}

// Método auxiliar para filtrar por relações com outras entidades
func (r *salesOrderRepository) applyRelatedEntityFilters(ctx context.Context, query *gorm.DB, filter SalesOrderFilter) *gorm.DB {
	// Filtro de invoice
	if filter.HasInvoice != nil {
		if orderIDs, err := r.getRelatedOrderIDs(ctx, "invoice", true); err == nil && len(orderIDs) > 0 {
			if *filter.HasInvoice {
				query = query.Where("id IN ?", orderIDs)
			} else {
				query = query.Where("id NOT IN ?", orderIDs)
			}
		}
	}

	// Filtro de purchase order
	if filter.HasPurchaseOrder != nil {
		if orderIDs, err := r.getRelatedOrderIDs(ctx, "purchase_order", true); err == nil && len(orderIDs) > 0 {
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
