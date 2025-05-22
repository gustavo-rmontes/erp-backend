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

// Função auxiliar para criar sales order de teste
func createTestSalesOrder(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.SalesOrder {
	// Cria repositório de sales order
	repo := repository.NewSalesOrderRepository(db, logger)

	// Cria um contexto para as chamadas
	ctx := context.Background()

	// Cria o sales order (sem QuotationID para evitar problemas de FK)
	salesOrder := &models.SalesOrder{
		ContactID: 1,
		// QuotationID omitido (será tratado como NULL)
		Status:          "",
		ExpectedDate:    time.Now().AddDate(0, 0, 30), // 30 dias
		SubTotal:        1000.0,
		TaxTotal:        180.0,
		DiscountTotal:   50.0,
		GrandTotal:      1130.0,
		Notes:           "Sales order de teste via testes automatizados",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua de Teste, 123 - Cidade Teste",
	}

	err := repo.CreateSalesOrder(ctx, salesOrder)
	assert.NoError(t, err)
	assert.NotZero(t, salesOrder.ID)
	assert.NotEmpty(t, salesOrder.SONo)
	assert.Equal(t, models.SOStatusDraft, salesOrder.Status)

	return salesOrder
}

// Função auxiliar para criar sales order com itens
func createTestSalesOrderWithItems(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.SalesOrder {
	// Limpa dados existentes para evitar conflitos
	err := db.Exec("DELETE FROM sales_order_items").Error
	assert.NoError(t, err)

	// Cria um sales order básico primeiro
	salesOrder := createTestSalesOrder(t, db, logger)

	// Adiciona itens ao sales order
	repo := repository.NewSalesOrderRepository(db, logger)
	ctx := context.Background()

	// Criamos itens manualmente sem definir IDs
	items := []models.SOItem{
		{
			SalesOrderID: salesOrder.ID,
			ProductID:    1,
			ProductName:  "Produto de Teste 1",
			ProductCode:  "P001",
			Description:  "Descrição do produto 1",
			Quantity:     2,
			UnitPrice:    100.0,
			Discount:     10.0,
			Tax:          18.0,
			Total:        208.0, // (2 * 100 - 10) * 1.18
		},
		{
			SalesOrderID: salesOrder.ID,
			ProductID:    2,
			ProductName:  "Produto de Teste 2",
			ProductCode:  "P002",
			Description:  "Descrição do produto 2",
			Quantity:     1,
			UnitPrice:    50.0,
			Discount:     0.0,
			Tax:          18.0,
			Total:        59.0, // (1 * 50) * 1.18
		},
	}

	// Adiciona os itens diretamente no banco (sem IDs definidos)
	for _, item := range items {
		err := db.Create(&item).Error
		assert.NoError(t, err)
	}

	// Atualiza o valor total do sales order
	salesOrder.SubTotal = 240.0   // (2*100) + (1*50) - 10
	salesOrder.TaxTotal = 43.2    // 240 * 0.18
	salesOrder.GrandTotal = 283.2 // 240 + 43.2

	err = repo.UpdateSalesOrder(ctx, salesOrder.ID, salesOrder)
	assert.NoError(t, err)

	// Busca o sales order novamente para ter os itens carregados
	updatedSalesOrder, err := repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.NoError(t, err)

	return updatedSalesOrder
}

// Função auxiliar para criar sales order a partir de uma quotation
func createTestSalesOrderFromQuotation(t *testing.T, db *gorm.DB, logger *zap.Logger, quotationID int) *models.SalesOrder {
	// Cria repositório de sales order
	repo := repository.NewSalesOrderRepository(db, logger)

	// Cria um contexto para as chamadas
	ctx := context.Background()

	// Cria o sales order baseado na quotation
	salesOrder := &models.SalesOrder{
		QuotationID:     quotationID,
		ContactID:       1,
		Status:          models.SOStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30),
		SubTotal:        1000.0,
		TaxTotal:        180.0,
		DiscountTotal:   50.0,
		GrandTotal:      1130.0,
		Notes:           "Sales order criado a partir de quotation",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua de Entrega, 456 - Cidade Entrega",
	}

	err := repo.CreateSalesOrder(ctx, salesOrder)
	assert.NoError(t, err)
	assert.NotZero(t, salesOrder.ID)
	assert.NotEmpty(t, salesOrder.SONo)
	assert.Equal(t, quotationID, salesOrder.QuotationID)

	return salesOrder
}

// Função auxiliar para criar múltiplos sales orders para teste de paginação
func createMultipleSalesOrders(t *testing.T, db *gorm.DB, logger *zap.Logger, count int) []*models.SalesOrder {
	var salesOrders []*models.SalesOrder

	for i := 0; i < count; i++ {
		salesOrder := createTestSalesOrder(t, db, logger)

		// Varia alguns campos para tornar os dados mais realistas
		salesOrder.ContactID = (i % 3) + 1                      // Varia entre contatos 1, 2, 3
		salesOrder.ExpectedDate = time.Now().AddDate(0, 0, i*7) // Varia datas de entrega
		salesOrder.GrandTotal = 1000.0 + float64(i*100)         // Varia valores

		if i%2 == 0 {
			salesOrder.Status = models.SOStatusConfirmed
		} else {
			salesOrder.Status = models.SOStatusDraft
		}

		// Atualiza no banco
		repo := repository.NewSalesOrderRepository(db, logger)
		ctx := context.Background()
		err := repo.UpdateSalesOrder(ctx, salesOrder.ID, salesOrder)
		assert.NoError(t, err)

		salesOrders = append(salesOrders, salesOrder)
	}

	return salesOrders
}

// Função auxiliar para limpar sales orders de teste
func cleanupSalesOrders(t *testing.T, db *gorm.DB, logger *zap.Logger, salesOrders []*models.SalesOrder) {
	repo := repository.NewSalesOrderRepository(db, logger)
	ctx := context.Background()

	for _, salesOrder := range salesOrders {
		err := repo.DeleteSalesOrder(ctx, salesOrder.ID)
		if err != nil {
			t.Logf("Aviso: Não foi possível deletar sales order %d: %v", salesOrder.ID, err)
		}
	}
}

// Função auxiliar para criar purchase order de teste
func createTestPurchaseOrder(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.PurchaseOrder {
	// Cria repositório de purchase order
	repo := repository.NewPurchaseOrderRepository(db, logger)

	// Cria um contexto para as chamadas
	ctx := context.Background()

	// Cria o purchase order (sem SalesOrderID para evitar problemas de FK)
	purchaseOrder := &models.PurchaseOrder{
		ContactID: 1,
		// SalesOrderID omitido (será tratado como NULL)
		Status:          "",
		ExpectedDate:    time.Now().AddDate(0, 0, 30), // 30 dias
		SubTotal:        2000.0,
		TaxTotal:        360.0,
		DiscountTotal:   100.0,
		GrandTotal:      2260.0,
		Notes:           "Purchase order de teste via testes automatizados",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua de Fornecedor, 456 - Cidade Fornecedor",
	}

	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.NoError(t, err)
	assert.NotZero(t, purchaseOrder.ID)
	assert.NotEmpty(t, purchaseOrder.PONo)
	assert.Equal(t, models.POStatusDraft, purchaseOrder.Status)

	return purchaseOrder
}

// Função auxiliar para criar purchase order com itens
func createTestPurchaseOrderWithItems(t *testing.T, db *gorm.DB, logger *zap.Logger) *models.PurchaseOrder {
	// Limpa dados existentes para evitar conflitos
	err := db.Exec("DELETE FROM purchase_order_items").Error
	assert.NoError(t, err)

	// Cria um purchase order básico primeiro
	purchaseOrder := createTestPurchaseOrder(t, db, logger)

	// Adiciona itens ao purchase order
	repo := repository.NewPurchaseOrderRepository(db, logger)
	ctx := context.Background()

	// Criamos itens manualmente sem definir IDs
	items := []models.POItem{
		{
			PurchaseOrderID: purchaseOrder.ID,
			ProductID:       1,
			ProductName:     "Produto de Compra 1",
			ProductCode:     "PC001",
			Description:     "Descrição do produto para compra 1",
			Quantity:        5,
			UnitPrice:       200.0,
			Discount:        20.0,
			Tax:             18.0,
			Total:           1144.0, // (5 * 200 - 20) * 1.18
		},
		{
			PurchaseOrderID: purchaseOrder.ID,
			ProductID:       2,
			ProductName:     "Produto de Compra 2",
			ProductCode:     "PC002",
			Description:     "Descrição do produto para compra 2",
			Quantity:        3,
			UnitPrice:       150.0,
			Discount:        0.0,
			Tax:             18.0,
			Total:           531.0, // (3 * 150) * 1.18
		},
	}

	// Adiciona os itens diretamente no banco (sem IDs definidos)
	for _, item := range items {
		err := db.Create(&item).Error
		assert.NoError(t, err)
	}

	// Atualiza o valor total do purchase order
	purchaseOrder.SubTotal = 1480.0   // (5*200) + (3*150) - 20
	purchaseOrder.TaxTotal = 266.4    // 1480 * 0.18
	purchaseOrder.GrandTotal = 1746.4 // 1480 + 266.4

	err = repo.UpdatePurchaseOrder(ctx, purchaseOrder.ID, purchaseOrder)
	assert.NoError(t, err)

	// Busca o purchase order novamente para ter os itens carregados
	updatedPurchaseOrder, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)

	return updatedPurchaseOrder
}

// Função auxiliar para criar purchase order a partir de um sales order
func createTestPurchaseOrderFromSalesOrder(t *testing.T, db *gorm.DB, logger *zap.Logger, salesOrderID int) *models.PurchaseOrder {
	// Cria repositório de purchase order
	repo := repository.NewPurchaseOrderRepository(db, logger)

	// Cria um contexto para as chamadas
	ctx := context.Background()

	// Busca o sales order para pegar alguns dados
	salesRepo := repository.NewSalesOrderRepository(db, logger)
	salesOrder, err := salesRepo.GetSalesOrderByID(ctx, salesOrderID)
	assert.NoError(t, err)

	// Cria o purchase order baseado no sales order
	purchaseOrder := &models.PurchaseOrder{
		SalesOrderID:    salesOrderID,
		SONo:            salesOrder.SONo,
		ContactID:       salesOrder.ContactID, // Pode ser diferente em casos reais
		Status:          models.POStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 45), // 45 dias para compra
		SubTotal:        salesOrder.SubTotal,
		TaxTotal:        salesOrder.TaxTotal,
		DiscountTotal:   salesOrder.DiscountTotal,
		GrandTotal:      salesOrder.GrandTotal,
		Notes:           "Purchase order criado a partir de sales order " + salesOrder.SONo,
		PaymentTerms:    "45 dias",
		ShippingAddress: "Endereço do Fornecedor, 789 - Cidade Fornecedor",
	}

	err = repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.NoError(t, err)
	assert.NotZero(t, purchaseOrder.ID)
	assert.NotEmpty(t, purchaseOrder.PONo)
	assert.Equal(t, salesOrderID, purchaseOrder.SalesOrderID)

	return purchaseOrder
}

// Função auxiliar para criar múltiplos purchase orders para teste de paginação
func createMultiplePurchaseOrders(t *testing.T, db *gorm.DB, logger *zap.Logger, count int) []*models.PurchaseOrder {
	var purchaseOrders []*models.PurchaseOrder

	for i := 0; i < count; i++ {
		purchaseOrder := createTestPurchaseOrder(t, db, logger)

		// Varia alguns campos para tornar os dados mais realistas
		purchaseOrder.ContactID = (i % 3) + 1                       // Varia entre contatos 1, 2, 3
		purchaseOrder.ExpectedDate = time.Now().AddDate(0, 0, i*10) // Varia datas de entrega
		purchaseOrder.GrandTotal = 2000.0 + float64(i*200)          // Varia valores

		if i%3 == 0 {
			purchaseOrder.Status = models.POStatusConfirmed
		} else if i%3 == 1 {
			purchaseOrder.Status = models.POStatusSent
		} else {
			purchaseOrder.Status = models.POStatusDraft
		}

		// Atualiza no banco
		repo := repository.NewPurchaseOrderRepository(db, logger)
		ctx := context.Background()
		err := repo.UpdatePurchaseOrder(ctx, purchaseOrder.ID, purchaseOrder)
		assert.NoError(t, err)

		purchaseOrders = append(purchaseOrders, purchaseOrder)
	}

	return purchaseOrders
}

// Função auxiliar para limpar purchase orders de teste
func cleanupPurchaseOrders(t *testing.T, db *gorm.DB, logger *zap.Logger, purchaseOrders []*models.PurchaseOrder) {
	repo := repository.NewPurchaseOrderRepository(db, logger)
	ctx := context.Background()

	for _, purchaseOrder := range purchaseOrders {
		err := repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)
		if err != nil {
			t.Logf("Aviso: Não foi possível deletar purchase order %d: %v", purchaseOrder.ID, err)
		}
	}
}
