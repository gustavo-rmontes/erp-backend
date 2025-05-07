package repository

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/logger"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PaymentRepository define a interface para operações de repositório de pagamentos
type PaymentRepository interface {
	// Operações CRUD básicas
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id int) (*models.Payment, error)
	GetAllPayments(params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	UpdatePayment(id int, payment *models.Payment) error
	DeletePayment(id int) error

	// Métodos adicionais específicos
	GetPaymentsByInvoice(invoiceID int) ([]models.Payment, error)
	GetPaymentsByDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
}

// gormPaymentRepository é a implementação concreta usando GORM
type gormPaymentRepository struct {
	db  *gorm.DB
	log *zap.Logger
}

// Singleton para o repositório
var (
	paymentRepoInstance *gormPaymentRepository
	paymentRepoOnce     sync.Once
)

// NewPaymentRepository retorna uma instância do repositório de pagamentos
func NewPaymentRepository() (PaymentRepository, error) {
	var initErr error

	paymentRepoOnce.Do(func() {
		conn, err := db.OpenGormDB()
		if err != nil {
			initErr = fmt.Errorf("%w: %v", errors.ErrDatabaseConnection, err)
			return
		}

		// Usar o logger centralizado
		log := logger.WithModule("PaymentRepository")

		paymentRepoInstance = &gormPaymentRepository{
			db:  conn,
			log: log,
		}
	})

	if initErr != nil {
		return nil, initErr
	}

	return paymentRepoInstance, nil
}

// CreatePayment cria um novo pagamento no banco de dados
func (r *gormPaymentRepository) CreatePayment(payment *models.Payment) error {
	r.log.Info("Iniciando criação de pagamento",
		zap.Int("invoice_id", payment.InvoiceID),
		zap.Float64("amount", payment.Amount),
		zap.String("operation", "CreatePayment"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Cria o pagamento
	if err := tx.Create(payment).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao criar pagamento", zap.Error(err))
		return fmt.Errorf("falha ao criar pagamento: %w", err)
	}

	// Atualiza o valor pago da fatura
	var invoice models.Invoice
	if err := tx.First(&invoice, payment.InvoiceID).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao buscar fatura para atualização", zap.Error(err))
		return fmt.Errorf("falha ao buscar fatura: %w", err)
	}

	// Atualiza o valor pago
	invoice.AmountPaid += payment.Amount

	// Atualiza o status da fatura com base no valor pago
	if invoice.AmountPaid >= invoice.GrandTotal {
		invoice.Status = models.InvoiceStatusPaid
	} else if invoice.AmountPaid > 0 {
		invoice.Status = models.InvoiceStatusPartial
	}

	// Salva as alterações na fatura
	if err := tx.Save(&invoice).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar fatura após pagamento", zap.Error(err))
		return fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pagamento criado com sucesso",
		zap.Int("payment_id", payment.ID),
		zap.Int("invoice_id", payment.InvoiceID),
		zap.Float64("amount", payment.Amount),
		zap.String("invoice_status", invoice.Status),
		zap.Float64("invoice_paid", invoice.AmountPaid),
		zap.Float64("invoice_total", invoice.GrandTotal),
	)

	return nil
}

// GetPaymentByID recupera um pagamento pelo seu ID
func (r *gormPaymentRepository) GetPaymentByID(id int) (*models.Payment, error) {
	r.log.Info("Buscando pagamento por ID",
		zap.Int("payment_id", id),
		zap.String("operation", "GetPaymentByID"),
	)

	var payment models.Payment
	if err := r.db.First(&payment, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pagamento não encontrado", zap.Int("payment_id", id))
			return nil, fmt.Errorf("%w: ID %d", errors.ErrPaymentNotFound, id)
		}
		r.log.Error("Erro ao buscar pagamento", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pagamento: %w", err)
	}

	// Carrega informações da fatura se necessário
	if err := r.db.Model(&payment).Association("Invoice").Find(&payment.Invoice); err != nil {
		r.log.Error("Erro ao carregar fatura do pagamento", zap.Error(err))
		return nil, fmt.Errorf("erro ao carregar fatura: %w", err)
	}

	r.log.Info("Pagamento recuperado com sucesso", zap.Int("payment_id", id))
	return &payment, nil
}

// GetAllPayments recupera todos os pagamentos do banco de dados com paginação
func (r *gormPaymentRepository) GetAllPayments(params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Valor padrão para paginação
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	if params != nil {
		if !params.Validate() {
			return nil, errors.ErrInvalidPagination
		}
		page = params.Page
		pageSize = params.PageSize
	}

	r.log.Info("Buscando pagamentos com paginação",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetAllPayments"),
	)

	var totalItems int64
	if err := r.db.Model(&models.Payment{}).Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar total de pagamentos", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pagamentos: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var payments []models.Payment
	if err := r.db.Offset(offset).Limit(pageSize).Order("payment_date DESC").Find(&payments).Error; err != nil {
		r.log.Error("Erro ao buscar pagamentos paginados", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pagamentos: %w", err)
	}

	// Carrega as faturas relacionadas aos pagamentos
	for i := range payments {
		if err := r.db.Model(&payments[i]).Association("Invoice").Find(&payments[i].Invoice); err != nil {
			r.log.Error("Erro ao carregar faturas dos pagamentos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar faturas: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, payments)

	r.log.Info("Pagamentos recuperados com sucesso",
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}

// UpdatePayment atualiza um pagamento existente
func (r *gormPaymentRepository) UpdatePayment(id int, payment *models.Payment) error {
	r.log.Info("Iniciando atualização de pagamento",
		zap.Int("payment_id", id),
		zap.String("operation", "UpdatePayment"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Verifica se o pagamento existe e obtém informações atuais
	var existing models.Payment
	if err := tx.First(&existing, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pagamento não encontrado para atualização", zap.Int("payment_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrPaymentNotFound, id)
		}
		r.log.Error("Erro ao verificar existência do pagamento", zap.Error(err))
		return fmt.Errorf("erro ao verificar pagamento: %w", err)
	}

	// Busca a fatura atual para atualizar os valores
	var invoice models.Invoice
	if err := tx.First(&invoice, existing.InvoiceID).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao buscar fatura atual", zap.Error(err))
		return fmt.Errorf("falha ao buscar fatura: %w", err)
	}

	// Calcula a diferença entre o valor anterior e o novo
	amountDiff := payment.Amount - existing.Amount

	// Atualiza o valor pago da fatura
	invoice.AmountPaid += amountDiff

	// Atualiza o status da fatura com base no valor pago
	if invoice.AmountPaid >= invoice.GrandTotal {
		invoice.Status = models.InvoiceStatusPaid
	} else if invoice.AmountPaid > 0 {
		invoice.Status = models.InvoiceStatusPartial
	} else {
		invoice.Status = models.InvoiceStatusSent
	}

	// Salva as alterações na fatura
	if err := tx.Save(&invoice).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar fatura após modificação de pagamento", zap.Error(err))
		return fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	// Atualiza o pagamento
	payment.ID = id
	if err := tx.Model(&existing).Updates(payment).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar pagamento", zap.Error(err))
		return fmt.Errorf("falha ao atualizar pagamento: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pagamento atualizado com sucesso",
		zap.Int("payment_id", id),
		zap.Float64("amount_diff", amountDiff),
		zap.Int("invoice_id", invoice.ID),
		zap.String("invoice_status", invoice.Status),
		zap.Float64("invoice_paid", invoice.AmountPaid),
	)

	return nil
}

// DeletePayment exclui um pagamento pelo seu ID
func (r *gormPaymentRepository) DeletePayment(id int) error {
	r.log.Info("Iniciando exclusão de pagamento",
		zap.Int("payment_id", id),
		zap.String("operation", "DeletePayment"),
	)

	// Inicia uma transação
	tx := r.db.Begin()
	if tx.Error != nil {
		r.log.Error("Falha ao iniciar transação", zap.Error(tx.Error))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, tx.Error)
	}

	// Recupera o pagamento para obter o valor e o ID da fatura
	var payment models.Payment
	if err := tx.First(&payment, id).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			r.log.Warn("Pagamento não encontrado para exclusão", zap.Int("payment_id", id))
			return fmt.Errorf("%w: ID %d", errors.ErrPaymentNotFound, id)
		}
		r.log.Error("Erro ao buscar pagamento para exclusão", zap.Error(err))
		return fmt.Errorf("erro ao buscar pagamento: %w", err)
	}

	// Recupera a fatura para atualizar o valor pago
	var invoice models.Invoice
	if err := tx.First(&invoice, payment.InvoiceID).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao buscar fatura relacionada ao pagamento", zap.Error(err))
		return fmt.Errorf("falha ao buscar fatura: %w", err)
	}

	// Atualiza o valor pago da fatura
	invoice.AmountPaid -= payment.Amount
	if invoice.AmountPaid < 0 {
		invoice.AmountPaid = 0
	}

	// Atualiza o status da fatura
	if invoice.AmountPaid <= 0 {
		invoice.Status = models.InvoiceStatusSent
	} else if invoice.AmountPaid < invoice.GrandTotal {
		invoice.Status = models.InvoiceStatusPartial
	}

	// Salva as alterações na fatura
	if err := tx.Save(&invoice).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao atualizar fatura após exclusão de pagamento", zap.Error(err))
		return fmt.Errorf("falha ao atualizar fatura: %w", err)
	}

	// Exclui o pagamento
	if err := tx.Delete(&models.Payment{}, id).Error; err != nil {
		tx.Rollback()
		r.log.Error("Falha ao excluir pagamento", zap.Error(err))
		return fmt.Errorf("falha ao excluir pagamento: %w", err)
	}

	// Confirma a transação
	if err := tx.Commit().Error; err != nil {
		r.log.Error("Falha ao confirmar transação", zap.Error(err))
		return fmt.Errorf("%w: %v", errors.ErrTransactionFailed, err)
	}

	r.log.Info("Pagamento excluído com sucesso",
		zap.Int("payment_id", id),
		zap.Float64("amount", payment.Amount),
		zap.Int("invoice_id", invoice.ID),
		zap.String("invoice_status", invoice.Status),
		zap.Float64("invoice_paid", invoice.AmountPaid),
	)

	return nil
}

// GetPaymentsByInvoice recupera pagamentos por ID de fatura
func (r *gormPaymentRepository) GetPaymentsByInvoice(invoiceID int) ([]models.Payment, error) {
	r.log.Info("Buscando pagamentos por fatura",
		zap.Int("invoice_id", invoiceID),
		zap.String("operation", "GetPaymentsByInvoice"),
	)

	var payments []models.Payment
	if err := r.db.Where("invoice_id = ?", invoiceID).Order("payment_date DESC").Find(&payments).Error; err != nil {
		r.log.Error("Erro ao buscar pagamentos por fatura", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pagamentos por fatura: %w", err)
	}

	r.log.Info("Pagamentos por fatura recuperados com sucesso",
		zap.Int("invoice_id", invoiceID),
		zap.Int("count", len(payments)),
	)

	return payments, nil
}

// GetPaymentsByDateRange recupera pagamentos por intervalo de datas com paginação
func (r *gormPaymentRepository) GetPaymentsByDateRange(startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	// Valor padrão para paginação
	page := pagination.DefaultPage
	pageSize := pagination.DefaultPageSize

	if params != nil {
		if !params.Validate() {
			return nil, errors.ErrInvalidPagination
		}
		page = params.Page
		pageSize = params.PageSize
	}

	r.log.Info("Buscando pagamentos por intervalo de datas",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("operation", "GetPaymentsByDateRange"),
	)

	// Ajusta o final do dia para a data final
	endDateAdjusted := endDate.Add(24*time.Hour - time.Second)

	var totalItems int64
	if err := r.db.Model(&models.Payment{}).
		Where("payment_date BETWEEN ? AND ?", startDate, endDateAdjusted).
		Count(&totalItems).Error; err != nil {
		r.log.Error("Erro ao contar pagamentos por intervalo de datas", zap.Error(err))
		return nil, fmt.Errorf("erro ao contar pagamentos: %w", err)
	}

	offset := pagination.CalculateOffset(page, pageSize)

	var payments []models.Payment
	if err := r.db.Where("payment_date BETWEEN ? AND ?", startDate, endDateAdjusted).
		Order("payment_date DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&payments).Error; err != nil {
		r.log.Error("Erro ao buscar pagamentos por intervalo de datas", zap.Error(err))
		return nil, fmt.Errorf("erro ao buscar pagamentos: %w", err)
	}

	// Carrega as faturas relacionadas aos pagamentos
	for i := range payments {
		if err := r.db.Model(&payments[i]).Association("Invoice").Find(&payments[i].Invoice); err != nil {
			r.log.Error("Erro ao carregar faturas dos pagamentos", zap.Error(err))
			return nil, fmt.Errorf("erro ao carregar faturas: %w", err)
		}
	}

	result := pagination.NewPaginatedResult(totalItems, page, pageSize, payments)

	r.log.Info("Pagamentos por intervalo de datas recuperados com sucesso",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.Int64("total_items", totalItems),
		zap.Int("total_pages", result.TotalPages),
	)

	return result, nil
}
