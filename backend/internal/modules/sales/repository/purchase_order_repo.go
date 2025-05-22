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

// PurchaseOrderRepository define as operações do repositório de purchase orders
type PurchaseOrderRepository interface {
	// CRUD básico
	CreatePurchaseOrder(purchaseOrder *models.PurchaseOrder) error
	GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error)
	UpdatePurchaseOrder(id int, purchaseOrder *models.PurchaseOrder) error
	DeletePurchaseOrder(id int) error

	// Consultas com paginação
	GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersBySalesOrder(salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByExpectedDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPurchaseOrdersByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPendingPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverduePurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada (opcional, considere mover para serviço se contiver muita lógica de negócio)
	SearchPurchaseOrders(filter PurchaseOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

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

// PurchaseOrderStats representa estatísticas de purchase orders
type PurchaseOrderStats struct {
	TotalOrders     int            `json:"total_orders"`
	TotalValue      float64        `json:"total_value"`
	TotalDraft      float64        `json:"total_draft"`
	TotalSent       float64        `json:"total_sent"`
	TotalConfirmed  float64        `json:"total_confirmed"`
	TotalReceived   float64        `json:"total_received"`
	TotalCancelled  float64        `json:"total_cancelled"`
	CountByStatus   map[string]int `json:"count_by_status"`
	FulfillmentRate float64        `json:"fulfillment_rate"`
}

// ContactPurchaseOrdersSummary representa um resumo dos purchase orders de um contato
type ContactPurchaseOrdersSummary struct {
	ContactID       int       `json:"contact_id"`
	ContactName     string    `json:"contact_name"`
	ContactType     string    `json:"contact_type"`
	TotalOrders     int       `json:"total_orders"`
	TotalValue      float64   `json:"total_value"`
	TotalReceived   float64   `json:"total_received"`
	TotalCancelled  float64   `json:"total_cancelled"`
	PendingCount    int       `json:"pending_count"`
	PendingValue    float64   `json:"pending_value"`
	OverdueCount    int       `json:"overdue_count"`
	OverdueValue    float64   `json:"overdue_value"`
	FulfillmentRate float64   `json:"fulfillment_rate"`
	LastOrderDate   time.Time `json:"last_order_date"`
}

type purchaseOrderRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPurchaseOrderRepository cria uma nova instância do repositório
func NewPurchaseOrderRepository() (PurchaseOrderRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &purchaseOrderRepository{
		db:     db,
		logger: logger.WithModule("purchase_order_repository"),
	}, nil
}

// CreatePurchaseOrder cria um novo purchase order no banco
func (r *purchaseOrderRepository) CreatePurchaseOrder(purchaseOrder *models.PurchaseOrder) error {
	// Gera o número do purchase order se não foi fornecido
	if purchaseOrder.PONo == "" {
		purchaseOrder.PONo = r.generatePurchaseOrderNumber()
	}

	// Define status padrão se não foi fornecido
	if purchaseOrder.Status == "" {
		purchaseOrder.Status = models.POStatusDraft
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria o purchase order
	if err := tx.Create(purchaseOrder).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar purchase order", zap.Error(err))
		return errors.WrapError(err, "falha ao criar purchase order")
	}

	// Se houver itens, cria os itens
	if len(purchaseOrder.Items) > 0 {
		for i := range purchaseOrder.Items {
			purchaseOrder.Items[i].PurchaseOrderID = purchaseOrder.ID
			if err := tx.Create(&purchaseOrder.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item do purchase order", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d do purchase order", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("purchase order criado com sucesso", zap.Int("id", purchaseOrder.ID), zap.String("po_no", purchaseOrder.PONo))
	return nil
}

// GetPurchaseOrderByID busca um purchase order pelo ID
func (r *purchaseOrderRepository) GetPurchaseOrderByID(id int) (*models.PurchaseOrder, error) {
	var purchaseOrder models.PurchaseOrder

	query := r.db.Preload("Contact").
		Preload("SalesOrder").
		Preload("Items").
		Preload("Items.Product")

	if err := query.First(&purchaseOrder, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPurchaseOrderNotFound
		}
		r.logger.Error("erro ao buscar purchase order por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar purchase order")
	}

	return &purchaseOrder, nil
}

// GetAllPurchaseOrders retorna todos os purchase orders com paginação
func (r *purchaseOrderRepository) GetAllPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

// UpdatePurchaseOrder atualiza um purchase order existente
func (r *purchaseOrderRepository) UpdatePurchaseOrder(id int, purchaseOrder *models.PurchaseOrder) error {
	// Verifica se o purchase order existe
	var existing models.PurchaseOrder
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPurchaseOrderNotFound
		}
		return errors.WrapError(err, "falha ao verificar purchase order existente")
	}

	// Atualiza os campos
	purchaseOrder.ID = id
	if err := r.db.Save(purchaseOrder).Error; err != nil {
		r.logger.Error("erro ao atualizar purchase order", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar purchase order")
	}

	r.logger.Info("purchase order atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeletePurchaseOrder remove um purchase order
func (r *purchaseOrderRepository) DeletePurchaseOrder(id int) error {
	// Verifica se existem deliveries relacionadas
	var deliveryCount int64
	if err := r.db.Model(&models.Delivery{}).Where("purchase_order_id = ?", id).Count(&deliveryCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar deliveries relacionadas")
	}

	if deliveryCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove o purchase order (cascade removerá os itens)
	result := r.db.Delete(&models.PurchaseOrder{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar purchase order", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar purchase order")
	}

	if result.RowsAffected == 0 {
		return errors.ErrPurchaseOrderNotFound
	}

	r.logger.Info("purchase order deletado com sucesso", zap.Int("id", id))
	return nil
}

// GetPurchaseOrdersByStatus busca purchase orders por status
func (r *purchaseOrderRepository) GetPurchaseOrdersByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetPurchaseOrdersByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetPurchaseOrdersBySalesOrder(salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetPurchaseOrdersByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetPurchaseOrdersByExpectedDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

// SearchPurchaseOrders busca purchase orders com filtros combinados
func (r *purchaseOrderRepository) SearchPurchaseOrders(filter PurchaseOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

// GetPurchaseOrdersByContactType busca purchase orders por tipo de contato
func (r *purchaseOrderRepository) GetPurchaseOrdersByContactType(contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetPendingPurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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
func (r *purchaseOrderRepository) GetOverduePurchaseOrders(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
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

// generatePurchaseOrderNumber gera um número único para o purchase order
func (r *purchaseOrderRepository) generatePurchaseOrderNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastPurchaseOrder models.PurchaseOrder

	r.db.Order("id DESC").First(&lastPurchaseOrder)

	year := time.Now().Year()
	sequence := lastPurchaseOrder.ID + 1

	return fmt.Sprintf("PO-%d-%06d", year, sequence)
}
