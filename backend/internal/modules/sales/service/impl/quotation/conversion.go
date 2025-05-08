package quotation

import (
	"context"
	"fmt"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// ConvertToSalesOrder converte uma cotação em um pedido de venda
func (s *Service) ConvertToSalesOrder(ctx context.Context, id int, data *dto.SalesOrderFromQuotationCreate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Convertendo cotação em pedido de venda", zap.Int("quotation_id", id))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para conversão", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para conversão: %w", err)
	}

	// Verificar status da cotação
	if quotation.Status != models.QuotationStatusSent &&
		quotation.Status != models.QuotationStatusDraft {
		s.logger.Error("Cotação em status inválido para conversão",
			zap.String("status", quotation.Status),
			zap.Int("quotation_id", id))
		return nil, fmt.Errorf("cotação em status inválido para conversão: %s", quotation.Status)
	}

	// Gerar número de SO
	soNumber := fmt.Sprintf(SequenceFormat, PrefixSalesOrder, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar pedido de venda a partir da cotação
	salesOrder := &models.SalesOrder{
		SONo:            soNumber,
		QuotationID:     quotation.ID,
		ContactID:       quotation.ContactID,
		Status:          models.SOStatusDraft,
		ExpectedDate:    data.ExpectedDate,
		SubTotal:        quotation.SubTotal,
		TaxTotal:        quotation.TaxTotal,
		DiscountTotal:   quotation.DiscountTotal,
		GrandTotal:      quotation.GrandTotal,
		Notes:           quotation.Notes,
		PaymentTerms:    data.PaymentTerms,
		ShippingAddress: data.ShippingAddress,
	}

	// Copiar itens da cotação para o pedido de venda
	for _, quoteItem := range quotation.Items {
		soItem := models.SOItem{
			ProductID:   quoteItem.ProductID,
			ProductName: quoteItem.ProductName,
			ProductCode: quoteItem.ProductCode,
			Description: quoteItem.Description,
			Quantity:    quoteItem.Quantity,
			UnitPrice:   quoteItem.UnitPrice,
			Discount:    quoteItem.Discount,
			Tax:         quoteItem.Tax,
			Total:       quoteItem.Total,
		}
		salesOrder.Items = append(salesOrder.Items, soItem)
	}

	// Salvar pedido de venda
	if err := s.salesOrderRepo.CreateSalesOrder(salesOrder); err != nil {
		s.logger.Error("Erro ao criar pedido de venda a partir da cotação",
			zap.Error(err),
			zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao criar pedido de venda a partir da cotação: %w", err)
	}

	// Atualizar status da cotação se necessário
	updateStatus := UpdateQuotationOnConversion
	if updateStatus {
		quotation.Status = models.QuotationStatusAccepted
		if err := s.quotationRepo.UpdateQuotation(id, quotation); err != nil {
			s.logger.Warn("Erro ao atualizar status da cotação após conversão",
				zap.Error(err),
				zap.Int("quotation_id", id))
		}
	}

	// Obter ID do processo de vendas se necessário
	salesProcessID := 0 // Valor padrão ou obter de algum lugar adequado

	// Se existir um processo de vendas, vincular a cotação e o pedido
	if salesProcessID > 0 {
		s.logger.Info("Vinculando cotação e pedido ao processo de vendas",
			zap.Int("sales_process_id", salesProcessID),
			zap.Int("quotation_id", id),
			zap.Int("sales_order_id", salesOrder.ID))

		// Vincular cotação ao processo
		if err := s.salesProcessRepo.LinkQuotationToProcess(salesProcessID, id); err != nil {
			s.logger.Warn("Erro ao vincular cotação ao processo de vendas",
				zap.Error(err),
				zap.Int("sales_process_id", salesProcessID),
				zap.Int("quotation_id", id))
		}

		// Vincular pedido de venda ao processo
		if err := s.salesProcessRepo.LinkSalesOrderToProcess(salesProcessID, salesOrder.ID); err != nil {
			s.logger.Warn("Erro ao vincular pedido de venda ao processo de vendas",
				zap.Error(err),
				zap.Int("sales_process_id", salesProcessID),
				zap.Int("sales_order_id", salesOrder.ID))
		}
	}

	s.logger.Info("Cotação convertida em pedido de venda com sucesso",
		zap.Int("quotation_id", id),
		zap.Int("sales_order_id", salesOrder.ID))
	return s.convertToSalesOrderResponse(salesOrder), nil
}

// Clone clona uma cotação
func (s *Service) Clone(ctx context.Context, id int, options *dto.QuotationCloneOptions) (*dto.QuotationResponse, error) {
	s.logger.Info("Clonando cotação", zap.Int("quotation_id", id), zap.Any("options", options))

	// Buscar cotação original
	original, err := s.quotationRepo.GetQuotationByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para clonagem", zap.Error(err), zap.Int("quotation_id", id))
		return nil, fmt.Errorf("falha ao buscar cotação para clonagem: %w", err)
	}

	// Gerar número para nova cotação
	newQuotationNo := fmt.Sprintf("QT-%s-%04d", time.Now().Format("20060102"), s.generateSequence())

	// Definir nova data de expiração
	newExpiryDate := time.Now().AddDate(0, DefaultExpiryMonths, 0)

	if !options.ExpiryDate.IsZero() {
		newExpiryDate = options.ExpiryDate
	}

	// Criar nova cotação baseada na original
	clone := &models.Quotation{
		QuotationNo:   newQuotationNo,
		ContactID:     original.ContactID,
		Status:        models.QuotationStatusDraft,
		ExpiryDate:    newExpiryDate,
		SubTotal:      original.SubTotal,
		TaxTotal:      original.TaxTotal,
		DiscountTotal: original.DiscountTotal,
		GrandTotal:    original.GrandTotal,
		Notes:         original.Notes,
		Terms:         original.Terms,
	}

	// Aplicar opções de clonagem
	if options.ContactID > 0 {
		clone.ContactID = options.ContactID
	}

	if options.CopyNotes {
		clone.Notes = original.Notes
	}

	// Copiar itens
	for _, item := range original.Items {
		newItem := models.QuotationItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Discount:    item.Discount,
			Tax:         item.Tax,
			Total:       item.Total,
		}

		// Aplicar ajuste de preço se necessário
		if options.AdjustPrices && options.PriceAdjustment != 0 {
			newItem.UnitPrice *= (1 + options.PriceAdjustment/100)
			newItem.Total = calculateItemTotal(newItem.Quantity, newItem.UnitPrice, newItem.Discount, newItem.Tax)
		}

		clone.Items = append(clone.Items, newItem)
	}

	// Recalcular totais
	s.calculateQuotationTotals(clone)

	// Salvar nova cotação
	if err := s.quotationRepo.CreateQuotation(clone); err != nil {
		s.logger.Error("Erro ao criar cotação clonada", zap.Error(err), zap.Int("original_id", id))
		return nil, fmt.Errorf("falha ao criar cotação clonada: %w", err)
	}

	s.logger.Info("Cotação clonada com sucesso",
		zap.Int("original_id", id),
		zap.Int("clone_id", clone.ID))
	return s.convertToQuotationResponse(clone), nil
}
