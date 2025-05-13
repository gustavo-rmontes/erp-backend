package dtos

import "time"

// QuotationCreateDTO representa os dados para criar uma quotation
type QuotationCreateDTO struct {
	ContactID  int                      `json:"contact_id" validate:"required"`
	ExpiryDate time.Time                `json:"expiry_date" validate:"required"`
	Notes      string                   `json:"notes,omitempty"`
	Terms      string                   `json:"terms,omitempty"`
	Items      []QuotationItemCreateDTO `json:"items" validate:"required,min=1,dive"`
}

// QuotationUpdateDTO representa os dados para atualizar uma quotation
type QuotationUpdateDTO struct {
	ExpiryDate *time.Time `json:"expiry_date,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
	Terms      *string    `json:"terms,omitempty"`
}

// QuotationResponseDTO representa os dados retornados de uma quotation
type QuotationResponseDTO struct {
	ID            int                        `json:"id"`
	QuotationNo   string                     `json:"quotation_no"`
	ContactID     int                        `json:"contact_id"`
	Contact       *ContactBasicInfo          `json:"contact,omitempty"`
	Status        string                     `json:"status"`
	CreatedAt     time.Time                  `json:"created_at"`
	UpdatedAt     time.Time                  `json:"updated_at"`
	ExpiryDate    time.Time                  `json:"expiry_date"`
	SubTotal      float64                    `json:"subtotal"`
	TaxTotal      float64                    `json:"tax_total"`
	DiscountTotal float64                    `json:"discount_total"`
	GrandTotal    float64                    `json:"grand_total"`
	Notes         string                     `json:"notes,omitempty"`
	Terms         string                     `json:"terms,omitempty"`
	Items         []QuotationItemResponseDTO `json:"items,omitempty"`
	IsExpired     bool                       `json:"is_expired"`
	DaysToExpiry  int                        `json:"days_to_expiry,omitempty"`
}

// QuotationListItemDTO representa uma versão resumida para listagens
type QuotationListItemDTO struct {
	ID           int               `json:"id"`
	QuotationNo  string            `json:"quotation_no"`
	ContactID    int               `json:"contact_id"`
	Contact      *ContactBasicInfo `json:"contact,omitempty"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	ExpiryDate   time.Time         `json:"expiry_date"`
	GrandTotal   float64           `json:"grand_total"`
	IsExpired    bool              `json:"is_expired"`
	DaysToExpiry int               `json:"days_to_expiry,omitempty"`
}

// QuotationItemCreateDTO representa os dados para criar um item de quotation
type QuotationItemCreateDTO struct {
	ProductID   int     `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name,omitempty"`
	ProductCode string  `json:"product_code,omitempty"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" validate:"min=0,max=100"`
	Tax         float64 `json:"tax" validate:"min=0"`
}

// QuotationItemUpdateDTO representa os dados para atualizar um item
type QuotationItemUpdateDTO struct {
	Quantity    *int     `json:"quantity,omitempty" validate:"omitempty,gt=0"`
	UnitPrice   *float64 `json:"unit_price,omitempty" validate:"omitempty,gt=0"`
	Discount    *float64 `json:"discount,omitempty" validate:"omitempty,min=0,max=100"`
	Tax         *float64 `json:"tax,omitempty" validate:"omitempty,min=0"`
	Description *string  `json:"description,omitempty"`
}

// QuotationItemResponseDTO representa os dados retornados de um item
type QuotationItemResponseDTO struct {
	ID          int     `json:"id"`
	QuotationID int     `json:"quotation_id"`
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Discount    float64 `json:"discount"`
	Tax         float64 `json:"tax"`
	Total       float64 `json:"total"`
}

// QuotationStatusUpdateDTO representa dados para atualizar status
type QuotationStatusUpdateDTO struct {
	Status string `json:"status" validate:"required,oneof=draft sent accepted rejected expired cancelled"`
	Notes  string `json:"notes,omitempty"`
}

// ConvertToSODTO representa dados para converter em sales order
type ConvertToSODTO struct {
	ExpectedDate    time.Time `json:"expected_date" validate:"required"`
	PaymentTerms    string    `json:"payment_terms,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	IncludeAllItems bool      `json:"include_all_items"`
}

// QuotationSendDTO representa dados para enviar quotation
type QuotationSendDTO struct {
	EmailTo      []string `json:"email_to" validate:"required,min=1,dive,email"`
	EmailCC      []string `json:"email_cc,omitempty" validate:"omitempty,dive,email"`
	EmailSubject string   `json:"email_subject,omitempty"`
	EmailBody    string   `json:"email_body,omitempty"`
	AttachPDF    bool     `json:"attach_pdf"`
}

// QuotationCloneDTO representa dados para clonar quotation
type QuotationCloneDTO struct {
	ContactID  int       `json:"contact_id,omitempty"`
	ExpiryDate time.Time `json:"expiry_date" validate:"required"`
	Notes      string    `json:"notes,omitempty"`
	Terms      string    `json:"terms,omitempty"`
}

// QuotationRevisionDTO representa dados para criar revisão
type QuotationRevisionDTO struct {
	ParentQuotationID int       `json:"parent_quotation_id" validate:"required"`
	ExpiryDate        time.Time `json:"expiry_date" validate:"required"`
	Notes             string    `json:"notes,omitempty"`
	RevisionNotes     string    `json:"revision_notes" validate:"required"`
}

// QuotationFollowUpDTO representa dados para follow-up
type QuotationFollowUpDTO struct {
	QuotationID  int       `json:"quotation_id" validate:"required"`
	FollowUpDate time.Time `json:"follow_up_date" validate:"required"`
	Method       string    `json:"method" validate:"required,oneof=email phone meeting"`
	Notes        string    `json:"notes,omitempty"`
	Reminder     bool      `json:"reminder"`
}

// QuotationComparisonDTO representa comparação entre quotations
type QuotationComparisonDTO struct {
	QuotationIDs []int `json:"quotation_ids" validate:"required,min=2"`
	CompareItems bool  `json:"compare_items"`
	CompareTerms bool  `json:"compare_terms"`
}

// QuotationTemplateDTO representa template de quotation
type QuotationTemplateDTO struct {
	Name         string                   `json:"name" validate:"required"`
	Description  string                   `json:"description,omitempty"`
	Terms        string                   `json:"terms,omitempty"`
	ValidityDays int                      `json:"validity_days" validate:"required,gt=0"`
	Items        []QuotationItemCreateDTO `json:"items,omitempty"`
	IsActive     bool                     `json:"is_active"`
}
