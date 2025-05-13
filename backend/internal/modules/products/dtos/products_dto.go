package dtos

import "time"

// ProductCreateDTO representa os dados para criar um product
type ProductCreateDTO struct {
	Name         string `json:"name" validate:"required"`
	DetailedName string `json:"detailed_name" validate:"required"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status" validate:"required,oneof=ativo desativado descontinuado"`
	SKU          string `json:"sku,omitempty"`
	Barcode      string `json:"barcode,omitempty"`
	ExternalID   string `json:"external_id,omitempty"`

	// Price related
	Coin       string  `json:"coin" validate:"required,oneof=BRL USD EUR CAD ADOBE_USD"`
	Price      float64 `json:"price" validate:"required,gte=0"`
	SalesPrice float64 `json:"sales_price" validate:"gte=0"`
	CostPrice  float64 `json:"cost_price" validate:"gte=0"`

	// Inventory
	Stock int `json:"stock" validate:"gte=0"`

	// Classification
	Type               string   `json:"type,omitempty"`
	ProductGroup       string   `json:"product_group,omitempty"`
	ProductCategory    string   `json:"product_category,omitempty"`
	ProductSubcategory string   `json:"product_subcategory,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Manufacturer       string   `json:"manufacturer,omitempty"`
	ManufacturerCode   string   `json:"manufacturer_code,omitempty"`

	// Fiscal
	NCM    string `json:"ncm,omitempty"`
	CEST   string `json:"cest,omitempty"`
	CNAE   string `json:"cnae,omitempty"`
	Origin string `json:"origin,omitempty"`

	// Multimedia
	Images    []string `json:"images,omitempty"`
	Documents []string `json:"documents,omitempty"`
}

// ProductUpdateDTO representa os dados para atualizar um product
type ProductUpdateDTO struct {
	Name         *string `json:"name,omitempty"`
	DetailedName *string `json:"detailed_name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Status       *string `json:"status,omitempty" validate:"omitempty,oneof=ativo desativado descontinuado"`
	SKU          *string `json:"sku,omitempty"`
	Barcode      *string `json:"barcode,omitempty"`
	ExternalID   *string `json:"external_id,omitempty"`

	// Price related
	Coin       *string  `json:"coin,omitempty" validate:"omitempty,oneof=BRL USD EUR CAD ADOBE_USD"`
	Price      *float64 `json:"price,omitempty" validate:"omitempty,gte=0"`
	SalesPrice *float64 `json:"sales_price,omitempty" validate:"omitempty,gte=0"`
	CostPrice  *float64 `json:"cost_price,omitempty" validate:"omitempty,gte=0"`

	// Inventory
	Stock *int `json:"stock,omitempty" validate:"omitempty,gte=0"`

	// Classification
	Type               *string   `json:"type,omitempty"`
	ProductGroup       *string   `json:"product_group,omitempty"`
	ProductCategory    *string   `json:"product_category,omitempty"`
	ProductSubcategory *string   `json:"product_subcategory,omitempty"`
	Tags               *[]string `json:"tags,omitempty"`
	Manufacturer       *string   `json:"manufacturer,omitempty"`
	ManufacturerCode   *string   `json:"manufacturer_code,omitempty"`

	// Fiscal
	NCM    *string `json:"ncm,omitempty"`
	CEST   *string `json:"cest,omitempty"`
	CNAE   *string `json:"cnae,omitempty"`
	Origin *string `json:"origin,omitempty"`

	// Multimedia
	Images    *[]string `json:"images,omitempty"`
	Documents *[]string `json:"documents,omitempty"`
}

// ProductResponseDTO representa os dados retornados de um product
type ProductResponseDTO struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	DetailedName string `json:"detailed_name"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status"`
	SKU          string `json:"sku,omitempty"`
	Barcode      string `json:"barcode,omitempty"`
	ExternalID   string `json:"external_id,omitempty"`

	// Price related
	Coin       string  `json:"coin"`
	Price      float64 `json:"price"`
	SalesPrice float64 `json:"sales_price"`
	CostPrice  float64 `json:"cost_price"`

	// Inventory
	Stock int `json:"stock"`

	// Classification
	Type               string   `json:"type,omitempty"`
	ProductGroup       string   `json:"product_group,omitempty"`
	ProductCategory    string   `json:"product_category,omitempty"`
	ProductSubcategory string   `json:"product_subcategory,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Manufacturer       string   `json:"manufacturer,omitempty"`
	ManufacturerCode   string   `json:"manufacturer_code,omitempty"`

	// Fiscal
	NCM    string `json:"ncm,omitempty"`
	CEST   string `json:"cest,omitempty"`
	CNAE   string `json:"cnae,omitempty"`
	Origin string `json:"origin,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Multimedia
	Images    []string `json:"images,omitempty"`
	Documents []string `json:"documents,omitempty"`
}

// ProductListItemDTO representa uma versão resumida para listagens
type ProductListItemDTO struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku,omitempty"`
	Barcode      string  `json:"barcode,omitempty"`
	Status       string  `json:"status"`
	Price        float64 `json:"price"`
	SalesPrice   float64 `json:"sales_price"`
	Stock        int     `json:"stock"`
	Category     string  `json:"category,omitempty"`
	Manufacturer string  `json:"manufacturer,omitempty"`
	ImageURL     string  `json:"image_url,omitempty"`
}

// WarrantyCreateDTO representa os dados para criar uma warranty
type WarrantyCreateDTO struct {
	ProductID      int     `json:"product_id" validate:"required"`
	DurationMonths int     `json:"duration_months" validate:"required,gt=0"`
	Price          float64 `json:"price" validate:"required,gt=0"`
}

// WarrantyResponseDTO representa os dados retornados de uma warranty
type WarrantyResponseDTO struct {
	ID             int                 `json:"id"`
	ProductID      int                 `json:"product_id"`
	Product        *ProductListItemDTO `json:"product,omitempty"`
	DurationMonths int                 `json:"duration_months"`
	Price          float64             `json:"price"`
}
