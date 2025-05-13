package mapper

import (
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
	"ERP-ONSMART/backend/internal/modules/sales/models"
	"time"
)

// ToSalesProcessResponseDTO converte SalesProcess model para SalesProcessResponseDTO
func ToSalesProcessResponseDTO(sp *models.SalesProcess) *dtos.SalesProcessResponseDTO {
	if sp == nil {
		return nil
	}

	dto := &dtos.SalesProcessResponseDTO{
		ID:         sp.ID,
		ContactID:  sp.ContactID,
		Status:     sp.Status,
		CreatedAt:  sp.CreatedAt,
		UpdatedAt:  sp.UpdatedAt,
		TotalValue: sp.TotalValue,
		TotalCost:  sp.TotalValue - sp.Profit, // Calculado
		Profit:     sp.Profit,
		Notes:      sp.Notes,
	}

	// Mapear Contact
	if sp.Contact != nil {
		dto.Contact = ToContactBasicInfo(sp.Contact)
	}

	// Calcular margem de lucro
	if sp.TotalValue > 0 {
		dto.ProfitMargin = (sp.Profit / sp.TotalValue) * 100
	}

	// Campos calculados - seriam obtidos de outros lugares
	dto.CurrentStage = "negotiation" // Exemplo
	dto.CompletionRate = 0.5         // 50% exemplo

	// LinkedDocuments seria preenchido através de queries específicas
	dto.LinkedDocuments = dtos.LinkedDocumentsDTO{}

	return dto
}

// ToSalesProcessListItemDTO converte SalesProcess para versão resumida
func ToSalesProcessListItemDTO(sp *models.SalesProcess) *dtos.SalesProcessListItemDTO {
	if sp == nil {
		return nil
	}

	dto := &dtos.SalesProcessListItemDTO{
		ID:           sp.ID,
		ContactID:    sp.ContactID,
		Status:       sp.Status,
		CreatedAt:    sp.CreatedAt,
		TotalValue:   sp.TotalValue,
		Profit:       sp.Profit,
		CurrentStage: "negotiation", // Exemplo
		LastActivity: sp.UpdatedAt,
	}

	// Mapear Contact
	if sp.Contact != nil {
		dto.Contact = ToContactBasicInfo(sp.Contact)
	}

	// Calcular margem de lucro
	if sp.TotalValue > 0 {
		dto.ProfitMargin = (sp.Profit / sp.TotalValue) * 100
	}

	// Taxa de conclusão seria calculada
	dto.CompletionRate = 0.5 // 50% exemplo

	return dto
}

// FromSalesProcessCreateDTO converte SalesProcessCreateDTO para SalesProcess model
func FromSalesProcessCreateDTO(dto *dtos.SalesProcessCreateDTO) *models.SalesProcess {
	if dto == nil {
		return nil
	}

	return &models.SalesProcess{
		ContactID:  dto.ContactID,
		Notes:      dto.Notes,
		TotalValue: dto.InitialValue,
		Status:     "open", // Status inicial
		Profit:     0,      // Inicialmente sem lucro
	}
}

// FromSalesProcessUpdateDTO aplica updates parciais
func FromSalesProcessUpdateDTO(dto *dtos.SalesProcessUpdateDTO, sp *models.SalesProcess) {
	if dto == nil || sp == nil {
		return
	}

	if dto.Notes != nil {
		sp.Notes = *dto.Notes
	}

	if dto.TotalValue != nil {
		sp.TotalValue = *dto.TotalValue
	}

	if dto.Profit != nil {
		sp.Profit = *dto.Profit
	}
}

// ToLinkedDocumentInfo converte informações de documento vinculado
func ToLinkedDocumentInfo(docType string, id int, docNo string, status string, value float64, date time.Time) *dtos.LinkedDocumentInfo {
	return &dtos.LinkedDocumentInfo{
		ID:         id,
		DocumentNo: docNo,
		Status:     status,
		Value:      value,
		Date:       date,
		LinkedAt:   time.Now(), // Exemplo
	}
}

// Helper functions para listas
func ToSalesProcessResponseDTOList(processes []models.SalesProcess) []dtos.SalesProcessResponseDTO {
	if processes == nil {
		return nil
	}

	result := make([]dtos.SalesProcessResponseDTO, len(processes))
	for i, process := range processes {
		result[i] = *ToSalesProcessResponseDTO(&process)
	}
	return result
}

func ToSalesProcessListItemDTOList(processes []models.SalesProcess) []dtos.SalesProcessListItemDTO {
	if processes == nil {
		return nil
	}

	result := make([]dtos.SalesProcessListItemDTO, len(processes))
	for i, process := range processes {
		result[i] = *ToSalesProcessListItemDTO(&process)
	}
	return result
}
