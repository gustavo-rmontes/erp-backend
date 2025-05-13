package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentRepository define as operações do repositório de payments
type PaymentRepository interface {
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id int) (*models.Payment, error)
	GetAllPayments(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdatePayment(id int, payment *models.Payment) error
	DeletePayment(id int) error
	GetPaymentsByInvoice(invoiceID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPaymentsByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPaymentsByMethod(method string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	SearchPayments(filter PaymentFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetPaymentStats(filter PaymentFilter) (*PaymentStats, error)
	GetPaymentMethodStats(startDate, endDate time.Time) (*PaymentMethodStats, error)
	GetDailyPaymentSummary(date time.Time) (*DailyPaymentSummary, error)
	GetMonthlyPaymentSummary(year int, month int) (*MonthlyPaymentSummary, error)
	GetPendingReconciliations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	ReconcilePayment(paymentID int, reference string) error
	ProcessInvoicePayment(invoiceID int, amount float64, method string, reference string) error
	GetPaymentHistory(invoiceID int) ([]models.Payment, error)
}

// PaymentFilter define os filtros para busca avançada
type PaymentFilter struct {
	InvoiceID      int
	ContactID      int
	DateRangeStart time.Time
	DateRangeEnd   time.Time
	MinAmount      float64
	MaxAmount      float64
	PaymentMethod  []string
	HasReference   *bool
	SearchQuery    string
}

// PaymentStats representa estatísticas de payments
type PaymentStats struct {
	TotalPayments     int                `json:"total_payments"`
	TotalAmount       float64            `json:"total_amount"`
	AverageAmount     float64            `json:"average_amount"`
	CountByMethod     map[string]int     `json:"count_by_method"`
	AmountByMethod    map[string]float64 `json:"amount_by_method"`
	TodayPayments     int                `json:"today_payments"`
	TodayAmount       float64            `json:"today_amount"`
	ThisMonthPayments int                `json:"this_month_payments"`
	ThisMonthAmount   float64            `json:"this_month_amount"`
}

// PaymentMethodStats representa estatísticas por método de pagamento
type PaymentMethodStats struct {
	Method        string  `json:"method"`
	Count         int     `json:"count"`
	TotalAmount   float64 `json:"total_amount"`
	AverageAmount float64 `json:"average_amount"`
	Percentage    float64 `json:"percentage"`
}

// DailyPaymentSummary representa resumo diário de pagamentos
type DailyPaymentSummary struct {
	Date          time.Time                `json:"date"`
	TotalPayments int                      `json:"total_payments"`
	TotalAmount   float64                  `json:"total_amount"`
	ByMethod      []PaymentMethodStats     `json:"by_method"`
	ByHour        map[int]PaymentHourStats `json:"by_hour"`
}

// PaymentHourStats representa estatísticas por hora
type PaymentHourStats struct {
	Hour   int     `json:"hour"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// MonthlyPaymentSummary representa resumo mensal de pagamentos
type MonthlyPaymentSummary struct {
	Year          int                     `json:"year"`
	Month         int                     `json:"month"`
	TotalPayments int                     `json:"total_payments"`
	TotalAmount   float64                 `json:"total_amount"`
	ByMethod      []PaymentMethodStats    `json:"by_method"`
	ByDay         map[int]DayPaymentStats `json:"by_day"`
	Comparison    PaymentComparisonStats  `json:"comparison"`
}

// DayPaymentStats representa estatísticas por dia
type DayPaymentStats struct {
	Day    int     `json:"day"`
	Count  int     `json:"count"`
	Amount float64 `json:"amount"`
}

// PaymentComparisonStats representa comparação com período anterior
type PaymentComparisonStats struct {
	PreviousMonthAmount float64 `json:"previous_month_amount"`
	PreviousMonthCount  int     `json:"previous_month_count"`
	AmountGrowth        float64 `json:"amount_growth_percentage"`
	CountGrowth         float64 `json:"count_growth_percentage"`
}

type paymentRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPaymentRepository cria uma nova instância do repositório
func NewPaymentRepository() (PaymentRepository, error) {
	db, err := db.OpenGormDB()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao abrir conexão com o banco")
	}

	return &paymentRepository{
		db:     db,
		logger: logger.WithModule("payment_repository"),
	}, nil
}

// CreatePayment cria um novo payment no banco
func (r *paymentRepository) CreatePayment(payment *models.Payment) error {
	// Valida se a invoice existe
	var invoice models.Invoice
	if err := r.db.First(&invoice, payment.InvoiceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrInvoiceNotFound
		}
		return errors.WrapError(err, "falha ao verificar invoice")
	}

	// Inicia transação
	tx := r.db.Begin()

	// Cria o payment
	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar payment", zap.Error(err))
		return errors.WrapError(err, "falha ao criar payment")
	}

	// Atualiza o valor pago na invoice
	totalPaid := invoice.AmountPaid + payment.Amount
	updateData := map[string]interface{}{
		"amount_paid": totalPaid,
	}

	// Atualiza o status da invoice se necessário
	if totalPaid >= invoice.GrandTotal {
		updateData["status"] = models.InvoiceStatusPaid
	} else if totalPaid > 0 {
		updateData["status"] = models.InvoiceStatusPartial
	}

	if err := tx.Model(&models.Invoice{}).Where("id = ?", payment.InvoiceID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao atualizar invoice", zap.Error(err))
		return errors.WrapError(err, "falha ao atualizar invoice")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("payment criado com sucesso", zap.Int("id", payment.ID), zap.Float64("amount", payment.Amount))
	return nil
}

// GetPaymentByID busca um payment pelo ID
func (r *paymentRepository) GetPaymentByID(id int) (*models.Payment, error) {
	var payment models.Payment

	query := r.db.Preload("Invoice").
		Preload("Invoice.Contact")

	if err := query.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrPaymentNotFound
		}
		r.logger.Error("erro ao buscar payment por ID", zap.Error(err), zap.Int("id", id))
		return nil, errors.WrapError(err, "falha ao buscar payment")
	}

	return &payment, nil
}

// GetAllPayments retorna todos os payments com paginação
func (r *paymentRepository) GetAllPayments(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	// Query base
	query := r.db.Model(&models.Payment{})

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar payments", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar payments")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Invoice").
		Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar payments", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar payments")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// UpdatePayment atualiza um payment existente
func (r *paymentRepository) UpdatePayment(id int, payment *models.Payment) error {
	// Verifica se o payment existe
	var existing models.Payment
	if err := r.db.First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPaymentNotFound
		}
		return errors.WrapError(err, "falha ao verificar payment existente")
	}

	// Busca a invoice para atualizar o valor pago
	var invoice models.Invoice
	if err := r.db.First(&invoice, existing.InvoiceID).Error; err != nil {
		return errors.WrapError(err, "falha ao buscar invoice")
	}

	// Inicia transação
	tx := r.db.Begin()

	// Calcula a diferença do valor
	diff := payment.Amount - existing.Amount
	newAmountPaid := invoice.AmountPaid + diff

	// Atualiza o payment
	payment.ID = id
	if err := tx.Save(payment).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao atualizar payment", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar payment")
	}

	// Atualiza a invoice
	updateData := map[string]interface{}{
		"amount_paid": newAmountPaid,
	}

	// Atualiza o status da invoice se necessário
	if newAmountPaid >= invoice.GrandTotal {
		updateData["status"] = models.InvoiceStatusPaid
	} else if newAmountPaid > 0 {
		updateData["status"] = models.InvoiceStatusPartial
	} else {
		updateData["status"] = models.InvoiceStatusSent
	}

	if err := tx.Model(&models.Invoice{}).Where("id = ?", existing.InvoiceID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao atualizar invoice", zap.Error(err))
		return errors.WrapError(err, "falha ao atualizar invoice")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("payment atualizado com sucesso", zap.Int("id", id))
	return nil
}

// DeletePayment remove um payment
func (r *paymentRepository) DeletePayment(id int) error {
	// Busca o payment
	var payment models.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPaymentNotFound
		}
		return errors.WrapError(err, "falha ao buscar payment")
	}

	// Busca a invoice para atualizar o valor pago
	var invoice models.Invoice
	if err := r.db.First(&invoice, payment.InvoiceID).Error; err != nil {
		return errors.WrapError(err, "falha ao buscar invoice")
	}

	// Inicia transação
	tx := r.db.Begin()

	// Remove o payment
	if err := tx.Delete(&payment).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao deletar payment", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao deletar payment")
	}

	// Atualiza a invoice
	newAmountPaid := invoice.AmountPaid - payment.Amount
	updateData := map[string]interface{}{
		"amount_paid": newAmountPaid,
	}

	// Atualiza o status da invoice se necessário
	if newAmountPaid >= invoice.GrandTotal {
		updateData["status"] = models.InvoiceStatusPaid
	} else if newAmountPaid > 0 {
		updateData["status"] = models.InvoiceStatusPartial
	} else {
		updateData["status"] = models.InvoiceStatusSent
	}

	if err := tx.Model(&models.Invoice{}).Where("id = ?", payment.InvoiceID).Updates(updateData).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao atualizar invoice", zap.Error(err))
		return errors.WrapError(err, "falha ao atualizar invoice")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit da transação", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("payment deletado com sucesso", zap.Int("id", id))
	return nil
}

// GetPaymentsByInvoice busca payments por invoice
func (r *paymentRepository) GetPaymentsByInvoice(invoiceID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("invoice_id = ?", invoiceID)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar payments por invoice", zap.Error(err), zap.Int("invoice_id", invoiceID))
		return nil, errors.WrapError(err, "falha ao contar payments por invoice")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar payments por invoice", zap.Error(err), zap.Int("invoice_id", invoiceID))
		return nil, errors.WrapError(err, "falha ao buscar payments por invoice")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// GetPaymentsByPeriod busca payments por período
func (r *paymentRepository) GetPaymentsByPeriod(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar payments por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar payments por período")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Invoice").
		Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar payments por período", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar payments por período")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// GetPaymentsByMethod busca payments por método de pagamento
func (r *paymentRepository) GetPaymentsByMethod(method string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{}).Where("payment_method = ?", method)

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar payments por método", zap.Error(err), zap.String("method", method))
		return nil, errors.WrapError(err, "falha ao contar payments por método")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Invoice").
		Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar payments por método", zap.Error(err), zap.String("method", method))
		return nil, errors.WrapError(err, "falha ao buscar payments por método")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// SearchPayments busca payments com filtros combinados
func (r *paymentRepository) SearchPayments(filter PaymentFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{})

	// Aplica os filtros
	if filter.InvoiceID > 0 {
		query = query.Where("invoice_id = ?", filter.InvoiceID)
	}

	// Filtro por contato (através da invoice)
	if filter.ContactID > 0 {
		invoiceSubquery := r.db.Model(&models.Invoice{}).Select("id").Where("contact_id = ?", filter.ContactID)
		query = query.Where("invoice_id IN (?)", invoiceSubquery)
	}

	// Filtros de data
	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("payment_date >= ? AND payment_date <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Filtros de valor
	if filter.MinAmount > 0 {
		query = query.Where("amount >= ?", filter.MinAmount)
	}

	if filter.MaxAmount > 0 {
		query = query.Where("amount <= ?", filter.MaxAmount)
	}

	// Filtro por método de pagamento
	if len(filter.PaymentMethod) > 0 {
		query = query.Where("payment_method IN ?", filter.PaymentMethod)
	}

	// Filtro por referência
	if filter.HasReference != nil {
		if *filter.HasReference {
			query = query.Where("reference IS NOT NULL AND reference != ''")
		} else {
			query = query.Where("reference IS NULL OR reference = ''")
		}
	}

	// Busca textual
	if filter.SearchQuery != "" {
		searchPattern := "%" + filter.SearchQuery + "%"
		query = query.Where("reference LIKE ? OR notes LIKE ?", searchPattern, searchPattern)
	}

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar payments na busca", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar payments na busca")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Invoice").
		Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar payments", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar payments")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// GetPaymentStats retorna estatísticas de payments
func (r *paymentRepository) GetPaymentStats(filter PaymentFilter) (*PaymentStats, error) {
	stats := &PaymentStats{
		CountByMethod:  make(map[string]int),
		AmountByMethod: make(map[string]float64),
	}

	query := r.db.Model(&models.Payment{})

	// Aplica filtros básicos
	if filter.InvoiceID > 0 {
		query = query.Where("invoice_id = ?", filter.InvoiceID)
	}

	if !filter.DateRangeStart.IsZero() && !filter.DateRangeEnd.IsZero() {
		query = query.Where("payment_date >= ? AND payment_date <= ?", filter.DateRangeStart, filter.DateRangeEnd)
	}

	// Contagem total e valores
	var result struct {
		Count       int
		TotalAmount float64
		AvgAmount   float64
	}

	if err := query.Select("COUNT(*) as count, SUM(amount) as total_amount, AVG(amount) as avg_amount").
		Scan(&result).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas")
	}

	stats.TotalPayments = result.Count
	stats.TotalAmount = result.TotalAmount
	stats.AverageAmount = result.AvgAmount

	// Estatísticas por método de pagamento
	rows, err := query.Select("payment_method, COUNT(*) as count, SUM(amount) as total").
		Group("payment_method").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas por método")
	}
	defer rows.Close()

	for rows.Next() {
		var method string
		var count int
		var total float64
		if err := rows.Scan(&method, &count, &total); err != nil {
			continue
		}
		stats.CountByMethod[method] = count
		stats.AmountByMethod[method] = total
	}

	// Estatísticas de hoje
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	var todayStats struct {
		Count int
		Total float64
	}
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", today, tomorrow).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Scan(&todayStats).Error; err != nil {
		r.logger.Warn("erro ao calcular estatísticas de hoje", zap.Error(err))
	}
	stats.TodayPayments = todayStats.Count
	stats.TodayAmount = todayStats.Total

	// Estatísticas do mês
	firstDay := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	lastDay := firstDay.AddDate(0, 1, 0)

	var monthStats struct {
		Count int
		Total float64
	}
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", firstDay, lastDay).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Scan(&monthStats).Error; err != nil {
		r.logger.Warn("erro ao calcular estatísticas do mês", zap.Error(err))
	}
	stats.ThisMonthPayments = monthStats.Count
	stats.ThisMonthAmount = monthStats.Total

	return stats, nil
}

// GetPaymentMethodStats retorna estatísticas por método de pagamento
func (r *paymentRepository) GetPaymentMethodStats(startDate, endDate time.Time) (*PaymentMethodStats, error) {
	// Query base com período
	query := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date <= ?", startDate, endDate)

	// Total geral para calcular percentuais
	var totalGeneral struct {
		Count int
		Total float64
	}
	if err := query.Select("COUNT(*) as count, SUM(amount) as total").
		Scan(&totalGeneral).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular total geral")
	}

	// Estatísticas por método
	var methodStats []PaymentMethodStats
	rows, err := query.Select("payment_method, COUNT(*) as count, SUM(amount) as total_amount, AVG(amount) as average_amount").
		Group("payment_method").
		Order("total_amount DESC").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas por método")
	}
	defer rows.Close()

	for rows.Next() {
		var stat PaymentMethodStats
		if err := rows.Scan(&stat.Method, &stat.Count, &stat.TotalAmount, &stat.AverageAmount); err != nil {
			continue
		}

		// Calcula percentual
		if totalGeneral.Total > 0 {
			stat.Percentage = (stat.TotalAmount / totalGeneral.Total) * 100
		}

		methodStats = append(methodStats, stat)
	}

	// Retorna o primeiro método (mais usado) ou vazio se não houver dados
	if len(methodStats) > 0 {
		return &methodStats[0], nil
	}

	return &PaymentMethodStats{}, nil
}

// GetDailyPaymentSummary retorna resumo diário de pagamentos
func (r *paymentRepository) GetDailyPaymentSummary(date time.Time) (*DailyPaymentSummary, error) {
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	summary := &DailyPaymentSummary{
		Date:     startOfDay,
		ByHour:   make(map[int]PaymentHourStats),
		ByMethod: make([]PaymentMethodStats, 0),
	}

	// Total do dia
	var dayTotal struct {
		Count int
		Total float64
	}
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", startOfDay, endOfDay).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Scan(&dayTotal).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular total do dia")
	}
	summary.TotalPayments = dayTotal.Count
	summary.TotalAmount = dayTotal.Total

	// Por método de pagamento
	methodQuery := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", startOfDay, endOfDay)

	rows, err := methodQuery.Select("payment_method, COUNT(*) as count, SUM(amount) as total_amount, AVG(amount) as average_amount").
		Group("payment_method").
		Order("total_amount DESC").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas por método")
	}
	defer rows.Close()

	for rows.Next() {
		var stat PaymentMethodStats
		if err := rows.Scan(&stat.Method, &stat.Count, &stat.TotalAmount, &stat.AverageAmount); err != nil {
			continue
		}

		// Calcula percentual
		if summary.TotalAmount > 0 {
			stat.Percentage = (stat.TotalAmount / summary.TotalAmount) * 100
		}

		summary.ByMethod = append(summary.ByMethod, stat)
	}

	// Por hora
	hourRows, err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", startOfDay, endOfDay).
		Select("HOUR(payment_date) as hour, COUNT(*) as count, SUM(amount) as amount").
		Group("HOUR(payment_date)").
		Rows()
	if err != nil {
		r.logger.Warn("erro ao calcular estatísticas por hora", zap.Error(err))
	} else {
		defer hourRows.Close()
		for hourRows.Next() {
			var hour int
			var stat PaymentHourStats
			if err := hourRows.Scan(&hour, &stat.Count, &stat.Amount); err != nil {
				continue
			}
			stat.Hour = hour
			summary.ByHour[hour] = stat
		}
	}

	return summary, nil
}

// GetMonthlyPaymentSummary retorna resumo mensal de pagamentos
func (r *paymentRepository) GetMonthlyPaymentSummary(year int, month int) (*MonthlyPaymentSummary, error) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, 0)

	summary := &MonthlyPaymentSummary{
		Year:     year,
		Month:    month,
		ByDay:    make(map[int]DayPaymentStats),
		ByMethod: make([]PaymentMethodStats, 0),
	}

	// Total do mês
	var monthTotal struct {
		Count int
		Total float64
	}
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", firstDay, lastDay).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Scan(&monthTotal).Error; err != nil {
		return nil, errors.WrapError(err, "falha ao calcular total do mês")
	}
	summary.TotalPayments = monthTotal.Count
	summary.TotalAmount = monthTotal.Total

	// Por método de pagamento
	methodQuery := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", firstDay, lastDay)

	rows, err := methodQuery.Select("payment_method, COUNT(*) as count, SUM(amount) as total_amount, AVG(amount) as average_amount").
		Group("payment_method").
		Order("total_amount DESC").
		Rows()
	if err != nil {
		return nil, errors.WrapError(err, "falha ao calcular estatísticas por método")
	}
	defer rows.Close()

	for rows.Next() {
		var stat PaymentMethodStats
		if err := rows.Scan(&stat.Method, &stat.Count, &stat.TotalAmount, &stat.AverageAmount); err != nil {
			continue
		}

		// Calcula percentual
		if summary.TotalAmount > 0 {
			stat.Percentage = (stat.TotalAmount / summary.TotalAmount) * 100
		}

		summary.ByMethod = append(summary.ByMethod, stat)
	}

	// Por dia
	dayRows, err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", firstDay, lastDay).
		Select("DAY(payment_date) as day, COUNT(*) as count, SUM(amount) as amount").
		Group("DAY(payment_date)").
		Rows()
	if err != nil {
		r.logger.Warn("erro ao calcular estatísticas por dia", zap.Error(err))
	} else {
		defer dayRows.Close()
		for dayRows.Next() {
			var day int
			var stat DayPaymentStats
			if err := dayRows.Scan(&day, &stat.Count, &stat.Amount); err != nil {
				continue
			}
			stat.Day = day
			summary.ByDay[day] = stat
		}
	}

	// Comparação com mês anterior
	prevFirstDay := firstDay.AddDate(0, -1, 0)
	prevLastDay := firstDay

	var prevMonthStats struct {
		Count int
		Total float64
	}
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date >= ? AND payment_date < ?", prevFirstDay, prevLastDay).
		Select("COUNT(*) as count, COALESCE(SUM(amount), 0) as total").
		Scan(&prevMonthStats).Error; err != nil {
		r.logger.Warn("erro ao calcular estatísticas do mês anterior", zap.Error(err))
	}

	summary.Comparison.PreviousMonthAmount = prevMonthStats.Total
	summary.Comparison.PreviousMonthCount = prevMonthStats.Count

	// Calcula crescimento
	if prevMonthStats.Total > 0 {
		summary.Comparison.AmountGrowth = ((summary.TotalAmount - prevMonthStats.Total) / prevMonthStats.Total) * 100
	}
	if prevMonthStats.Count > 0 {
		summary.Comparison.CountGrowth = ((float64(summary.TotalPayments) - float64(prevMonthStats.Count)) / float64(prevMonthStats.Count)) * 100
	}

	return summary, nil
}

// GetPendingReconciliations busca pagamentos pendentes de reconciliação
func (r *paymentRepository) GetPendingReconciliations(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var payments []models.Payment
	var total int64

	// Pagamentos sem referência
	query := r.db.Model(&models.Payment{}).Where("reference IS NULL OR reference = ''")

	// Conta o total
	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("erro ao contar pagamentos pendentes de reconciliação", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao contar pagamentos pendentes")
	}

	// Aplica paginação e busca os dados
	offset := pagination.CalculateOffset(params.Page, params.PageSize)
	if err := query.Preload("Invoice").
		Order("payment_date DESC").
		Limit(params.PageSize).
		Offset(offset).
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar pagamentos pendentes", zap.Error(err))
		return nil, errors.WrapError(err, "falha ao buscar pagamentos pendentes")
	}

	result := pagination.NewPaginatedResult(total, params.Page, params.PageSize, payments)
	return result, nil
}

// ReconcilePayment reconcilia um pagamento com uma referência
func (r *paymentRepository) ReconcilePayment(paymentID int, reference string) error {
	// Busca o payment
	var payment models.Payment
	if err := r.db.First(&payment, paymentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrPaymentNotFound
		}
		return errors.WrapError(err, "falha ao buscar payment")
	}

	// Atualiza a referência
	payment.Reference = reference
	if err := r.db.Save(&payment).Error; err != nil {
		r.logger.Error("erro ao reconciliar payment", zap.Error(err), zap.Int("payment_id", paymentID))
		return errors.WrapError(err, "falha ao reconciliar payment")
	}

	r.logger.Info("payment reconciliado com sucesso", zap.Int("payment_id", paymentID), zap.String("reference", reference))
	return nil
}

// ProcessInvoicePayment processa um pagamento para uma invoice
func (r *paymentRepository) ProcessInvoicePayment(invoiceID int, amount float64, method string, reference string) error {
	payment := &models.Payment{
		InvoiceID:     invoiceID,
		Amount:        amount,
		PaymentMethod: method,
		Reference:     reference,
		PaymentDate:   time.Now(),
	}

	return r.CreatePayment(payment)
}

// GetPaymentHistory retorna o histórico de pagamentos de uma invoice
func (r *paymentRepository) GetPaymentHistory(invoiceID int) ([]models.Payment, error) {
	var payments []models.Payment

	if err := r.db.Where("invoice_id = ?", invoiceID).
		Order("payment_date DESC").
		Find(&payments).Error; err != nil {
		r.logger.Error("erro ao buscar histórico de pagamentos", zap.Error(err), zap.Int("invoice_id", invoiceID))
		return nil, errors.WrapError(err, "falha ao buscar histórico de pagamentos")
	}

	return payments, nil
}
