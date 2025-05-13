package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"time"
)

// ToQuotationResponseDTO converte Quotation model para QuotationResponseDTO
func ToQuotationResponseDTO(quotation *models.Quotation) *dtos.QuotationResponseDTO {
	if quotation == nil {
		return nil
	}

	dto := &dtos.QuotationResponseDTO{
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
	}

	// Mapear Contact
	if quotation.Contact != nil {
		dto.Contact = ToContactBasicInfo(quotation.Contact)
	}

	// Mapear itens
	dto.Items = ToQuotationItemResponseDTOList(quotation.Items)

	// Calcular campos derivados
	now := time.Now()
	dto.IsExpired = quotation.Status != models.QuotationStatusAccepted &&
		quotation.Status != models.QuotationStatusCancelled &&
		now.After(quotation.ExpiryDate)

	if !dto.IsExpired && now.Before(quotation.ExpiryDate) {
		dto.DaysToExpiry = int(quotation.ExpiryDate.Sub(now).Hours() / 24)
	}

	return dto
}

// ToQuotationListItemDTO converte Quotation para versão resumida
func ToQuotationListItemDTO(quotation *models.Quotation) *dtos.QuotationListItemDTO {
	if quotation == nil {
		return nil
	}

	dto := &dtos.QuotationListItemDTO{
		ID:          quotation.ID,
		QuotationNo: quotation.QuotationNo,
		ContactID:   quotation.ContactID,
		Status:      quotation.Status,
		CreatedAt:   quotation.CreatedAt,
		ExpiryDate:  quotation.ExpiryDate,
		GrandTotal:  quotation.GrandTotal,
	}

	// Mapear Contact
	if quotation.Contact != nil {
		dto.Contact = ToContactBasicInfo(quotation.Contact)
	}

	// Calcular campos derivados
	now := time.Now()
	dto.IsExpired = quotation.Status != models.QuotationStatusAccepted &&
		quotation.Status != models.QuotationStatusCancelled &&
		now.After(quotation.ExpiryDate)

	if !dto.IsExpired && now.Before(quotation.ExpiryDate) {
		dto.DaysToExpiry = int(quotation.ExpiryDate.Sub(now).Hours() / 24)
	}

	return dto
}

// ToQuotationItemResponseDTO converte QuotationItem model para QuotationItemResponseDTO
func ToQuotationItemResponseDTO(item *models.QuotationItem) *dtos.QuotationItemResponseDTO {
	if item == nil {
		return nil
	}

	return &dtos.QuotationItemResponseDTO{
		ID:          item.ID,
		QuotationID: item.QuotationID,
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

// FromQuotationCreateDTO converte QuotationCreateDTO para Quotation model
func FromQuotationCreateDTO(dto *dtos.QuotationCreateDTO) *models.Quotation {
	if dto == nil {
		return nil
	}

	quotation := &models.Quotation{
		ContactID:  dto.ContactID,
		ExpiryDate: dto.ExpiryDate,
		Notes:      dto.Notes,
		Terms:      dto.Terms,
		Status:     models.QuotationStatusDraft,
	}

	// Mapear itens
	quotation.Items = make([]models.QuotationItem, len(dto.Items))
	for i, itemDTO := range dto.Items {
		quotation.Items[i] = *FromQuotationItemCreateDTO(&itemDTO)
	}

	return quotation
}

// FromQuotationItemCreateDTO converte QuotationItemCreateDTO para QuotationItem model
func FromQuotationItemCreateDTO(dto *dtos.QuotationItemCreateDTO) *models.QuotationItem {
	if dto == nil {
		return nil
	}

	return &models.QuotationItem{
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

// FromQuotationUpdateDTO aplica updates parciais
func FromQuotationUpdateDTO(dto *dtos.QuotationUpdateDTO, quotation *models.Quotation) {
	if dto == nil || quotation == nil {
		return
	}

	if dto.ExpiryDate != nil {
		quotation.ExpiryDate = *dto.ExpiryDate
	}

	if dto.Notes != nil {
		quotation.Notes = *dto.Notes
	}

	if dto.Terms != nil {
		quotation.Terms = *dto.Terms
	}
}

// Helper functions
func ToQuotationResponseDTOList(quotations []models.Quotation) []dtos.QuotationResponseDTO {
	if quotations == nil {
		return nil
	}

	result := make([]dtos.QuotationResponseDTO, len(quotations))
	for i, quotation := range quotations {
		result[i] = *ToQuotationResponseDTO(&quotation)
	}
	return result
}

func ToQuotationItemResponseDTOList(items []models.QuotationItem) []dtos.QuotationItemResponseDTO {
	if items == nil {
		return nil
	}

	result := make([]dtos.QuotationItemResponseDTO, len(items))
	for i, item := range items {
		result[i] = *ToQuotationItemResponseDTO(&item)
	}
	return result
}
