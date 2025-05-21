package repository_test

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Função auxiliar para criar cotação de teste
func createTestQuotation(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.Quotation {
	// Cria repositório de quotation
	repo := repository.NewQuotationRepository(db, logger)

	// Cria um contexto para as chamadas
	ctx := context.Background()

	// Cria a cotação
	quotation := &models.Quotation{
		ContactID:     1,
		Status:        "",
		ExpiryDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:      1000.0,
		TaxTotal:      100.0,
		DiscountTotal: 50.0,
		GrandTotal:    1050.0,
		Notes:         "Cotação de teste via testes automatizados",
		Terms:         "Condições de pagamento: 30 dias",
	}

	// Adiciona parâmetro de contexto
	err := repo.CreateQuotation(ctx, quotation)
	assert.NoError(t, err)
	assert.NotZero(t, quotation.ID)
	assert.NotEmpty(t, quotation.QuotationNo)
	assert.Equal(t, models.QuotationStatusDraft, quotation.Status)

	return quotation
}

// Cria uma cotação com itens
func createTestQuotationWithItems(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.Quotation {
	// Cria uma cotação básica
	quotation := createTestQuotation(t, db, logger)

	// Adiciona itens à cotação
	repo := repository.NewQuotationRepository(db, logger)

	// Cria contexto para as operações
	ctx := context.Background()

	// Busca a cotação para ter o ID correto
	var err error
	quotation, err = repo.GetQuotationByID(ctx, quotation.ID)
	assert.NoError(t, err)

	// Criamos itens manualmente
	items := []models.QuotationItem{
		{
			QuotationID: quotation.ID,
			ProductID:   1, // Assume que existe um produto com ID 1
			ProductName: "Produto de Teste 1",
			ProductCode: "P001",
			Description: "Descrição do produto 1",
			Quantity:    2,
			UnitPrice:   100.0,
			Discount:    10.0,
			Tax:         18.0,
			Total:       208.0, // (2 * 100 - 10) * 1.18
		},
		{
			QuotationID: quotation.ID,
			ProductID:   2, // Assume que existe um produto com ID 2
			ProductName: "Produto de Teste 2",
			ProductCode: "P002",
			Description: "Descrição do produto 2",
			Quantity:    1,
			UnitPrice:   50.0,
			Discount:    0.0,
			Tax:         18.0,
			Total:       59.0, // (1 * 50) * 1.18
		},
	}

	// Adiciona os itens diretamente no banco
	for _, item := range items {
		err := db.Create(&item).Error
		assert.NoError(t, err)
	}

	// Atualiza o valor total da cotação
	quotation.SubTotal = 240.0   // (2*100) + (1*50) - 10
	quotation.TaxTotal = 43.2    // 240 * 0.18
	quotation.GrandTotal = 283.2 // 240 + 43.2

	// Adiciona contexto ao updateQuotation
	err = repo.UpdateQuotation(ctx, quotation.ID, quotation)
	assert.NoError(t, err)

	// Busca a cotação novamente para ter os itens carregados
	updatedQuotation, err := repo.GetQuotationByID(ctx, quotation.ID)
	assert.NoError(t, err)

	return updatedQuotation
}

// Função auxiliar para criar contato do tipo cliente
func createTestClient(t *testing.T, db *gorm.DB, logger *zap.Logger) *contact.Contact {
	t.Helper()
	cli := &contact.Contact{
		PersonType: "pf",
		Type:       "cliente",
		Name:       "Cliente Teste",
		Document:   "123.456.789-00",
		Email:      "cliente@teste.com",
		ZipCode:    "01001-000",
		CreatedAt:  time.Now(),
	}
	err := db.Create(cli).Error
	assert.NoError(t, err)
	assert.NotZero(t, cli.ID)
	return cli
}

// Função auxiliar para criar contato do tipo fornecedor
func createTestSupplier(t *testing.T, db *gorm.DB, logger *zap.Logger) *contact.Contact {
	t.Helper()
	fn := &contact.Contact{
		PersonType: "pj",
		Type:       "fornecedor",
		Name:       "Fornecedor Teste",
		Document:   "12.345.678/0001-99",
		Email:      "fornecedor@teste.com",
		ZipCode:    "20010-000",
		CreatedAt:  time.Now(),
	}
	err := db.Create(fn).Error
	assert.NoError(t, err)
	assert.NotZero(t, fn.ID)
	return fn
}
