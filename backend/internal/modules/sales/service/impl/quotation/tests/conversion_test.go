// conversion_test.go
package tests

import (
	"testing"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestConvertToSalesOrder testa a conversão de cotação para pedido de venda
func TestConvertToSalesOrder(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, mockSalesOrderRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusSent

	data := &dto.SalesOrderFromQuotationCreate{
		ExpectedDate:    time.Now().AddDate(0, 0, 15),
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua Exemplo, 123",
	}

	// Expectativa: buscar cotação e criar pedido
	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	mockSalesOrderRepo.EXPECT().
		CreateSalesOrder(gomock.Any()).
		DoAndReturn(func(salesOrder *models.SalesOrder) error {
			// Verificar dados do pedido de venda
			assert.Equal(t, quotationID, salesOrder.QuotationID)
			assert.Equal(t, quotation.ContactID, salesOrder.ContactID)
			assert.Equal(t, models.SOStatusDraft, salesOrder.Status)
			assert.Equal(t, data.ExpectedDate, salesOrder.ExpectedDate)
			assert.Equal(t, data.PaymentTerms, salesOrder.PaymentTerms)
			assert.Equal(t, data.ShippingAddress, salesOrder.ShippingAddress)
			assert.Equal(t, quotation.GrandTotal, salesOrder.GrandTotal)
			assert.Equal(t, len(quotation.Items), len(salesOrder.Items))

			// Simular atribuição de ID
			salesOrder.ID = 5

			return nil
		})

	// Act
	result, err := service.ConvertToSalesOrder(ctx, quotationID, data)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 5, result.ID)
	assert.Equal(t, quotationID, result.QuotationID)
	assert.Equal(t, quotation.GrandTotal, result.GrandTotal)
	assert.Equal(t, data.PaymentTerms, result.PaymentTerms)
}

// TestConvertToSalesOrder_InvalidStatus testa conversão com status inválido
func TestConvertToSalesOrder_InvalidStatus(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	quotation := createTestQuotation(quotationID)
	quotation.Status = models.QuotationStatusRejected // Status inválido para conversão

	data := &dto.SalesOrderFromQuotationCreate{
		ExpectedDate:    time.Now().AddDate(0, 0, 15),
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua Exemplo, 123",
	}

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(quotation, nil)

	// Act
	result, err := service.ConvertToSalesOrder(ctx, quotationID, data)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cotação em status inválido para conversão")
}

// TestClone testa a clonagem de uma cotação
func TestClone(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	quotationID := 1
	original := createTestQuotation(quotationID)

	options := &dto.QuotationCloneOptions{
		ContactID:       original.ContactID,
		ExpiryDate:      time.Now().AddDate(0, 2, 0),
		CopyNotes:       true,
		AdjustPrices:    true,
		PriceAdjustment: 10.0, // Aumento de 10%
	}

	mockQuotationRepo.EXPECT().
		GetQuotationByID(quotationID).
		Return(original, nil)

	mockQuotationRepo.EXPECT().
		CreateQuotation(gomock.Any()).
		DoAndReturn(func(quotation *models.Quotation) error {
			// Verificar dados da cotação clonada
			assert.Equal(t, original.ContactID, quotation.ContactID)
			assert.Equal(t, models.QuotationStatusDraft, quotation.Status)
			assert.Equal(t, options.ExpiryDate, quotation.ExpiryDate)
			assert.NotEqual(t, original.QuotationNo, quotation.QuotationNo)

			// Verificar ajuste de preço
			if len(quotation.Items) > 0 {
				expectedPrice := original.Items[0].UnitPrice * 1.1
				assert.InDelta(t, expectedPrice, quotation.Items[0].UnitPrice, 0.01)
			}

			// Simular atribuição de ID
			quotation.ID = 2

			return nil
		})

	// Act
	result, err := service.Clone(ctx, quotationID, options)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.ID)
	assert.Equal(t, original.ContactID, result.ContactID)
	assert.Equal(t, models.QuotationStatusDraft, result.Status)
}
