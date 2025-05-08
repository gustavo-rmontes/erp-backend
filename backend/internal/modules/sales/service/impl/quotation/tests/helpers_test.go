// test_helpers.go
package tests

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/mocks"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/service/impl/quotation"
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"
)

// setupTest configura o ambiente de teste com mocks
func setupTest(t *testing.T) (
	context.Context,
	*quotation.Service,
	*mocks.MockQuotationRepository,
	*mocks.MockSalesOrderRepository,
	*mocks.MockSalesProcessRepository,
	*gomock.Controller,
) {
	ctrl := gomock.NewController(t)

	// Criar mocks para os repositórios
	mockQuotationRepo := mocks.NewMockQuotationRepository(ctrl)
	mockSalesOrderRepo := mocks.NewMockSalesOrderRepository(ctrl)
	mockSalesProcessRepo := mocks.NewMockSalesProcessRepository(ctrl)

	// Criar logger de teste
	logger, _ := zap.NewDevelopment()

	// Criar o serviço
	serviceInstance := quotation.New(
		mockQuotationRepo,
		mockSalesOrderRepo,
		mockSalesProcessRepo,
		logger,
	).(*quotation.Service)

	ctx := context.Background()

	return ctx, serviceInstance, mockQuotationRepo, mockSalesOrderRepo, mockSalesProcessRepo, ctrl
}

// createTestQuotation cria uma cotação para testes
func createTestQuotation(id int) *models.Quotation {
	return &models.Quotation{
		ID:            id,
		QuotationNo:   "TEST-QT-2025-" + fmt.Sprintf("%04d", id),
		ContactID:     1,
		Status:        models.QuotationStatusDraft,
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:      1000.00,
		TaxTotal:      150.00,
		DiscountTotal: 50.00,
		GrandTotal:    1100.00,
		Notes:         "Notas de teste",
		Terms:         "Termos de teste",
		Items: []models.QuotationItem{
			{
				ID:          1,
				QuotationID: id,
				ProductID:   101,
				ProductName: "Produto de Teste",
				ProductCode: "PROD-101",
				Description: "Descrição do produto de teste",
				Quantity:    10,
				UnitPrice:   100.00,
				Discount:    5.00,
				Tax:         15.00,
				Total:       1100.00,
			},
		},
		Contact: &contact.Contact{
			ID:    1,
			Name:  "Cliente de Teste",
			Email: "cliente@teste.com",
			Phone: "11999999999",
		},
	}
}

// createTestContact cria um contato para testes
func createTestContact(id int) *contact.Contact {
	// Certifique-se de que está usando a estrutura Contact correta
	// Se estiver em outro pacote, importe adequadamente
	return &contact.Contact{
		ID:    id,
		Name:  "Cliente de Teste " + strconv.Itoa(id),
		Email: "cliente" + strconv.Itoa(id) + "@teste.com",
		Phone: "11999999" + fmt.Sprintf("%03d", id),
	}
}
