package dtos

import "time"

// DateRange representa um intervalo de datas
type DateRange struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
}

// AmountRange representa um intervalo de valores
type AmountRange struct {
	MinAmount float64 `json:"min_amount" validate:"min=0"`
	MaxAmount float64 `json:"max_amount" validate:"min=0,gtefield=MinAmount"`
}

// PaginationRequest representa os parâmetros de paginação da requisição
type PaginationRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	PageSize int    `json:"page_size" validate:"min=1,max=100"`
	Sort     string `json:"sort,omitempty"`
	Order    string `json:"order,omitempty" validate:"omitempty,oneof=asc desc"`
}

// PaginationResponse representa os dados de paginação da resposta
type PaginationResponse struct {
	Page         int `json:"page"`
	PageSize     int `json:"page_size"`
	TotalPages   int `json:"total_pages"`
	TotalRecords int `json:"total_records"`
}

// SortOptions representa as opções de ordenação
type SortOptions struct {
	Field string `json:"field"`
	Order string `json:"order" validate:"oneof=asc desc"`
}

// ItemDTO representa um item genérico usado em vários documentos
type ItemDTO struct {
	ID          int     `json:"id,omitempty"`
	ProductID   int     `json:"product_id" validate:"required"`
	ProductName string  `json:"product_name"`
	ProductCode string  `json:"product_code"`
	Description string  `json:"description,omitempty"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
	Discount    float64 `json:"discount" validate:"min=0,max=100"`
	Tax         float64 `json:"tax" validate:"min=0"`
	Total       float64 `json:"total"`
}

// StatusCount representa contagem por status
type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// AmountSummary representa um resumo de valores
type AmountSummary struct {
	Total   float64 `json:"total"`
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
}

// ContactBasicInfo representa informações básicas de um contato
type ContactBasicInfo struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	CompanyName string `json:"company_name,omitempty"`
	Type        string `json:"type"`
	PersonType  string `json:"person_type,omitempty"`
}

// AddressDTO representa um endereço
type AddressDTO struct {
	Street     string `json:"street"`
	Number     string `json:"number"`
	Complement string `json:"complement,omitempty"`
	District   string `json:"district"`
	City       string `json:"city"`
	State      string `json:"state"`
	ZipCode    string `json:"zip_code"`
	Country    string `json:"country"`
}

// TimeRange representa um intervalo de tempo
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
