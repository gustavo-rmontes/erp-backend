package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	// Identification fields
	ID           int    `gorm:"primaryKey" json:"id"`
	Name         string `gorm:"column:name" json:"name" binding:"required"`
	DetailedName string `gorm:"column:detailed_name" json:"detailed_name" binding:"required"`
	Description  string `gorm:"column:description" json:"description"`
	Status       string `gorm:"column:status" json:"status" binding:"required,oneof=ativo desativado descontinuado"`
	SKU          string `gorm:"column:sku" json:"sku"`
	Barcode      string `gorm:"column:barcode" json:"barcode"`
	ExternalID   string `gorm:"column:external_id" json:"external_id,omitempty"`

	// Price related
	Coin       string  `gorm:"column:coin" json:"coin" binding:"required,oneof=BRL USD EUR CAD ADOBE_USD"`
	Price      float64 `gorm:"column:price" json:"price" binding:"required,gte=0"`
	SalesPrice float64 `gorm:"column:sales_price" json:"sales_price" binding:"gte=0"`
	CostPrice  float64 `gorm:"column:cost_price" json:"cost_price" binding:"gte=0"`

	// Inventory related
	Stock int `gorm:"column:stock" json:"stock" binding:"gte=0"`

	// Classification
	Type               string         `gorm:"column:type" json:"type"`
	ProductGroup       string         `gorm:"column:product_group" json:"product_group"`
	ProductCategory    string         `gorm:"column:product_category" json:"product_category"`
	ProductSubcategory string         `gorm:"column:product_subcategory" json:"product_subcategory"`
	Tags               pq.StringArray `gorm:"column:tags;type:text[]" json:"tags"`
	Manufacturer       string         `gorm:"column:manufacturer" json:"manufacturer"`
	ManufacturerCode   string         `gorm:"column:manufacturer_code" json:"manufacturer_code"`

	// Fiscal related
	NCM    string `gorm:"column:ncm" json:"ncm"`
	CEST   string `gorm:"column:cest" json:"cest"`
	CNAE   string `gorm:"column:cnae" json:"cnae"`
	Origin string `gorm:"column:origin" json:"origin"`

	// Campos temporais
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Recursos multimídia
	Images    pq.StringArray `gorm:"column:images;type:text[]" json:"images,omitempty"`
	Documents pq.StringArray `gorm:"column:documents;type:text[]" json:"documents,omitempty"`
}

// Warranty representa a garantia do produto.
type Warranty struct {
	ID             int     `json:"id"`
	ProductID      int     `json:"product_id" binding:"required"`
	DurationMonths int     `json:"duration_months"`
	Price          float64 `json:"price" binding:"required,gt=0"`
}
