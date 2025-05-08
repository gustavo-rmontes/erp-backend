// stats_test.go
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

// TestGetStats testa a obtenção de estatísticas de cotações
func TestGetStats(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, _, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	// Definir o intervalo de datas para o teste
	startDate := time.Now().AddDate(0, -1, 0) // 1 mês atrás
	endDate := time.Now().AddDate(0, 1, 0)    // 1 mês à frente

	// Criar cotações para estatísticas COM DATAS DENTRO DO INTERVALO
	// Essa é a parte crítica - as datas devem estar entre startDate e endDate
	quotation1 := createTestQuotation(1)
	quotation1.Status = models.QuotationStatusDraft
	quotation1.GrandTotal = 1000.00
	quotation1.CreatedAt = time.Now() // Atual (dentro do intervalo)

	quotation2 := createTestQuotation(2)
	quotation2.Status = models.QuotationStatusSent
	quotation2.GrandTotal = 2000.00
	quotation2.CreatedAt = time.Now().AddDate(0, 0, -5) // 5 dias atrás (dentro do intervalo)

	quotation3 := createTestQuotation(3)
	quotation3.Status = models.QuotationStatusAccepted
	quotation3.GrandTotal = 3000.00
	quotation3.CreatedAt = time.Now().AddDate(0, 0, -10) // 10 dias atrás (dentro do intervalo)

	// Configurar mock para retornar as cotações (observando o tipo de parâmetro correto)
	mockQuotationRepo.EXPECT().
		GetAllQuotations(gomock.Any()).
		Return(&pagination.PaginatedResult{
			Items:       []models.Quotation{*quotation1, *quotation2, *quotation3},
			TotalItems:  3,
			CurrentPage: 1,
			PageSize:    10,
			TotalPages:  1,
		}, nil)

	// Criar filtro de data
	dateRange := &dto.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Act
	stats, err := service.GetStats(ctx, dateRange)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalCount)
	assert.Equal(t, 6000.00, stats.TotalValue)
	assert.Equal(t, 2000.00, stats.AverageValue)

	// Verificar contagem por status
	assert.Equal(t, 1, stats.CountByStatus[models.QuotationStatusDraft])
	assert.Equal(t, 1, stats.CountByStatus[models.QuotationStatusSent])
	assert.Equal(t, 1, stats.CountByStatus[models.QuotationStatusAccepted])

	// Verificar valores por status
	assert.Equal(t, 1000.00, stats.TotalValueByStatus[models.QuotationStatusDraft])
	assert.Equal(t, 2000.00, stats.TotalValueByStatus[models.QuotationStatusSent])
	assert.Equal(t, 3000.00, stats.TotalValueByStatus[models.QuotationStatusAccepted])
}

// TestGetConversionStats testa as estatísticas de conversão
func TestGetConversionStats(t *testing.T) {
	// Arrange
	ctx, service, mockQuotationRepo, mockSalesOrderRepo, _, ctrl := setupTest(t)
	defer ctrl.Finish()

	// Definir o intervalo de datas para o teste
	startDate := time.Now().AddDate(0, -1, 0) // 1 mês atrás
	endDate := time.Now().AddDate(0, 1, 0)    // 1 mês à frente

	// Criar cotações para estatísticas COM DATAS DENTRO DO INTERVALO
	quotation1 := createTestQuotation(1)
	quotation1.Status = models.QuotationStatusSent
	quotation1.GrandTotal = 1000.00
	quotation1.CreatedAt = time.Now().AddDate(0, 0, -15) // 15 dias atrás (dentro do intervalo)

	quotation2 := createTestQuotation(2)
	quotation2.Status = models.QuotationStatusAccepted
	quotation2.GrandTotal = 2000.00
	quotation2.CreatedAt = time.Now().AddDate(0, 0, -10) // 10 dias atrás (dentro do intervalo)

	quotation3 := createTestQuotation(3)
	quotation3.Status = models.QuotationStatusAccepted
	quotation3.GrandTotal = 3000.00
	quotation3.CreatedAt = time.Now().AddDate(0, 0, -5) // 5 dias atrás (dentro do intervalo)

	// Criar pedidos de venda para as cotações
	salesOrder2 := &models.SalesOrder{
		ID:          1,
		SONo:        "SO-2025-0001",
		QuotationID: 2,
		ContactID:   quotation2.ContactID,
		Status:      models.SOStatusDraft,
		GrandTotal:  2000.00,
		CreatedAt:   quotation2.CreatedAt.Add(24 * time.Hour), // 1 dia depois
	}

	salesOrder3 := &models.SalesOrder{
		ID:          2,
		SONo:        "SO-2025-0002",
		QuotationID: 3,
		ContactID:   quotation3.ContactID,
		Status:      models.SOStatusDraft,
		GrandTotal:  3000.00,
		CreatedAt:   quotation3.CreatedAt.Add(48 * time.Hour), // 2 dias depois
	}

	// Configurar mock para retornar as cotações
	mockQuotationRepo.EXPECT().
		GetAllQuotations(gomock.Any()).
		Return(&pagination.PaginatedResult{
			Items:       []models.Quotation{*quotation1, *quotation2, *quotation3},
			TotalItems:  3,
			CurrentPage: 1,
			PageSize:    10,
			TotalPages:  1,
		}, nil)

	// Verifique se o método está correto: único objeto ou slice
	// Investigando o código em stats.go, parece que espera um único objeto
	mockSalesOrderRepo.EXPECT().
		GetSalesOrdersByQuotation(2).
		Return(salesOrder2, nil)

	mockSalesOrderRepo.EXPECT().
		GetSalesOrdersByQuotation(3).
		Return(salesOrder3, nil)

	// Criar filtro de data
	dateRange := &dto.DateRange{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Act
	stats, err := service.GetConversionStats(ctx, dateRange)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalQuotations)
	assert.Equal(t, 2, stats.ConvertedQuotations)
	assert.Equal(t, 66.66666666666666, stats.ConversionRate)      // 2/3 * 100 = 66.67%
	assert.Equal(t, 1, stats.AverageTimeToConvert)                // Média de 1.5 dias (arredondada para 1)
	assert.Equal(t, 83.33333333333334, stats.ValueConversionRate) // (2000+3000)/(1000+2000+3000) * 100 = 83.33%
}
