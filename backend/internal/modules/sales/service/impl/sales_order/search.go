// search.go (com funções obsoletas removidas)
package sales_order

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"go.uber.org/zap"
)

// extractSalesOrdersFromResult extrai pedidos de venda do resultado paginado
func (s *Service) extractSalesOrdersFromResult(result *pagination.PaginatedResult) []models.SalesOrder {
	orders := make([]models.SalesOrder, 0)
	if items, ok := result.Items.([]models.SalesOrder); ok {
		orders = items
	}
	return orders
}

// applySalesOrderFilter aplica filtros comuns e retorna o resultado paginado
func (s *Service) applySalesOrderFilter(filter *dto.SalesOrderFilter, params *pagination.PaginationParams) (*pagination.PaginatedResult, error) {
	var result *pagination.PaginatedResult
	var err error

	// Aplicar filtros de acordo com os critérios
	if filter == nil {
		result, err = s.salesOrderRepo.GetAllSalesOrders(params)
	} else if filter.ContactID > 0 {
		result, err = s.salesOrderRepo.GetSalesOrdersByContact(filter.ContactID, params)
	} else if filter.QuotationID > 0 {
		order, err := s.salesOrderRepo.GetSalesOrdersByQuotation(filter.QuotationID)
		if err != nil {
			return nil, err
		}
		// Criar um resultado paginado manualmente com o pedido encontrado
		if order != nil {
			result = &pagination.PaginatedResult{
				Items:       []models.SalesOrder{*order},
				TotalItems:  1,
				TotalPages:  1,
				CurrentPage: params.Page,
				PageSize:    params.PageSize,
			}
		} else {
			result = &pagination.PaginatedResult{
				Items:       []models.SalesOrder{},
				TotalItems:  0,
				TotalPages:  0,
				CurrentPage: params.Page,
				PageSize:    params.PageSize,
			}
		}
		err = nil
	} else if len(filter.Status) > 0 && filter.Status[0] != "" {
		// Pegar apenas o primeiro status da lista para simplificar
		result, err = s.salesOrderRepo.GetSalesOrdersByStatus(filter.Status[0], params)
	} else if filter.HasDelivery != nil {
		// Usar o novo método implementado
		result, err = s.salesOrderRepo.GetSalesOrdersWithDeliveries(*filter.HasDelivery, params)
	} else if filter.HasInvoice != nil {
		// Usar o novo método implementado
		result, err = s.salesOrderRepo.GetSalesOrdersWithInvoices(*filter.HasInvoice, params)
	} else {
		result, err = s.salesOrderRepo.GetAllSalesOrders(params)
	}

	return result, err
}

// Search busca pedidos de venda por termo de pesquisa
func (s *Service) Search(ctx context.Context, query string, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderShortResponse, error) {
	s.logger.Info("Pesquisando pedidos de venda", zap.String("query", query), zap.Any("pagination", params))

	// Usar o novo método de busca no banco de dados em vez da filtragem em memória
	result, err := s.salesOrderRepo.SearchSalesOrders(query, params)
	if err != nil {
		s.logger.Error("Erro ao pesquisar pedidos de venda", zap.Error(err), zap.String("query", query))
		return nil, fmt.Errorf("falha ao pesquisar pedidos de venda: %w", err)
	}

	// Extrair os pedidos do resultado paginado
	orders := s.extractSalesOrdersFromResult(result)

	// Criar resposta paginada
	response := &dto.PaginatedSalesOrderShortResponse{
		Items:      make([]dto.SalesOrderShortResponse, 0, len(orders)),
		Pagination: createPaginationResponse(result),
	}

	// Converter cada pedido para DTO resumido
	for i := range orders {
		shortResp := s.convertToSalesOrderShortResponse(&orders[i])
		response.Items = append(response.Items, *shortResp)
	}

	s.logger.Info("Pesquisa de pedidos de venda concluída com sucesso",
		zap.Int64("total_results", result.TotalItems))
	return response, nil
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
