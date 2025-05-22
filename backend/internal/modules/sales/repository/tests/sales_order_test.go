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

func Test_SalesOrderRepository_NotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	// Cria o repositório com um logger noop (sem saída)
	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())

	_, err := repo.GetSalesOrderByID(context.Background(), 999999)
	assert.ErrorIs(t, err, errors.ErrSalesOrderNotFound)
}

func Test_SalesOrderRepository_Create(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	salesOrder := &models.SalesOrder{
		ContactID: 1,
		// QuotationID não definido (será 0, que precisa ser tratado como NULL)
		Status:          models.SOStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30),
		SubTotal:        1000.0,
		TaxTotal:        100.0,
		DiscountTotal:   50.0,
		GrandTotal:      1050.0,
		Notes:           "Sales order de teste",
		PaymentTerms:    "30 dias",
		ShippingAddress: "Rua de Teste, 123",
	}

	err := repo.CreateSalesOrder(ctx, salesOrder)
	assert.NoError(t, err)
	assert.NotZero(t, salesOrder.ID)
	assert.NotEmpty(t, salesOrder.SONo)
	assert.Equal(t, models.SOStatusDraft, salesOrder.Status)

	// Cleanup
	err = repo.DeleteSalesOrder(ctx, salesOrder.ID)
	assert.NoError(t, err)
}

func Test_SalesOrderRepository_GetByID(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um sales order primeiro
	salesOrder := createTestSalesOrder(t, dbTest.GormDB, zap.NewNop())
	defer repo.DeleteSalesOrder(ctx, salesOrder.ID)

	// Testa a busca
	foundSalesOrder, err := repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundSalesOrder)
	assert.Equal(t, salesOrder.ID, foundSalesOrder.ID)
	assert.Equal(t, salesOrder.SONo, foundSalesOrder.SONo)
	assert.Equal(t, salesOrder.ContactID, foundSalesOrder.ContactID)
	assert.Equal(t, salesOrder.GrandTotal, foundSalesOrder.GrandTotal)

	// Verifica se o slice de Items é inicializado (pode estar vazio)
	assert.NotNil(t, foundSalesOrder.Items, "Items deve ser inicializado")

	// Para Contact, vamos verificar se o preload está funcionando
	// mas sem falhar se o ContactID dos seeds não existir
	t.Logf("ContactID do sales order: %d", foundSalesOrder.ContactID)
	if foundSalesOrder.Contact != nil {
		t.Logf("Contact carregado: ID=%d, Name=%s", foundSalesOrder.Contact.ID, foundSalesOrder.Contact.Name)
		assert.Equal(t, salesOrder.ContactID, foundSalesOrder.Contact.ID)
	} else {
		t.Logf("Contact não carregado - verificando se ContactID existe no banco...")

		// Verifica se o contato existe no banco
		var existingContact contact.Contact
		err := dbTest.GormDB.First(&existingContact, salesOrder.ContactID).Error
		if err != nil {
			t.Logf("ContactID %d não existe no banco: %v", salesOrder.ContactID, err)
			t.Logf("Isso é esperado se os seeds não criaram contatos suficientes")
		} else {
			t.Errorf("ContactID %d existe no banco mas não foi carregado via preload", salesOrder.ContactID)
		}
	}
}

func Test_SalesOrderRepository_Update(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um sales order primeiro
	salesOrder := createTestSalesOrder(t, dbTest.GormDB, zap.NewNop())
	defer repo.DeleteSalesOrder(ctx, salesOrder.ID)

	// Atualiza o sales order
	salesOrder.Status = models.SOStatusConfirmed
	salesOrder.Notes = "Sales order atualizado"
	salesOrder.GrandTotal = 2000.0

	err := repo.UpdateSalesOrder(ctx, salesOrder.ID, salesOrder)
	assert.NoError(t, err)

	// Verifica se a atualização foi persistida
	updatedSalesOrder, err := repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.SOStatusConfirmed, updatedSalesOrder.Status)
	assert.Equal(t, "Sales order atualizado", updatedSalesOrder.Notes)
	assert.Equal(t, 2000.0, updatedSalesOrder.GrandTotal)
}

func Test_SalesOrderRepository_Delete(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// Cria um sales order primeiro
	salesOrder := createTestSalesOrder(t, dbTest.GormDB, zap.NewNop())

	// Verifica que existe
	foundSalesOrder, err := repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.NoError(t, err)
	assert.NotNil(t, foundSalesOrder)

	// Deleta
	err = repo.DeleteSalesOrder(ctx, salesOrder.ID)
	assert.NoError(t, err)

	// Verifica que foi deletado
	_, err = repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.ErrorIs(t, err, errors.ErrSalesOrderNotFound)
}

func Test_SalesOrderRepository_FullWorkflow(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	// 1. Cria um sales order
	salesOrder := createTestSalesOrder(t, dbTest.GormDB, zap.NewNop())

	// 2. Verifica que foi criado com status draft
	assert.Equal(t, models.SOStatusDraft, salesOrder.Status)

	// 3. Atualiza para confirmado
	salesOrder.Status = models.SOStatusConfirmed
	err := repo.UpdateSalesOrder(ctx, salesOrder.ID, salesOrder)
	assert.NoError(t, err)

	// 4. Verifica a mudança de status
	confirmedSO, err := repo.GetSalesOrderByID(ctx, salesOrder.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.SOStatusConfirmed, confirmedSO.Status)

	// 5. Atualiza para processando
	confirmedSO.Status = models.SOStatusProcessing
	err = repo.UpdateSalesOrder(ctx, confirmedSO.ID, confirmedSO)
	assert.NoError(t, err)

	// 6. Finaliza como completed
	processingSO, err := repo.GetSalesOrderByID(ctx, confirmedSO.ID)
	assert.NoError(t, err)
	processingSO.Status = models.SOStatusCompleted
	err = repo.UpdateSalesOrder(ctx, processingSO.ID, processingSO)
	assert.NoError(t, err)

	// 7. Verifica status final
	completedSO, err := repo.GetSalesOrderByID(ctx, processingSO.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.SOStatusCompleted, completedSO.Status)

	// 8. Cleanup
	err = repo.DeleteSalesOrder(ctx, completedSO.ID)
	assert.NoError(t, err)
}

func Test_SalesOrderRepository_UpdateNotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	salesOrder := &models.SalesOrder{
		ContactID: 1,
		Status:    models.SOStatusConfirmed,
	}

	err := repo.UpdateSalesOrder(ctx, 999999, salesOrder)
	assert.ErrorIs(t, err, errors.ErrSalesOrderNotFound)
}

func Test_SalesOrderRepository_DeleteNotFound(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())
	ctx := context.Background()

	err := repo.DeleteSalesOrder(ctx, 999999)
	assert.ErrorIs(t, err, errors.ErrSalesOrderNotFound)
}

func Test_SalesOrderRepository_ContextTimeout(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())

	// Cria um contexto já cancelado
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancela imediatamente

	salesOrder := &models.SalesOrder{
		ContactID: 1,
		// QuotationID omitido
		Status: models.SOStatusDraft,
	}

	// Testa operações com contexto cancelado
	err := repo.CreateSalesOrder(ctx, salesOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	_, err = repo.GetSalesOrderByID(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	err = repo.UpdateSalesOrder(ctx, 1, salesOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")

	err = repo.DeleteSalesOrder(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelada")
}

func Test_SalesOrderRepository_ContextDeadline(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	repo := repository.NewSalesOrderRepository(dbTest.GormDB, zap.NewNop())

	// Cria um contexto com deadline já expirado
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	salesOrder := &models.SalesOrder{
		ContactID: 1,
		// QuotationID omitido
		Status: models.SOStatusDraft,
	}

	// Testa operações com contexto expirado
	err := repo.CreateSalesOrder(ctx, salesOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	_, err = repo.GetSalesOrderByID(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	err = repo.UpdateSalesOrder(ctx, 1, salesOrder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")

	err = repo.DeleteSalesOrder(ctx, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}
