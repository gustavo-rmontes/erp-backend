package tests

import "ERP-ONSMART/backend/internal/modules/sales/dto"

// calculateItemTotal calcula o total de um item
func calculateItemTotal(quantity int, unitPrice float64, discount float64, tax float64) float64 {
	subtotal := float64(quantity) * unitPrice
	discountAmount := subtotal * (discount / 100.0)
	taxAmount := (subtotal - discountAmount) * (tax / 100.0)
	return subtotal - discountAmount + taxAmount
}

// Adapte este exemplo com base nas suas definições reais
type testQuotationItemCreate struct {
	ProductID   int
	ProductName string
	Description string
	Quantity    int
	UnitPrice   float64
	Discount    float64
	Tax         float64
}

type testEmailOptions struct {
	To        []string
	Subject   string
	Body      string
	AttachPDF bool
}

// Converte testQuotationItemCreate para dto.QuotationItemCreate
func convertToRealQuotationItemCreate(item testQuotationItemCreate) dto.QuotationItemCreate {
	return dto.QuotationItemCreate{
		BaseItem: dto.BaseItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Discount:    item.Discount,
			Tax:         item.Tax,
		},
	}
}

func convertToRealEmailOptions(opts testEmailOptions) dto.EmailOptions {
	return dto.EmailOptions{
		To:        opts.To,
		Subject:   opts.Subject,
		Message:   opts.Body, // Use Message em vez de Content
		AttachPDF: opts.AttachPDF,
	}
}
