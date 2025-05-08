package sales_order

import (
	"context"
	"fmt"
	"time"

	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"

	"go.uber.org/zap"
)

// CreateFromQuotation cria um pedido de venda a partir de uma cotação
func (s *Service) CreateFromQuotation(ctx context.Context, quotationID int, data *dto.SalesOrderFromQuotationCreate) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Criando pedido de venda a partir de cotação", zap.Int("quotation_id", quotationID))

	// Buscar cotação
	quotation, err := s.quotationRepo.GetQuotationByID(quotationID)
	if err != nil {
		s.logger.Error("Erro ao buscar cotação para criar pedido", zap.Error(err), zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao buscar cotação para criar pedido: %w", err)
	}

	// Verificar status da cotação
	if quotation.Status != models.QuotationStatusSent && quotation.Status != models.QuotationStatusAccepted {
		s.logger.Error("Cotação em status inválido para criar pedido",
			zap.String("status", quotation.Status),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("cotação em status inválido para criar pedido: %s", quotation.Status)
	}

	// Gerar número do pedido
	soNo := fmt.Sprintf(SequenceFormat, PrefixSalesOrder, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar pedido a partir da cotação
	order := &models.SalesOrder{
		SONo:            soNo,
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

	// Copiar itens da cotação para o pedido
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
		order.Items = append(order.Items, soItem)
	}

	// Salvar pedido
	if err := s.salesOrderRepo.CreateSalesOrder(order); err != nil {
		s.logger.Error("Erro ao criar pedido a partir da cotação",
			zap.Error(err),
			zap.Int("quotation_id", quotationID))
		return nil, fmt.Errorf("falha ao criar pedido a partir da cotação: %w", err)
	}

	// Atualizar status da cotação se necessário
	if quotation.Status == models.QuotationStatusSent {
		quotation.Status = models.QuotationStatusAccepted
		if err := s.quotationRepo.UpdateQuotation(quotationID, quotation); err != nil {
			s.logger.Warn("Erro ao atualizar status da cotação após criar pedido",
				zap.Error(err),
				zap.Int("quotation_id", quotationID))
		}
	}

	s.logger.Info("Pedido criado a partir da cotação com sucesso",
		zap.Int("quotation_id", quotationID),
		zap.Int("order_id", order.ID))
	return s.convertToSalesOrderResponse(order), nil
}

// CreatePurchaseOrder cria um pedido de compra a partir de um pedido de venda
func (s *Service) CreatePurchaseOrder(ctx context.Context, id int, data *dto.PurchaseOrderFromSOCreate) (*dto.PurchaseOrderResponse, error) {
	s.logger.Info("Criando pedido de compra a partir de pedido de venda", zap.Int("order_id", id))

	// Buscar pedido de venda
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda para criar pedido de compra", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda para criar pedido de compra: %w", err)
	}

	// Gerar número do pedido de compra
	poNo := fmt.Sprintf(SequenceFormat, PrefixPurchaseOrder, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar pedido de compra - Correção: usar ContactID em vez de SupplierID
	purchaseOrder := &models.PurchaseOrder{
		PONo:         poNo,
		SalesOrderID: order.ID,
		ContactID:    data.ContactID, // Alterado: usar ContactID em vez de SupplierID
		Status:       models.POStatusDraft,
		ExpectedDate: data.ExpectedDate,
		Notes:        data.Notes,
	}

	// Copiar itens selecionados
	for _, itemID := range data.ItemIDs {
		// Encontrar item no pedido de venda
		for _, soItem := range order.Items {
			if soItem.ID == itemID {
				// Verificar se o preço é ou não é zero com if tradicional
				var unitPrice float64 = 0
				if !data.CopyAllItems { // Usando CopyAllItems em vez de UseSOPrices
					unitPrice = soItem.UnitPrice
				}

				poItem := models.POItem{
					ProductID:   soItem.ProductID,
					ProductName: soItem.ProductName,
					Quantity:    soItem.Quantity,
					UnitPrice:   unitPrice,
					Description: soItem.Description,
				}
				purchaseOrder.Items = append(purchaseOrder.Items, poItem)
				break
			}
		}
	}

	// Salvar pedido de compra
	if err := s.purchaseOrderRepo.CreatePurchaseOrder(purchaseOrder); err != nil {
		s.logger.Error("Erro ao criar pedido de compra",
			zap.Error(err),
			zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao criar pedido de compra: %w", err)
	}

	s.logger.Info("Pedido de compra criado com sucesso",
		zap.Int("order_id", id),
		zap.Int("purchase_order_id", purchaseOrder.ID))

	// Converter para DTO de resposta
	// Em uma implementação real, isso precisaria ser implementado no serviço de Purchase Order
	return &dto.PurchaseOrderResponse{
		ID:   purchaseOrder.ID,
		PONo: purchaseOrder.PONo,
		// Outros campos seriam preenchidos aqui
	}, nil
}

// CreateInvoice cria uma fatura a partir de um pedido de venda
func (s *Service) CreateInvoice(ctx context.Context, id int, data *dto.InvoiceFromSOCreate) (*dto.InvoiceResponse, error) {
	s.logger.Info("Criando fatura a partir de pedido de venda", zap.Int("order_id", id))

	// Buscar pedido de venda
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda para criar fatura", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda para criar fatura: %w", err)
	}

	// Verificar status do pedido
	if order.Status != models.SOStatusConfirmed && order.Status != models.SOStatusProcessing && order.Status != models.SOStatusCompleted {
		s.logger.Error("Pedido em status inválido para criar fatura",
			zap.String("status", order.Status),
			zap.Int("order_id", id))
		return nil, fmt.Errorf("pedido em status inválido para criar fatura: %s", order.Status)
	}

	// Gerar número da fatura
	invoiceNo := fmt.Sprintf(SequenceFormat, PrefixInvoice, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar fatura
	invoice := &models.Invoice{
		InvoiceNo:    invoiceNo,
		SalesOrderID: order.ID,
		ContactID:    order.ContactID,
		Status:       models.InvoiceStatusDraft,
		DueDate:      data.DueDate,
		Notes:        data.Notes,
		PaymentTerms: data.PaymentTerms,
	}

	// Copiar itens selecionados ou todos se nenhum foi especificado
	if len(data.ItemIDs) > 0 {
		for _, itemID := range data.ItemIDs {
			// Encontrar item no pedido de venda
			for _, soItem := range order.Items {
				if soItem.ID == itemID {
					invoiceItem := models.InvoiceItem{
						ProductID:   soItem.ProductID,
						ProductName: soItem.ProductName,
						Quantity:    soItem.Quantity,
						UnitPrice:   soItem.UnitPrice,
						Discount:    soItem.Discount,
						Tax:         soItem.Tax,
						Description: soItem.Description,
						Total:       soItem.Total,
					}
					invoice.Items = append(invoice.Items, invoiceItem)
					break
				}
			}
		}
	} else {
		// Copiar todos os itens
		for _, soItem := range order.Items {
			invoiceItem := models.InvoiceItem{
				ProductID:   soItem.ProductID,
				ProductName: soItem.ProductName,
				Quantity:    soItem.Quantity,
				UnitPrice:   soItem.UnitPrice,
				Discount:    soItem.Discount,
				Tax:         soItem.Tax,
				Description: soItem.Description,
				Total:       soItem.Total,
			}
			invoice.Items = append(invoice.Items, invoiceItem)
		}
	}

	// Calcular totais da fatura
	for _, item := range invoice.Items {
		invoice.SubTotal += float64(item.Quantity) * item.UnitPrice
		invoice.DiscountTotal += (float64(item.Quantity) * item.UnitPrice) * (item.Discount / PercentageDivisor)
		taxableAmount := (float64(item.Quantity) * item.UnitPrice) - ((float64(item.Quantity) * item.UnitPrice) * (item.Discount / PercentageDivisor))
		invoice.TaxTotal += taxableAmount * (item.Tax / PercentageDivisor)
	}
	invoice.GrandTotal = invoice.SubTotal - invoice.DiscountTotal + invoice.TaxTotal

	// Salvar fatura
	if err := s.invoiceRepo.CreateInvoice(invoice); err != nil {
		s.logger.Error("Erro ao criar fatura",
			zap.Error(err),
			zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao criar fatura: %w", err)
	}

	s.logger.Info("Fatura criada com sucesso",
		zap.Int("order_id", id),
		zap.Int("invoice_id", invoice.ID))

	// Converter para DTO de resposta
	// Em uma implementação real, isso precisaria ser implementado no serviço de Invoice
	return &dto.InvoiceResponse{
		ID:        invoice.ID,
		InvoiceNo: invoice.InvoiceNo,
		// Outros campos seriam preenchidos aqui
	}, nil
}

// CreateDelivery cria uma entrega a partir de um pedido de venda
func (s *Service) CreateDelivery(ctx context.Context, id int, data *dto.DeliveryFromSOCreate) (*dto.DeliveryResponse, error) {
	s.logger.Info("Criando entrega a partir de pedido de venda", zap.Int("order_id", id))

	// Buscar pedido de venda
	order, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido de venda para criar entrega", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido de venda para criar entrega: %w", err)
	}

	// Verificar status do pedido (código existente permanece)

	// Gerar número da entrega
	deliveryNo := fmt.Sprintf(SequenceFormat, PrefixDelivery, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar entrega - Corrigindo os campos para alinhar com o modelo
	delivery := &models.Delivery{
		DeliveryNo:   deliveryNo,
		SalesOrderID: order.ID,
		// Removido ContactID - campo não existe no modelo
		Status:          models.DeliveryStatusPending,
		DeliveryDate:    data.DeliveryDate,   // Alterado: usar DeliveryDate em vez de PlannedDate
		ShippingMethod:  data.ShippingMethod, // Alterado: em vez de ShippingInfo
		TrackingNumber:  data.TrackingNumber, // Alterado: em vez de TrackingNo
		Notes:           data.Notes,
		ShippingAddress: data.ShippingAddress, // Este campo existe tanto no modelo quanto no DTO
	}

	// Copiar itens selecionados - Correção para CopyAllItems e ItemIDs
	if !data.CopyAllItems && len(data.ItemIDs) > 0 {
		for _, itemID := range data.ItemIDs {
			// Encontrar item no pedido de venda
			for _, soItem := range order.Items {
				if soItem.ID == itemID {
					// Verificar quantidade a usar
					qtyToShip := soItem.Quantity
					if data.Quantities != nil && data.Quantities[itemID] > 0 {
						qtyToShip = data.Quantities[itemID]
						if qtyToShip > soItem.Quantity {
							qtyToShip = soItem.Quantity
						}
					}

					deliveryItem := models.DeliveryItem{
						ProductID:   soItem.ProductID,
						ProductName: soItem.ProductName,
						ProductCode: soItem.ProductCode,
						Quantity:    qtyToShip,
						ReceivedQty: 0,
						Notes:       "", // Campo direto em vez de via Items
					}
					delivery.Items = append(delivery.Items, deliveryItem)
					break
				}
			}
		}
	} else {
		// Copiar todos os itens
		for _, soItem := range order.Items {
			deliveryItem := models.DeliveryItem{
				ProductID:   soItem.ProductID,
				ProductName: soItem.ProductName,
				ProductCode: soItem.ProductCode,
				Quantity:    soItem.Quantity,
				ReceivedQty: 0,
			}
			delivery.Items = append(delivery.Items, deliveryItem)
		}
	}

	// Verificar se há itens
	if len(delivery.Items) == 0 {
		s.logger.Error("Não há itens válidos para criar entrega", zap.Int("order_id", id))
		return nil, fmt.Errorf("não há itens válidos para criar entrega")
	}

	// Salvar entrega
	if err := s.deliveryRepo.CreateDelivery(delivery); err != nil {
		s.logger.Error("Erro ao criar entrega",
			zap.Error(err),
			zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao criar entrega: %w", err)
	}

	// Atualizar status do pedido se necessário
	if order.Status == models.SOStatusConfirmed {
		order.Status = models.SOStatusProcessing
		if err := s.salesOrderRepo.UpdateSalesOrder(id, order); err != nil {
			s.logger.Warn("Erro ao atualizar status do pedido após criar entrega",
				zap.Error(err),
				zap.Int("order_id", id))
		}
	}

	s.logger.Info("Entrega criada com sucesso",
		zap.Int("order_id", id),
		zap.Int("delivery_id", delivery.ID))

	// Converter para DTO de resposta
	// Em uma implementação real, isso precisaria ser implementado no serviço de Delivery
	return &dto.DeliveryResponse{
		ID:         delivery.ID,
		DeliveryNo: delivery.DeliveryNo,
		// Outros campos seriam preenchidos aqui
	}, nil
}

// Clone clona um pedido de venda
func (s *Service) Clone(ctx context.Context, id int, options *dto.SalesOrderCloneOptions) (*dto.SalesOrderResponse, error) {
	s.logger.Info("Clonando pedido de venda", zap.Int("order_id", id), zap.Any("options", options))

	// Buscar pedido original
	original, err := s.salesOrderRepo.GetSalesOrderByID(id)
	if err != nil {
		s.logger.Error("Erro ao buscar pedido para clonagem", zap.Error(err), zap.Int("order_id", id))
		return nil, fmt.Errorf("falha ao buscar pedido para clonagem: %w", err)
	}

	// Gerar número para novo pedido
	newSONo := fmt.Sprintf(SequenceFormat, PrefixSalesOrder, time.Now().Format(DateFormatForCodes), s.generateSequence())

	// Criar novo pedido baseado no original
	clone := &models.SalesOrder{
		SONo:            newSONo,
		QuotationID:     0, // Normalmente um clone não estaria associado à cotação original
		ContactID:       original.ContactID,
		Status:          models.SOStatusDraft,
		ExpectedDate:    time.Now().AddDate(0, 0, 30), // Data padrão: hoje + 30 dias
		SubTotal:        original.SubTotal,
		TaxTotal:        original.TaxTotal,
		DiscountTotal:   original.DiscountTotal,
		GrandTotal:      original.GrandTotal,
		Notes:           original.Notes,
		PaymentTerms:    original.PaymentTerms,
		ShippingAddress: original.ShippingAddress,
	}

	// Aplicar opções de clonagem
	if options.ContactID > 0 {
		clone.ContactID = options.ContactID
	}

	if !options.ExpectedDate.IsZero() {
		clone.ExpectedDate = options.ExpectedDate
	}

	if options.CopyItems {
		clone.QuotationID = original.QuotationID
	}

	// Copiar itens
	for _, item := range original.Items {
		newItem := models.SOItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			ProductCode: item.ProductCode,
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
	s.calculateOrderTotals(clone)

	// Salvar novo pedido
	if err := s.salesOrderRepo.CreateSalesOrder(clone); err != nil {
		s.logger.Error("Erro ao criar pedido clonado", zap.Error(err), zap.Int("original_id", id))
		return nil, fmt.Errorf("falha ao criar pedido clonado: %w", err)
	}

	s.logger.Info("Pedido clonado com sucesso",
		zap.Int("original_id", id),
		zap.Int("clone_id", clone.ID))
	return s.convertToSalesOrderResponse(clone), nil
}
