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

// UpdateStatus atualiza o status de uma cotação
func (s *Service) UpdateStatus(ctx context.Context, id int, req *dto.QuotationStatusUpdateRequest) (*dto.QuotationResponse, error) {
	s.logger.Info("Atualizando status da cotação", zap.Int("quotation_id", id), zap.String("new_status", req.Status))

	// Verificar se a cotação existe
	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para atualização de status", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para atualização de status: %w", err)
	}

	// Validar transição de status
	if !s.isValidStatusTransition(quotation.Status, req.Status) {
		s.logger.Error("Transição de status inválida",
			zap.String("current_status", quotation.Status),
			zap.String("requested_status", req.Status))
		return nil, fmt.Errorf("transição de status inválida: %s -> %s", quotation.Status, req.Status)
	}

	// Atualizar status
	quotation.Status = req.Status

	// Atualizar motivo de rejeição se for o caso
	if req.Status == models.QuotationStatusRejected && req.Reason != "" {
		quotation.Notes = fmt.Sprintf(RejectionNoteFormat, quotation.Notes, req.Reason)
	}

	// Salvar alterações
	if err := s.quotationRepo.UpdateQuotation(id, quotation); err != nil {
		s.logger.Error("Erro ao atualizar status da cotação", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao atualizar status da cotação: %w", err)
	}

	// Buscar cotação atualizada
	updatedQuotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação após atualização de status", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação após atualização de status: %w", err)
	}

	s.logger.Info("Status da cotação atualizado com sucesso",
		zap.Int("quotation_id", id),
		zap.String("new_status", req.Status))
	return s.convertToQuotationResponse(updatedQuotation), nil
}

// ProcessExpirations processa cotações expiradas
func (s *Service) ProcessExpirations(ctx context.Context) (int, error) {
	s.logger.Info("Processando cotações expiradas")

	// Obter cotações expiradas
	result, err := s.quotationRepo.GetExpiredQuotations(&pagination.PaginationParams{
		Page:     FirstPage,
		PageSize: DefaultBatchSize,
	})

	if err != nil {
		s.logger.Error("Erro ao buscar cotações expiradas", zap.Error(err))
		return 0, fmt.Errorf("falha ao buscar cotações expiradas: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Atualizar status das cotações expiradas
	count := 0
	for _, quotation := range quotations {
		if quotation.Status != models.QuotationStatusExpired &&
			quotation.Status != models.QuotationStatusAccepted &&
			quotation.Status != models.QuotationStatusRejected {

			quotation.Status = models.QuotationStatusExpired
			if err := s.quotationRepo.UpdateQuotation(quotation.ID, &quotation); err != nil {
				s.logger.Error("Erro ao atualizar status da cotação expirada",
					zap.Error(err),
					zap.Int("quotation_id", quotation.ID))
				continue
			}
			count++
		}
	}

	s.logger.Info("Processamento de cotações expiradas concluído", zap.Int("updated_count", count))
	return count, nil
}

// NotifyExpiringQuotations notifica sobre cotações prestes a expirar
func (s *Service) NotifyExpiringQuotations(ctx context.Context, daysBeforeExpiry int) (int, error) {
	s.logger.Info("Verificando cotações prestes a expirar", zap.Int("days_before_expiry", daysBeforeExpiry))

	// Obter todas as cotações
	result, err := s.quotationRepo.GetAllQuotations(&pagination.PaginationParams{
		Page:     FirstPage,
		PageSize: DefaultBatchSize,
	})

	if err != nil {
		s.logger.Error("Erro ao buscar cotações para verificação de expiração", zap.Error(err))
		return 0, fmt.Errorf("falha ao buscar cotações para verificação de expiração: %w", err)
	}

	// Extrair as cotações do resultado paginado
	quotations := s.extractQuotationsFromResult(result)

	// Data limite
	limitDate := time.Now().AddDate(0, 0, daysBeforeExpiry)

	// Identificar cotações prestes a expirar
	count := 0
	for _, quotation := range quotations {
		// Verificar apenas cotações no status "enviada"
		if quotation.Status == models.QuotationStatusSent {
			// Verificar se a data de expiração está próxima
			if quotation.ExpiryDate.After(time.Now()) && quotation.ExpiryDate.Before(limitDate) {
				// Aqui implementaríamos a lógica de notificação
				s.logger.Info("Cotação prestes a expirar",
					zap.Int("quotation_id", quotation.ID),
					zap.Time("expiry_date", quotation.ExpiryDate))
				count++
			}
		}
	}

	s.logger.Info("Verificação de cotações prestes a expirar concluída",
		zap.Int("expiring_count", count))
	return count, nil
}

// isValidStatusTransition verifica se uma transição de status é válida
func (s *Service) isValidStatusTransition(currentStatus string, newStatus string) bool {
	// Definir transições de status válidas
	validTransitions := map[string][]string{
		models.QuotationStatusDraft: {
			models.QuotationStatusSent,
			models.QuotationStatusCancelled,
		},
		models.QuotationStatusSent: {
			models.QuotationStatusAccepted,
			models.QuotationStatusRejected,
			models.QuotationStatusExpired,
			models.QuotationStatusCancelled,
		},
		models.QuotationStatusExpired: {
			models.QuotationStatusCancelled,
		},
		models.QuotationStatusRejected: {
			models.QuotationStatusCancelled,
		},
		models.QuotationStatusAccepted: {
			models.QuotationStatusCancelled,
		},
		models.QuotationStatusCancelled: {},
	}

	// Se o status atual e o novo são iguais, é válido
	if currentStatus == newStatus {
		return true
	}

	// Verificar se a transição é válida
	for _, validStatus := range validTransitions[currentStatus] {
		if validStatus == newStatus {
			return true
		}
	}

	return false
}
