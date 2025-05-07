package dto

import "time"

// ContactCreate DTO for creating new contacts
type ContactCreate struct {
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
	ZipCode      string `json:"zip_code,omitempty"`
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
}

// ContactUpdate DTO for updating existing contacts
type ContactUpdate struct {
	PersonType   string `json:"person_type" validate:"omitempty,oneof=pf pj"`
	Type         string `json:"type" validate:"omitempty,oneof=cliente fornecedor lead"`
	Name         string `json:"name" validate:"omitempty"`
	CompanyName  string `json:"company_name,omitempty"`
	TradeName    string `json:"trade_name,omitempty"`
	Document     string `json:"document" validate:"omitempty"`
	SecondaryDoc string `json:"secondary_doc,omitempty"`
	Suframa      string `json:"suframa,omitempty"`
	Isento       *bool  `json:"isento,omitempty"`
	CCM          string `json:"ccm,omitempty"`
	Email        string `json:"email" validate:"omitempty,email"`
	Phone        string `json:"phone,omitempty"`
	ZipCode      string `json:"zip_code,omitempty"`
	Street       string `json:"street,omitempty"`
	Number       string `json:"number,omitempty"`
	Complement   string `json:"complement,omitempty"`
	Neighborhood string `json:"neighborhood,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
}

// ContactResponse DTO for contact data in responses
type ContactResponse struct {
	ID           int       `json:"id"`
	PersonType   string    `json:"person_type"`
	Type         string    `json:"type"`
	Name         string    `json:"name"`
	CompanyName  string    `json:"company_name,omitempty"`
	TradeName    string    `json:"trade_name,omitempty"`
	Document     string    `json:"document"`
	SecondaryDoc string    `json:"secondary_doc,omitempty"`
	Suframa      string    `json:"suframa,omitempty"`
	Isento       bool      `json:"isento"`
	CCM          string    `json:"ccm,omitempty"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	ZipCode      string    `json:"zip_code,omitempty"`
	Street       string    `json:"street,omitempty"`
	Number       string    `json:"number,omitempty"`
	Complement   string    `json:"complement,omitempty"`
	Neighborhood string    `json:"neighborhood,omitempty"`
	City         string    `json:"city,omitempty"`
	State        string    `json:"state,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ContactShortResponse DTO for minimized contact data when needed
type ContactShortResponse struct {
	ID          int    `json:"id"`
	PersonType  string `json:"person_type"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Document    string `json:"document"`
	Email       string `json:"email"`
	Phone       string `json:"phone,omitempty"`
	CompanyName string `json:"company_name,omitempty"`
}

// PaginatedContactResponse DTO for paginated responses
type PaginatedContactResponse struct {
	Items       []ContactResponse `json:"items"`
	TotalItems  int64             `json:"total_items"`
	TotalPages  int               `json:"total_pages"`
	CurrentPage int               `json:"current_page"`
	PageSize    int               `json:"page_size"`
}
