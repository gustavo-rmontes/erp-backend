// helpers.go
package quotation

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

// calculateQuotationTotals recalcula os totais de uma cotação
func (s *Service) calculateQuotationTotals(quotation *models.Quotation) {
	var subtotal, taxTotal, discountTotal, grandTotal float64

	for i := range quotation.Items {
		item := &quotation.Items[i]

		// Recalcular total do item
		itemSubtotal := float64(item.Quantity) * item.UnitPrice
		itemDiscount := itemSubtotal * (item.Discount / PercentageDivisor)
		itemTax := (itemSubtotal - itemDiscount) * (item.Tax / PercentageDivisor)
		itemTotal := itemSubtotal - itemDiscount + itemTax

		item.Total = itemTotal

		// Adicionar aos totais da cotação
		subtotal += itemSubtotal
		discountTotal += itemDiscount
		taxTotal += itemTax
		grandTotal += itemTotal
	}

	quotation.SubTotal = subtotal
	quotation.DiscountTotal = discountTotal
	quotation.TaxTotal = taxTotal
	quotation.GrandTotal = grandTotal
}

// convertToQuotationResponse converte um modelo de cotação para seu DTO correspondente
func (s *Service) convertToQuotationResponse(quotation *models.Quotation) *dto.QuotationResponse {
	response := &dto.QuotationResponse{
		ID:            quotation.ID,
		QuotationNo:   quotation.QuotationNo,
		ContactID:     quotation.ContactID,
		Status:        quotation.Status,
		CreatedAt:     quotation.CreatedAt,
		UpdatedAt:     quotation.UpdatedAt,
		ExpiryDate:    quotation.ExpiryDate,
		SubTotal:      quotation.SubTotal,
		TaxTotal:      quotation.TaxTotal,
		DiscountTotal: quotation.DiscountTotal,
		GrandTotal:    quotation.GrandTotal,
		Notes:         quotation.Notes,
		Terms:         quotation.Terms,
		Contact:       dto.ContactResponse{},         // Inicializa com valor vazio
		Items:         []dto.QuotationItemResponse{}, // Inicializa como slice vazio
	}

	// Converter contato se presente
	if quotation.Contact != nil {
		response.Contact = dto.ContactResponse{
			ID:    quotation.Contact.ID,
			Name:  quotation.Contact.Name,
			Email: quotation.Contact.Email,
			Phone: quotation.Contact.Phone,
		}
	}

	// Converter itens
	for _, item := range quotation.Items {
		baseResponse := dto.BaseItemResponse{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			Description: item.Description,
		}
		itemResponse := dto.QuotationItemResponse{
			BaseItemResponse: baseResponse,
		}

		response.Items = append(response.Items, itemResponse)
	}

	return response
}

// convertToQuotationShortResponse converte um modelo de cotação para seu DTO resumido
func (s *Service) convertToQuotationShortResponse(quotation *models.Quotation) *dto.QuotationShortResponse {
	response := &dto.QuotationShortResponse{
		ID:          quotation.ID,
		QuotationNo: quotation.QuotationNo,
		ContactID:   quotation.ContactID,
		Status:      quotation.Status,
		ExpiryDate:  quotation.ExpiryDate,
		GrandTotal:  quotation.GrandTotal,
		ItemsCount:  len(quotation.Items),
		Contact:     dto.ContactShortResponse{},
	}

	// Criar contato resumido se disponível
	if quotation.Contact != nil {
		response.Contact = dto.ContactShortResponse{
			ID:   quotation.Contact.ID,
			Name: quotation.Contact.Name,
		}
	}

	return response
}

// convertToSalesOrderResponse converte um modelo de pedido de venda para seu DTO
func (s *Service) convertToSalesOrderResponse(salesOrder *models.SalesOrder) *dto.SalesOrderResponse {
	response := &dto.SalesOrderResponse{
		ID:              salesOrder.ID,
		SONo:            salesOrder.SONo,
		QuotationID:     salesOrder.QuotationID,
		ContactID:       salesOrder.ContactID,
		Status:          salesOrder.Status,
		CreatedAt:       salesOrder.CreatedAt,
		UpdatedAt:       salesOrder.UpdatedAt,
		ExpectedDate:    salesOrder.ExpectedDate,
		SubTotal:        salesOrder.SubTotal,
		TaxTotal:        salesOrder.TaxTotal,
		DiscountTotal:   salesOrder.DiscountTotal,
		GrandTotal:      salesOrder.GrandTotal,
		Notes:           salesOrder.Notes,
		PaymentTerms:    salesOrder.PaymentTerms,
		ShippingAddress: salesOrder.ShippingAddress,
	}

	return response
}
