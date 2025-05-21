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

	// Cria várias cotações com o mesmo contactID
	contactID := 1 // Assume que existe um contato com ID 1
	var createdQuotations []*models.Quotation

	// Cria 3 cotações para o mesmo contato
	for i := 0; i < 3; i++ {
		quotation := createTestQuotation(t, dbTest.GormDB, logger)
		quotation.ContactID = contactID
		err := repo.UpdateQuotation(ctx, quotation.ID, quotation)
		assert.NoError(t, err)
		createdQuotations = append(createdQuotations, quotation)
	}

	// Cria uma cotação para um contato diferente
	otherQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	otherQuotation.ContactID = 2 // Contato diferente
	err := repo.UpdateQuotation(ctx, otherQuotation.ID, otherQuotation)
	assert.NoError(t, err)

	// Use defer para garantir limpeza, mesmo em caso de falha
	defer func() {
		// Limpa as cotações criadas
		for _, q := range createdQuotations {
			repo.DeleteQuotation(ctx, q.ID)
		}
		repo.DeleteQuotation(ctx, otherQuotation.ID)
	}()

	// Busca cotações por contato
	params := &pagination.PaginationParams{
		Page:     1,
		PageSize: 10,
	}

	result, err := repo.GetQuotationsByContact(ctx, contactID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3))

	// Verifica se todas as cotações retornadas pertencem ao contato correto
	quotations := result.Items.([]models.Quotation)
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
	result, err = repo.GetQuotationsByContact(ctx, contactID, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.GreaterOrEqual(t, result.TotalItems, int64(3))            // O total deve ser o mesmo
	assert.LessOrEqual(t, len(result.Items.([]models.Quotation)), 2) // Mas o número de itens deve ser limitado pelo pageSize

	// Verifica se o preload de Contact e Items está funcionando
	if len(quotations) > 0 {
		firstQuotation := quotations[0]
		assert.NotNil(t, firstQuotation.Contact, "O relacionamento Contact deve ser carregado")
	}
}

func Test_GetQuotationsByPeriod(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)

	ctx := context.Background()

	// Cria cotações com diferentes datas
	quotations := []*models.Quotation{}

	// Cotação dentro do período (atual)
	currentQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	quotations = append(quotations, currentQuotation)

	// Cotação antiga - manipula created_at diretamente no banco
	oldQuotation := createTestQuotation(t, dbTest.GormDB, logger)

	// Define períodos para teste APÓS criar as cotações
	now := time.Now()                     // Deve ser definido APÓS criar cotações
	pastDate := now.AddDate(0, -1, 0)     // 1 mês atrás
	futureDate := now.AddDate(0, 1, 0)    // 1 mês no futuro
	veryPastDate := now.AddDate(0, -2, 0) // 2 meses atrás

	// Manipula a data de criação da cotação antiga
	err := repo.SetCreatedAtForTesting(ctx, oldQuotation.ID, veryPastDate)
	assert.NoError(t, err)
	quotations = append(quotations, oldQuotation)
	// Use defer para garantir limpeza
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
	result, err := repo.GetQuotationsByPeriod(ctx, pastDate, now, params)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	currentPeriodQuotations := result.Items.([]models.Quotation)

	// Verifica se cotações do período atual estão nos resultados
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
	resultAll, err := repo.GetQuotationsByPeriod(ctx, veryPastDate, futureDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, resultAll)
	assert.GreaterOrEqual(t, resultAll.TotalItems, int64(2), "Deveria encontrar pelo menos as duas cotações criadas")

	// Testa paginação
	params.PageSize = 1
	resultPaginated, err := repo.GetQuotationsByPeriod(ctx, veryPastDate, futureDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, resultPaginated)
	assert.Equal(t, int64(resultAll.TotalItems), resultPaginated.TotalItems, "Total de itens deve ser o mesmo")
	assert.Equal(t, 1, len(resultPaginated.Items.([]models.Quotation)), "Número de itens deve ser limitado pelo pageSize")

	// Testa se o preload está funcionando
	if len(currentPeriodQuotations) > 0 {
		assert.NotNil(t, currentPeriodQuotations[0].Contact, "O relacionamento Contact deve ser carregado")
	}
}

func Test_GetQuotationsByExpiryDateRange(t *testing.T) {
	dbTest := testutils.NewDBTest(t)
	defer dbTest.Cleanup()
	logger := zap.NewNop()

	err := dbTest.GormDB.Exec("DELETE FROM quotations").Error
	assert.NoError(t, err, "deveria conseguir apagar todas as quotations antes do teste")

	// Inicializa o repositório
	repo := repository.NewQuotationRepository(dbTest.GormDB, logger)

	ctx := context.Background()

	// Cria cotações com diferentes datas de expiração
	quotations := []*models.Quotation{}

	// Cotação que expira em 1 mês
	upcomingQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	upcomingQuotation.ExpiryDate = time.Now().AddDate(0, 1, 0) // Expira em 1 mês
	err = repo.UpdateQuotation(ctx, upcomingQuotation.ID, upcomingQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, upcomingQuotation)

	// Cotação que expira em 2 semanas
	midRangeQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	midRangeQuotation.ExpiryDate = time.Now().AddDate(0, 0, 14) // Expira em 2 semanas
	err = repo.UpdateQuotation(ctx, midRangeQuotation.ID, midRangeQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, midRangeQuotation)

	// Cotação que já expirou
	expiredQuotation := createTestQuotation(t, dbTest.GormDB, logger)
	expiredQuotation.ExpiryDate = time.Now().AddDate(0, 0, -10) // Expirou há 10 dias
	err = repo.UpdateQuotation(ctx, expiredQuotation.ID, expiredQuotation)
	assert.NoError(t, err)
	quotations = append(quotations, expiredQuotation)

	// Define o intervalo de datas para a busca - apenas as próximas 3 semanas
	now := time.Now()
	startDate := now
	endDate := now.AddDate(0, 0, 21) // 3 semanas à frente

	// Use defer para garantir limpeza
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

	result, err := repo.GetQuotationsByExpiryDateRange(ctx, startDate, endDate, params)
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

	largerResult, err := repo.GetQuotationsByExpiryDateRange(ctx, largerStartDate, largerEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, largerResult)

	// Deve incluir todas as três cotações no intervalo maior
	assert.Equal(t, int64(3), largerResult.TotalItems, "O intervalo maior deveria encontrar 3 cotações")

	// Testa paginação
	params.PageSize = 1
	pagedResult, err := repo.GetQuotationsByExpiryDateRange(ctx, largerStartDate, largerEndDate, params)
	assert.NoError(t, err)
	assert.NotNil(t, pagedResult)
	assert.Equal(t, int64(3), pagedResult.TotalItems, "Total de itens deve ser o mesmo")
	assert.Equal(t, 1, len(pagedResult.Items.([]models.Quotation)), "Número de itens deve ser limitado pelo pageSize")

	// Testa se o preload está funcionando
	if len(rangeQuotations) > 0 {
		assert.NotNil(t, rangeQuotations[0].Contact, "O relacionamento Contact deve ser carregado")
		if len(rangeQuotations[0].Items) > 0 {
			assert.NotNil(t, rangeQuotations[0].Items, "Os Items devem ser carregados")
		}
	}
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
