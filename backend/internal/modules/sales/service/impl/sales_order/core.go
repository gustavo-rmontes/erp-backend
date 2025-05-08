package sales_order

import (
	"context"
	"fmt"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"go.uber.org/zap"
)

// Create cria um novo pedido de venda
func (s *Service) Create(ctx context.Context, orderDTO *dto.SalesOrderCreate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Criando novo pedido de venda", zap.Any("order", orderDTO))

	// Gerar número do pedido
	soNo := fmt.Sprintf(SequenceFormat, PrefixSalesOrder, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Converter DTO para modelo
	order := &models.SalesOrder{
		SONo:            soNo,
		ContactID:       orderDTO.ContactID,
		QuotationID:     orderDTO.QuotationID,
		Status:          models.SOStatusDraft,
		ExpectedDate:    orderDTO.ExpectedDate,
		PaymentTerms:    orderDTO.PaymentTerms,
		ShippingAddress: orderDTO.ShippingAddress,
		Notes:           orderDTO.Notes,
		SubTotal:        0,
		TaxTotal:        0,
		DiscountTotal:   0,
		GrandTotal:      0,
	}

	// Converter itens do DTO para itens do modelo
	for _, itemDTO := range orderDTO.Items {
		item := models.SOItem{
			ProductID:   itemDTO.ProductID,
			ProductName: itemDTO.ProductName,
			Description: itemDTO.Description,
			Quantity:    itemDTO.Quantity,
			UnitPrice:   itemDTO.UnitPrice,
			Discount:    itemDTO.Discount,
			Tax:         itemDTO.Tax,
			Total:       calculateItemTotal(itemDTO.Quantity, itemDTO.UnitPrice, itemDTO.Discount, itemDTO.Tax),
		}
		order.Items = append(order.Items, item)
	}

	// Calcular totais
	s.calculateOrderTotals(order)

	// Salvar no banco de dados
	if err := s.salesOrderRepo.CreateSalesOrder(order); err != nil {
		s.logger.Error("Erro ao criar pedido de venda", zap.Error(err))
		return nil, fmt.Errorf("falha ao criar pedido de venda: %w", err)
	}

	// Retornar resposta
	s.logger.Info("Pedido de venda criado com sucesso", zap.Int("order_id", order.ID))
	return s.convertToSalesOrderResponse(order), nil
}

// GetByID obtém um pedido de venda pelo ID
func (s *Service) GetByID(ctx context.Context, id int) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Buscando pedido de venda por ID", zap.Int("order_id", id))

	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda: %w", err)
	}

	return s.convertToSalesOrderResponse(order), nil
}

// GetShortByID obtém uma versão resumida do pedido de venda pelo ID
func (s *Service) GetShortByID(ctx context.Context, id int) (*dto.SalesOrderShortResponse, error) {
	s.logger.Info("Buscando pedido de venda resumido por ID", zap.Int("order_id", id))

	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda resumido", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda: %w", err)
	}

	return s.convertToSalesOrderShortResponse(order), nil
}

// Update atualiza um pedido de venda existente
func (s *Service) Update(ctx context.Context, id int, orderDTO *dto.SalesOrderUpdate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Atualizando pedido de venda", zap.Int("order_id", id), zap.Any("order", orderDTO))

	// Verificar se o pedido existe
	existingOrder, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda para atualização", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda para atualização: %w", err)
	}

	// Atualizar campos - conservando os que não podem ser atualizados
	order := &models.SalesOrder{
		ID:              id,
		SONo:            existingOrder.SONo,
		ContactID:       existingOrder.ContactID,
		QuotationID:     existingOrder.QuotationID,
		Status:          existingOrder.Status,
		CreatedAt:       existingOrder.CreatedAt,
		ExpectedDate:    orderDTO.ExpectedDate,
		PaymentTerms:    orderDTO.PaymentTerms,
		ShippingAddress: orderDTO.ShippingAddress,
		Notes:           orderDTO.Notes,
		SubTotal:        0,
		TaxTotal:        0,
		DiscountTotal:   0,
		GrandTotal:      0,
	}

	// Se novos itens foram fornecidos, substituir os existentes
	if orderDTO.Items != nil {
		order.Items = []models.SOItem{}
		for _, itemDTO := range orderDTO.Items {
			item := models.SOItem{
				ProductID:   itemDTO.ProductID,
				ProductName: itemDTO.ProductName,
				Description: itemDTO.Description,
				Quantity:    itemDTO.Quantity,
				UnitPrice:   itemDTO.UnitPrice,
				Discount:    itemDTO.Discount,
				Tax:         itemDTO.Tax,
				Total:       calculateItemTotal(itemDTO.Quantity, itemDTO.UnitPrice, itemDTO.Discount, itemDTO.Tax),
			}
			order.Items = append(order.Items, item)
		}
	} else {
		// Manter os itens existentes
		order.Items = existingOrder.Items
	}

	// Recalcular totais
	s.calculateOrderTotals(order)

	// Atualizar no banco de dados
	if err := s.salesOrderRepo.UpdateSalesOrder(id, order); err != nil {
		s.logger.Error("Erro ao atualizar pedido de venda", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao atualizar pedido de venda: %w", err)
	}

	// Buscar pedido atualizado
	updatedOrder, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda atualizado", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda atualizado: %w", err)
	}

	s.logger.Info("Pedido de venda atualizado com sucesso", zap.Int("order_id", id))
	return s.convertToSalesOrderResponse(updatedOrder), nil
}

// Delete exclui um pedido de venda
func (s *Service) Delete(ctx context.Context, id int) error {
	s.logger.Info("Excluindo pedido de venda", zap.Int("order_id", id))

	// Verificar se o pedido pode ser excluído (ex: não pode ter documentos derivados)
	if err := s.canDeleteOrder(id); err != nil {
		s.logger.Error("Não é possível excluir o pedido de venda", zap.Error(err), zap.Int("order_id", id))
		return err
	}

	if err := s.salesOrderRepo.DeleteSalesOrder(id); err != nil {
		s.logger.Error("Erro ao excluir pedido de venda", zap.Error(err), zap.Int("order_id", id))
		return fmt.Errorf("falha ao excluir pedido de venda: %w", err)
	}

	s.logger.Info("Pedido de venda excluído com sucesso", zap.Int("order_id", id))
	return nil
}

// canDeleteOrder verifica se um pedido de venda pode ser excluído
func (s *Service) canDeleteOrder(id int) error {
	// Verificar se existem documentos derivados (faturas, entregas, etc.)
	// Implementação depende das necessidades específicas do negócio
	return nil
}

// Find busca pedidos de venda com filtros
func (s *Service) Find(ctx context.Context, filter *dto.SalesOrderFilter, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderResponse, error) {
	s.logger.Info("Buscando pedidos de venda com filtros", zap.Any("filter", filter), zap.Any("pagination", params))

	result, err := s.applySalesOrderFilter(filter, params)
	if err != nil {
		s.logger.Error("Erro ao buscar pedidos de venda", zap.Error(err), zap.Any("filter", filter))
		return nil, fmt.Errorf("falha ao buscar pedidos de venda: %w", err)
	}

	// Extrair os pedidos do resultado paginado
	orders := s.extractSalesOrdersFromResult(result)

	// Criar resposta paginada
	response := &dto.PaginatedSalesOrderResponse{
		Items:      make([]dto.SalesOrderResponse, 0, len(orders)),
		Pagination: createPaginationResponse(result),
	}

	// Converter cada pedido para DTO
	for i := range orders {
		if i < len(orders) {
			orderResp := s.convertToSalesOrderResponse(&orders[i])
			response.Items = append(response.Items, *orderResp)
		}
	}

	s.logger.Info("Pedidos de venda recuperados com sucesso",
		zap.Int64("total_items", result.TotalItems),
		zap.Int("total_pages", result.TotalPages))
	return response, nil
}

// FindShort busca versões resumidas de pedidos de venda com filtros
func (s *Service) FindShort(ctx context.Context, filter *dto.SalesOrderFilter, params *pagination.PaginationParams) (*dto.PaginatedSalesOrderShortResponse, error) {
	s.logger.Info("Buscando pedidos de venda resumidos com filtros", zap.Any("filter", filter), zap.Any("pagination", params))

	result, err := s.applySalesOrderFilter(filter, params)
	if err != nil {
		s.logger.Error("Erro ao buscar pedidos de venda resumidos", zap.Error(err), zap.Any("filter", filter))
		return nil, fmt.Errorf("falha ao buscar pedidos de venda resumidos: %w", err)
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
		if i < len(orders) {
			shortResp := s.convertToSalesOrderShortResponse(&orders[i])
			response.Items = append(response.Items, *shortResp)
		}
	}

	s.logger.Info("Pedidos de venda resumidos recuperados com sucesso",
		zap.Int64("total_items", result.TotalItems),
		zap.Int("total_pages", result.TotalPages))
	return response, nil
}
