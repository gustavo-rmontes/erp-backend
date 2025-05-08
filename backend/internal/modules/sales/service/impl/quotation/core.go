package quotation

import (
	"context"
	"fmt"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"ERP-ONSMART/backend/internal/utils/pagination"

	"go.uber.org/zap"
)

// Create cria uma nova cotação
func (s *Service) Create(ctx context.Context, quotationDTO *dto.QuotationCreate) (*dto.QuotationResponse, error) {
	s.logger.Info("Criando nova cotação", zap.Any("quotation", quotationDTO))

	// Gerar número da cotação
	quotationNo := fmt.Sprintf(SequenceFormat, PrefixQuotation, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Converter DTO para modelo
	quotation := &models.Quotation{
		QuotationNo:   quotationNo,
		ContactID:     quotationDTO.ContactID,
		Status:        models.QuotationStatusDraft,
		ExpiryDate:    quotationDTO.ExpiryDate,
		SubTotal:      0,
		TaxTotal:      0,
		DiscountTotal: 0,
		GrandTotal:    0,
		Notes:         quotationDTO.Notes,
		Terms:         quotationDTO.Terms,
	}

	// Converter itens do DTO para itens do modelo
	for _, itemDTO := range quotationDTO.Items {
		item := models.QuotationItem{
			ProductID:   itemDTO.ProductID,
			ProductName: itemDTO.ProductName,
			Description: itemDTO.Description,
			Quantity:    itemDTO.Quantity,
			UnitPrice:   itemDTO.UnitPrice,
			Discount:    itemDTO.Discount,
			Tax:         itemDTO.Tax,
			Total:       calculateItemTotal(itemDTO.Quantity, itemDTO.UnitPrice, itemDTO.Discount, itemDTO.Tax),
		}
		quotation.Items = append(quotation.Items, item)
	}

	// Calcular totais
	s.calculateQuotationTotals(quotation)

	// Salvar no banco de dados
	if err := s.quotationRepo.CreateQuotation(quotation); err != nil {
		s.logger.Error("Erro ao criar cotação", zap.Error(err))
		return nil, fmt.Errorf("falha ao criar cotação: %w", err)
	}

	// Retornar resposta
	s.logger.Info("Cotação criada com sucesso", zap.Int("quotation_id", quotation.ID))
	return s.convertToQuotationResponse(quotation), nil
}

// GetByID obtém uma cotação pelo ID
func (s *Service) GetByID(ctx context.Context, id int) (*dto.QuotationResponse, error) {
	s.logger.Info("Buscando cotação por ID", zap.Int("quotation_id", id))

	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação: %w", err)
	}

	return s.convertToQuotationResponse(quotation), nil
}

// GetShortByID obtém uma versão resumida da cotação pelo ID
func (s *Service) GetShortByID(ctx context.Context, id int) (*dto.QuotationShortResponse, error) {
	s.logger.Info("Buscando cotação resumida por ID", zap.Int("quotation_id", id))

	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação resumida", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação: %w", err)
	}

	return s.convertToQuotationShortResponse(quotation), nil
}

// Update atualiza uma cotação existente
func (s *Service) Update(ctx context.Context, id int, quotationDTO *dto.QuotationUpdate) (*dto.QuotationResponse, error) {
	s.logger.Info("Atualizando cotação", zap.Int("quotation_id", id), zap.Any("quotation", quotationDTO))

	// Verificar se a cotação existe
	existingQuotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para atualização", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para atualização: %w", err)
	}

	// Atualizar campos - conservando os que não podem ser atualizados
	quotation := &models.Quotation{
		ID:            id,
		QuotationNo:   existingQuotation.QuotationNo,
		ContactID:     existingQuotation.ContactID,
		Status:        existingQuotation.Status,
		ExpiryDate:    quotationDTO.ExpiryDate,
		SubTotal:      0,
		TaxTotal:      0,
		DiscountTotal: 0,
		GrandTotal:    0,
		Notes:         quotationDTO.Notes,
		Terms:         quotationDTO.Terms,
		CreatedAt:     existingQuotation.CreatedAt,
	}

	// Converter itens do DTO para itens do modelo
	for _, itemDTO := range quotationDTO.Items {
		item := models.QuotationItem{
			ProductID:   itemDTO.ProductID,
			ProductName: itemDTO.ProductName,
			Description: itemDTO.Description,
			Quantity:    itemDTO.Quantity,
			UnitPrice:   itemDTO.UnitPrice,
			Discount:    itemDTO.Discount,
			Tax:         itemDTO.Tax,
			Total:       calculateItemTotal(itemDTO.Quantity, itemDTO.UnitPrice, itemDTO.Discount, itemDTO.Tax),
		}
		quotation.Items = append(quotation.Items, item)
	}

	// Recalcular totais
	s.calculateQuotationTotals(quotation)

	// Atualizar no banco de dados
	if err := s.quotationRepo.UpdateQuotation(id, quotation); err != nil {
		s.logger.Error("Erro ao atualizar cotação", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao atualizar cotação: %w", err)
	}

	// Buscar cotação atualizada
	updatedQuotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação atualizada", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação atualizada: %w", err)
	}

	s.logger.Info("Cotação atualizada com sucesso", zap.Int("quotation_id", id))
	return s.convertToQuotationResponse(updatedQuotation), nil
}

// Delete exclui uma cotação
func (s *Service) Delete(ctx context.Context, id int) error {
	s.logger.Info("Excluindo cotação", zap.Int("quotation_id", id))

	if err := s.quotationRepo.DeleteQuotation(id); err != nil {
		s.logger.Error("Erro ao excluir cotação", zap.Error(err), zap.Int("quotation_id", id))
		return fmt.Errorf("falha ao excluir cotação: %w", err)
	}

	s.logger.Info("Cotação excluída com sucesso", zap.Int("quotation_id", id))
	return nil
}

// Find busca cotações com filtros
func (s *Service) Find(ctx context.Context, filter *dto.QuotationFilter, params *pagination.PaginationParams) (*dto.PaginatedQuotationResponse, error) {
	s.logger.Info("Buscando cotações com filtros", zap.Any("filter", filter), zap.Any("pagination", params))

	result, err := s.applyQuotationFilter(filter, params)
	if err != nil {
		s.logger.Error("Erro ao buscar cotações", zap.Error(err), zap.Any("filter", filter))
		return nil, fmt.Errorf("falha ao buscar cotações: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Criar resposta paginada
	response := &dto.PaginatedQuotationResponse{
		Items:      make([]dto.QuotationResponse, 0, len(quotations)),
		Pagination: createPaginationResponse(result),
	}

	// Converter cada cotação para DTO
	for i := range quotations {
		if i < len(quotations) {
			quotationResp := s.convertToQuotationResponse(&quotations[i])
			response.Items = append(response.Items, *quotationResp)
		}
	}

	s.logger.Info("Cotações recuperadas com sucesso",
		zap.Int64("total_items", result.TotalItems),
		zap.Int("total_pages", result.TotalPages))
	return response, nil
}

// FindShort busca versões resumidas de cotações com filtros
func (s *Service) FindShort(ctx context.Context, filter *dto.QuotationFilter, params *pagination.PaginationParams) (*dto.PaginatedQuotationShortResponse, error) {
	s.logger.Info("Buscando cotações resumidas com filtros", zap.Any("filter", filter), zap.Any("pagination", params))

	result, err := s.applyQuotationFilter(filter, params)
	if err != nil {
		s.logger.Error("Erro ao buscar cotações resumidas", zap.Error(err), zap.Any("filter", filter))
		return nil, fmt.Errorf("falha ao buscar cotações resumidas: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Criar resposta paginada
	response := &dto.PaginatedQuotationShortResponse{
		Items:      make([]dto.QuotationShortResponse, 0, len(quotations)),
		Pagination: createPaginationResponse(result),
	}

	// Converter cada cotação para DTO resumido
	for i := range quotations {
		if i < len(quotations) {
			shortResp := s.convertToQuotationShortResponse(&quotations[i])
			response.Items = append(response.Items, *shortResp)
		}
	}

	s.logger.Info("Cotações resumidas recuperadas com sucesso",
		zap.Int64("total_items", result.TotalItems),
		zap.Int("total_pages", result.TotalPages))
	return response, nil
}
