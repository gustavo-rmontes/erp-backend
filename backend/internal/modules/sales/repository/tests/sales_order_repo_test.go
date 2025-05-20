package repository_test

import (
	"ERP-ONSMART/backend/internal/db"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/utils/pagination"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	viper.SetConfigFile("../../../../../../.env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		panic("Erro ao carregar .env: " + err.Error())
	}
	os.Exit(m.Run())
}

// createTestSalesOrder cria um pedido de venda para testes
func createTestSalesOrder(t *testing.T) *models.SalesOrder {
	// Cria uma cotação válida primeiro
	quotation := createTestQuotation(t) // Esta função já gerencia seu próprio repositório

	// Cria um repositório de pedidos de venda
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria um pedido de venda de teste com QuotationID válido
	salesOrder := &models.SalesOrder{
		QuotationID:     quotation.ID, // Aqui está a correção
		ContactID:       1,
		Status:          "",
		ExpectedDate:    time.Now().AddDate(0, 1, 0),
		SubTotal:        1000.0,
		TaxTotal:        100.0,
		DiscountTotal:   50.0,
		GrandTotal:      1050.0,
		Notes:           "Pedido de teste via testes automatizados",
		PaymentTerms:    "Condições de pagamento: 30 dias",
		ShippingAddress: "Rua de Testes, 123, Cidade Teste",
	}

	// Salva o pedido
	err = repo.CreateSalesOrder(salesOrder)
	assert.NoError(t, err)

	return salesOrder
}

// createTestSalesOrderWithItems cria um pedido de venda com itens para testes
func createTestSalesOrderWithItems(t *testing.T) *models.SalesOrder {
	// Cria um pedido básico
	salesOrder := createTestSalesOrder(t)

	// Adiciona itens ao pedido
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Busca o pedido para ter o ID correto
	salesOrder, err = repo.GetSalesOrderByID(salesOrder.ID)
	assert.NoError(t, err)

	// Criamos itens manualmente (normalmente você buscaria produtos reais do banco)
	items := []models.SOItem{
		{
			SalesOrderID: salesOrder.ID,
			ProductID:    1, // Assume que existe um produto com ID 1
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
			ProductID:    2, // Assume que existe um produto com ID 2
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

	// Abre conexão com o banco para adicionar os itens diretamente
	db, err := db.OpenGormDB()
	assert.NoError(t, err)

	// Adiciona os itens diretamente no banco
	for _, item := range items {
		err := db.Create(&item).Error
		assert.NoError(t, err)
	}

	// Atualiza o valor total do pedido
	salesOrder.SubTotal = 240.0   // (2*100) + (1*50) - 10
	salesOrder.TaxTotal = 43.2    // 240 * 0.18
	salesOrder.GrandTotal = 283.2 // 240 + 43.2
	err = repo.UpdateSalesOrder(salesOrder.ID, salesOrder)
	assert.NoError(t, err)

	// Busca o pedido novamente para ter os itens carregados
	updatedSalesOrder, err := repo.GetSalesOrderByID(salesOrder.ID)
	assert.NoError(t, err)

	return updatedSalesOrder
}

// createSalesOrderFromQuotation cria um pedido de venda a partir de uma cotação
func createSalesOrderFromQuotation(t *testing.T) (*models.Quotation, *models.SalesOrder) {
	// Cria uma cotação aceita com itens
	quotationRepo, err := repository.NewQuotationRepository()
	assert.NoError(t, err)

	quotation := createTestQuotationWithItems(t) // Essa função deve estar disponível no pacote de teste
	quotation.Status = models.QuotationStatusAccepted
	err = quotationRepo.UpdateQuotation(quotation.ID, quotation)
	assert.NoError(t, err)

	// Converte para pedido de venda
	err = quotationRepo.ConvertToSalesOrder(quotation.ID)
	assert.NoError(t, err)

	// Busca o pedido de venda criado
	db, err := db.OpenGormDB()
	assert.NoError(t, err)

	var salesOrder models.SalesOrder
	err = db.Where("quotation_id = ?", quotation.ID).First(&salesOrder).Error
	assert.NoError(t, err)

	return quotation, &salesOrder
}

func Test_GetAllSalesOrders(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria alguns pedidos para teste
	var createdOrders []*models.SalesOrder
	for i := 0; i < 3; i++ {
		so := createTestSalesOrder(t)
		createdOrders = append(createdOrders, so)
	}

	// Limpa os pedidos no final do teste
	defer func() {
		for _, so := range createdOrders {
			repo.DeleteSalesOrder(so.ID)
		}
	}()

	// Busca todos os pedidos com paginação
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetAllSalesOrders(params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3))
	salesOrders := result.Items.([]models.SalesOrder)
	assert.NotEmpty(t, salesOrders)
}

func Test_GetSalesOrdersByStatus(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria um pedido com status específico
	salesOrder := createTestSalesOrder(t)
	salesOrder.Status = models.SOStatusConfirmed
	err = repo.UpdateSalesOrder(salesOrder.ID, salesOrder)
	assert.NoError(t, err)

	// Limpa o pedido no final do teste
	defer func() {
		repo.DeleteSalesOrder(salesOrder.ID)
	}()

	// Busca por status
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetSalesOrdersByStatus(models.SOStatusConfirmed, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(1))

	// Verifica se o pedido está nos resultados
	salesOrders := result.Items.([]models.SalesOrder)
	found := false
	for _, so := range salesOrders {
		if so.ID == salesOrder.ID {
			found = true
			assert.Equal(t, models.SOStatusConfirmed, so.Status)
			break
		}
	}
	assert.True(t, found, "O pedido com status confirmado deveria estar nos resultados")
}

func Test_SearchSalesOrders(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria um pedido para pesquisa
	searchOrder := createTestSalesOrder(t)
	searchOrder.Notes = "Pedido PESQUISÁVEL especial"
	err = repo.UpdateSalesOrder(searchOrder.ID, searchOrder)
	assert.NoError(t, err)

	// Limpa o pedido no final do teste
	defer func() {
		repo.DeleteSalesOrder(searchOrder.ID)
	}()

	// Define filtros de pesquisa
	filter := repository.SalesOrderFilter{
		Status:      []string{models.SOStatusDraft},
		ContactID:   1,
		SearchQuery: "PESQUISÁVEL",
	}

	// Busca com filtros
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.SearchSalesOrders(filter, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verifica se o pedido está nos resultados
	salesOrders := result.Items.([]models.SalesOrder)
	found := false
	for _, so := range salesOrders {
		if so.ID == searchOrder.ID {
			found = true
			assert.Contains(t, so.Notes, "PESQUISÁVEL")
			break
		}
	}
	assert.True(t, found, "O pedido pesquisável deveria estar nos resultados")
}

func Test_SalesOrderRepository_NotFound(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Testa busca com ID inválido
	_, err = repo.GetSalesOrderByID(999999)
	assert.Error(t, err)

	// Tenta deletar com ID inválido
	err = repo.DeleteSalesOrder(999999)
	assert.Error(t, err)

	// Tenta atualizar com ID inválido
	err = repo.UpdateSalesOrder(999999, &models.SalesOrder{})
	assert.Error(t, err)
}

func Test_GetSalesOrdersByContact(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria vários pedidos com o mesmo contactID
	contactID := 1 // Assume que existe um contato com ID 1
	var createdOrders []*models.SalesOrder

	// Cria 3 pedidos para o mesmo contato
	for i := 0; i < 3; i++ {
		salesOrder := createTestSalesOrder(t)
		salesOrder.ContactID = contactID
		err = repo.UpdateSalesOrder(salesOrder.ID, salesOrder)
		assert.NoError(t, err)
		createdOrders = append(createdOrders, salesOrder)
	}

	// Cria um pedido para um contato diferente
	otherOrder := createTestSalesOrder(t)
	otherOrder.ContactID = 2 // Contato diferente
	err = repo.UpdateSalesOrder(otherOrder.ID, otherOrder)
	assert.NoError(t, err)

	// Limpa os pedidos no final do teste
	defer func() {
		for _, so := range createdOrders {
			repo.DeleteSalesOrder(so.ID)
		}
		repo.DeleteSalesOrder(otherOrder.ID)
	}()

	// Busca pedidos por contato
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetSalesOrdersByContact(contactID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3))

	// Verifica se todos os pedidos retornados pertencem ao contato correto
	salesOrders := result.Items.([]models.SalesOrder)
	for _, so := range salesOrders {
		assert.Equal(t, contactID, so.ContactID, "Todos os pedidos devem pertencer ao contato especificado")
	}

	// Verifica se o pedido do outro contato não está nos resultados
	found := false
	for _, so := range salesOrders {
		if so.ID == otherOrder.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "O pedido do outro contato não deveria estar nos resultados")
}

func Test_GetSalesOrdersByPeriod(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria pedidos para o teste
	currentOrder := createTestSalesOrder(t)

	// Busca o pedido por período
	now := time.Now()
	pastDate := now.AddDate(0, -1, 0) // 1 mês atrás
	// futureDate := now.AddDate(0, 1, 0) // 1 mês no futuro

	// Limpa o pedido no final do teste
	defer func() {
		repo.DeleteSalesOrder(currentOrder.ID)
	}()

	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	// Testa busca no período atual
	result, err := repo.GetSalesOrdersByPeriod(pastDate, now, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verifica se o pedido está nos resultados
	found := false
	salesOrders := result.Items.([]models.SalesOrder)
	for _, so := range salesOrders {
		if so.ID == currentOrder.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "O pedido atual deveria estar nos resultados do período")
}

func Test_GetSalesOrdersByQuotation(t *testing.T) {
	// Inicializa repositórios
	soRepo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria um pedido a partir de uma cotação
	quotation, salesOrder := createSalesOrderFromQuotation(t)

	// Limpa no final do teste
	defer func() {
		// Limpa primeiro o pedido, depois a cotação
		soRepo.DeleteSalesOrder(salesOrder.ID)
		quoRepo, _ := repository.NewQuotationRepository()
		quoRepo.DeleteQuotation(quotation.ID)
	}()

	// Busca pedidos por cotação
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := soRepo.GetSalesOrdersByQuotation(quotation.ID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalItems, "Deveria encontrar exatamente 1 pedido")

	// Verifica se o pedido está corretamente vinculado à cotação
	salesOrders := result.Items.([]models.SalesOrder)
	assert.Equal(t, 1, len(salesOrders), "Deveria haver exatamente 1 pedido no resultado")
	assert.Equal(t, quotation.ID, salesOrders[0].QuotationID, "O pedido deve referenciar a cotação correta")
}

func Test_GetSalesOrdersByExpectedDate(t *testing.T) {
	// Inicializa o repositório
	repo, err := repository.NewSalesOrderRepository()
	assert.NoError(t, err)

	// Cria pedidos com diferentes datas esperadas
	// 1. Pedido com data esperada no passado
	pastOrder := createTestSalesOrder(t)
	pastExpectedDate := time.Now().AddDate(0, -1, 0) // 1 mês atrás
	pastOrder.ExpectedDate = pastExpectedDate
	err = repo.UpdateSalesOrder(pastOrder.ID, pastOrder)
	assert.NoError(t, err)

	// 2. Pedido com data esperada na próxima semana
	nearFutureOrder := createTestSalesOrder(t)
	nearFutureExpectedDate := time.Now().AddDate(0, 0, 7) // 7 dias no futuro
	nearFutureOrder.ExpectedDate = nearFutureExpectedDate
	err = repo.UpdateSalesOrder(nearFutureOrder.ID, nearFutureOrder)
	assert.NoError(t, err)

	// 3. Pedido com data esperada daqui a 2 meses
	farFutureOrder := createTestSalesOrder(t)
	farFutureExpectedDate := time.Now().AddDate(0, 2, 0) // 2 meses no futuro
	farFutureOrder.ExpectedDate = farFutureExpectedDate
	err = repo.UpdateSalesOrder(farFutureOrder.ID, farFutureOrder)
	assert.NoError(t, err)

	// Limpa os pedidos no final do teste
	defer func() {
		repo.DeleteSalesOrder(pastOrder.ID)
		repo.DeleteSalesOrder(nearFutureOrder.ID)
		repo.DeleteSalesOrder(farFutureOrder.ID)
	}()

	// Testa busca por intervalo que deve incluir apenas os dois primeiros pedidos
	// (o do passado e o da próxima semana, excluindo o de 2 meses no futuro)
	startDate := time.Now().AddDate(0, -2, 0) // 2 meses atrás
	endDate := time.Now().AddDate(0, 1, 0)    // 1 mês no futuro

	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetSalesOrdersByExpectedDate(startDate, endDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(2), "Deveria haver pelo menos 2 pedidos no intervalo de datas")

	// Verifica se os pedidos corretos estão nos resultados
	salesOrders := result.Items.([]models.SalesOrder)

	// Verifica se cada pedido esperado está nos resultados
	pastOrderFound := false
	nearFutureOrderFound := false
	farFutureOrderFound := false

	for _, so := range salesOrders {
		if so.ID == pastOrder.ID {
			pastOrderFound = true
			assert.Equal(t, pastExpectedDate.Format("2006-01-02"), so.ExpectedDate.Format("2006-01-02"),
				"A data esperada do pedido do passado deve ser preservada")
		}
		if so.ID == nearFutureOrder.ID {
			nearFutureOrderFound = true
			assert.Equal(t, nearFutureExpectedDate.Format("2006-01-02"), so.ExpectedDate.Format("2006-01-02"),
				"A data esperada do pedido da próxima semana deve ser preservada")
		}
		if so.ID == farFutureOrder.ID {
			farFutureOrderFound = true
		}
	}

	// Verifica os resultados esperados
	assert.True(t, pastOrderFound, "O pedido com data esperada no passado deveria estar nos resultados")
	assert.True(t, nearFutureOrderFound, "O pedido com data esperada na próxima semana deveria estar nos resultados")
	assert.False(t, farFutureOrderFound, "O pedido com data esperada em 2 meses NÃO deveria estar nos resultados")

	// Teste opcional: verificar a ordenação por data esperada (ASC)
	if len(salesOrders) >= 2 {
		for i := 0; i < len(salesOrders)-1; i++ {
			assert.LessOrEqual(t,
				salesOrders[i].ExpectedDate.Format("2006-01-02"),
				salesOrders[i+1].ExpectedDate.Format("2006-01-02"),
				"Os pedidos devem estar ordenados por data esperada (ASC)")
		}
	}
}
