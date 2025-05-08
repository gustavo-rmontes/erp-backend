package quotation

import (
	"context"
	"fmt"
	"strings"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"go.uber.org/zap"
)

// extractQuotationsFromResult extrai cotações do resultado paginado
func (s *Service) extractQuotationsFromResult(result *pagination.PaginatedResult) []models.Quotation {
	quotations := make([]models.Quotation, 0)
	if items, ok := result.Items.([]models.Quotation); ok {
		quotations = items
	}
	return quotations
}

// applyManualPagination aplica paginação manual a uma slice de cotações
func (s *Service) applyManualPagination(items []models.Quotation, params *pagination.PaginationParams) []models.Quotation {
	startIndex := (params.Page - 1) * params.PageSize
	endIndex := startIndex + params.PageSize

	if startIndex >= len(items) {
		return []models.Quotation{}
	}

	if endIndex > len(items) {
		endIndex = len(items)
	}

	return items[startIndex:endIndex]
}

// applyQuotationFilter aplica filtros comuns e retorna o resultado paginado
func (s *Service) applyQuotationFilter(filter *dto.QuotationFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var result *pagination.PaginatedResult
	var err error

	// Aplicar filtros de acordo com os critérios
	if filter == nil {
		result, err = s.quotationRepo.GetAllQuotations(params)
	} else if filter.ContactID > 0 {
		result, err = s.quotationRepo.GetQuotationsByContact(filter.ContactID, params)
	} else if len(filter.Status) > 0 && filter.Status[0] != "" {
		// Pegar apenas o primeiro status da lista para simplificar
		result, err = s.quotationRepo.GetQuotationsByStatus(filter.Status[0], params)
	} else if filter.IsExpired != nil && *filter.IsExpired {
		result, err = s.quotationRepo.GetExpiredQuotations(params)
	} else {
		result, err = s.quotationRepo.GetAllQuotations(params)
	}

	return result, err
}

// Search busca cotações por termo de pesquisa
func (s *Service) Search(ctx context.Context, query string, params *pagination.PaginationParams) (*dto.PaginatedQuotationShortResponse, error) {
	s.logger.Info("Pesquisando cotações", zap.String("query", query), zap.Any("pagination", params))

	// Implementação simplificada da pesquisa (em um caso real, idealmente seria implementada no nível do banco de dados)
	result, err := s.quotationRepo.GetAllQuotations(params)
	if err != nil {
		s.logger.Error("Erro ao pesquisar cotações", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("falha ao pesquisar cotações: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Filtrar cotações com base na consulta
	filteredQuotations := make([]models.Quotation, 0)
	query = strings.ToLower(query)

	for _, quotation := range quotations {
		if s.quotationMatchesSearch(&quotation, query) {
			filteredQuotations = append(filteredQuotations, quotation)
		}
	}

	// Aplicar paginação manualmente aos resultados filtrados
	paginatedQuotations := s.applyManualPagination(filteredQuotations, params)

	// Criar resposta paginada
	response := &dto.PaginatedQuotationShortResponse{
		Items: make([]dto.QuotationShortResponse, 0, len(paginatedQuotations)),
		Pagination: dto.Pagination{
			TotalItems:  int64(len(filteredQuotations)),
			TotalPages:  (len(filteredQuotations) + params.PageSize - 1) / params.PageSize,
			CurrentPage: params.Page,
			PageSize:    params.PageSize,
		},
	}

	// Converter cada cotação para DTO resumido
	for i := range paginatedQuotations {
		shortResp := s.convertToQuotationShortResponse(&paginatedQuotations[i])
		response.Items = append(response.Items, *shortResp)
	}

	s.logger.Info("Pesquisa de cotações concluída com sucesso",
		zap.Int("total_results", len(filteredQuotations)))
	return response, nil
}

// quotationMatchesSearch verifica se uma cotação corresponde a um termo de pesquisa
func (s *Service) quotationMatchesSearch(quotation *models.Quotation, query string) bool {
	// Verificar correspondência em campos da cotação
	if strings.Contains(strings.ToLower(quotation.QuotationNo), query) {
		return true
	}

	if strings.Contains(strings.ToLower(quotation.Notes), query) {
		return true
	}

	if strings.Contains(strings.ToLower(quotation.Terms), query) {
		return true
	}

	// Verificar correspondência no contato
	if quotation.Contact != nil {
		if strings.Contains(strings.ToLower(quotation.Contact.Name), query) {
			return true
		}

		if strings.Contains(strings.ToLower(quotation.Contact.Email), query) {
			return true
		}
	}

	// Verificar correspondência nos itens
	for _, item := range quotation.Items {
		if strings.Contains(strings.ToLower(item.ProductName), query) {
			return true
		}

		if strings.Contains(strings.ToLower(item.ProductCode), query) {
			return true
		}

		if strings.Contains(strings.ToLower(item.Description), query) {
			return true
		}
	}

	return false
}

// createPaginationResponse cria uma estrutura de paginação comum
func createPaginationResponse(result *pagination.PaginatedResult) dto.Pagination {
	return dto.Pagination{
		TotalItems:  result.TotalItems,
		TotalPages:  result.TotalPages,
		CurrentPage: result.CurrentPage,
		PageSize:    result.PageSize,
	}
}
