package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"time"
)

// ToInvoiceResponseDTO converte Invoice model para InvoiceResponseDTO
func ToInvoiceResponseDTO(invoice *models.Invoice) *dtos.InvoiceResponseDTO {
	if invoice == nil {
		return nil
	}

	dto := &dtos.InvoiceResponseDTO{
		ID:            invoice.ID,
		InvoiceNo:     invoice.InvoiceNo,
		SalesOrderID:  invoice.SalesOrderID,
		ContactID:     invoice.ContactID,
		Status:        invoice.Status,
		CreatedAt:     invoice.CreatedAt,
		UpdatedAt:     invoice.UpdatedAt,
		IssueDate:     invoice.IssueDate,
		DueDate:       invoice.DueDate,
		SubTotal:      invoice.SubTotal,
		TaxTotal:      invoice.TaxTotal,
		DiscountTotal: invoice.DiscountTotal,
		GrandTotal:    invoice.GrandTotal,
		AmountPaid:    invoice.AmountPaid,
		BalanceDue:    invoice.GrandTotal - invoice.AmountPaid, // Calculado
		PaymentTerms:  invoice.PaymentTerms,
		Notes:         invoice.Notes,
	}

	// Mapear relações
	if invoice.SalesOrder != nil {
		dto.SONo = invoice.SalesOrder.SONo
	}

	if invoice.Contact != nil {
		dto.Contact = ToContactBasicInfo(invoice.Contact)
	}

	// Mapear itens
	dto.Items = ToInvoiceItemResponseDTOList(invoice.Items)

	// Mapear pagamentos
	dto.Payments = ToPaymentResponseDTOList(invoice.Payments)

	// Calcular campos derivados
	now := time.Now()
	dto.IsOverdue = invoice.Status != models.InvoiceStatusPaid &&
		invoice.Status != models.InvoiceStatusCancelled &&
		now.After(invoice.DueDate)

	if dto.IsOverdue {
		dto.DaysOverdue = int(now.Sub(invoice.DueDate).Hours() / 24)
	}

	return dto
}

// ToInvoiceItemResponseDTO converte InvoiceItem model para InvoiceItemResponseDTO
func ToInvoiceItemResponseDTO(item *models.InvoiceItem) *dtos.InvoiceItemResponseDTO {
	if item == nil {
		return nil
	}

	return &dtos.InvoiceItemResponseDTO{
		ID:          item.ID,
		InvoiceID:   item.InvoiceID,
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
}

// FromInvoiceCreateDTO converte InvoiceCreateDTO para Invoice model
func FromInvoiceCreateDTO(dto *dtos.InvoiceCreateDTO) *models.Invoice {
	if dto == nil {
		return nil
	}

	invoice := &models.Invoice{
		SalesOrderID: dto.SalesOrderID,
		ContactID:    dto.ContactID,
		IssueDate:    dto.IssueDate,
		DueDate:      dto.DueDate,
		PaymentTerms: dto.PaymentTerms,
		Notes:        dto.Notes,
		Status:       models.InvoiceStatusDraft, // Status inicial
	}

	// Mapear itens
	invoice.Items = make([]models.InvoiceItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		invoice.Items[i] = *FromInvoiceItemCreateDTO(&itemDTO)
	}

	return invoice
}

// FromInvoiceItemCreateDTO converte InvoiceItemCreateDTO para InvoiceItem model
func FromInvoiceItemCreateDTO(dto *dtos.InvoiceItemCreateDTO) *models.InvoiceItem {
	if dto == nil {
		return nil
	}

	return &models.InvoiceItem{
		ProductID:   dto.ProductID,
		ProductName: dto.ProductName,
		ProductCode: dto.ProductCode,
		Description: dto.Description,
		Quantity:    dto.Quantity,
		UnitPrice:   dto.UnitPrice,
		Discount:    dto.Discount,
		Tax:         dto.Tax,
	}
}

// Helper functions
func ToInvoiceResponseDTOList(invoices []models.Invoice) []dtos.InvoiceResponseDTO {
	if invoices == nil {
		return nil
	}

	result := make([]dtos.InvoiceResponseDTO, len(invoices))
	for i, invoice := range invoices {
		result[i] = *ToInvoiceResponseDTO(&invoice)
	}
	return result
}

func ToInvoiceItemResponseDTOList(items []models.InvoiceItem) []dtos.InvoiceItemResponseDTO {
	if items == nil {
		return nil
	}

	result := make([]dtos.InvoiceItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = *ToInvoiceItemResponseDTO(&item)
	}
	return result
}
