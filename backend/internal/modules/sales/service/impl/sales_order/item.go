package sales_order

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// AddItem adiciona um item a um pedido de venda
func (s *Service) AddItem(ctx context.Context, id int, item *dto.SalesOrderItemCreate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Adicionando item ao pedido", zap.Int("order_id", id), zap.Any("item", item))

	// Buscar pedido
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para adicionar item", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido para adicionar item: %w", err)
	}

	// Verificar se o pedido pode ser alterado
	if order.Status != models.SOStatusDraft && order.Status != models.SOStatusConfirmed {
		s.logger.Error("Pedido em status que não permite adição de itens",
			zap.String("status", order.Status),
			zap.Int("order_id", id))
		return nil, fmt.Errorf("não é possível adicionar itens a um pedido com status %s", order.Status)
	}

	// Criar novo item
	newItem := models.SOItem{
		// OrderID:     id,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Discount:    item.Discount,
		Tax:         item.Tax,
		Total:       calculateItemTotal(item.Quantity, item.UnitPrice, item.Discount, item.Tax),
	}

	// Adicionar item ao pedido
	order.Items = append(order.Items, newItem)

	// Recalcular totais
	s.calculateOrderTotals(order)

	// Atualizar pedido
	if err := s.salesOrderRepo.UpdateSalesOrder(id, order); err != nil {
		s.logger.Error("Erro ao atualizar pedido com novo item", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao atualizar pedido com novo item: %w", err)
	}

	// Buscar pedido atualizado
	updatedOrder, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido após adicionar item", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido após adicionar item: %w", err)
	}

	s.logger.Info("Item adicionado ao pedido com sucesso", zap.Int("order_id", id))
	return s.convertToSalesOrderResponse(updatedOrder), nil
}

// UpdateItem atualiza um item de um pedido de venda
func (s *Service) UpdateItem(ctx context.Context, orderID int, itemID int, item *dto.SalesOrderItemCreate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Atualizando item do pedido",
		zap.Int("order_id", orderID),
		zap.Int("item_id", itemID),
		zap.Any("item", item))

	// Buscar pedido
	order, err := s.salesOrderRepo.GetSalesOrderByID(orderID)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para atualizar item",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao buscar pedido para atualizar item: %w", err)
	}

	// Verificar se o pedido pode ser alterado
	if order.Status != models.SOStatusDraft && order.Status != models.SOStatusConfirmed {
		s.logger.Error("Pedido em status que não permite atualização de itens",
			zap.String("status", order.Status),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("não é possível atualizar itens de um pedido com status %s", order.Status)
	}

	// Encontrar item
	itemIndex := -1
	for i, existingItem := range order.Items {
		if existingItem.ID == itemID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		s.logger.Error("Item não encontrado no pedido",
			zap.Int("order_id", orderID),
			zap.Int("item_id", itemID))
		return nil, fmt.Errorf("item não encontrado no pedido")
	}

	// Atualizar item
	order.Items[itemIndex].ProductID = item.ProductID
	order.Items[itemIndex].ProductName = item.ProductName
	order.Items[itemIndex].Description = item.Description
	order.Items[itemIndex].Quantity = item.Quantity
	order.Items[itemIndex].UnitPrice = item.UnitPrice
	order.Items[itemIndex].Discount = item.Discount
	order.Items[itemIndex].Tax = item.Tax
	order.Items[itemIndex].Total = calculateItemTotal(item.Quantity, item.UnitPrice, item.Discount, item.Tax)

	// Recalcular totais
	s.calculateOrderTotals(order)

	// Atualizar pedido
	if err := s.salesOrderRepo.UpdateSalesOrder(orderID, order); err != nil {
		s.logger.Error("Erro ao atualizar pedido com item modificado",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao atualizar pedido com item modificado: %w", err)
	}

	// Buscar pedido atualizado
	updatedOrder, err := s.salesOrderRepo.GetSalesOrderByID(orderID)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido após atualizar item",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao buscar pedido após atualizar item: %w", err)
	}

	s.logger.Info("Item do pedido atualizado com sucesso",
		zap.Int("order_id", orderID),
		zap.Int("item_id", itemID))
	return s.convertToSalesOrderResponse(updatedOrder), nil
}

// RemoveItem remove um item de um pedido de venda
func (s *Service) RemoveItem(ctx context.Context, orderID int, itemID int) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Removendo item do pedido", zap.Int("order_id", orderID), zap.Int("item_id", itemID))

	// Buscar pedido
	order, err := s.salesOrderRepo.GetSalesOrderByID(orderID)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para remover item",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao buscar pedido para remover item: %w", err)
	}

	// Verificar se o pedido pode ser alterado
	if order.Status != models.SOStatusDraft && order.Status != models.SOStatusConfirmed {
		s.logger.Error("Pedido em status que não permite remoção de itens",
			zap.String("status", order.Status),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("não é possível remover itens de um pedido com status %s", order.Status)
	}

	// Encontrar item
	itemIndex := -1
	for i, existingItem := range order.Items {
		if existingItem.ID == itemID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		s.logger.Error("Item não encontrado no pedido",
			zap.Int("order_id", orderID),
			zap.Int("item_id", itemID))
		return nil, fmt.Errorf("item não encontrado no pedido")
	}

	// Remover item
	order.Items = append(order.Items[:itemIndex], order.Items[itemIndex+1:]...)

	// Recalcular totais
	s.calculateOrderTotals(order)

	// Atualizar pedido
	if err := s.salesOrderRepo.UpdateSalesOrder(orderID, order); err != nil {
		s.logger.Error("Erro ao atualizar pedido após remover item",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao atualizar pedido após remover item: %w", err)
	}

	// Buscar pedido atualizado
	updatedOrder, err := s.salesOrderRepo.GetSalesOrderByID(orderID)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido após remover item",
			zap.Error(err),
			zap.Int("order_id", orderID))
		return nil, fmt.Errorf("falha ao buscar pedido após remover item: %w", err)
	}

	s.logger.Info("Item removido do pedido com sucesso",
		zap.Int("order_id", orderID),
		zap.Int("item_id", itemID))
	return s.convertToSalesOrderResponse(updatedOrder), nil
}
