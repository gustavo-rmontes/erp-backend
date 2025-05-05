package models

import "time"

type Contact struct {
	ID           int    `json:"id"`
	PersonType   string `json:"person_type" binding:"required,oneof=pf pj"`
	Type         string `json:"type" binding:"required,oneof=cliente fornecedor lead"`
	Name         string `json:"name" binding:"required"`
	CompanyName  string `json:"company_name"`
	TradeName    string `json:"trade_name"`
	Document     string `json:"document" binding:"required"`
	SecondaryDoc string `json:"secondary_doc"`
	Suframa      string `json:"suframa"`
	Isento       bool   `json:"isento"`
	CCM          string `json:"ccm"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone"`

	ZipCode      string `json:"zip_code" binding:"required"`
	Street       string `json:"street"`
	Number       string `json:"number"`
	Complement   string `json:"complement"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
