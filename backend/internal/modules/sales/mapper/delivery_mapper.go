package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
)

// ToDeliveryResponseDTO converte Delivery model para DeliveryResponseDTO
func ToDeliveryResponseDTO(delivery *models.Delivery) *dtos.DeliveryResponseDTO {
	if delivery == nil {
		return nil
	}

	dto := &dtos.DeliveryResponseDTO{
		ID:              delivery.ID,
		DeliveryNo:      delivery.DeliveryNo,
		PurchaseOrderID: delivery.PurchaseOrderID,
		PONo:            delivery.PONo,
		SalesOrderID:    delivery.SalesOrderID,
		SONo:            delivery.SONo,
		Status:          delivery.Status,
		CreatedAt:       delivery.CreatedAt,
		UpdatedAt:       delivery.UpdatedAt,
		DeliveryDate:    delivery.DeliveryDate,
		ShippingMethod:  delivery.ShippingMethod,
		TrackingNumber:  delivery.TrackingNumber,
		ShippingAddress: delivery.ShippingAddress,
		Notes:           delivery.Notes,
	}

	// ReceivedDate pode estar vazio
	if !delivery.ReceivedDate.IsZero() {
		dto.ReceivedDate = &delivery.ReceivedDate
	}

	// Mapear itens
	dto.Items = ToDeliveryItemResponseDTOList(delivery.Items)

	// Contact precisa ser obtido através do PO ou SO
	if delivery.PurchaseOrder != nil && delivery.PurchaseOrder.Contact != nil {
		dto.Contact = ToContactBasicInfo(delivery.PurchaseOrder.Contact)
	} else if delivery.SalesOrder != nil && delivery.SalesOrder.Contact != nil {
		dto.Contact = ToContactBasicInfo(delivery.SalesOrder.Contact)
	}

	return dto
}

// ToDeliveryItemResponseDTO converte DeliveryItem model para DeliveryItemResponseDTO
func ToDeliveryItemResponseDTO(item *models.DeliveryItem) *dtos.DeliveryItemResponseDTO {
	if item == nil {
		return nil
	}

	dto := &dtos.DeliveryItemResponseDTO{
		ID:          item.ID,
		DeliveryID:  item.DeliveryID,
		ProductID:   item.ProductID,
		ProductName: item.ProductName,
		ProductCode: item.ProductCode,
		Description: item.Description,
		Quantity:    item.Quantity,
		ReceivedQty: item.ReceivedQty,
		Notes:       item.Notes,
	}

	// Calcular status do item
	if item.ReceivedQty == 0 {
		dto.Status = "pending"
	} else if item.ReceivedQty < item.Quantity {
		dto.Status = "partial"
	} else {
		dto.Status = "complete"
	}

	return dto
}

// FromDeliveryCreateDTO converte DeliveryCreateDTO para Delivery model
func FromDeliveryCreateDTO(dto *dtos.DeliveryCreateDTO) *models.Delivery {
	if dto == nil {
		return nil
	}

	delivery := &models.Delivery{
		PurchaseOrderID: dto.PurchaseOrderID,
		SalesOrderID:    dto.SalesOrderID,
		DeliveryDate:    dto.DeliveryDate,
		ShippingMethod:  dto.ShippingMethod,
		ShippingAddress: dto.ShippingAddress,
		Notes:           dto.Notes,
		Status:          models.DeliveryStatusPending,
	}

	// Mapear itens
	delivery.Items = make([]models.DeliveryItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		delivery.Items[i] = *FromDeliveryItemCreateDTO(&itemDTO)
	}

	return delivery
}

// FromDeliveryItemCreateDTO converte DeliveryItemCreateDTO para DeliveryItem model
func FromDeliveryItemCreateDTO(dto *dtos.DeliveryItemCreateDTO) *models.DeliveryItem {
	if dto == nil {
		return nil
	}

	return &models.DeliveryItem{
		ProductID:   dto.ProductID,
		ProductName: dto.ProductName,
		ProductCode: dto.ProductCode,
		Description: dto.Description,
		Quantity:    dto.Quantity,
		Notes:       dto.Notes,
		ReceivedQty: 0, // Inicialmente não recebido
	}
}

// Helper functions
func ToDeliveryResponseDTOList(deliveries []models.Delivery) []dtos.DeliveryResponseDTO {
	if deliveries == nil {
		return nil
	}

	result := make([]dtos.DeliveryResponseDTO, len(deliveries))
	for i, delivery := range deliveries {
		result[i] = *ToDeliveryResponseDTO(&delivery)
	}
	return result
}

func ToDeliveryItemResponseDTOList(items []models.DeliveryItem) []dtos.DeliveryItemResponseDTO {
	if items == nil {
		return nil
	}

	result := make([]dtos.DeliveryItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = *ToDeliveryItemResponseDTO(&item)
	}
	return result
}
