package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// SalesProcessRepository define as operações do repositório de sales process
type SalesProcessRepository interface {
	CreateSalesProcess(salesProcess *models.SalesProcess) error
	GetSalesProcessByID(id int) (*models.SalesProcess, error)
	GetAllSalesProcesses(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdateSalesProcess(id int, salesProcess *models.SalesProcess) error
	DeleteSalesProcess(id int) error
	GetSalesProcessesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesProcessesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesProcessesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchSalesProcesses(filter SalesProcessFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetSalesProcessStats(filter SalesProcessFilter) (*SalesProcessStats, error)
	GetContactSalesProcessSummary(contactID int) (*ContactSalesProcessSummary, error)

	// Process flow methods
	InitiateFromQuotation(quotationID int) (*models.SalesProcess, error)
	LinkQuotation(processID int, quotationID int) error
	LinkSalesOrder(processID int, salesOrderID int) error
	LinkPurchaseOrder(processID int, purchaseOrderID int) error
	LinkDelivery(processID int, deliveryID int) error
	LinkInvoice(processID int, invoiceID int) error

	// Status transitions
	UpdateProcessStatus(id int, status string) error
	CalculateProfitability(id int) error

	// Complex queries
	GetCompleteProcessFlow(id int) (*CompleteProcessFlow, error)
	GetProcessTimeline(id int) (*ProcessTimeline, error)
	GetProfitabilityAnalysis(filter SalesProcessFilter) (*ProfitabilityAnalysis, error)
	GetSalesConversionMetrics(filter SalesProcessFilter) (*SalesConversionMetrics, error)
	GetProcessesByStage(stage string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetAbandonedProcesses(days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// SalesProcessFilter define os filtros para busca avançada
type SalesProcessFilter struct {
	Status           []string
	ContactID        int
	ContactType      string
	DateRangeStart   time.Time
	DateRangeEnd     time.Time
	MinValue         float64
	MaxValue         float64
	MinProfit        float64
	MaxProfit        float64
	HasQuotation     *bool
	HasSalesOrder    *bool
	HasPurchaseOrder *bool
	HasInvoice       *bool
	IsComplete       *bool
	SearchQuery      string
}

// SalesProcessStats representa estatísticas de sales processes
type SalesProcessStats struct {
	TotalProcesses   int            `json:"total_processes"`
	TotalValue       float64        `json:"total_value"`
	TotalProfit      float64        `json:"total_profit"`
	AverageValue     float64        `json:"average_value"`
	AverageProfit    float64        `json:"average_profit"`
	ProfitMargin     float64        `json:"profit_margin_percentage"`
	CountByStatus    map[string]int `json:"count_by_status"`
	CompletionRate   float64        `json:"completion_rate"`
	AverageCycleTime float64        `json:"average_cycle_time_days"`
}

// ContactSalesProcessSummary representa um resumo dos processos de um contato
type ContactSalesProcessSummary struct {
	ContactID          int       `json:"contact_id"`
	ContactName        string    `json:"contact_name"`
	ContactType        string    `json:"contact_type"`
	TotalProcesses     int       `json:"total_processes"`
	ActiveProcesses    int       `json:"active_processes"`
	CompletedProcesses int       `json:"completed_processes"`
	TotalValue         float64   `json:"total_value"`
	TotalProfit        float64   `json:"total_profit"`
	AverageValue       float64   `json:"average_value"`
	ConversionRate     float64   `json:"conversion_rate"`
	LastProcessDate    time.Time `json:"last_process_date"`
}

// CompleteProcessFlow representa o fluxo completo de um processo
type CompleteProcessFlow struct {
	Process        *models.SalesProcess   `json:"process"`
	Quotation      *models.Quotation      `json:"quotation,omitempty"`
	SalesOrder     *models.SalesOrder     `json:"sales_order,omitempty"`
	PurchaseOrders []models.PurchaseOrder `json:"purchase_orders,omitempty"`
	Deliveries     []models.Delivery      `json:"deliveries,omitempty"`
	Invoices       []models.Invoice       `json:"invoices,omitempty"`
	Payments       []models.Payment       `json:"payments,omitempty"`
	Timeline       []ProcessEvent         `json:"timeline"`
}

// ProcessTimeline representa a linha do tempo de eventos do processo
type ProcessTimeline struct {
	ProcessID int            `json:"process_id"`
	Events    []ProcessEvent `json:"events"`
	Duration  int            `json:"duration_days"`
	Status    string         `json:"current_status"`
}

// ProcessEvent representa um evento na linha do tempo
type ProcessEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	DocumentID  int       `json:"document_id,omitempty"`
	DocumentNo  string    `json:"document_no,omitempty"`
	Value       float64   `json:"value,omitempty"`
}

// ProfitabilityAnalysis representa análise de lucratividade
type ProfitabilityAnalysis struct {
	TotalRevenue    float64                 `json:"total_revenue"`
	TotalCosts      float64                 `json:"total_costs"`
	TotalProfit     float64                 `json:"total_profit"`
	ProfitMargin    float64                 `json:"profit_margin_percentage"`
	ByProduct       []ProductProfitability  `json:"by_product"`
	ByCustomer      []CustomerProfitability `json:"by_customer"`
	ByPeriod        []PeriodProfitability   `json:"by_period"`
	TopProfitable   []models.SalesProcess   `json:"top_profitable_processes"`
	LeastProfitable []models.SalesProcess   `json:"least_profitable_processes"`
}

// ProductProfitability representa lucratividade por produto
type ProductProfitability struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	Revenue     float64 `json:"revenue"`
	Cost        float64 `json:"cost"`
	Profit      float64 `json:"profit"`
	Margin      float64 `json:"margin_percentage"`
	Quantity    int     `json:"quantity_sold"`
}

// CustomerProfitability representa lucratividade por cliente
type CustomerProfitability struct {
	ContactID    int     `json:"contact_id"`
	ContactName  string  `json:"contact_name"`
	Revenue      float64 `json:"revenue"`
	Cost         float64 `json:"cost"`
	Profit       float64 `json:"profit"`
	Margin       float64 `json:"margin_percentage"`
	ProcessCount int     `json:"process_count"`
}

// PeriodProfitability representa lucratividade por período
type PeriodProfitability struct {
	Period  string  `json:"period"`
	Revenue float64 `json:"revenue"`
	Cost    float64 `json:"cost"`
	Profit  float64 `json:"profit"`
	Margin  float64 `json:"margin_percentage"`
}

// SalesConversionMetrics representa métricas de conversão de vendas
type SalesConversionMetrics struct {
	TotalQuotations       int                     `json:"total_quotations"`
	QuotationToSORate     float64                 `json:"quotation_to_so_rate"`
	SOToInvoiceRate       float64                 `json:"so_to_invoice_rate"`
	InvoiceToPaymentRate  float64                 `json:"invoice_to_payment_rate"`
	OverallConversionRate float64                 `json:"overall_conversion_rate"`
	AverageConversionTime float64                 `json:"average_conversion_time_days"`
	ByStage               map[string]StageMetrics `json:"by_stage"`
}

// StageMetrics representa métricas por estágio do processo
type StageMetrics struct {
	Count           int     `json:"count"`
	ConversionRate  float64 `json:"conversion_rate"`
	AverageTime     float64 `json:"average_time_days"`
	AbandonmentRate float64 `json:"abandonment_rate"`
}

// ProcessStatus define os status possíveis do processo
const (
	ProcessStatusDraft      = "draft"
	ProcessStatusQuotation  = "quotation"
	ProcessStatusSalesOrder = "sales_order"
	ProcessStatusPurchase   = "purchase"
	ProcessStatusDelivery   = "delivery"
	ProcessStatusInvoicing  = "invoicing"
	ProcessStatusPayment    = "payment"
	ProcessStatusCompleted  = "completed"
	ProcessStatusCancelled  = "cancelled"
)

type salesProcessRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewSalesProcessRepository cria uma nova instância do repositório
func NewSalesProcessRepository() (SalesProcessRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &salesProcessRepository{
		db:     db,
		logger: logger.WithModule("sales_process_repository"),
	}, nil
}

// CreateSalesProcess cria um novo sales process no banco
func (r *salesProcessRepository) CreateSalesProcess(salesProcess *models.SalesProcess) error {
	// Define status padrão se não foi fornecido
	if salesProcess.Status == "" {
		salesProcess.Status = ProcessStatusDraft
	}

	// Cria o sales process
	if err := r.db.Create(salesProcess).Error; err != nil {
		r.logger.Error("erro ao criar sales process", zap.Error(err))
		return errors.WrapError(err, "falha ao criar sales process")
	}

	r.logger.Info("sales process criado com sucesso", zap.Int("id", salesProcess.ID))
	return nil
}

// GetSalesProcessByID busca um sales process pelo ID
func (r *salesProcessRepository) GetSalesProcessByID(id int) (*models.SalesProcess, error) {
	var salesProcess models.SalesProcess

	query := r.db.Preload("Contact")

	if err := query.First(&salesProcess, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrSalesProcessNotFound
		}
		r.logger.Error("erro ao buscar sales process por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar sales process")
	}

	// Carrega os documentos relacionados
	if err := r.loadRelatedDocuments(&salesProcess); err != nil {
		r.logger.Warn("erro ao carregar documentos relacionados", zap.Error(err))
	}

	return &salesProcess, nil
}

// GetAllSalesProcesses retorna todos os sales processes com paginação
func (r *salesProcessRepository) GetAllSalesProcesses(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	// Query base
	query := r.db.Model(&models.SalesProcess{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales processes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales processes")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar sales processes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales processes")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// UpdateSalesProcess atualiza um sales process existente
func (r *salesProcessRepository) UpdateSalesProcess(id int, salesProcess *models.SalesProcess) error {
	// Verifica se o sales process existe
	var existing models.SalesProcess
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao verificar sales process existente")
	}

	// Atualiza os campos
	salesProcess.ID = id
	if err := r.db.Save(salesProcess).Error; err != nil {
		r.logger.Error("erro ao atualizar sales process", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar sales process")
	}

	r.logger.Info("sales process atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeleteSalesProcess remove um sales process
func (r *salesProcessRepository) DeleteSalesProcess(id int) error {
	// Verifica se o sales process existe
	var existing models.SalesProcess
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao verificar sales process")
	}

	// Não permite deletar processos com documentos vinculados
	if existing.Status != ProcessStatusDraft {
		return errors.WrapError(gorm.ErrInvalidData, "não é possível deletar processos com documentos vinculados")
	}

	// Remove o sales process
	result := r.db.Delete(&models.SalesProcess{}, id)
	if result.Error != nil {
		r.logger.Error("erro ao deletar sales process", zap.Error(result.Error), zap.Int("id", id))
		return errors.WrapError(result.Error, "falha ao deletar sales process")
	}

	if result.RowsAffected == 0 {
		return errors.ErrSalesProcessNotFound
	}

	r.logger.Info("sales process deletado com sucesso", zap.Int("id", id))
	return nil
}

// GetSalesProcessesByStatus busca sales processes por status
func (r *salesProcessRepository) GetSalesProcessesByStatus(status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	query := r.db.Model(&models.SalesProcess{}).Where("status = ?", status)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales processes por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao contar sales processes por status")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar sales processes por status", zap.Error(err), zap.String("status", status))
		return nil, errors.WrapError(err, "falha ao buscar sales processes por status")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// GetSalesProcessesByContact busca sales processes por contato
func (r *salesProcessRepository) GetSalesProcessesByContact(contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	query := r.db.Model(&models.SalesProcess{}).Where("contact_id = ?", contactID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales processes por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao contar sales processes por contato")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar sales processes por contato", zap.Error(err), zap.Int("contact_id", contactID))
		return nil, errors.WrapError(err, "falha ao buscar sales processes por contato")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// GetSalesProcessesByPeriod busca sales processes por período
func (r *salesProcessRepository) GetSalesProcessesByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	query := r.db.Model(&models.SalesProcess{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales processes por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales processes por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar sales processes por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales processes por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// SearchSalesProcesses busca sales processes com filtros combinados
func (r *salesProcessRepository) SearchSalesProcesses(filter SalesProcessFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	query := r.db.Model(&models.SalesProcess{})

	// Aplica os filtros
	if len(filter.Status) > 0 {
		query = query.Where("status IN ?", filter.Status)
	}

	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	// Filtro por tipo de contato
	if filter.ContactType != "" {
		contactQuery := r.db.Model(&contact.Contact{}).Select("id").Where("type = ?", filter.ContactType)
		query = query.Where("contact_id IN (?)", contactQuery)
	}

	// Filtros de data
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Filtros de valor
	if filter.MinValue > 0 {
		query = query.Where("total_value >= ?", filter.MinValue)
	}

	if filter.MaxValue > 0 {
		query = query.Where("total_value <= ?", filter.MaxValue)
	}

	// Filtros de lucro
	if filter.MinProfit > 0 {
		query = query.Where("profit >= ?", filter.MinProfit)
	}

	if filter.MaxProfit > 0 {
		query = query.Where("profit <= ?", filter.MaxProfit)
	}

	// Filtros de completude
	if filter.IsComplete != nil {
		if *filter.IsComplete {
			query = query.Where("status = ?", ProcessStatusCompleted)
		} else {
			query = query.Where("status != ?", ProcessStatusCompleted)
		}
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Joins("LEFT JOIN contacts ON contacts.id = sales_processes.contact_id").
			Where("sales_processes.notes LIKE ? OR contacts.name LIKE ? OR contacts.company_name LIKE ?",
				searchPattern, searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar sales processes na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar sales processes na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("created_at DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar sales processes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar sales processes")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// GetSalesProcessStats retorna estatísticas de sales processes
func (r *salesProcessRepository) GetSalesProcessStats(filter SalesProcessFilter) (*SalesProcessStats, error) {
	stats := &SalesProcessStats{
		CountByStatus: make(map[string]int),
	}

	query := r.db.Model(&models.SalesProcess{})

	// Aplica filtros básicos
	if filter.ContactID > 0 {
		query = query.Where("contact_id = ?", filter.ContactID)
	}

	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Contagem total e valores
	var result struct {
		Count       int
		TotalValue  float64
		TotalProfit float64
		AvgValue    float64
		AvgProfit   float64
	}

	if err := query.Select("COUNT(*) as count, SUM(total_value) as total_value, SUM(profit) as total_profit, AVG(total_value) as avg_value, AVG(profit) as avg_profit").
		Scan(&result).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas")
	}

	stats.TotalProcesses = result.Count
	stats.TotalValue = result.TotalValue
	stats.TotalProfit = result.TotalProfit
	stats.AverageValue = result.AvgValue
	stats.AverageProfit = result.AvgProfit

	// Calcula margem de lucro
	if stats.TotalValue > 0 {
		stats.ProfitMargin = (stats.TotalProfit / stats.TotalValue) * 100
	}

	// Contagem por status
	rows, err := query.Select("status, COUNT(*) as count").
		Group("status").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao contar por status")
	}
	defer rows.Close()

	var completedCount int
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		stats.CountByStatus[status] = count
		if status == ProcessStatusCompleted {
			completedCount = count
		}
	}

	// Calcula taxa de conclusão
	if stats.TotalProcesses > 0 {
		stats.CompletionRate = (float64(completedCount) / float64(stats.TotalProcesses)) * 100
	}

	// Calcula tempo médio de ciclo
	var avgCycleTime struct {
		AvgDays float64
	}
	if err := r.db.Model(&models.SalesProcess{}).
		Where("status = ?", ProcessStatusCompleted).
		Select("AVG(JULIANDAY(updated_at) - JULIANDAY(created_at)) as avg_days").
		Scan(&avgCycleTime).Error; err == nil {
		stats.AverageCycleTime = avgCycleTime.AvgDays
	}

	return stats, nil
}

// GetContactSalesProcessSummary retorna um resumo dos processos de um contato
func (r *salesProcessRepository) GetContactSalesProcessSummary(contactID int) (*ContactSalesProcessSummary, error) {
	summary := &ContactSalesProcessSummary{
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

	// Estatísticas dos processos
	var stats struct {
		Count       int
		TotalValue  float64
		TotalProfit float64
		AvgValue    float64
	}

	if err := r.db.Model(&models.SalesProcess{}).
		Where("contact_id = ?", contactID).
		Select("COUNT(*) as count, SUM(total_value) as total_value, SUM(profit) as total_profit, AVG(total_value) as avg_value").
		Scan(&stats).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas do contato")
	}

	summary.TotalProcesses = stats.Count
	summary.TotalValue = stats.TotalValue
	summary.TotalProfit = stats.TotalProfit
	summary.AverageValue = stats.AvgValue

	// Conta processos ativos e completos
	var activeCount int64
	if err := r.db.Model(&models.SalesProcess{}).
		Where("contact_id = ? AND status NOT IN ?", contactID, []string{ProcessStatusCompleted, ProcessStatusCancelled}).
		Count(&activeCount).Error; err != nil {
		r.logger.Warn("erro ao contar processos ativos", zap.Error(err))
	}
	summary.ActiveProcesses = int(activeCount)

	var completedCount int64
	if err := r.db.Model(&models.SalesProcess{}).
		Where("contact_id = ? AND status = ?", contactID, ProcessStatusCompleted).
		Count(&completedCount).Error; err != nil {
		r.logger.Warn("erro ao contar processos completos", zap.Error(err))
	}
	summary.CompletedProcesses = int(completedCount)

	// Calcula taxa de conversão
	if summary.TotalProcesses > 0 {
		summary.ConversionRate = (float64(summary.CompletedProcesses) / float64(summary.TotalProcesses)) * 100
	}

	// Último processo
	var lastProcess models.SalesProcess
	if err := r.db.Model(&models.SalesProcess{}).
		Where("contact_id = ?", contactID).
		Order("created_at DESC").
		First(&lastProcess).Error; err == nil {
		summary.LastProcessDate = lastProcess.CreatedAt
	}

	return summary, nil
}

// InitiateFromQuotation inicia um processo a partir de uma cotação
func (r *salesProcessRepository) InitiateFromQuotation(quotationID int) (*models.SalesProcess, error) {
	// Busca a quotation
	var quotation models.Quotation
	if err := r.db.Preload("Contact").First(&quotation, quotationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrQuotationNotFound
		}
		return nil, errors.WrapError(err, "falha ao buscar quotation")
	}

	// Cria o processo
	process := &models.SalesProcess{
		ContactID:  quotation.ContactID,
		Status:     ProcessStatusQuotation,
		TotalValue: quotation.GrandTotal,
		Notes:      fmt.Sprintf("Processo iniciado a partir da cotação %s", quotation.QuotationNo),
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria o processo
	if err := tx.Create(process).Error; err != nil {
		tx.Rollback()
		return nil, errors.WrapError(err, "falha ao criar processo")
	}

	// Vincula a quotation
	// Aqui precisaríamos de uma tabela de relacionamento ou campo no modelo
	// Por ora, vamos apenas registrar no log
	r.logger.Info("processo iniciado a partir de quotation",
		zap.Int("process_id", process.ID),
		zap.Int("quotation_id", quotationID))

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		return nil, errors.WrapError(err, "falha ao confirmar transação")
	}

	return process, nil
}

// LinkQuotation vincula uma quotation ao processo
func (r *salesProcessRepository) LinkQuotation(processID int, quotationID int) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Verifica se a quotation existe
	var quotation models.Quotation
	if err := r.db.First(&quotation, quotationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrQuotationNotFound
		}
		return errors.WrapError(err, "falha ao buscar quotation")
	}

	// Atualiza o status do processo
	process.Status = ProcessStatusQuotation
	process.TotalValue = quotation.GrandTotal

	if err := r.db.Save(&process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar processo")
	}

	r.logger.Info("quotation vinculada ao processo",
		zap.Int("process_id", processID),
		zap.Int("quotation_id", quotationID))

	return nil
}

// LinkSalesOrder vincula um sales order ao processo
func (r *salesProcessRepository) LinkSalesOrder(processID int, salesOrderID int) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Verifica se o sales order existe
	var salesOrder models.SalesOrder
	if err := r.db.First(&salesOrder, salesOrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesOrderNotFound
		}
		return errors.WrapError(err, "falha ao buscar sales order")
	}

	// Atualiza o status do processo
	process.Status = ProcessStatusSalesOrder
	process.TotalValue = salesOrder.GrandTotal

	if err := r.db.Save(&process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar processo")
	}

	r.logger.Info("sales order vinculado ao processo",
		zap.Int("process_id", processID),
		zap.Int("sales_order_id", salesOrderID))

	return nil
}

// LinkPurchaseOrder vincula um purchase order ao processo
func (r *salesProcessRepository) LinkPurchaseOrder(processID int, purchaseOrderID int) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Verifica se o purchase order existe
	var purchaseOrder models.PurchaseOrder
	if err := r.db.First(&purchaseOrder, purchaseOrderID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPurchaseOrderNotFound
		}
		return errors.WrapError(err, "falha ao buscar purchase order")
	}

	// Atualiza o status do processo se apropriado
	if process.Status == ProcessStatusSalesOrder {
		process.Status = ProcessStatusPurchase
	}

	// Calcula o custo (simplificado - você pode melhorar isso)
	cost := purchaseOrder.GrandTotal
	process.Profit = process.TotalValue - cost

	if err := r.db.Save(&process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar processo")
	}

	r.logger.Info("purchase order vinculado ao processo",
		zap.Int("process_id", processID),
		zap.Int("purchase_order_id", purchaseOrderID))

	return nil
}

// LinkDelivery vincula uma delivery ao processo
func (r *salesProcessRepository) LinkDelivery(processID int, deliveryID int) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Verifica se a delivery existe
	var delivery models.Delivery
	if err := r.db.First(&delivery, deliveryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrDeliveryNotFound
		}
		return errors.WrapError(err, "falha ao buscar delivery")
	}

	// Atualiza o status do processo se apropriado
	if process.Status == ProcessStatusPurchase || process.Status == ProcessStatusSalesOrder {
		process.Status = ProcessStatusDelivery
	}

	if err := r.db.Save(&process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar processo")
	}

	r.logger.Info("delivery vinculada ao processo",
		zap.Int("process_id", processID),
		zap.Int("delivery_id", deliveryID))

	return nil
}

// LinkInvoice vincula uma invoice ao processo
func (r *salesProcessRepository) LinkInvoice(processID int, invoiceID int) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, processID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Verifica se a invoice existe
	var invoice models.Invoice
	if err := r.db.First(&invoice, invoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrInvoiceNotFound
		}
		return errors.WrapError(err, "falha ao buscar invoice")
	}

	// Atualiza o status do processo
	process.Status = ProcessStatusInvoicing

	// Verifica se está totalmente paga
	if invoice.AmountPaid >= invoice.GrandTotal {
		process.Status = ProcessStatusCompleted
	}

	if err := r.db.Save(&process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar processo")
	}

	r.logger.Info("invoice vinculada ao processo",
		zap.Int("process_id", processID),
		zap.Int("invoice_id", invoiceID))

	return nil
}

// UpdateProcessStatus atualiza o status de um processo
func (r *salesProcessRepository) UpdateProcessStatus(id int, status string) error {
	// Verifica se o processo existe
	var process models.SalesProcess
	if err := r.db.First(&process, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrSalesProcessNotFound
		}
		return errors.WrapError(err, "falha ao buscar processo")
	}

	// Atualiza o status
	process.Status = status
	if err := r.db.Save(&process).Error; err != nil {
		r.logger.Error("erro ao atualizar status do processo", zap.Error(err), zap.Int("id", id), zap.String("status", status))
		return errors.WrapError(err, "falha ao atualizar status do processo")
	}

	r.logger.Info("status do processo atualizado", zap.Int("id", id), zap.String("status", status))
	return nil
}

// CalculateProfitability calcula a lucratividade de um processo
func (r *salesProcessRepository) CalculateProfitability(id int) error {
	// Busca o processo com todos os documentos relacionados
	process, err := r.GetCompleteProcessFlow(id)
	if err != nil {
		return err
	}

	// Calcula receita (invoices)
	var revenue float64
	for _, invoice := range process.Invoices {
		revenue += invoice.GrandTotal
	}

	// Calcula custos (purchase orders)
	var costs float64
	for _, po := range process.PurchaseOrders {
		costs += po.GrandTotal
	}

	// Atualiza o processo
	process.Process.TotalValue = revenue
	process.Process.Profit = revenue - costs

	if err := r.db.Save(process.Process).Error; err != nil {
		return errors.WrapError(err, "falha ao atualizar lucratividade")
	}

	r.logger.Info("lucratividade calculada",
		zap.Int("process_id", id),
		zap.Float64("revenue", revenue),
		zap.Float64("costs", costs),
		zap.Float64("profit", process.Process.Profit))

	return nil
}

// GetCompleteProcessFlow retorna o fluxo completo de um processo
func (r *salesProcessRepository) GetCompleteProcessFlow(id int) (*CompleteProcessFlow, error) {
	flow := &CompleteProcessFlow{
		Timeline: make([]ProcessEvent, 0),
	}

	// Busca o processo
	process, err := r.GetSalesProcessByID(id)
	if err != nil {
		return nil, err
	}
	flow.Process = process

	// Carrega todos os documentos relacionados
	// Nota: Em um cenário real, você precisaria de tabelas de relacionamento
	// ou campos de process_id em cada modelo para fazer essas queries

	// Busca quotations do contato (simplificado)
	if err := r.db.Where("contact_id = ?", process.ContactID).
		Order("created_at DESC").
		First(&flow.Quotation).Error; err != nil && err != gorm.ErrRecordNotFound {
		r.logger.Warn("erro ao buscar quotation", zap.Error(err))
	}

	// Busca sales orders
	if err := r.db.Where("contact_id = ?", process.ContactID).
		Order("created_at DESC").
		First(&flow.SalesOrder).Error; err != nil && err != gorm.ErrRecordNotFound {
		r.logger.Warn("erro ao buscar sales order", zap.Error(err))
	}

	// Busca purchase orders
	if err := r.db.Where("sales_order_id = ?", flow.SalesOrder.ID).
		Find(&flow.PurchaseOrders).Error; err != nil {
		r.logger.Warn("erro ao buscar purchase orders", zap.Error(err))
	}

	// Busca deliveries
	if err := r.db.Where("sales_order_id = ?", flow.SalesOrder.ID).
		Find(&flow.Deliveries).Error; err != nil {
		r.logger.Warn("erro ao buscar deliveries", zap.Error(err))
	}

	// Busca invoices
	if err := r.db.Where("sales_order_id = ?", flow.SalesOrder.ID).
		Find(&flow.Invoices).Error; err != nil {
		r.logger.Warn("erro ao buscar invoices", zap.Error(err))
	}

	// Busca payments
	for _, invoice := range flow.Invoices {
		var payments []models.Payment
		if err := r.db.Where("invoice_id = ?", invoice.ID).
			Find(&payments).Error; err == nil {
			flow.Payments = append(flow.Payments, payments...)
		}
	}

	// Monta a timeline
	flow.Timeline = r.buildTimeline(flow)

	return flow, nil
}

// GetProcessTimeline retorna a linha do tempo de um processo
func (r *salesProcessRepository) GetProcessTimeline(id int) (*ProcessTimeline, error) {
	flow, err := r.GetCompleteProcessFlow(id)
	if err != nil {
		return nil, err
	}

	timeline := &ProcessTimeline{
		ProcessID: id,
		Events:    flow.Timeline,
		Status:    flow.Process.Status,
	}

	// Calcula duração
	if len(timeline.Events) > 0 {
		firstEvent := timeline.Events[0]
		lastEvent := timeline.Events[len(timeline.Events)-1]
		duration := lastEvent.Timestamp.Sub(firstEvent.Timestamp)
		timeline.Duration = int(duration.Hours() / 24)
	}

	return timeline, nil
}

// GetProfitabilityAnalysis retorna análise de lucratividade
func (r *salesProcessRepository) GetProfitabilityAnalysis(filter SalesProcessFilter) (*ProfitabilityAnalysis, error) {
	analysis := &ProfitabilityAnalysis{
		ByProduct:  make([]ProductProfitability, 0),
		ByCustomer: make([]CustomerProfitability, 0),
		ByPeriod:   make([]PeriodProfitability, 0),
	}

	// Query base com filtros
	query := r.db.Model(&models.SalesProcess{})
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Totais gerais
	var totals struct {
		Revenue float64
		Profit  float64
	}
	if err := query.Select("SUM(total_value) as revenue, SUM(profit) as profit").
		Scan(&totals).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular totais")
	}

	analysis.TotalRevenue = totals.Revenue
	analysis.TotalProfit = totals.Profit
	analysis.TotalCosts = totals.Revenue - totals.Profit

	if analysis.TotalRevenue > 0 {
		analysis.ProfitMargin = (analysis.TotalProfit / analysis.TotalRevenue) * 100
	}

	// Por cliente
	customerRows, err := query.Joins("JOIN contacts ON contacts.id = sales_processes.contact_id").
		Select("contacts.id, contacts.name, contacts.company_name, COUNT(*) as process_count, SUM(total_value) as revenue, SUM(total_value - profit) as cost, SUM(profit) as profit").
		Group("contacts.id").
		Order("profit DESC").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao calcular lucratividade por cliente")
	}
	defer customerRows.Close()

	for customerRows.Next() {
		var prof CustomerProfitability
		var companyName sql.NullString
		if err := customerRows.Scan(&prof.ContactID, &prof.ContactName, &companyName, &prof.ProcessCount, &prof.Revenue, &prof.Cost, &prof.Profit); err != nil {
			continue
		}

		if companyName.Valid && companyName.String != "" {
			prof.ContactName = companyName.String
		}

		if prof.Revenue > 0 {
			prof.Margin = (prof.Profit / prof.Revenue) * 100
		}

		analysis.ByCustomer = append(analysis.ByCustomer, prof)
	}

	// Processos mais e menos lucrativos
	var topProcesses, bottomProcesses []models.SalesProcess

	query.Order("profit DESC").Limit(5).Find(&topProcesses)
	analysis.TopProfitable = topProcesses

	query.Order("profit ASC").Limit(5).Find(&bottomProcesses)
	analysis.LeastProfitable = bottomProcesses

	return analysis, nil
}

// GetSalesConversionMetrics retorna métricas de conversão
func (r *salesProcessRepository) GetSalesConversionMetrics(filter SalesProcessFilter) (*SalesConversionMetrics, error) {
	metrics := &SalesConversionMetrics{
		ByStage: make(map[string]StageMetrics),
	}

	// Query base
	query := r.db.Model(&models.SalesProcess{})
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("created_at >= ? AND created_at <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Conta total de quotations (simplificado - assumindo que todo processo começa com uma)
	var totalProcesses int64
	query.Count(&totalProcesses)
	metrics.TotalQuotations = int(totalProcesses)

	// Conta por estágio
	stages := []string{
		ProcessStatusQuotation,
		ProcessStatusSalesOrder,
		ProcessStatusPurchase,
		ProcessStatusDelivery,
		ProcessStatusInvoicing,
		ProcessStatusPayment,
		ProcessStatusCompleted,
	}

	previousCount := metrics.TotalQuotations
	for i, stage := range stages {
		var count int64
		query.Where("status = ?", stage).Count(&count)

		stageMetric := StageMetrics{
			Count: int(count),
		}

		if previousCount > 0 {
			stageMetric.ConversionRate = (float64(count) / float64(previousCount)) * 100
			stageMetric.AbandonmentRate = 100 - stageMetric.ConversionRate
		}

		metrics.ByStage[stage] = stageMetric

		// Calcula taxa de conversão específica
		switch stage {
		case ProcessStatusSalesOrder:
			if metrics.TotalQuotations > 0 {
				metrics.QuotationToSORate = (float64(count) / float64(metrics.TotalQuotations)) * 100
			}
		case ProcessStatusInvoicing:
			soCount := metrics.ByStage[ProcessStatusSalesOrder].Count
			if soCount > 0 {
				metrics.SOToInvoiceRate = (float64(count) / float64(soCount)) * 100
			}
		case ProcessStatusCompleted:
			invoiceCount := metrics.ByStage[ProcessStatusInvoicing].Count
			if invoiceCount > 0 {
				metrics.InvoiceToPaymentRate = (float64(count) / float64(invoiceCount)) * 100
			}
			if metrics.TotalQuotations > 0 {
				metrics.OverallConversionRate = (float64(count) / float64(metrics.TotalQuotations)) * 100
			}
		}

		if i > 0 {
			previousCount = int(count)
		}
	}

	// Tempo médio de conversão
	var avgCycleTime struct {
		AvgDays float64
	}
	if err := r.db.Model(&models.SalesProcess{}).
		Where("status = ?", ProcessStatusCompleted).
		Select("AVG(JULIANDAY(updated_at) - JULIANDAY(created_at)) as avg_days").
		Scan(&avgCycleTime).Error; err == nil {
		metrics.AverageConversionTime = avgCycleTime.AvgDays
	}

	return metrics, nil
}

// GetProcessesByStage busca processos por estágio
func (r *salesProcessRepository) GetProcessesByStage(stage string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Mapeia estágio para status
	statusMap := map[string]string{
		"quotation":   ProcessStatusQuotation,
		"sales_order": ProcessStatusSalesOrder,
		"purchase":    ProcessStatusPurchase,
		"delivery":    ProcessStatusDelivery,
		"invoicing":   ProcessStatusInvoicing,
		"payment":     ProcessStatusPayment,
		"completed":   ProcessStatusCompleted,
		"cancelled":   ProcessStatusCancelled,
	}

	status, ok := statusMap[stage]
	if !ok {
		return nil, errors.WrapError(gorm.ErrInvalidData, "estágio inválido")
	}

	return r.GetSalesProcessesByStatus(status, params)
}

// GetAbandonedProcesses busca processos abandonados
func (r *salesProcessRepository) GetAbandonedProcesses(days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var salesProcesses []models.SalesProcess
	var total int64

	cutoffDate := time.Now().AddDate(0, 0, -days)

	query := r.db.Model(&models.SalesProcess{}).
		Where("updated_at < ? AND status NOT IN ?", cutoffDate, []string{ProcessStatusCompleted, ProcessStatusCancelled})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar processos abandonados", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar processos abandonados")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Contact").
		Order("updated_at ASC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&salesProcesses).Error; err != nil {
		r.logger.Error("erro ao buscar processos abandonados", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar processos abandonados")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, salesProcesses)
	return result, nil
}

// Funções auxiliares privadas

// loadRelatedDocuments carrega os documentos relacionados ao processo
func (r *salesProcessRepository) loadRelatedDocuments(process *models.SalesProcess) error {
	// Esta é uma implementação simplificada
	// Em um cenário real, você precisaria de relacionamentos apropriados no banco

	// Carrega quotation
	if err := r.db.Where("contact_id = ?", process.ContactID).
		Order("created_at DESC").
		First(&process.Quotation).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Carrega sales order
	if err := r.db.Where("contact_id = ?", process.ContactID).
		Order("created_at DESC").
		First(&process.SalesOrder).Error; err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// Carrega purchase orders
	if process.SalesOrder != nil {
		if err := r.db.Where("sales_order_id = ?", process.SalesOrder.ID).
			Find(&process.PurchaseOrder).Error; err != nil {
			return err
		}
	}

	// Carrega deliveries
	if process.SalesOrder != nil {
		if err := r.db.Where("sales_order_id = ?", process.SalesOrder.ID).
			Find(&process.Deliveries).Error; err != nil {
			return err
		}
	}

	// Carrega invoices
	if process.SalesOrder != nil {
		if err := r.db.Where("sales_order_id = ?", process.SalesOrder.ID).
			Find(&process.Invoices).Error; err != nil {
			return err
		}
	}

	return nil
}

// buildTimeline constrói a linha do tempo do processo
func (r *salesProcessRepository) buildTimeline(flow *CompleteProcessFlow) []ProcessEvent {
	timeline := make([]ProcessEvent, 0)

	// Adiciona evento de criação do processo
	timeline = append(timeline, ProcessEvent{
		Timestamp:   flow.Process.CreatedAt,
		EventType:   "process_created",
		Description: "Processo de venda iniciado",
		DocumentID:  flow.Process.ID,
	})

	// Adiciona evento da quotation
	if flow.Quotation != nil {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   flow.Quotation.CreatedAt,
			EventType:   "quotation_created",
			Description: fmt.Sprintf("Cotação %s criada", flow.Quotation.QuotationNo),
			DocumentID:  flow.Quotation.ID,
			DocumentNo:  flow.Quotation.QuotationNo,
			Value:       flow.Quotation.GrandTotal,
		})
	}

	// Adiciona evento do sales order
	if flow.SalesOrder != nil {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   flow.SalesOrder.CreatedAt,
			EventType:   "sales_order_created",
			Description: fmt.Sprintf("Pedido de venda %s criado", flow.SalesOrder.SONo),
			DocumentID:  flow.SalesOrder.ID,
			DocumentNo:  flow.SalesOrder.SONo,
			Value:       flow.SalesOrder.GrandTotal,
		})
	}

	// Adiciona eventos de purchase orders
	for _, po := range flow.PurchaseOrders {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   po.CreatedAt,
			EventType:   "purchase_order_created",
			Description: fmt.Sprintf("Ordem de compra %s criada", po.PONo),
			DocumentID:  po.ID,
			DocumentNo:  po.PONo,
			Value:       po.GrandTotal,
		})
	}

	// Adiciona eventos de deliveries
	for _, delivery := range flow.Deliveries {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   delivery.CreatedAt,
			EventType:   "delivery_created",
			Description: fmt.Sprintf("Entrega %s criada", delivery.DeliveryNo),
			DocumentID:  delivery.ID,
			DocumentNo:  delivery.DeliveryNo,
		})
	}

	// Adiciona eventos de invoices
	for _, invoice := range flow.Invoices {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   invoice.CreatedAt,
			EventType:   "invoice_created",
			Description: fmt.Sprintf("Fatura %s criada", invoice.InvoiceNo),
			DocumentID:  invoice.ID,
			DocumentNo:  invoice.InvoiceNo,
			Value:       invoice.GrandTotal,
		})
	}

	// Adiciona eventos de payments
	for _, payment := range flow.Payments {
		timeline = append(timeline, ProcessEvent{
			Timestamp:   payment.PaymentDate,
			EventType:   "payment_received",
			Description: fmt.Sprintf("Pagamento de %.2f recebido", payment.Amount),
			DocumentID:  payment.ID,
			Value:       payment.Amount,
		})
	}

	// Ordena a timeline por timestamp
	// Aqui você usaria um sort.Slice para ordenar

	return timeline
}
