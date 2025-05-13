package dtos

import "time"

// ContactCreateDTO representa os dados para criar um contact
type ContactCreateDTO struct {
	PersonType   string `json:"person_type" validate:"required,oneof=pf pj"`
	Type         string `json:"type" validate:"required,oneof=cliente fornecedor lead"`
	Name         string `json:"name" validate:"required"`
	CompanyName  string `json:"company_name,omitempty"`
	TradeName    string `json:"trade_name,omitempty"`
	Document     string `json:"document" validate:"required"`
	SecondaryDoc string `json:"secondary_doc,omitempty"`
	Suframa      string `json:"suframa,omitempty"`
	Isento       bool   `json:"isento"`
	CCM          string `json:"ccm,omitempty"`
	Email        string `json:"email" validate:"required,email"`
	Phone        string `json:"phone,omitempty"`

	// Address fields
	ZipCode      string `json:"zip_code" validate:"required"`
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
}

// ContactUpdateDTO representa os dados para atualizar um contact
type ContactUpdateDTO struct {
	PersonType   *string `json:"person_type,omitempty" validate:"omitempty,oneof=pf pj"`
	Type         *string `json:"type,omitempty" validate:"omitempty,oneof=cliente fornecedor lead"`
	Name         *string `json:"name,omitempty"`
	CompanyName  *string `json:"company_name,omitempty"`
	TradeName    *string `json:"trade_name,omitempty"`
	Document     *string `json:"document,omitempty"`
	SecondaryDoc *string `json:"secondary_doc,omitempty"`
	Suframa      *string `json:"suframa,omitempty"`
	Isento       *bool   `json:"isento,omitempty"`
	CCM          *string `json:"ccm,omitempty"`
	Email        *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone        *string `json:"phone,omitempty"`

	// Address fields
	ZipCode      *string `json:"zip_code,omitempty"`
	Street       *string `json:"street,omitempty"`
	Number       *string `json:"number,omitempty"`
	Complement   *string `json:"complement,omitempty"`
	Neighborhood *string `json:"neighborhood,omitempty"`
	City         *string `json:"city,omitempty"`
	State        *string `json:"state,omitempty"`
}

// ContactResponseDTO representa os dados retornados de um contact
type ContactResponseDTO struct {
	ID           int    `json:"id"`
	PersonType   string `json:"person_type"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	CompanyName  string `json:"company_name,omitempty"`
	TradeName    string `json:"trade_name,omitempty"`
	Document     string `json:"document"`
	SecondaryDoc string `json:"secondary_doc,omitempty"`
	Suframa      string `json:"suframa,omitempty"`
	Isento       bool   `json:"isento"`
	CCM          string `json:"ccm,omitempty"`
	Email        string `json:"email"`
	Phone        string `json:"phone,omitempty"`

	// Address fields
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContactListItemDTO representa uma versão resumida para listagens
type ContactListItemDTO struct {
	ID          int    `json:"id"`
	PersonType  string `json:"person_type"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	CompanyName string `json:"company_name,omitempty"`
	Document    string `json:"document"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
}
