package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
)

// ToPaymentResponseDTO converte Payment model para PaymentResponseDTO
func ToPaymentResponseDTO(payment *models.Payment) *dtos.PaymentResponseDTO {
	if payment == nil {
		return nil
	}

	dto := &dtos.PaymentResponseDTO{
		ID:            payment.ID,
		InvoiceID:     payment.InvoiceID,
		Amount:        payment.Amount,
		PaymentDate:   payment.PaymentDate,
		PaymentMethod: payment.PaymentMethod,
		Reference:     payment.Reference,
		Notes:         payment.Notes,
	}

	// Mapear relações opcionais
	if payment.Invoice != nil {
		dto.InvoiceNo = payment.Invoice.InvoiceNo

		// Contact vem através da Invoice
		if payment.Invoice.Contact != nil {
			dto.Contact = ToContactBasicInfo(payment.Invoice.Contact)
		}
	}

	return dto
}

// ToPaymentListItemDTO converte Payment model para PaymentListItemDTO (versão resumida)
func ToPaymentListItemDTO(payment *models.Payment) *dtos.PaymentListItemDTO {
	if payment == nil {
		return nil
	}

	dto := &dtos.PaymentListItemDTO{
		ID:            payment.ID,
		InvoiceID:     payment.InvoiceID,
		Amount:        payment.Amount,
		PaymentDate:   payment.PaymentDate,
		PaymentMethod: payment.PaymentMethod,
		Reference:     payment.Reference,
	}

	if payment.Invoice != nil {
		dto.InvoiceNo = payment.Invoice.InvoiceNo

		if payment.Invoice.Contact != nil {
			dto.Contact = ToContactBasicInfo(payment.Invoice.Contact)
		}
	}

	return dto
}

// FromPaymentCreateDTO converte PaymentCreateDTO para Payment model
func FromPaymentCreateDTO(dto *dtos.PaymentCreateDTO) *models.Payment {
	if dto == nil {
		return nil
	}

	return &models.Payment{
		InvoiceID:     dto.InvoiceID,
		Amount:        dto.Amount,
		PaymentDate:   dto.PaymentDate,
		PaymentMethod: dto.PaymentMethod,
		Reference:     dto.Reference,
		Notes:         dto.Notes,
	}
}

// FromPaymentUpdateDTO converte PaymentUpdateDTO para Payment model (apenas campos não nulos)
func FromPaymentUpdateDTO(dto *dtos.PaymentUpdateDTO, payment *models.Payment) {
	if dto == nil || payment == nil {
		return
	}

	if dto.Amount != nil {
		payment.Amount = *dto.Amount
	}

	if dto.PaymentDate != nil {
		payment.PaymentDate = *dto.PaymentDate
	}

	if dto.PaymentMethod != nil {
		payment.PaymentMethod = *dto.PaymentMethod
	}

	if dto.Reference != nil {
		payment.Reference = *dto.Reference
	}

	if dto.Notes != nil {
		payment.Notes = *dto.Notes
	}
}

// Helper functions para listas
func ToPaymentResponseDTOList(payments []models.Payment) []dtos.PaymentResponseDTO {
	if payments == nil {
		return nil
	}

	result := make([]dtos.PaymentResponseDTO, len(payments))
	for i, payment := range payments {
		result[i] = *ToPaymentResponseDTO(&payment)
	}
	return result
}

func ToPaymentListItemDTOList(payments []models.Payment) []dtos.PaymentListItemDTO {
	if payments == nil {
		return nil
	}

	result := make([]dtos.PaymentListItemDTO, len(payments))
	for i, payment := range payments {
		result[i] = *ToPaymentListItemDTO(&payment)
	}
	return result
}
