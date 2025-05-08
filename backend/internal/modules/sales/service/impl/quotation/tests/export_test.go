// export_test.go
package tests

import (
	"testing"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestGeneratePDF testa a geração de PDF de cotação
func TestGeneratePDF(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)

	// Expectativa: buscar cotação
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	// Act
	pdfBytes, err := service.GeneratePDF(ctx, quotationID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pdfBytes)
	assert.NotEmpty(t, pdfBytes)
}

// TestExportToCSV testa a exportação de cotações para CSV
func TestExportToCSV(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationIDs := []int{1, 2}
	quotation1 := createTestQuotation(1)
	quotation2 := createTestQuotation(2)

	// Expectativa: buscar cada cotação
	mockQuotationRepo.EXPECT().
		GetQuotationByID(1).
		Return(quotation1, nil)

	mockQuotationRepo.EXPECT().
		GetQuotationByID(2).
		Return(quotation2, nil)

	// Act
	csvBytes, err := service.ExportToCSV(ctx, quotationIDs)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, csvBytes)
	assert.NotEmpty(t, csvBytes)
}

// TestSendByEmail testa o envio de cotação por e-mail
func TestSendByEmail(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusDraft

	options := &dto.EmailOptions{
		To:        []string{"cliente@exemplo.com"},
		Subject:   "Cotação TEST-QT-2025-0001",
		Message:   "Segue em anexo a cotação solicitada.", // Use Message em vez de Body/Content
		AttachPDF: true,
	}
	// Expectativa: buscar cotação e atualizar status
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockQuotationRepo.EXPECT().
		UpdateQuotation(quotationID, gomock.Any()).
		DoAndReturn(func(id int, updatedQuotation *models.Quotation) error {
			// Verificar se o status foi atualizado para enviado
			assert.Equal(t, models.QuotationStatusSent, updatedQuotation.Status)
			return nil
		})

	// Act
	err := service.SendByEmail(ctx, quotationID, options)

	// Assert
	assert.NoError(t, err)
}

// TestSendByEmail_WithoutAttachPDF testa o envio sem alteração de status
func TestSendByEmail_WithoutAttachPDF(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusDraft

	options := &dto.EmailOptions{
		To:        []string{"cliente@exemplo.com"},
		Subject:   "Cotação TEST-QT-2025-0001",
		Message:   "Informações sobre a cotação.",
		AttachPDF: false, // Sem anexar PDF, não deve alterar status
	}

	// Expectativa: apenas buscar cotação, sem atualizar
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	// Act
	err := service.SendByEmail(ctx, quotationID, options)

	// Assert
	assert.NoError(t, err)
}
