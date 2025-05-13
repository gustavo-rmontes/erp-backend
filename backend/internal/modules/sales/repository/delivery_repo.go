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

// DeliveryRepository define as operações do repositório de deliveries
type DeliveryRepository interface {
	CreateDelivery(delivery *models.Delivery) error
	GetDeliveryByID(id int) (*models.Delivery, error)
	GetAllDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateDelivery(id int, delivery *models.Delivery) error
	DeleteDelivery(id int) error
	GetDeliveriesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesByPurchaseOrder(purchaseOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesBySalesOrder(salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesByDeliveryDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveriesByReceivedDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchDeliveries(filter DeliveryFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveryStats(filter DeliveryFilter) (*DeliveryStats, error)
	GetContactDeliveriesSummary(contactID int, deliveryType string) (*ContactDeliveriesSummary, error)
	UpdateDeliveryStatus(id int, status string) error
	UpdateDeliveryItem(deliveryID int, itemID int, receivedQty int) error
	MarkAsShipped(id int, trackingNumber string) error
	MarkAsDelivered(id int) error
	MarkAsReturned(id int, reason string) error
	GetPendingDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetOverdueDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetDeliveryTrackingInfo(id int) (*DeliveryTrackingInfo, error)
}

// DeliveryFilter define os filtros para busca avançada
type DeliveryFilter struct {
	Status            []string
	PurchaseOrderID   int
	SalesOrderID      int
	ContactID         int
	DateRangeStart    time.Time
	DateRangeEnd      time.Time
	DeliveryDateStart time.Time
	DeliveryDateEnd   time.Time
	ReceivedDateStart time.Time
	ReceivedDateEnd   time.Time
	ShippingMethod    string
	HasTrackingNumber *bool
	IsOverdue         *bool
	SearchQuery       string
	DeliveryType      string // "incoming" (from PO) or "outgoing" (from SO)
}

// DeliveryStats representa estatísticas de deliveries
type DeliveryStats struct {
	TotalDeliveries     int            `json:"total_deliveries"`
	TotalPending        int            `json:"total_pending"`
	TotalShipped        int            `json:"total_shipped"`
	TotalDelivered      int            `json:"total_delivered"`
	TotalReturned       int            `json:"total_returned"`
	CountByStatus       map[string]int `json:"count_by_status"`
	DeliveryRate        float64        `json:"delivery_rate"`
	ReturnRate          float64        `json:"return_rate"`
	AverageDeliveryTime float64        `json:"average_delivery_time_days"`
}

// ContactDeliveriesSummary representa um resumo das deliveries de um contato
type ContactDeliveriesSummary struct {
	ContactID        int       `json:"contact_id"`
	ContactName      string    `json:"contact_name"`
	ContactType      string    `json:"contact_type"`
	DeliveryType     string    `json:"delivery_type"` // incoming/outgoing
	TotalDeliveries  int       `json:"total_deliveries"`
	PendingCount     int       `json:"pending_count"`
	ShippedCount     int       `json:"shipped_count"`
	DeliveredCount   int       `json:"delivered_count"`
	ReturnedCount    int       `json:"returned_count"`
	OverdueCount     int       `json:"overdue_count"`
	DeliveryRate     float64   `json:"delivery_rate"`
	ReturnRate       float64   `json:"return_rate"`
	LastDeliveryDate time.Time `json:"last_delivery_date"`
}

// DeliveryTrackingInfo representa informações de rastreamento da entrega
type DeliveryTrackingInfo struct {
	DeliveryID      int                  `json:"delivery_id"`
	DeliveryNo      string               `json:"delivery_no"`
	Status          string               `json:"status"`
	TrackingNumber  string               `json:"tracking_number"`
	ShippingMethod  string               `json:"shipping_method"`
	ShippingAddress string               `json:"shipping_address"`
	DeliveryDate    time.Time            `json:"delivery_date"`
	ReceivedDate    time.Time            `json:"received_date"`
	Items           []DeliveryItemStatus `json:"items"`
}

// DeliveryItemStatus representa o status de um item na entrega
type DeliveryItemStatus struct {
	ItemID      int    `json:"item_id"`
	ProductName string `json:"product_name"`
	ProductCode string `json:"product_code"`
	Quantity    int    `json:"quantity"`
	ReceivedQty int    `json:"received_qty"`
	PendingQty  int    `json:"pending_qty"`
	Status      string `json:"status"` // pending, partial, complete
}

type deliveryRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewDeliveryRepository cria uma nova instância do repositório
func NewDeliveryRepository() (DeliveryRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &deliveryRepository{
		db:     db,
		logger: logger.WithModule("delivery_repository"),
	}, nil
}

// CreateDelivery cria uma nova delivery no banco
func (r *deliveryRepository) CreateDelivery(delivery *models.Delivery) error {
	// Gera o número da delivery se não foi fornecido
	if delivery.DeliveryNo == "" {
		delivery.DeliveryNo = r.generateDeliveryNumber()
	}

	// Define status padrão se não foi fornecido
	if delivery.Status == "" {
		delivery.Status = models.DeliveryStatusPending
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria a delivery
	if err := tx.Create(delivery).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar delivery", zap.Error(err))
		return errors.WrapError(err, "falha ao criar delivery")
	}

	// Se houver itens, cria os itens
	if len(delivery.Items) > 0 {
		for i := range delivery.Items {
			delivery.Items[i].DeliveryID = delivery.ID
			// Inicializa ReceivedQty como 0 se não foi fornecido
			if delivery.Items[i].ReceivedQty == 0 {
				delivery.Items[i].ReceivedQty = 0
			}
			if err := tx.Create(&delivery.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item da delivery", zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d da delivery", i))
			}
		}
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("delivery criada com sucesso", zap.Int("id", delivery.ID), zap.String("delivery_no", delivery.DeliveryNo))
	return nil
}

// GetDeliveryByID busca uma delivery pelo ID
func (r *deliveryRepository) GetDeliveryByID(id int) (*models.Delivery, error) {
	var delivery models.Delivery

	query := r.db.Preload("PurchaseOrder").
		Preload("PurchaseOrder.Contact").
		Preload("SalesOrder").
		Preload("SalesOrder.Contact").
		Preload("Items").
		Preload("Items.Product")

	if err := query.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrDeliveryNotFound
		}
		r.logger.Error("erro ao buscar delivery por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar delivery")
	}

	return &delivery, nil
}

// GetAllDeliveries retorna todas as deliveries com paginação
func (r *deliveryRepository) GetAllDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	// Query base
	query := r.db.Model(&models.Delivery{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// UpdateDelivery atualiza uma delivery existente
func (r *deliveryRepository) UpdateDelivery(id int, delivery *models.Delivery) error {
	// Verifica se a delivery existe
	var existing models.Delivery
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao verificar delivery existente")
	}

	// Atualiza os campos
	delivery.ID = id
	if err := r.db.Save(delivery).Error; err != nil {
		r.logger.Error("erro ao atualizar delivery", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar delivery")
	}

	r.logger.Info("delivery atualizada com sucesso", zap.Int("id", id))
	return nil
}

// DeleteDelivery remove uma delivery
func (r *deliveryRepository) DeleteDelivery(id int) error {
	// Verifica o status da delivery
	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao verificar delivery")
	}

	// Não permite deletar deliveries que já foram entregues
	if delivery.Status == models.DeliveryStatusDelivered {
		return errors.WrapError(gorm.ErrInvalidData, "não é possível deletar entregas concluídas")
	}

	// Remove a delivery (cascade removerá os itens)
	result := r.db.Delete(&models.Delivery{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar delivery", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar delivery")
	}

	if result.RowsAffected == 0 {
		return errors.ErrDeliveryNotFound
	}

	r.logger.Info("delivery deletada com sucesso", zap.Int("id", id))
	return nil
}

// GetDeliveriesByStatus busca deliveries por status
func (r *deliveryRepository) GetDeliveriesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar deliveries por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveriesByPurchaseOrder busca deliveries por purchase order
func (r *deliveryRepository) GetDeliveriesByPurchaseOrder(purchaseOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).Where("purchase_order_id = ?", purchaseOrderID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por purchase order", zap.Error(err), zap.Int("purchase_order_id", purchaseOrderID))
		return nil, errors.WrapError(err, "falha ao contar deliveries por purchase order")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por purchase order", zap.Error(err), zap.Int("purchase_order_id", purchaseOrderID))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por purchase order")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveriesBySalesOrder busca deliveries por sales order
func (r *deliveryRepository) GetDeliveriesBySalesOrder(salesOrderID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).Where("sales_order_id = ?", salesOrderID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por sales order", zap.Error(err), zap.Int("sales_order_id", salesOrderID))
		return nil, errors.WrapError(err, "falha ao contar deliveries por sales order")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("SalesOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por sales order", zap.Error(err), zap.Int("sales_order_id", salesOrderID))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por sales order")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveriesByPeriod busca deliveries por período (usando created_at)
func (r *deliveryRepository) GetDeliveriesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveriesByDeliveryDate busca deliveries por data de entrega
func (r *deliveryRepository) GetDeliveriesByDeliveryDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).
		Where("delivery_date >= ? AND delivery_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por data de entrega", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries por data de entrega")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Preload("Items").
		Order("delivery_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por data de entrega", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por data de entrega")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveriesByReceivedDate busca deliveries por data de recebimento
func (r *deliveryRepository) GetDeliveriesByReceivedDate(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{}).
		Where("received_date >= ? AND received_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries por data de recebimento", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries por data de recebimento")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Preload("Items").
		Order("received_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries por data de recebimento", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries por data de recebimento")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// SearchDeliveries busca deliveries com filtros combinados
func (r *deliveryRepository) SearchDeliveries(filter DeliveryFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	query := r.db.Model(&models.Delivery{})

	// Aplica os filtros
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if filter.PurchaseOrderID > 0 {
		query = query.Where("purchase_order_id = ?", filter.PurchaseOrderID)
	}

	if filter.SalesOrderID > 0 {
		query = query.Where("sales_order_id = ?", filter.SalesOrderID)
	}

	// Filtro por tipo de delivery (incoming/outgoing)
	if filter.DeliveryType == "incoming" {
		query = query.Where("purchase_order_id IS NOT NULL")
	} else if filter.DeliveryType == "outgoing" {
		query = query.Where("sales_order_id IS NOT NULL")
	}

	// Filtro por contato (através de PO ou SO)
	if filter.ContactID > 0 {
		poSubquery := r.db.Model(&models.PurchaseOrder{}).Select("id").Where("contact_id = ?", filter.ContactID)
		soSubquery := r.db.Model(&models.SalesOrder{}).Select("id").Where("contact_id = ?", filter.ContactID)
		query = query.Where("purchase_order_id IN (?) OR sales_order_id IN (?)", poSubquery, soSubquery)
	}

	// Filtros de data
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	if !filter.DeliveryDateStart.IsZero() && !filter.DeliveryDateEnd.IsZero() {
		query = query.Where("delivery_date >= ? AND delivery_date <= ?", filter.DeliveryDateStart, filter.DeliveryDateEnd)
	}

	if !filter.ReceivedDateStart.IsZero() && !filter.ReceivedDateEnd.IsZero() {
		query = query.Where("received_date >= ? AND received_date <= ?", filter.ReceivedDateStart, filter.ReceivedDateEnd)
	}

	// Filtro por método de envio
	if filter.ShippingMethod != "" {
		query = query.Where("shipping_method = ?", filter.ShippingMethod)
	}

	// Filtro por tracking number
	if filter.HasTrackingNumber != nil {
		if *filter.HasTrackingNumber {
			query = query.Where("tracking_number IS NOT NULL AND tracking_number != ''")
		} else {
			query = query.Where("tracking_number IS NULL OR tracking_number = ''")
		}
	}

	// Filtro de overdue (vencido)
	if filter.IsOverdue != nil && *filter.IsOverdue {
		now := time.Now()
		query = query.Where("delivery_date < ? AND status IN ?", now, []string{models.DeliveryStatusPending, models.DeliveryStatusShipped})
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Where("delivery_no LIKE ? OR po_no LIKE ? OR so_no LIKE ? OR tracking_number LIKE ? OR notes LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Preload("Items").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveryStats retorna estatísticas de deliveries
func (r *deliveryRepository) GetDeliveryStats(filter DeliveryFilter) (*DeliveryStats, error) {
	stats := &DeliveryStats{
		CountByStatus: make(map[string]int),
	}

	query := r.db.Model(&models.Delivery{})

	// Aplica filtros básicos
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Contagem total
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao contar deliveries")
	}
	stats.TotalDeliveries = int(totalCount)

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

		// Atualiza contadores específicos
		switch status {
		case models.DeliveryStatusPending:
			stats.TotalPending = count
		case models.DeliveryStatusShipped:
			stats.TotalShipped = count
		case models.DeliveryStatusDelivered:
			stats.TotalDelivered = count
		case models.DeliveryStatusReturned:
			stats.TotalReturned = count
		}
	}

	// Calcula taxas
	if stats.TotalDeliveries > 0 {
		stats.DeliveryRate = float64(stats.TotalDelivered) / float64(stats.TotalDeliveries) * 100
		stats.ReturnRate = float64(stats.TotalReturned) / float64(stats.TotalDeliveries) * 100
	}

	// Calcula tempo médio de entrega
	var avgDeliveryTime struct {
		AvgDays float64
	}
	if err := r.db.Model(&models.Delivery{}).
		Where("status = ? AND received_date IS NOT NULL AND delivery_date IS NOT NULL", models.DeliveryStatusDelivered).
		Select("AVG(JULIANDAY(received_date) - JULIANDAY(delivery_date)) as avg_days").
		Scan(&avgDeliveryTime).Error; err == nil {
		stats.AverageDeliveryTime = avgDeliveryTime.AvgDays
	}

	return stats, nil
}

// GetContactDeliveriesSummary retorna um resumo das deliveries de um contato
func (r *deliveryRepository) GetContactDeliveriesSummary(contactID int, deliveryType string) (*ContactDeliveriesSummary, error) {
	summary := &ContactDeliveriesSummary{
		ContactID:    contactID,
		DeliveryType: deliveryType,
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

	// Query base dependendo do tipo de delivery
	query := r.db.Model(&models.Delivery{})
	if deliveryType == "incoming" {
		// Deliveries de Purchase Orders (entrada)
		poSubquery := r.db.Model(&models.PurchaseOrder{}).Select("id").Where("contact_id = ?", contactID)
		query = query.Where("purchase_order_id IN (?)", poSubquery)
	} else if deliveryType == "outgoing" {
		// Deliveries de Sales Orders (saída)
		soSubquery := r.db.Model(&models.SalesOrder{}).Select("id").Where("contact_id = ?", contactID)
		query = query.Where("sales_order_id IN (?)", soSubquery)
	}

	// Contagem total
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao contar deliveries do contato")
	}
	summary.TotalDeliveries = int(totalCount)

	// Contagem por status
	var statusCounts []struct {
		Status string
		Count  int
	}
	if err := query.Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao contar deliveries por status")
	}

	for _, sc := range statusCounts {
		switch sc.Status {
		case models.DeliveryStatusPending:
			summary.PendingCount = sc.Count
		case models.DeliveryStatusShipped:
			summary.ShippedCount = sc.Count
		case models.DeliveryStatusDelivered:
			summary.DeliveredCount = sc.Count
		case models.DeliveryStatusReturned:
			summary.ReturnedCount = sc.Count
		}
	}

	// Deliveries vencidas
	now := time.Now()
	var overdueCount int64
	if err := query.Where("delivery_date < ? AND status IN ?", now, []string{models.DeliveryStatusPending, models.DeliveryStatusShipped}).
		Count(&overdueCount).Error; err != nil {
		r.logger.Warn("erro ao contar deliveries vencidas", zap.Error(err))
	}
	summary.OverdueCount = int(overdueCount)

	// Calcula taxas
	if summary.TotalDeliveries > 0 {
		summary.DeliveryRate = float64(summary.DeliveredCount) / float64(summary.TotalDeliveries) * 100
		summary.ReturnRate = float64(summary.ReturnedCount) / float64(summary.TotalDeliveries) * 100
	}

	// Última delivery
	var lastDelivery models.Delivery
	if err := query.Order("created_at DESC").First(&lastDelivery).Error; err == nil {
		summary.LastDeliveryDate = lastDelivery.CreatedAt
	}

	return summary, nil
}

// UpdateDeliveryStatus atualiza o status de uma delivery
func (r *deliveryRepository) UpdateDeliveryStatus(id int, status string) error {
	// Verifica se a delivery existe
	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao buscar delivery")
	}

	// Atualiza o status
	delivery.Status = status

	// Se estiver marcando como entregue, atualiza a data de recebimento
	if status == models.DeliveryStatusDelivered && delivery.ReceivedDate.IsZero() {
		delivery.ReceivedDate = time.Now()
	}

	if err := r.db.Save(&delivery).Error; err != nil {
		r.logger.Error("erro ao atualizar status da delivery", zap.Error(err), zap.Int("id", id), zap.String("status", status))
		return errors.WrapError(err, "falha ao atualizar status da delivery")
	}

	r.logger.Info("status da delivery atualizado", zap.Int("id", id), zap.String("status", status))
	return nil
}

// UpdateDeliveryItem atualiza a quantidade recebida de um item
func (r *deliveryRepository) UpdateDeliveryItem(deliveryID int, itemID int, receivedQty int) error {
	// Busca o item
	var item models.DeliveryItem
	if err := r.db.Where("delivery_id = ? AND id = ?", deliveryID, itemID).First(&item).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryItemNotFound
		}
		return errors.WrapError(err, "falha ao buscar item da delivery")
	}

	// Valida a quantidade
	if receivedQty < 0 || receivedQty > item.Quantity {
		return errors.WrapError(gorm.ErrInvalidData, "quantidade recebida inválida")
	}

	// Atualiza a quantidade recebida
	item.ReceivedQty = receivedQty
	if err := r.db.Save(&item).Error; err != nil {
		r.logger.Error("erro ao atualizar item da delivery", zap.Error(err), zap.Int("delivery_id", deliveryID), zap.Int("item_id", itemID))
		return errors.WrapError(err, "falha ao atualizar item da delivery")
	}

	// Verifica se todos os itens foram recebidos para atualizar o status da delivery
	var pendingItems int64
	if err := r.db.Model(&models.DeliveryItem{}).
		Where("delivery_id = ? AND received_qty < quantity", deliveryID).
		Count(&pendingItems).Error; err != nil {
		r.logger.Warn("erro ao contar itens pendentes", zap.Error(err))
	}

	// Se todos os itens foram recebidos, atualiza o status da delivery para delivered
	if pendingItems == 0 {
		if err := r.UpdateDeliveryStatus(deliveryID, models.DeliveryStatusDelivered); err != nil {
			r.logger.Warn("erro ao atualizar status da delivery para delivered", zap.Error(err))
		}
	}

	r.logger.Info("item da delivery atualizado", zap.Int("delivery_id", deliveryID), zap.Int("item_id", itemID), zap.Int("received_qty", receivedQty))
	return nil
}

// MarkAsShipped marca uma delivery como enviada
func (r *deliveryRepository) MarkAsShipped(id int, trackingNumber string) error {
	// Busca a delivery
	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao buscar delivery")
	}

	// Verifica se o status permite marcação como shipped
	if delivery.Status != models.DeliveryStatusPending {
		return errors.WrapError(gorm.ErrInvalidData, "apenas deliveries pendentes podem ser marcadas como enviadas")
	}

	// Atualiza o status e o tracking number
	delivery.Status = models.DeliveryStatusShipped
	delivery.TrackingNumber = trackingNumber
	if delivery.DeliveryDate.IsZero() {
		delivery.DeliveryDate = time.Now()
	}

	if err := r.db.Save(&delivery).Error; err != nil {
		r.logger.Error("erro ao marcar delivery como shipped", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao marcar delivery como shipped")
	}

	r.logger.Info("delivery marcada como shipped", zap.Int("id", id), zap.String("tracking_number", trackingNumber))
	return nil
}

// MarkAsDelivered marca uma delivery como entregue
func (r *deliveryRepository) MarkAsDelivered(id int) error {
	// Busca a delivery
	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao buscar delivery")
	}

	// Verifica se o status permite marcação como delivered
	if delivery.Status != models.DeliveryStatusShipped {
		return errors.WrapError(gorm.ErrInvalidData, "apenas deliveries enviadas podem ser marcadas como entregues")
	}

	// Atualiza o status e a data de recebimento
	delivery.Status = models.DeliveryStatusDelivered
	delivery.ReceivedDate = time.Now()

	if err := r.db.Save(&delivery).Error; err != nil {
		r.logger.Error("erro ao marcar delivery como delivered", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao marcar delivery como delivered")
	}

	// Atualiza todos os itens como recebidos (quantidade total)
	if err := r.db.Model(&models.DeliveryItem{}).
		Where("delivery_id = ?", id).
		Updates(map[string]interface{}{
			"received_qty": gorm.Expr("quantity"),
		}).Error; err != nil {
		r.logger.Warn("erro ao atualizar itens como recebidos", zap.Error(err))
	}

	r.logger.Info("delivery marcada como delivered", zap.Int("id", id))
	return nil
}

// MarkAsReturned marca uma delivery como devolvida
func (r *deliveryRepository) MarkAsReturned(id int, reason string) error {
	// Busca a delivery
	var delivery models.Delivery
	if err := r.db.First(&delivery, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao buscar delivery")
	}

	// Atualiza o status e adiciona a razão nas notas
	delivery.Status = models.DeliveryStatusReturned
	if reason != "" {
		if delivery.Notes != "" {
			delivery.Notes += " | "
		}
		delivery.Notes += "Devolvido: " + reason
	}

	if err := r.db.Save(&delivery).Error; err != nil {
		r.logger.Error("erro ao marcar delivery como returned", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao marcar delivery como returned")
	}

	r.logger.Info("delivery marcada como returned", zap.Int("id", id), zap.String("reason", reason))
	return nil
}

// GetPendingDeliveries busca deliveries pendentes
func (r *deliveryRepository) GetPendingDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	return r.GetDeliveriesByStatus(models.DeliveryStatusPending, params)
}

// GetOverdueDeliveries busca deliveries vencidas
func (r *deliveryRepository) GetOverdueDeliveries(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var deliveries []models.Delivery
	var total int64

	now := time.Now()
	query := r.db.Model(&models.Delivery{}).
		Where("delivery_date < ? AND status IN ?", now, []string{models.DeliveryStatusPending, models.DeliveryStatusShipped})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar deliveries vencidas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar deliveries vencidas")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("PurchaseOrder").
		Preload("SalesOrder").
		Order("delivery_date ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&deliveries).Error; err != nil {
		r.logger.Error("erro ao buscar deliveries vencidas", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar deliveries vencidas")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, deliveries)
	return result, nil
}

// GetDeliveryTrackingInfo retorna informações detalhadas de rastreamento
func (r *deliveryRepository) GetDeliveryTrackingInfo(id int) (*DeliveryTrackingInfo, error) {
	// Busca a delivery com todos os relacionamentos
	delivery, err := r.GetDeliveryByID(id)
	if err != nil {
		return nil, err
	}

	tracking := &DeliveryTrackingInfo{
		DeliveryID:      delivery.ID,
		DeliveryNo:      delivery.DeliveryNo,
		Status:          delivery.Status,
		TrackingNumber:  delivery.TrackingNumber,
		ShippingMethod:  delivery.ShippingMethod,
		ShippingAddress: delivery.ShippingAddress,
		DeliveryDate:    delivery.DeliveryDate,
		ReceivedDate:    delivery.ReceivedDate,
		Items:           make([]DeliveryItemStatus, 0),
	}

	// Processa os itens
	for _, item := range delivery.Items {
		itemStatus := DeliveryItemStatus{
			ItemID:      item.ID,
			ProductName: item.ProductName,
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
			ReceivedQty: item.ReceivedQty,
			PendingQty:  item.Quantity - item.ReceivedQty,
		}

		// Determina o status do item
		if item.ReceivedQty == 0 {
			itemStatus.Status = "pending"
		} else if item.ReceivedQty < item.Quantity {
			itemStatus.Status = "partial"
		} else {
			itemStatus.Status = "complete"
		}

		tracking.Items = append(tracking.Items, itemStatus)
	}

	return tracking, nil
}

// generateDeliveryNumber gera um número único para a delivery
func (r *deliveryRepository) generateDeliveryNumber() string {
	// Implementação simples - você pode melhorar isso
	var lastDelivery models.Delivery

	r.db.Order("id DESC").First(&lastDelivery)

	year := time.Now().Year()
	sequence := lastDelivery.ID + 1

	return fmt.Sprintf("DLV-%d-%06d", year, sequence)
}
