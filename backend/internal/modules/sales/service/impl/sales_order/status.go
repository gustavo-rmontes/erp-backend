package sales_order

import (
	"context"
	"fmt"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// UpdateStatus atualiza o status de um pedido de venda
func (s *Service) UpdateStatus(ctx context.Context, id int, req *dto.SalesOrderStatusUpdateRequest) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Atualizando status do pedido de venda", zap.Int("order_id", id), zap.String("new_status", req.Status))

	// Verificar se o pedido existe
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para atualização de status", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido para atualização de status: %w", err)
	}

	// Validar transição de status
	if !s.isValidStatusTransition(order.Status, req.Status) {
		s.logger.Error("Transição de status inválida",
			zap.String("current_status", order.Status),
			zap.String("requested_status", req.Status))
		return nil, fmt.Errorf("transição de status inválida: %s -> %s", order.Status, req.Status)
	}

	// Atualizar status
	order.Status = req.Status

	// Atualizar motivo de cancelamento se for o caso
	if req.Status == models.SOStatusCancelled && req.Reason != "" {
		order.Notes = fmt.Sprintf("%s\n\nCancellation reason: %s", order.Notes, req.Reason)
	}

	// Salvar alterações
	if err := s.salesOrderRepo.UpdateSalesOrder(id, order); err != nil {
		s.logger.Error("Erro ao atualizar status do pedido", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao atualizar status do pedido: %w", err)
	}

	// Buscar pedido atualizado
	updatedOrder, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido após atualização de status", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido após atualização de status: %w", err)
	}

	s.logger.Info("Status do pedido atualizado com sucesso",
		zap.Int("order_id", id),
		zap.String("new_status", req.Status))
	return s.convertToSalesOrderResponse(updatedOrder), nil
}

// isValidStatusTransition verifica se uma transição de status é válida
func (s *Service) isValidStatusTransition(currentStatus string, newStatus string) bool {
	// Definir transições de status válidas
	validTransitions := map[string][]string{
		models.SOStatusDraft: {
			models.SOStatusConfirmed,
			models.SOStatusCancelled,
		},
		models.SOStatusConfirmed: {
			models.SOStatusProcessing,
			models.SOStatusCancelled,
		},
		models.SOStatusProcessing: {
			models.SOStatusCompleted,
			models.SOStatusCancelled,
		},
		models.SOStatusCompleted: {},
		models.SOStatusCancelled: {},
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
