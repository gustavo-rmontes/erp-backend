package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
)

// ToSalesOrderResponseDTO converte SalesOrder model para SalesOrderResponseDTO
func ToSalesOrderResponseDTO(so *models.SalesOrder) *dtos.SalesOrderResponseDTO {
	if so == nil {
		return nil
	}

	dto := &dtos.SalesOrderResponseDTO{
		ID:              so.ID,
		SONo:            so.SONo,
		QuotationID:     so.QuotationID,
		ContactID:       so.ContactID,
		Status:          so.Status,
		CreatedAt:       so.CreatedAt,
		UpdatedAt:       so.UpdatedAt,
		ExpectedDate:    so.ExpectedDate,
		SubTotal:        so.SubTotal,
		TaxTotal:        so.TaxTotal,
		DiscountTotal:   so.DiscountTotal,
		GrandTotal:      so.GrandTotal,
		Notes:           so.Notes,
		PaymentTerms:    so.PaymentTerms,
		ShippingAddress: so.ShippingAddress,
	}

	// Mapear Contact
	if so.Contact != nil {
		dto.Contact = ToContactBasicInfo(so.Contact)
	}

	// Mapear itens
	dto.Items = ToSOItemResponseDTOList(so.Items)

	// Calcular campos derivados - estes seriam calculados através de repositories
	// Por ora, deixamos como 0
	dto.InvoiceCount = 0
	dto.POCount = 0
	dto.DeliveryCount = 0
	dto.FulfillmentRate = 0

	return dto
}

// ToSalesOrderListItemDTO converte SalesOrder para versão resumida
func ToSalesOrderListItemDTO(so *models.SalesOrder) *dtos.SalesOrderListItemDTO {
	if so == nil {
		return nil
	}

	dto := &dtos.SalesOrderListItemDTO{
		ID:           so.ID,
		SONo:         so.SONo,
		ContactID:    so.ContactID,
		Status:       so.Status,
		CreatedAt:    so.CreatedAt,
		ExpectedDate: so.ExpectedDate,
		GrandTotal:   so.GrandTotal,
		ItemCount:    len(so.Items),
	}

	// Mapear Contact
	if so.Contact != nil {
		dto.Contact = ToContactBasicInfo(so.Contact)
	}

	// Campos calculados - por ora como 0
	dto.InvoiceCount = 0
	dto.DeliveryCount = 0
	dto.FulfillmentRate = 0

	return dto
}

// ToSOItemResponseDTO converte SOItem model para SOItemResponseDTO
func ToSOItemResponseDTO(item *models.SOItem) *dtos.SOItemResponseDTO {
	if item == nil {
		return nil
	}

	dto := &dtos.SOItemResponseDTO{
		ID:           item.ID,
		SalesOrderID: item.SalesOrderID,
		ProductID:    item.ProductID,
		ProductName:  item.ProductName,
		ProductCode:  item.ProductCode,
		Description:  item.Description,
		Quantity:     item.Quantity,
		UnitPrice:    item.UnitPrice,
		Discount:     item.Discount,
		Tax:          item.Tax,
		Total:        item.Total,
	}

	// Campos calculados - por ora como 0 ou valores default
	dto.DeliveredQty = 0
	dto.InvoicedQty = 0
	dto.PendingQty = item.Quantity

	return dto
}

// FromSalesOrderCreateDTO converte SalesOrderCreateDTO para SalesOrder model
func FromSalesOrderCreateDTO(dto *dtos.SalesOrderCreateDTO) *models.SalesOrder {
	if dto == nil {
		return nil
	}

	so := &models.SalesOrder{
		QuotationID:     dto.QuotationID,
		ContactID:       dto.ContactID,
		ExpectedDate:    dto.ExpectedDate,
		PaymentTerms:    dto.PaymentTerms,
		ShippingAddress: dto.ShippingAddress,
		Notes:           dto.Notes,
		Status:          models.SOStatusDraft,
	}

	// Mapear itens
	so.Items = make([]models.SOItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		so.Items[i] = *FromSOItemCreateDTO(&itemDTO)
	}

	return so
}

// FromSOItemCreateDTO converte SOItemCreateDTO para SOItem model
func FromSOItemCreateDTO(dto *dtos.SOItemCreateDTO) *models.SOItem {
	if dto == nil {
		return nil
	}

	return &models.SOItem{
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

// FromSalesOrderUpdateDTO aplica updates parciais
func FromSalesOrderUpdateDTO(dto *dtos.SalesOrderUpdateDTO, so *models.SalesOrder) {
	if dto == nil || so == nil {
		return
	}

	if dto.ExpectedDate != nil {
		so.ExpectedDate = *dto.ExpectedDate
	}

	if dto.PaymentTerms != nil {
		so.PaymentTerms = *dto.PaymentTerms
	}

	if dto.ShippingAddress != nil {
		so.ShippingAddress = *dto.ShippingAddress
	}

	if dto.Notes != nil {
		so.Notes = *dto.Notes
	}
}

// Helper functions
func ToSalesOrderResponseDTOList(sos []models.SalesOrder) []dtos.SalesOrderResponseDTO {
	if sos == nil {
		return nil
	}

	result := make([]dtos.SalesOrderResponseDTO, len(sos))
	for i, so := range sos {
		result[i] = *ToSalesOrderResponseDTO(&so)
	}
	return result
}

func ToSOItemResponseDTOList(items []models.SOItem) []dtos.SOItemResponseDTO {
	if items == nil {
		return nil
	}

	result := make([]dtos.SOItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = *ToSOItemResponseDTO(&item)
	}
	return result
}
