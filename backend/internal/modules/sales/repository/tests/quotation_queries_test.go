package repository_test

import (
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/modules/sales/repository"
	"ERP-ONSMART/backend/internal/utils/pagination"
	testutils "ERP-ONSMART/backend/internal/utils/test_utils"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_GetAllQuotations(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()

	err := dbTest.GormDB.
		Exec("DELETE FROM quotations").
		Error
	assert.NoError(t, err, "deveria conseguir apagar todas as quotations antes do teste")

	logger := zap.NewNop()

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, zap.NewNop())

	ctx := context.Background()

	// Cria algumas cotações para teste
	for i := 0; i < 3; i++ {
		createTestQuotation(t, dbTest.GormDB, logger)
	}

	// Busca todas as cotações com paginação
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetAllQuotations(ctx, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3))

	quotations, ok := result.Items.([]models.Quotation)
	assert.True(t, ok, "Items deve ser um slice de Quotation")
	assert.Len(t, quotations, 3)
	assert.NotEmpty(t, quotations)
}

func Test_GetQuotationsByStatus(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)

	ctx := context.Background()

	// Cria uma cotação com status específico
	quotation := createTestQuotation(t, dbTest.GormDB, logger)
	quotation.Status = models.QuotationStatusSent
	err := repo.UpdateQuotation(ctx, quotation.ID, quotation)
	assert.NoError(t, err)

	// Busca por status
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetQuotationsByStatus(ctx, models.QuotationStatusSent, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(1))

	// Verifica se a cotação está nos resultados
	quotations := result.Items.([]models.Quotation)
	found := false
	for _, q := range quotations {
		if q.ID == quotation.ID {
			found = true
			assert.Equal(t, models.QuotationStatusSent, q.Status)
			break
		}
	}
	assert.True(t, found, "A cotação com status enviado deveria estar nos resultados")
}

func Test_GetQuotationsByContact(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)
	ctx := context.Background()

	// Cria contatos explicitamente para garantir que existem
	testContact := createTestClient(t, dbTest.GormDB, logger)
	contactID := testContact.ID

	otherContact := createTestSupplier(t, dbTest.GormDB, logger)
	otherContactID := otherContact.ID

	var createdQuotations []*models.Quotation

	// Cria 3 cotações para o contato principal
	for i := 0; i < 3; i++ {
		quotation := createTestQuotation(t, dbTest.GormDB, logger)
		quotation.ContactID = contactID
		err := repo.UpdateQuotation(ctx, quotation.ID, quotation)
		assert.NoError(t, err)
		createdQuotations = append(createdQuotations, quotation)
	}

	// Cria uma cotação para um contato diferente
	otherQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	otherQuotation.ContactID = otherContactID
	err := repo.UpdateQuotation(ctx, otherQuotation.ID, otherQuotation)
	assert.NoError(t, err)

	// Garante limpeza, mesmo em caso de falha
	defer func() {
		// Limpa as cotações criadas
		for _, q := range createdQuotations {
			repo.DeleteQuotation(ctx, q.ID)
		}
		repo.DeleteQuotation(ctx, otherQuotation.ID)
	}()

	// Testa busca de cotações por contato
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetQuotationsByContact(ctx, contactID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3), "Deve encontrar pelo menos 3 cotações")

	// Verifica se todas as cotações retornadas pertencem ao contato correto
	quotations := result.Items.([]models.Quotation)
	assert.GreaterOrEqual(t, len(quotations), 3, "Deve retornar pelo menos 3 cotações")

	for _, q := range quotations {
		assert.Equal(t, contactID, q.ContactID, "Todas as cotações devem pertencer ao contato especificado")
	}

	// Verifica se a cotação do outro contato não está nos resultados
	found := false
	for _, q := range quotations {
		if q.ID == otherQuotation.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "A cotação do outro contato não deveria estar nos resultados")

	// Testa a paginação
	params.PageSize = 2
	paginatedResult, err := repo.GetQuotationsByContact(ctx, contactID, params)
	assert.NoError(t, err)
	assert.NotNil(t, paginatedResult)
	assert.GreaterOrEqual(t, paginatedResult.TotalItems, int64(3), "O total deve ser o mesmo")
	assert.LessOrEqual(t, len(paginatedResult.Items.([]models.Quotation)), 2, "Número de itens deve ser limitado pelo pageSize")

	// Verifica se o preload de Contact está funcionando corretamente
	if len(quotations) > 0 {
		firstQuotation := quotations[0]

		// Verifica se o Contact foi carregado
		assert.NotNil(t, firstQuotation.Contact, "O relacionamento Contact deve ser carregado")
		assert.Equal(t, contactID, firstQuotation.Contact.ID, "O Contact carregado deve ter o ID correto")
		assert.NotEmpty(t, firstQuotation.Contact.Name, "O Contact deve ter um nome preenchido")
		assert.Equal(t, testContact.Name, firstQuotation.Contact.Name, "O nome do Contact deve corresponder ao contato criado")
		assert.Equal(t, testContact.Type, firstQuotation.Contact.Type, "O tipo do Contact deve corresponder ao contato criado")
	}

	// Testa busca com contato que não possui cotações
	emptyContactResult, err := repo.GetQuotationsByContact(ctx, 99999, params)
	assert.NoError(t, err)
	assert.NotNil(t, emptyContactResult)
	assert.Equal(t, int64(0), emptyContactResult.TotalItems, "Contato sem cotações deve retornar lista vazia")
	assert.Empty(t, emptyContactResult.Items.([]models.Quotation), "Lista de cotações deve estar vazia")
}

func Test_GetQuotationsByDateRange(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)
	ctx := context.Background()

	// Cria contatos explicitamente para garantir que existem
	testContact := createTestClient(t, dbTest.GormDB, logger)
	contactID := testContact.ID

	// Cria cotações com diferentes datas
	quotations := []*models.Quotation{}

	// Cotação atual (dentro do período de teste)
	currentQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	currentQuotation.ContactID = contactID
	err := repo.UpdateQuotation(ctx, currentQuotation.ID, currentQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, currentQuotation)

	// Cotação antiga - vamos manipular a data de criação
	oldQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	oldQuotation.ContactID = contactID
	err = repo.UpdateQuotation(ctx, oldQuotation.ID, oldQuotation)
	assert.NoError(t, err)

	// Define períodos para teste APÓS criar as cotações
	now := time.Now()
	pastDate := now.AddDate(0, -1, 0)     // 1 mês atrás
	futureDate := now.AddDate(0, 1, 0)    // 1 mês no futuro
	veryPastDate := now.AddDate(0, -2, 0) // 2 meses atrás

	// Manipula a data de criação da cotação antiga usando o método de teste
	err = repo.SetCreatedAtForTesting(ctx, oldQuotation.ID, veryPastDate)
	assert.NoError(t, err)
	quotations = append(quotations, oldQuotation)

	// Garante limpeza das cotações criadas
	defer func() {
		for _, q := range quotations {
			repo.DeleteQuotation(ctx, q.ID)
		}
	}()

	// Testa busca no período que deve incluir apenas a cotação atual
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	// Período: de 1 mês atrás até agora (deve incluir cotações atuais, excluir antigas)
	result, err := repo.GetQuotationsByDateRange(ctx, pastDate, now, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	currentPeriodQuotations := result.Items.([]models.Quotation)

	// Verifica se a cotação atual está nos resultados
	foundCurrent := false
	for _, q := range currentPeriodQuotations {
		if q.ID == currentQuotation.ID {
			foundCurrent = true
			break
		}
	}
	assert.True(t, foundCurrent, "A cotação atual deveria estar nos resultados do período")

	// Verifica se a cotação antiga não está nos resultados do período atual
	foundOld := false
	for _, q := range currentPeriodQuotations {
		if q.ID == oldQuotation.ID {
			foundOld = true
			break
		}
	}
	assert.False(t, foundOld, "A cotação antiga não deveria estar nos resultados do período atual")

	// Testa busca com período que deve incluir todas as cotações
	resultAll, err := repo.GetQuotationsByDateRange(ctx, veryPastDate, futureDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, resultAll)
	assert.GreaterOrEqual(t, resultAll.TotalItems, int64(2), "Deveria encontrar pelo menos as duas cotações criadas")

	// Testa paginação
	params.PageSize = 1
	resultPaginated, err := repo.GetQuotationsByDateRange(ctx, veryPastDate, futureDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, resultPaginated)
	assert.Equal(t, resultAll.TotalItems, resultPaginated.TotalItems, "Total de itens deve ser o mesmo")
	assert.Equal(t, 1, len(resultPaginated.Items.([]models.Quotation)), "Número de itens deve ser limitado pelo pageSize")

	// Verifica se o preload está funcionando corretamente
	if len(currentPeriodQuotations) > 0 {
		firstQuotation := currentPeriodQuotations[0]

		// Verifica se o Contact foi carregado
		assert.NotNil(t, firstQuotation.Contact, "O relacionamento Contact deve ser carregado")
		assert.Equal(t, contactID, firstQuotation.Contact.ID, "O Contact carregado deve ter o ID correto")
		assert.NotEmpty(t, firstQuotation.Contact.Name, "O Contact deve ter um nome preenchido")
		assert.Equal(t, testContact.Name, firstQuotation.Contact.Name, "O nome do Contact deve corresponder ao contato criado")

		// Verifica se os Items foram carregados (se existem)
		assert.NotNil(t, firstQuotation.Items, "Os Items devem estar inicializados (mesmo que vazios)")
	}

	// Testa período vazio (sem cotações)
	emptyStartDate := now.AddDate(0, 2, 0) // 2 meses no futuro
	emptyEndDate := now.AddDate(0, 3, 0)   // 3 meses no futuro

	emptyResult, err := repo.GetQuotationsByDateRange(ctx, emptyStartDate, emptyEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, emptyResult)
	assert.Equal(t, int64(0), emptyResult.TotalItems, "Período vazio deve retornar zero cotações")
	assert.Empty(t, emptyResult.Items.([]models.Quotation), "Lista deve estar vazia para período sem cotações")
}

func Test_GetQuotationsByExpiryRange(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	err := dbTest.GormDB.Exec("DELETE FROM quotations").Error
	assert.NoError(t, err, "deveria conseguir apagar todas as quotations antes do teste")

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)
	ctx := context.Background()

	// Cria contatos explicitamente para garantir que existem
	testContact := createTestClient(t, dbTest.GormDB, logger)
	contactID := testContact.ID

	// Cria cotações com diferentes datas de expiração
	quotations := []*models.Quotation{}

	// Cotação que expira em 1 mês
	upcomingQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	upcomingQuotation.ContactID = contactID
	upcomingQuotation.ExpiryDate = time.Now().AddDate(0, 1, 0) // Expira em 1 mês
	err = repo.UpdateQuotation(ctx, upcomingQuotation.ID, upcomingQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, upcomingQuotation)

	// Cotação que expira em 2 semanas
	midRangeQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	midRangeQuotation.ContactID = contactID
	midRangeQuotation.ExpiryDate = time.Now().AddDate(0, 0, 14) // Expira em 2 semanas
	err = repo.UpdateQuotation(ctx, midRangeQuotation.ID, midRangeQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, midRangeQuotation)

	// Cotação que já expirou
	expiredQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	expiredQuotation.ContactID = contactID
	expiredQuotation.ExpiryDate = time.Now().AddDate(0, 0, -10) // Expirou há 10 dias
	err = repo.UpdateQuotation(ctx, expiredQuotation.ID, expiredQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, expiredQuotation)

	// Define o intervalo de datas para a busca - apenas as próximas 3 semanas
	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 0, 21) // 3 semanas à frente

	// Garante limpeza das cotações criadas
	defer func() {
		for _, q := range quotations {
			repo.DeleteQuotation(ctx, q.ID)
		}
	}()

	// Testa busca no intervalo definido (deve incluir apenas a cotação de 2 semanas)
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetQuotationsByExpiryRange(ctx, startDate, endDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verifica o número total de cotações no intervalo
	assert.Equal(t, int64(1), result.TotalItems, "Deveria encontrar apenas 1 cotação no intervalo de 3 semanas")

	// Obtém as cotações retornadas
	rangeQuotations := result.Items.([]models.Quotation)

	// Verifica se a cotação de meio-termo (2 semanas) está nos resultados
	foundMidRange := false
	for _, q := range rangeQuotations {
		if q.ID == midRangeQuotation.ID {
			foundMidRange = true
			break
		}
	}
	assert.True(t, foundMidRange, "A cotação que expira em 2 semanas deveria estar nos resultados")

	// Verifica se a cotação de 1 mês não está nos resultados (fora do intervalo)
	foundUpcoming := false
	for _, q := range rangeQuotations {
		if q.ID == upcomingQuotation.ID {
			foundUpcoming = true
			break
		}
	}
	assert.False(t, foundUpcoming, "A cotação que expira em 1 mês não deveria estar nos resultados")

	// Verifica se a cotação expirada não está nos resultados (fora do intervalo)
	foundExpired := false
	for _, q := range rangeQuotations {
		if q.ID == expiredQuotation.ID {
			foundExpired = true
			break
		}
	}
	assert.False(t, foundExpired, "A cotação expirada não deveria estar nos resultados")

	// Testa um intervalo maior que deve incluir todas as cotações
	largerStartDate := now.AddDate(0, 0, -15) // 15 dias atrás
	largerEndDate := now.AddDate(0, 2, 0)     // 2 meses à frente

	largerResult, err := repo.GetQuotationsByExpiryRange(ctx, largerStartDate, largerEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, largerResult)

	// Deve incluir todas as três cotações no intervalo maior
	assert.Equal(t, int64(3), largerResult.TotalItems, "O intervalo maior deveria encontrar 3 cotações")

	// Testa paginação
	params.PageSize = 1
	pagedResult, err := repo.GetQuotationsByExpiryRange(ctx, largerStartDate, largerEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, pagedResult)
	assert.Equal(t, int64(3), pagedResult.TotalItems, "Total de itens deve ser o mesmo")
	assert.Equal(t, 1, len(pagedResult.Items.([]models.Quotation)), "Número de itens deve ser limitado pelo pageSize")

	// Verifica se o preload está funcionando corretamente
	if len(rangeQuotations) > 0 {
		firstQuotation := rangeQuotations[0]

		// Verifica se o Contact foi carregado
		assert.NotNil(t, firstQuotation.Contact, "O relacionamento Contact deve ser carregado")
		assert.Equal(t, contactID, firstQuotation.Contact.ID, "O Contact carregado deve ter o ID correto")
		assert.NotEmpty(t, firstQuotation.Contact.Name, "O Contact deve ter um nome preenchido")
		assert.Equal(t, testContact.Name, firstQuotation.Contact.Name, "O nome do Contact deve corresponder ao contato criado")

		// Verifica se os Items foram carregados (mesmo que vazios)
		assert.NotNil(t, firstQuotation.Items, "Os Items devem estar inicializados (mesmo que vazios)")
	}

	// Testa intervalo vazio (sem cotações)
	emptyStartDate := now.AddDate(0, 3, 0) // 3 meses no futuro
	emptyEndDate := now.AddDate(0, 4, 0)   // 4 meses no futuro

	emptyResult, err := repo.GetQuotationsByExpiryRange(ctx, emptyStartDate, emptyEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, emptyResult)
	assert.Equal(t, int64(0), emptyResult.TotalItems, "Intervalo vazio deve retornar zero cotações")
	assert.Empty(t, emptyResult.Items.([]models.Quotation), "Lista deve estar vazia para intervalo sem cotações")
}

func Test_GetQuotationsByContactType(t *testing.T) {
	// 1) Prepara o DB de teste
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	// 2) Limpa as tabelas que vamos usar
	assert.NoError(t, dbTest.GormDB.Exec("DELETE FROM quotations").Error)
	assert.NoError(t, dbTest.GormDB.Exec("DELETE FROM contacts").Error)

	// 3) Cria dois contatos de teste com tipos diferentes
	cli := createTestClient(t, dbTest.GormDB, logger)

	fn := createTestSupplier(t, dbTest.GormDB, logger)

	// 4) Inicializa repositório e contexto
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)
	ctx := context.Background()

	// 5) Gera 3 cotações para cada contato
	for i := 0; i < 3; i++ {
		q := createTestQuotation(t, dbTest.GormDB, logger)
		q.ContactID = cli.ID
		assert.NoError(t, repo.UpdateQuotation(ctx, q.ID, q))
	}
	for i := 0; i < 2; i++ {
		q := createTestQuotation(t, dbTest.GormDB, logger)
		q.ContactID = fn.ID
		assert.NoError(t, repo.UpdateQuotation(ctx, q.ID, q))
	}

	// 6) Busca por tipo "cliente"
	params := &pagination.PaginationParams{Page: 1, PageSize: 10}
	cliRes, err := repo.GetQuotationsByContactType(ctx, "cliente", params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), cliRes.TotalItems)
	cliItems := cliRes.Items.([]models.Quotation)
	assert.Len(t, cliItems, 3)
	for _, q := range cliItems {
		assert.Equal(t, cli.ID, q.ContactID)
		assert.Equal(t, "cliente", q.Contact.Type)
	}

	// 7) Busca por tipo "fornecedor"
	fnRes, err := repo.GetQuotationsByContactType(ctx, "fornecedor", params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), fnRes.TotalItems)
	fnItems := fnRes.Items.([]models.Quotation)
	assert.Len(t, fnItems, 2)
	for _, q := range fnItems {
		assert.Equal(t, fn.ID, q.ContactID)
		assert.Equal(t, "fornecedor", q.Contact.Type)
	}

	// 8) Paginação (pageSize=1) mantém TotalItems, mas limita itens retornados
	params.PageSize = 1
	paged, err := repo.GetQuotationsByContactType(ctx, "cliente", params)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), paged.TotalItems)
	assert.Len(t, paged.Items.([]models.Quotation), 1)

	// 9) Tipo inexistente → nenhum erro, TotalItems=0, lista vazia
	empty, err := repo.GetQuotationsByContactType(ctx, "nenhum_tipo", &pagination.PaginationParams{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Zero(t, empty.TotalItems)
	assert.Empty(t, empty.Items.([]models.Quotation))
}

func Test_SearchQuotations(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	err := dbTest.GormDB.Exec("DELETE FROM quotations").Error
	assert.NoError(t, err, "deveria conseguir apagar todas as quotations antes do teste")

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)

	ctx := context.Background()

	// Cria uma cotação para pesquisa
	searchQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	searchQuotation.Notes = "Cotação PESQUISÁVEL especial"
	err = repo.UpdateQuotation(ctx, searchQuotation.ID, searchQuotation)
	assert.NoError(t, err)

	// Define filtros de pesquisa
	filter := repository.QuotationFilter{
		Status:      []string{models.QuotationStatusDraft},
		ContactID:   1,
		SearchQuery: "PESQUISÁVEL",
	}

	// Busca com filtros
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.SearchQuotations(ctx, filter, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verifica se a cotação está nos resultados
	quotations := result.Items.([]models.Quotation)
	found := false
	for _, q := range quotations {
		if q.ID == searchQuotation.ID {
			found = true
			assert.Contains(t, q.Notes, "PESQUISÁVEL")
			break
		}
	}
	assert.True(t, found, "A cotação pesquisável deveria estar nos resultados")

	// Limpa a cotação criada
	repo.DeleteQuotation(ctx, searchQuotation.ID)
}
