package repository

import (
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"fmt"
	"time"

	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// QuotationRepository define as operações do repositório de quotations
type QuotationRepository interface {
	// CRUD básico
	CreateQuotation(ctx context.Context, quotation *models.Quotation) error
	GetQuotationByID(ctx context.Context, id int) (*models.Quotation, error)
	UpdateQuotation(ctx context.Context, id int, quotation *models.Quotation) error
	DeleteQuotation(ctx context.Context, id int) error

	// Consultas com paginação
	GetAllQuotations(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByStatus(ctx context.Context, status string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByContact(ctx context.Context, contactID int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByDateRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByExpiryRange(ctx context.Context, startDate, endDate time.Time, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetQuotationsByContactType(ctx context.Context, contactType string, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetExpiredQuotations(ctx context.Context, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)
	GetExpiringQuotations(ctx context.Context, days int, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Busca avançada
	SearchQuotations(ctx context.Context, filter QuotationFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error)

	// Apenas para testes (poderia ser movido para um pacote de testes)
	SetCreatedAtForTesting(ctx context.Context, quotationID int, createdAt time.Time) error // mover para testes
}

type quotationRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewQuotationRepository cria uma nova instância do repositório
func NewQuotationRepository(db *gorm.DB, logger *zap.Logger) QuotationRepository {
	return &quotationRepository{
		db:     db,
		logger: logger.With(zap.String("module", "quotation_repository")),
	}
}

// CreateQuotation cria uma nova quotation no banco
func (r *quotationRepository) CreateQuotation(ctx context.Context, quotation *models.Quotation) error {
	// Verificação inicial do contexto
	if ctx.Err() != nil {
		switch ctx.Err() {
		case context.DeadlineExceeded:
			r.logger.Warn("timeout antes de iniciar operação", zap.String("op", "CreateQuotation"))
			return errors.WrapError(ctx.Err(), "timeout ao criar cotação")
		case context.Canceled:
			r.logger.Info("operação cancelada", zap.String("op", "CreateQuotation"))
			return errors.WrapError(ctx.Err(), "operação cancelada pelo cliente")
		default:
			return errors.WrapError(ctx.Err(), "erro de contexto ao criar cotação")
		}
	}

	// Preparação da cotação
	if quotation.QuotationNo == "" {
		// Ideal seria passar o contexto aqui também
		quotation.QuotationNo = r.generateQuotationNumber()
	}

	if quotation.Status == "" {
		quotation.Status = models.QuotationStatusDraft
	}

	// Inicia transação
	tx := r.db.WithContext(ctx).Begin()

	// Verifica novamente o contexto após iniciar transação (pode ter atingido timeout)
	if ctx.Err() != nil {
		tx.Rollback()
		return errors.WrapError(ctx.Err(), "contexto expirou após iniciar transação")
	}

	// Cria a quotation
	if err := tx.Create(quotation).Error; err != nil {
		tx.Rollback()
		r.logger.Error("erro ao criar quotation", zap.Error(err))
		return errors.WrapError(err, "falha ao criar quotation")
	}

	// Insere os itens da cotação
	if len(quotation.Items) > 0 {
		for i := range quotation.Items {
			// Verifica contexto frequentemente em loops longos
			if ctx.Err() != nil {
				tx.Rollback()
				return errors.WrapError(ctx.Err(), "contexto expirou durante criação de itens")
			}

			quotation.Items[i].QuotationID = quotation.ID
			if err := tx.Create(&quotation.Items[i]).Error; err != nil {
				tx.Rollback()
				r.logger.Error("erro ao criar item da quotation",
					zap.Error(err), zap.Int("item_index", i))
				return errors.WrapError(err, fmt.Sprintf("falha ao criar item %d da quotation", i))
			}
		}
	}

	// Verificação final do contexto antes do commit
	if ctx.Err() != nil {
		tx.Rollback()
		return errors.WrapError(ctx.Err(), "contexto expirou antes do commit")
	}

	// Commit da transação
	if err := tx.Commit().Error; err != nil {
		r.logger.Error("erro ao fazer commit", zap.Error(err))
		return errors.WrapError(err, "falha ao confirmar transação")
	}

	r.logger.Info("cotação criada com sucesso",
		zap.Int("id", quotation.ID),
		zap.String("quotation_no", quotation.QuotationNo))
	return nil
}

// GetQuotationByID busca uma quotation pelo ID
func (r *quotationRepository) GetQuotationByID(ctx context.Context, id int) (*models.Quotation, error) {
	var quotation models.Quotation

	query := r.db.WithContext(ctx).Preload("Contact").
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

// UpdateQuotation atualiza uma quotation existente
func (r *quotationRepository) UpdateQuotation(ctx context.Context, id int, quotation *models.Quotation) error {
	// Verifica se a quotation existe
	var existing models.Quotation
	if err := r.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrQuotationNotFound
		}
		return errors.WrapError(err, "falha ao verificar quotation existente")
	}

	// Atualiza os campos
	quotation.ID = id
	if err := r.db.WithContext(ctx).Save(quotation).Error; err != nil {
		r.logger.Error("erro ao atualizar quotation", zap.Error(err), zap.Int("id", id))
		return errors.WrapError(err, "falha ao atualizar quotation")
	}

	r.logger.Info("quotation atualizada com sucesso", zap.Int("id", id))
	return nil
}

// DeleteQuotation remove uma quotation
func (r *quotationRepository) DeleteQuotation(ctx context.Context, id int) error {
	// Verifica se existem sales orders relacionadas
	var salesOrderCount int64
	if err := r.db.WithContext(ctx).Model(&models.SalesOrder{}).Where("quotation_id = ?", id).Count(&salesOrderCount).Error; err != nil {
		return errors.WrapError(err, "falha ao verificar pedidos de venda relacionados")
	}

	if salesOrderCount > 0 {
		return errors.ErrRelatedRecordsExist
	}

	// Remove a quotation (cascade removerá os itens)
	result := r.db.WithContext(ctx).Delete(&models.Quotation{}, id)
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

// generateQuotationNumber gera um número único para a quotation
func (r *quotationRepository) generateQuotationNumber() string {
	var lastQuotation models.Quotation
	err := r.db.Order("id DESC").First(&lastQuotation).Error
	year := time.Now().Year()
	if err != nil {
		// Se não houver registro, inicia a sequência em 1
		if err == gorm.ErrRecordNotFound {
			return fmt.Sprintf("QT-%d-%06d", year, 1)
		}
		// Outras situações, se necessário tratar
	}
	sequence := lastQuotation.ID + 1
	return fmt.Sprintf("QT-%d-%06d", year, sequence)
}

func (r *quotationRepository) generateSalesOrderNumber(tx *gorm.DB) string { // --> mover para SalesOrder
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

// Apenas para uso em testes -> mover para testes
func (r *quotationRepository) SetCreatedAtForTesting(ctx context.Context, quotationID int, createdAt time.Time) error {
	return r.db.WithContext(ctx).Exec("UPDATE quotations SET created_at = ? WHERE id = ?", createdAt, quotationID).Error
}
