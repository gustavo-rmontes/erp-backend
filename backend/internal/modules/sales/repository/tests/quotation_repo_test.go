package repository_test

import (
	"ERP-ONSMART/backend/internal/errors"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/utils/pagination"
	testutils "ERP-ONSMART/backend/internal/utils/test_utils"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_QuotationRepository_NotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	// Cria o repositório com um logger noop (sem saída)
	repo := repository.NewQuotationRepository(dbTest.GormDB, zap.NewNop())

	_, err := repo.GetQuotationByID(context.Background(), 999999)
	assert.ErrorIs(t, err, errors.ErrQuotationNotFound)
}

func Test_ExpiredAndExpiringQuotations(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	// Inicializa o repositório usando testEnv
	repo := repository.NewQuotationRepository(dbTest.GormDB, zap.NewNop())

	// Criar um contexto para as operações
	ctx := context.Background()

	// Cria uma cotação que vai expirar em breve
	expiringQuotation := &models.Quotation{
		ContactID:  1,
		Status:     models.QuotationStatusSent,
		ExpiryDate: time.Now().AddDate(0, 0, 2), // Expira em 2 dias
		SubTotal:   500.0,
		GrandTotal: 500.0,
		Notes:      "Cotação a expirar em breve",
	}
	err := repo.CreateQuotation(ctx, expiringQuotation)
	assert.NoError(t, err)

	// Cria uma cotação já expirada
	expiredQuotation := &models.Quotation{
		ContactID:  1,
		Status:     models.QuotationStatusSent,
		ExpiryDate: time.Now().AddDate(0, 0, -5), // Expirou há 5 dias
		SubTotal:   300.0,
		GrandTotal: 300.0,
		Notes:      "Cotação já expirada",
	}
	err = repo.CreateQuotation(ctx, expiredQuotation)
	assert.NoError(t, err)

	// Busca cotações a expirar em 7 dias
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	expiringResult, err := repo.GetExpiringQuotations(ctx, 7, params)
	assert.NoError(t, err)
	assert.NotNil(t, expiringResult)
	assert.GreaterOrEqual(t, expiringResult.TotalItems, int64(1))

	// Busca cotações expiradas
	expiredResult, err := repo.GetExpiredQuotations(ctx, params)
	assert.NoError(t, err)
	assert.NotNil(t, expiredResult)
	assert.GreaterOrEqual(t, expiredResult.TotalItems, int64(1))

	// Limpa as cotações criadas
	err = repo.DeleteQuotation(ctx, expiringQuotation.ID)
	assert.NoError(t, err)
	err = repo.DeleteQuotation(ctx, expiredQuotation.ID)
	assert.NoError(t, err)
}
