package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"time"
)

// ToPurchaseOrderResponseDTO converte PurchaseOrder model para PurchaseOrderResponseDTO
func ToPurchaseOrderResponseDTO(po *models.PurchaseOrder) *dtos.PurchaseOrderResponseDTO {
	if po == nil {
		return nil
	}

	dto := &dtos.PurchaseOrderResponseDTO{
		ID:              po.ID,
		PONo:            po.PONo,
		SONo:            po.SONo,
		SalesOrderID:    po.SalesOrderID,
		ContactID:       po.ContactID,
		Status:          po.Status,
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
		ExpectedDate:    po.ExpectedDate,
		SubTotal:        po.SubTotal,
		TaxTotal:        po.TaxTotal,
		DiscountTotal:   po.DiscountTotal,
		GrandTotal:      po.GrandTotal,
		Notes:           po.Notes,
		PaymentTerms:    po.PaymentTerms,
		ShippingAddress: po.ShippingAddress,
	}

	// Mapear Contact
	if po.Contact != nil {
		dto.Contact = ToContactBasicInfo(po.Contact)
	}

	// Mapear itens
	dto.Items = ToPOItemResponseDTOList(po.Items)

	// Calcular campos derivados
	// DeliveryCount seria calculado através de um repository
	// Por ora, deixamos como 0
	dto.DeliveryCount = 0

	// Calcular se está atrasado
	now := time.Now()
	dto.IsOverdue = po.Status != models.POStatusReceived &&
		po.Status != models.POStatusCancelled &&
		now.After(po.ExpectedDate)

	if dto.IsOverdue {
		dto.DaysOverdue = int(now.Sub(po.ExpectedDate).Hours() / 24)
	}

	return dto
}

// ToPurchaseOrderListItemDTO converte PurchaseOrder para versão resumida
func ToPurchaseOrderListItemDTO(po *models.PurchaseOrder) *dtos.PurchaseOrderListItemDTO {
	if po == nil {
		return nil
	}

	dto := &dtos.PurchaseOrderListItemDTO{
		ID:           po.ID,
		PONo:         po.PONo,
		SONo:         po.SONo,
		ContactID:    po.ContactID,
		Status:       po.Status,
		CreatedAt:    po.CreatedAt,
		ExpectedDate: po.ExpectedDate,
		GrandTotal:   po.GrandTotal,
		ItemCount:    len(po.Items),
	}

	// Mapear Contact
	if po.Contact != nil {
		dto.Contact = ToContactBasicInfo(po.Contact)
	}

	// Calcular campos derivados
	now := time.Now()
	dto.IsOverdue = po.Status != models.POStatusReceived &&
		po.Status != models.POStatusCancelled &&
		now.After(po.ExpectedDate)

	if dto.IsOverdue {
		dto.DaysOverdue = int(now.Sub(po.ExpectedDate).Hours() / 24)
	}

	return dto
}

// ToPOItemResponseDTO converte POItem model para POItemResponseDTO
func ToPOItemResponseDTO(item *models.POItem) *dtos.POItemResponseDTO {
	if item == nil {
		return nil
	}

	return &dtos.POItemResponseDTO{
		ID:              item.ID,
		PurchaseOrderID: item.PurchaseOrderID,
		ProductID:       item.ProductID,
		ProductName:     item.ProductName,
		ProductCode:     item.ProductCode,
		Description:     item.Description,
		Quantity:        item.Quantity,
		UnitPrice:       item.UnitPrice,
		Discount:        item.Discount,
		Tax:             item.Tax,
		Total:           item.Total,
		// ReceivedQty e PendingQty seriam calculados através de deliveries
		// Por ora, deixamos como 0 e Quantity
		ReceivedQty: 0,
		PendingQty:  item.Quantity,
	}
}

// FromPurchaseOrderCreateDTO converte PurchaseOrderCreateDTO para PurchaseOrder model
func FromPurchaseOrderCreateDTO(dto *dtos.PurchaseOrderCreateDTO) *models.PurchaseOrder {
	if dto == nil {
		return nil
	}

	po := &models.PurchaseOrder{
		SalesOrderID:    dto.SalesOrderID,
		ContactID:       dto.ContactID,
		ExpectedDate:    dto.ExpectedDate,
		PaymentTerms:    dto.PaymentTerms,
		ShippingAddress: dto.ShippingAddress,
		Notes:           dto.Notes,
		Status:          models.POStatusDraft,
	}

	// Mapear itens
	po.Items = make([]models.POItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		po.Items[i] = *FromPOItemCreateDTO(&itemDTO)
	}

	return po
}

// FromPOItemCreateDTO converte POItemCreateDTO para POItem model
func FromPOItemCreateDTO(dto *dtos.POItemCreateDTO) *models.POItem {
	if dto == nil {
		return nil
	}

	return &models.POItem{
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

// FromPurchaseOrderUpdateDTO aplica updates parciais
func FromPurchaseOrderUpdateDTO(dto *dtos.PurchaseOrderUpdateDTO, po *models.PurchaseOrder) {
	if dto == nil || po == nil {
		return
	}

	if dto.ExpectedDate != nil {
		po.ExpectedDate = *dto.ExpectedDate
	}

	if dto.PaymentTerms != nil {
		po.PaymentTerms = *dto.PaymentTerms
	}

	if dto.ShippingAddress != nil {
		po.ShippingAddress = *dto.ShippingAddress
	}

	if dto.Notes != nil {
		po.Notes = *dto.Notes
	}

	if dto.Status != nil {
		po.Status = *dto.Status
	}
}

// Helper functions
func ToPurchaseOrderResponseDTOList(pos []models.PurchaseOrder) []dtos.PurchaseOrderResponseDTO {
	if pos == nil {
		return nil
	}

	result := make([]dtos.PurchaseOrderResponseDTO, len(pos))
	for i, po := range pos {
		result[i] = *ToPurchaseOrderResponseDTO(&po)
	}
	return result
}

func ToPOItemResponseDTOList(items []models.POItem) []dtos.POItemResponseDTO {
	if items == nil {
		return nil
	}

	result := make([]dtos.POItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = *ToPOItemResponseDTO(&item)
	}
	return result
}
