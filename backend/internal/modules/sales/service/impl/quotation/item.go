package quotation

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// AddItem adiciona um item a uma cotação
func (s *Service) AddItem(ctx context.Context, id int, item *dto.QuotationItemCreate) (*dto.QuotationResponse, error) {
	s.logger.Info("Adicionando item à cotação", zap.Int("quotation_id", id), zap.Any("item", item))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para adicionar item", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para adicionar item: %w", err)
	}

	// Criar novo item
	newItem := models.QuotationItem{
		QuotationID: id,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		Description: item.Description,
		Quantity:    item.Quantity,
		UnitPrice:   item.UnitPrice,
		Discount:    item.Discount,
		Tax:         item.Tax,
		Total:       calculateItemTotal(item.Quantity, item.UnitPrice, item.Discount, item.Tax),
	}

	// Adicionar item à cotação
	quotation.Items = append(quotation.Items, newItem)

	// Recalcular totais
	s.calculateQuotationTotals(quotation)

	// Atualizar cotação
	if err := s.quotationRepo.UpdateQuotation(id, quotation); err != nil {
		s.logger.Error("Erro ao atualizar cotação com novo item", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao atualizar cotação com novo item: %w", err)
	}

	// Buscar cotação atualizada
	updatedQuotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação após adicionar item", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação após adicionar item: %w", err)
	}

	s.logger.Info("Item adicionado à cotação com sucesso", zap.Int("quotation_id", id))
	return s.convertToQuotationResponse(updatedQuotation), nil
}

// UpdateItem atualiza um item de uma cotação
func (s *Service) UpdateItem(ctx context.Context, quotationID int, itemID int, item *dto.QuotationItemCreate) (*dto.QuotationResponse, error) {
	s.logger.Info("Atualizando item da cotação",
		zap.Int("quotation_id", quotationID),
		zap.Int("item_id", itemID),
		zap.Any("item", item))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para atualizar item",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao buscar cotação para atualizar item: %w", err)
	}

	// Encontrar item
	itemIndex := -1
	for i, existingItem := range quotation.Items {
		if existingItem.ID == itemID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		s.logger.Error("Item não encontrado na cotação",
			zap.Int("quotation_id", quotationID),
			zap.Int("item_id", itemID))
		return nil, fmt.Errorf("item não encontrado na cotação")
	}

	// Atualizar item
	quotation.Items[itemIndex].ProductID = item.ProductID
	quotation.Items[itemIndex].ProductName = item.ProductName
	quotation.Items[itemIndex].Description = item.Description
	quotation.Items[itemIndex].Quantity = item.Quantity
	quotation.Items[itemIndex].UnitPrice = item.UnitPrice
	quotation.Items[itemIndex].Discount = item.Discount
	quotation.Items[itemIndex].Tax = item.Tax
	quotation.Items[itemIndex].Total = calculateItemTotal(item.Quantity, item.UnitPrice, item.Discount, item.Tax)

	// Recalcular totais
	s.calculateQuotationTotals(quotation)

	// Atualizar cotação
	if err := s.quotationRepo.UpdateQuotation(quotationID, quotation); err != nil {
		s.logger.Error("Erro ao atualizar cotação com item modificado",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao atualizar cotação com item modificado: %w", err)
	}

	// Buscar cotação atualizada
	updatedQuotation, err := s.quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação após atualizar item",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao buscar cotação após atualizar item: %w", err)
	}

	s.logger.Info("Item da cotação atualizado com sucesso",
		zap.Int("quotation_id", quotationID),
		zap.Int("item_id", itemID))
	return s.convertToQuotationResponse(updatedQuotation), nil
}

// RemoveItem remove um item de uma cotação
func (s *Service) RemoveItem(ctx context.Context, quotationID int, itemID int) (*dto.QuotationResponse, error) {
	s.logger.Info("Removendo item da cotação", zap.Int("quotation_id", quotationID), zap.Int("item_id", itemID))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para remover item",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao buscar cotação para remover item: %w", err)
	}

	// Encontrar item
	itemIndex := -1
	for i, existingItem := range quotation.Items {
		if existingItem.ID == itemID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		s.logger.Error("Item não encontrado na cotação",
			zap.Int("quotation_id", quotationID),
			zap.Int("item_id", itemID))
		return nil, fmt.Errorf("item não encontrado na cotação")
	}

	// Remover item
	quotation.Items = append(quotation.Items[:itemIndex], quotation.Items[itemIndex+1:]...)

	// Recalcular totais
	s.calculateQuotationTotals(quotation)

	// Atualizar cotação
	if err := s.quotationRepo.UpdateQuotation(quotationID, quotation); err != nil {
		s.logger.Error("Erro ao atualizar cotação após remover item",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao atualizar cotação após remover item: %w", err)
	}

	// Buscar cotação atualizada
	updatedQuotation, err := s.quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação após remover item",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao buscar cotação após remover item: %w", err)
	}

	s.logger.Info("Item removido da cotação com sucesso",
		zap.Int("quotation_id", quotationID),
		zap.Int("item_id", itemID))
	return s.convertToQuotationResponse(updatedQuotation), nil
}
