package sales_order

import (
	"ERP-ONSMART/backend/internal/modules/sales/dto"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"time"
)

// generateSequence gera um número sequencial para identificação
func (s *Service) generateSequence() int {
	// Em uma implementação real, poderia usar um contador persistente
	// Para esta implementação, usamos o timestamp como base
	return int(time.Now().Unix() % SequenceModulo)
}

// calculateItemTotal calcula o total de um item
func calculateItemTotal(quantity int, unitPrice float64, discount float64, tax float64) float64 {
	subtotal := float64(quantity) * unitPrice
	discountAmount := subtotal * (discount / PercentageDivisor)
	taxAmount := (subtotal - discountAmount) * (tax / PercentageDivisor)
	return subtotal - discountAmount + taxAmount
}

// calculateOrderTotals recalcula os totais de um pedido
func (s *Service) calculateOrderTotals(order *models.SalesOrder) {
	var subtotal, taxTotal, discountTotal, grandTotal float64

	for i := range order.Items {
		item := &order.Items[i]

		// Recalcular total do item
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemDiscount := itemSubtotal * (item.Discount / PercentageDivisor)
		itemTax := (itemSubtotal - itemDiscount) * (item.Tax / PercentageDivisor)
		itemTotal := itemSubtotal - itemDiscount + itemTax

		item.Total = itemTotal

		// Adicionar aos totais do pedido
		subtotal += itemSubtotal
		discountTotal += itemDiscount
		taxTotal += itemTax
		grandTotal += itemTotal
	}

	order.SubTotal = subtotal
	order.DiscountTotal = discountTotal
	order.TaxTotal = taxTotal
	order.GrandTotal = grandTotal
}

// convertToSalesOrderResponse converte um modelo de pedido para seu DTO correspondente
func (s *Service) convertToSalesOrderResponse(order *models.SalesOrder) *dto.SalesOrderResponse {
	response := &dto.SalesOrderResponse{
		ID:              order.ID,
		SONo:            order.SONo,
		ContactID:       order.ContactID,
		QuotationID:     order.QuotationID,
		Status:          order.Status,
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
		ExpectedDate:    order.ExpectedDate,
		SubTotal:        order.SubTotal,
		TaxTotal:        order.TaxTotal,
		DiscountTotal:   order.DiscountTotal,
		GrandTotal:      order.GrandTotal,
		PaymentTerms:    order.PaymentTerms,
		ShippingAddress: order.ShippingAddress,
		Notes:           order.Notes,
		Contact:         dto.ContactResponse{},          // Inicializa com valor vazio
		Items:           []dto.SalesOrderItemResponse{}, // Inicializa como slice vazio
	}

	// Converter contato se presente
	if order.Contact != nil {
		response.Contact = dto.ContactResponse{
			ID:    order.Contact.ID,
			Name:  order.Contact.Name,
			Email: order.Contact.Email,
			Phone: order.Contact.Phone,
		}
	}

	// Converter cotação se presente
	if order.Quotation != nil {
		response.Quotation = &dto.QuotationShortResponse{
			ID:          order.Quotation.ID,
			QuotationNo: order.Quotation.QuotationNo,
			Status:      order.Quotation.Status,
		}
	}

	// Converter itens
	for _, item := range order.Items {
		baseResponse := dto.BaseItemResponse{
			ID:          item.ID,
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			ProductCode: item.ProductCode,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Discount:    item.Discount,
			Tax:         item.Tax,
			Total:       item.Total,
			Description: item.Description,
		}
		itemResponse := dto.SalesOrderItemResponse{
			BaseItemResponse: baseResponse,
		}

		response.Items = append(response.Items, itemResponse)
	}

	return response
}

// convertToSalesOrderShortResponse converte um modelo de pedido para seu DTO resumido
func (s *Service) convertToSalesOrderShortResponse(order *models.SalesOrder) *dto.SalesOrderShortResponse {
	response := &dto.SalesOrderShortResponse{
		ID:           order.ID,
		SONo:         order.SONo,
		ContactID:    order.ContactID,
		Status:       order.Status,
		ExpectedDate: order.ExpectedDate,
		GrandTotal:   order.GrandTotal,
		ItemsCount:   len(order.Items),
		Contact:      dto.ContactShortResponse{},
	}

	// Criar contato resumido se disponível
	if order.Contact != nil {
		response.Contact = dto.ContactShortResponse{
			ID:    order.Contact.ID,
			Name:  order.Contact.Name,
			Email: order.Contact.Email,
			Phone: order.Contact.Phone,
		}
	}

	return response
}
