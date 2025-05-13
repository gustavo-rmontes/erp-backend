package mapper

import (
	contact "ERP-ONSMART/backend/internal/modules/contact/models"
	"ERP-ONSMART/backend/internal/modules/sales/dtos"
)

// ToContactBasicInfo converte Contact model para ContactBasicInfo DTO
func ToContactBasicInfo(contact *contact.Contact) *dtos.ContactBasicInfo {
	if contact == nil {
		return nil
	}

	return &dtos.ContactBasicInfo{
		ID:          contact.ID,
		Name:        contact.Name,
		CompanyName: contact.CompanyName,
		Type:        contact.Type,
		PersonType:  contact.PersonType,
	}
}

// ToAddressDTO converte campos de endereço do Contact para AddressDTO
func ToAddressDTO(contact *contact.Contact) *dtos.AddressDTO {
	if contact == nil {
		return nil
	}

	return &dtos.AddressDTO{
		Street:     contact.Street,
		Number:     contact.Number,
		Complement: contact.Complement,
		District:   contact.Neighborhood,
		City:       contact.City,
		State:      contact.State,
		ZipCode:    contact.ZipCode,
		Country:    "Brasil", // Default, já que não temos esse campo no model
	}
}
