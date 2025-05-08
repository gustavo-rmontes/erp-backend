// status_test.go
package tests

import (
	"testing"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestUpdateStatus testa a atualização de status de uma cotação
func TestUpdateStatus(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusDraft

	updateReq := &dto.QuotationStatusUpdateRequest{
		Status: models.QuotationStatusSent,
	}

	// Expectativa: buscar cotação, atualizar e buscar novamente
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, updatedQuotation *models.Quotation) error {
			// Verificar se o status foi atualizado
			assert.Equal(t, updateReq.Status, updatedQuotation.Status)
			return nil
		})

	updatedQuotation := *quotation
	updatedQuotation.Status = models.QuotationStatusSent

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(&updatedQuotation, nil)

	// Act
	result, err := service.UpdateStatus(ctx, quotationID, updateReq)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, models.QuotationStatusSent, result.Status)
}

// TestUpdateStatus_InvalidTransition testa uma transição de status inválida
func TestUpdateStatus_InvalidTransition(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusRejected // Já está rejeitada

	updateReq := &dto.QuotationStatusUpdateRequest{
		Status: models.QuotationStatusSent, // Tentar enviar algo rejeitado (inválido)
	}

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	// Act
	result, err := service.UpdateStatus(ctx, quotationID, updateReq)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "transição de status inválida")
}

// TestProcessExpirations testa o processamento de cotações expiradas
func TestProcessExpirations(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	// Cotações expiradas (2 cotações)
	quotation1 := createTestQuotation(1)
	quotation1.Status = models.QuotationStatusSent

	quotation2 := createTestQuotation(2)
	quotation2.Status = models.QuotationStatusSent

	// Criar resultado paginado
	paginatedResult := &pagination.PaginatedResult{
		Items:       []models.Quotation{*quotation1, *quotation2},
		TotalItems:  2,
		CurrentPage: 1,
		PageSize:    10,
		TotalPages:  1,
	}

	// Expectativa: buscar cotações expiradas e atualizar status
	mockQuotationRepo.EXPECT().
		GetExpiredQuotations(gomock.Any()).
		Return(paginatedResult, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotation1.ID, gomock.Any()).
		DoAndReturn(func(id int, q *models.Quotation) error {
			assert.Equal(t, models.QuotationStatusExpired, q.Status)
			return nil
		})

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotation2.ID, gomock.Any()).
		DoAndReturn(func(id int, q *models.Quotation) error {
			assert.Equal(t, models.QuotationStatusExpired, q.Status)
			return nil
		})

	// Act
	count, err := service.ProcessExpirations(ctx)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, count) // 2 cotações processadas
}

// TestNotifyExpiringQuotations testa a notificação de cotações prestes a expirar
func TestNotifyExpiringQuotations(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	daysBeforeExpiry := 7

	// Cotações prestes a expirar
	quotation1 := createTestQuotation(1)
	quotation1.Status = models.QuotationStatusSent
	quotation1.ExpiryDate = time.Now().AddDate(0, 0, 3) // Expira em 3 dias

	quotation2 := createTestQuotation(2)
	quotation2.Status = models.QuotationStatusSent
	quotation2.ExpiryDate = time.Now().AddDate(0, 0, 5) // Expira em 5 dias

	// Cotação que não vai expirar em breve
	quotation3 := createTestQuotation(3)
	quotation3.Status = models.QuotationStatusSent
	quotation3.ExpiryDate = time.Now().AddDate(0, 0, 30) // Expira em 30 dias

	// Criar resultado paginado
	paginatedResult := &pagination.PaginatedResult{
		Items:       []models.Quotation{*quotation1, *quotation2, *quotation3},
		TotalItems:  3,
		CurrentPage: 1,
		PageSize:    10,
		TotalPages:  1,
	}

	// Expectativa: buscar todas as cotações
	mockQuotationRepo.EXPECT().
		GetAllQuotations(gomock.Any()).
		Return(paginatedResult, nil)

	// Act
	count, err := service.NotifyExpiringQuotations(ctx, daysBeforeExpiry)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, count) // 2 cotações prestes a expirar
}
