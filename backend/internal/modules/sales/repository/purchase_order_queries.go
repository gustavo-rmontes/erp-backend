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

// PurchaseOrderFilter define os filtros para busca avançada
type PurchaseOrderFilter struct {
	Status            []string
	ContactID         int
	ContactType       string // fornecedor principalmente
	PersonType        string // pf, pj
	DateRangeStart    time.Time
	DateRangeEnd      time.Time
	ExpectedDateStart time.Time
	ExpectedDateEnd   time.Time
	MinAmount         float64
	MaxAmount         float64
	HasDelivery       *bool
	IsOverdue         *bool
	SearchQuery       string
	SalesOrderID      int
}

// GetAllPurchaseOrders retorna todos os purchase orders com paginação
func (r *purchaseOrderRepository) GetAllPurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	// Query base
	query := r.db.Model(&models.PurchaseOrder{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersByStatus busca purchase orders por status
func (r *purchaseOrderRepository) GetPurchaseOrdersByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersByContact busca purchase orders por contato
func (r *purchaseOrderRepository) GetPurchaseOrdersByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersBySalesOrder busca purchase orders por sales order
func (r *purchaseOrderRepository) GetPurchaseOrdersBySalesOrder(ctx context.Context, salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).Where("sales_order_id = ?", salesOrderID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por sales order", zap.Error(err), zap.Int("sales_order_id", salesOrderID))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por sales order")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("SalesOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por sales order", zap.Error(err), zap.Int("sales_order_id", salesOrderID))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por sales order")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersByPeriod busca purchase orders por período (usando created_at)
func (r *purchaseOrderRepository) GetPurchaseOrdersByPeriod(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersByExpectedDateRange busca purchase orders por data esperada
func (r *purchaseOrderRepository) GetPurchaseOrdersByExpectedDateRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{}).
		Where("expected_date >= ? AND expected_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por data esperada", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por data esperada")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("expected_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por data esperada", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por data esperada")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPurchaseOrdersByContactType busca purchase orders por tipo de contato
func (r *purchaseOrderRepository) GetPurchaseOrdersByContactType(ctx context.Context, contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
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
		return pagination.NewPaginatedResult(0, params.Page, params.PageSize, []models.PurchaseOrder{}), nil
	}

	// Busca os purchase orders dos contatos encontrados
	query := r.db.Model(&models.PurchaseOrder{}).Where("contact_id IN ?", contactIDs)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao contar purchase orders por tipo de contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders por tipo de contato", zap.Error(err), zap.String("contact_type", contactType))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders por tipo de contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetPendingPurchaseOrders busca purchase orders pendentes
func (r *purchaseOrderRepository) GetPendingPurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	pendingStatuses := []string{models.POStatusDraft, models.POStatusSent, models.POStatusConfirmed}
	query := r.db.Model(&models.PurchaseOrder{}).Where("status IN ?", pendingStatuses)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders pendentes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders pendentes")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders pendentes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders pendentes")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// GetOverduePurchaseOrders busca purchase orders vencidos
func (r *purchaseOrderRepository) GetOverduePurchaseOrders(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	now := time.Now()
	query := r.db.Model(&models.PurchaseOrder{}).
		Where("expected_date < ? AND status IN ?", now, []string{models.POStatusDraft, models.POStatusSent, models.POStatusConfirmed})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders vencidos", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders vencidos")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("expected_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders vencidos", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders vencidos")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}

// SearchPurchaseOrders busca purchase orders com filtros combinados
func (r *purchaseOrderRepository) SearchPurchaseOrders(ctx context.Context, filter PurchaseOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var purchaseOrders []models.PurchaseOrder
	var total int64

	query := r.db.Model(&models.PurchaseOrder{})

	// Aplica os filtros
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	if filter.SalesOrderID > 0 {
		query = query.Where("sales_order_id = ?", filter.SalesOrderID)
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

	// Filtro de overdue (vencido)
	if filter.IsOverdue != nil && *filter.IsOverdue {
		now := time.Now()
		query = query.Where("expected_date < ? AND status IN ?", now, []string{models.POStatusDraft, models.POStatusSent, models.POStatusConfirmed})
	}

	// Filtro de delivery
	if filter.HasDelivery != nil {
		if *filter.HasDelivery {
			var poIDs []int
			r.db.Model(&models.Delivery{}).Distinct("purchase_order_id").Where("purchase_order_id IS NOT NULL").Pluck("purchase_order_id", &poIDs)
			if len(poIDs) > 0 {
				query = query.Where("id IN ?", poIDs)
			}
		} else {
			var poIDs []int
			r.db.Model(&models.Delivery{}).Distinct("purchase_order_id").Where("purchase_order_id IS NOT NULL").Pluck("purchase_order_id", &poIDs)
			if len(poIDs) > 0 {
				query = query.Where("id NOT IN ?", poIDs)
			}
		}
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Joins("LEFT JOIN contacts ON contacts.id = purchase_orders.contact_id").
			Where("purchase_orders.po_no LIKE ? OR purchase_orders.so_no LIKE ? OR purchase_orders.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar purchase orders na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar purchase orders na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&purchaseOrders).Error; err != nil {
		r.logger.Error("erro ao buscar purchase orders", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar purchase orders")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, purchaseOrders)
	return result, nil
}
