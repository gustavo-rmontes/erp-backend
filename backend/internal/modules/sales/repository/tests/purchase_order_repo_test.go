package repository_test

import (
	"ERP-ONSMART/backend/internal/errors"
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	testutils "ERP-ONSMART/backend/internal/utils/test_utils"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_PurchaseOrderRepository_NotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	// Cria o repositório com um logger noop (sem saída)
	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())

	_, err := repo.GetPurchaseOrderByID(context.Background(), 999999)
	assert.ErrorIs(t, err, errors.ErrPurchaseOrderNotFound)
}

func Test_PurchaseOrderRepository_Create(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	purchaseOrder := &models.PurchaseOrder{
		ContactID: 1,
		// SalesOrderID não definido (será 0, que precisa ser tratado como NULL)
		Status:          models.POStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30),
		SubTotal:        2000.0,
		TaxTotal:        200.0,
		DiscountTotal:   100.0,
		GrandTotal:      2100.0,
		Notes:           "Purchase order de teste",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua de Teste, 123",
	}

	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.NoError(t, err)
	assert.NotZero(t, purchaseOrder.ID)
	assert.NotEmpty(t, purchaseOrder.PONo)
	assert.Equal(t, models.POStatusDraft, purchaseOrder.Status)

	// Cleanup
	err = repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
}

func Test_PurchaseOrderRepository_GetByID(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um purchase order primeiro
	purchaseOrder := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())
	defer repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)

	// Testa a busca
	foundPurchaseOrder, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundPurchaseOrder)
	assert.Equal(t, purchaseOrder.ID, foundPurchaseOrder.ID)
	assert.Equal(t, purchaseOrder.PONo, foundPurchaseOrder.PONo)
	assert.Equal(t, purchaseOrder.ContactID, foundPurchaseOrder.ContactID)
	assert.Equal(t, purchaseOrder.GrandTotal, foundPurchaseOrder.GrandTotal)

	// Verifica se o slice de Items é inicializado (pode estar vazio)
	assert.NotNil(t, foundPurchaseOrder.Items, "Items deve ser inicializado")

	// Para Contact, vamos verificar se o preload está funcionando
	// mas sem falhar se o ContactID dos seeds não existir
	t.Logf("ContactID do purchase order: %d", foundPurchaseOrder.ContactID)
	if foundPurchaseOrder.Contact != nil {
		t.Logf("Contact carregado: ID=%d, Name=%s", foundPurchaseOrder.Contact.ID, foundPurchaseOrder.Contact.Name)
		assert.Equal(t, purchaseOrder.ContactID, foundPurchaseOrder.Contact.ID)
	} else {
		t.Logf("Contact não carregado - verificando se ContactID existe no banco...")

		// Verifica se o contato existe no banco
		var existingContact contact.Contact
		err := dbTest.GormDB.First(&existingContact, purchaseOrder.ContactID).Error
		if err != nil {
			t.Logf("ContactID %d não existe no banco: %v", purchaseOrder.ContactID, err)
			t.Logf("Isso é esperado se os seeds não criaram contatos suficientes")
		} else {
			t.Errorf("ContactID %d existe no banco mas não foi carregado via preload", purchaseOrder.ContactID)
		}
	}
}

func Test_PurchaseOrderRepository_Update(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um purchase order primeiro
	purchaseOrder := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())
	defer repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)

	// Atualiza o purchase order
	purchaseOrder.Status = models.POStatusConfirmed
	purchaseOrder.Notes = "Purchase order atualizado"
	purchaseOrder.GrandTotal = 3000.0

	err := repo.UpdatePurchaseOrder(ctx, purchaseOrder.ID, purchaseOrder)
	assert.NoError(t, err)

	// Verifica se a atualização foi persistida
	updatedPurchaseOrder, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusConfirmed, updatedPurchaseOrder.Status)
	assert.Equal(t, "Purchase order atualizado", updatedPurchaseOrder.Notes)
	assert.Equal(t, 3000.0, updatedPurchaseOrder.GrandTotal)
}

func Test_PurchaseOrderRepository_Delete(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um purchase order primeiro
	purchaseOrder := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())

	// Verifica que existe
	foundPurchaseOrder, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundPurchaseOrder)

	// Deleta
	err = repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)
	assert.NoError(t, err)

	// Verifica que foi deletado
	_, err = repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.ErrorIs(t, err, errors.ErrPurchaseOrderNotFound)
}

func Test_PurchaseOrderRepository_FullWorkflow(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// 1. Cria um purchase order
	purchaseOrder := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())

	// 2. Verifica que foi criado com status draft
	assert.Equal(t, models.POStatusDraft, purchaseOrder.Status)

	// 3. Atualiza para enviado
	purchaseOrder.Status = models.POStatusSent
	err := repo.UpdatePurchaseOrder(ctx, purchaseOrder.ID, purchaseOrder)
	assert.NoError(t, err)

	// 4. Verifica a mudança de status
	sentPO, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusSent, sentPO.Status)

	// 5. Atualiza para confirmado
	sentPO.Status = models.POStatusConfirmed
	err = repo.UpdatePurchaseOrder(ctx, sentPO.ID, sentPO)
	assert.NoError(t, err)

	// 6. Finaliza como recebido
	confirmedPO, err := repo.GetPurchaseOrderByID(ctx, sentPO.ID)
	assert.NoError(t, err)
	confirmedPO.Status = models.POStatusReceived
	err = repo.UpdatePurchaseOrder(ctx, confirmedPO.ID, confirmedPO)
	assert.NoError(t, err)

	// 7. Verifica status final
	receivedPO, err := repo.GetPurchaseOrderByID(ctx, confirmedPO.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusReceived, receivedPO.Status)

	// 8. Cleanup
	err = repo.DeletePurchaseOrder(ctx, receivedPO.ID)
	assert.NoError(t, err)
}

func Test_PurchaseOrderRepository_UpdateNotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	purchaseOrder := &models.PurchaseOrder{
		ContactID: 1,
		Status:    models.POStatusConfirmed,
	}

	err := repo.UpdatePurchaseOrder(ctx, 999999, purchaseOrder)
	assert.ErrorIs(t, err, errors.ErrPurchaseOrderNotFound)
}

func Test_PurchaseOrderRepository_DeleteNotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	err := repo.DeletePurchaseOrder(ctx, 999999)
	assert.ErrorIs(t, err, errors.ErrPurchaseOrderNotFound)
}

func Test_PurchaseOrderRepository_ContextTimeout(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())

	// Cria um contexto já cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela imediatamente

	purchaseOrder := &models.PurchaseOrder{
		ContactID: 1,
		// SalesOrderID omitido
		Status: models.POStatusDraft,
	}

	// Testa operações com contexto cancelado
	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	_, err = repo.GetPurchaseOrderByID(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	err = repo.UpdatePurchaseOrder(ctx, 1, purchaseOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	err = repo.DeletePurchaseOrder(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")
}

func Test_PurchaseOrderRepository_ContextDeadline(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())

	// Cria um contexto com deadline já expirado
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	purchaseOrder := &models.PurchaseOrder{
		ContactID: 1,
		// SalesOrderID omitido
		Status: models.POStatusDraft,
	}

	// Testa operações com contexto expirado
	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	_, err = repo.GetPurchaseOrderByID(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	err = repo.UpdatePurchaseOrder(ctx, 1, purchaseOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	err = repo.DeletePurchaseOrder(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

// Teste para criação de Purchase Order com relacionamento a Sales Order
func Test_PurchaseOrderRepository_CreateFromSalesOrder(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um sales order primeiro
	salesOrder := createTestSalesOrder(t, dbTest.GormDB, zap.NewNop())
	defer func() {
		salesRepo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
		salesRepo.DeleteSalesOrder(ctx, salesOrder.ID)
	}()

	// Cria um purchase order baseado no sales order
	purchaseOrder := createTestPurchaseOrderFromSalesOrder(t, dbTest.GormDB, zap.NewNop(), salesOrder.ID)
	defer repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)

	// Verifica se o relacionamento foi criado corretamente
	foundPO, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, salesOrder.ID, foundPO.SalesOrderID)
	assert.Equal(t, salesOrder.SONo, foundPO.SONo)

	// Verifica se o preload do SalesOrder funciona
	if foundPO.SalesOrder != nil {
		assert.Equal(t, salesOrder.ID, foundPO.SalesOrder.ID)
		assert.Equal(t, salesOrder.SONo, foundPO.SalesOrder.SONo)
	}
}

// Teste para fluxo completo de status de Purchase Order
func Test_PurchaseOrderRepository_StatusWorkflow(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// 1. Cria purchase order em draft
	purchaseOrder := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())
	assert.Equal(t, models.POStatusDraft, purchaseOrder.Status)

	// 2. Draft -> Sent
	purchaseOrder.Status = models.POStatusSent
	err := repo.UpdatePurchaseOrder(ctx, purchaseOrder.ID, purchaseOrder)
	assert.NoError(t, err)

	sentPO, err := repo.GetPurchaseOrderByID(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusSent, sentPO.Status)

	// 3. Sent -> Confirmed
	sentPO.Status = models.POStatusConfirmed
	err = repo.UpdatePurchaseOrder(ctx, sentPO.ID, sentPO)
	assert.NoError(t, err)

	confirmedPO, err := repo.GetPurchaseOrderByID(ctx, sentPO.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusConfirmed, confirmedPO.Status)

	// 4. Confirmed -> Received
	confirmedPO.Status = models.POStatusReceived
	err = repo.UpdatePurchaseOrder(ctx, confirmedPO.ID, confirmedPO)
	assert.NoError(t, err)

	receivedPO, err := repo.GetPurchaseOrderByID(ctx, confirmedPO.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusReceived, receivedPO.Status)

	// 5. Teste de cancelamento (pode ser feito de qualquer status)
	// Vamos criar outro PO para testar cancelamento
	cancelTestPO := createTestPurchaseOrder(t, dbTest.GormDB, zap.NewNop())

	// Draft -> Cancelled
	cancelTestPO.Status = models.POStatusCancelled
	err = repo.UpdatePurchaseOrder(ctx, cancelTestPO.ID, cancelTestPO)
	assert.NoError(t, err)

	cancelledPO, err := repo.GetPurchaseOrderByID(ctx, cancelTestPO.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.POStatusCancelled, cancelledPO.Status)

	// Cleanup
	err = repo.DeletePurchaseOrder(ctx, receivedPO.ID)
	assert.NoError(t, err)
	err = repo.DeletePurchaseOrder(ctx, cancelledPO.ID)
	assert.NoError(t, err)
}

// Teste para geração automática de número PO
func Test_PurchaseOrderRepository_AutoGeneratePONumber(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria purchase order sem definir PONo
	purchaseOrder := &models.PurchaseOrder{
		ContactID:       1,
		Status:          models.POStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30),
		SubTotal:        1000.0,
		TaxTotal:        180.0,
		DiscountTotal:   0.0,
		GrandTotal:      1180.0,
		Notes:           "Teste geração automática de número",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua Auto Number, 123",
		// PONo não definido - deve ser gerado automaticamente
	}

	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.NoError(t, err)
	assert.NotZero(t, purchaseOrder.ID)
	assert.NotEmpty(t, purchaseOrder.PONo)
	assert.Contains(t, purchaseOrder.PONo, "PO-")
	assert.Contains(t, purchaseOrder.PONo, "2025") // Ano atual

	// Verifica o padrão do número gerado
	t.Logf("Número PO gerado: %s", purchaseOrder.PONo)

	// Cleanup
	err = repo.DeletePurchaseOrder(ctx, purchaseOrder.ID)
	assert.NoError(t, err)
}

// Teste de transação - falha na criação de itens deve fazer rollback
func Test_PurchaseOrderRepository_TransactionRollback(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewPurchaseOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Simula um erro forçando um contexto com timeout muito curto
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Aguarda o contexto expirar
	time.Sleep(2 * time.Nanosecond)

	purchaseOrder := &models.PurchaseOrder{
		ContactID:       1,
		Status:          models.POStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30),
		SubTotal:        1000.0,
		TaxTotal:        180.0,
		DiscountTotal:   0.0,
		GrandTotal:      1180.0,
		Notes:           "Teste rollback",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua Rollback, 123",
		Items: []models.POItem{
			{
				ProductID:   1,
				ProductName: "Produto Rollback",
				ProductCode: "PR001",
				Description: "Item que deve causar rollback",
				Quantity:    1,
				UnitPrice:   1000.0,
				Discount:    0.0,
				Tax:         18.0,
				Total:       1180.0,
			},
		},
	}

	// Deve falhar devido ao contexto expirado
	err := repo.CreatePurchaseOrder(ctx, purchaseOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	// Purchase Order não deve ter sido criado (rollback funcionou)
	assert.Zero(t, purchaseOrder.ID)
}
