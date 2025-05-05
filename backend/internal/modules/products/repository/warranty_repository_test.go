package repository

import (
	"ERP-ONSMART/backend/internal/modules/products/models"
	"testing"
)

func TestInsertWarranty(t *testing.T) {
	// Cria um produto para associar à garantia
	p := models.Product{
		Name:        "Produto para Garantia",
		Description: "Produto para teste de garantia",
		Price:       100.0,
		Stock:       10,
	}
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para garantia: %v", err)
	}
	products, err := GetAllProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao buscar produtos para garantia: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	w := models.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          20.0,
	}
	if err := CreateWarranty(w); err != nil {
		t.Fatalf("Erro ao criar garantia: %v", err)
	}
}

func TestListWarranties(t *testing.T) {
	warranties, err := GetWarranties()
	if err != nil {
		t.Fatalf("Erro ao listar garantias: %v", err)
	}
	if len(warranties) == 0 {
		t.Fatalf("Nenhuma garantia encontrada. Certifique-se de que pelo menos uma garantia foi inserida.")
	}
}

func TestUpdateWarrantyByID(t *testing.T) {
	// Cria um produto para associar a garantia
	p := models.Product{
		Name:        "Produto para Atualização de Garantia",
		Description: "Produto para teste de atualização de garantia",
		Price:       100.0,
		Stock:       10,
	}
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para garantia update: %v", err)
	}
	products, err := GetAllProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao buscar produtos para garantia update: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	w := models.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          20.0,
	}
	if err := CreateWarranty(w); err != nil {
		t.Fatalf("Erro ao criar garantia para update: %v", err)
	}
	warranties, err := GetWarranties()
	if err != nil || len(warranties) == 0 {
		t.Fatalf("Erro ao buscar garantias para update: %v", err)
	}
	warrantyID := warranties[len(warranties)-1].ID

	// Atualiza a garantia
	updatedWarranty := models.Warranty{
		ProductID:      productID,
		DurationMonths: 24,
		Price:          25.5,
	}
	if err := UpdateWarrantyByID(warrantyID, updatedWarranty); err != nil {
		t.Fatalf("Erro ao atualizar garantia: %v", err)
	}
}

func TestDeleteWarrantyByID(t *testing.T) {
	// Cria um produto para associar a garantia
	p := models.Product{
		Name:        "Produto para Deleção de Garantia",
		Description: "Produto para teste de deleção de garantia",
		Price:       100.0,
		Stock:       10,
	}
	if err := CreateProduct(&p); err != nil {
		t.Fatalf("Erro ao criar produto para garantia delete: %v", err)
	}
	products, err := GetAllProducts()
	if err != nil || len(products) == 0 {
		t.Fatalf("Erro ao buscar produtos para garantia delete: %v", err)
	}
	productID := products[len(products)-1].ID

	// Cria uma garantia associada ao produto
	w := models.Warranty{
		ProductID:      productID,
		DurationMonths: 12,
		Price:          20.0,
	}
	if err := CreateWarranty(w); err != nil {
		t.Fatalf("Erro ao criar garantia para delete: %v", err)
	}
	warranties, err := GetWarranties()
	if err != nil || len(warranties) == 0 {
		t.Fatalf("Erro ao listar garantias para delete: %v", err)
	}
	warrantyID := warranties[len(warranties)-1].ID

	// Deleta a garantia
	if err := DeleteWarrantyByID(warrantyID); err != nil {
		t.Fatalf("Erro ao deletar garantia: %v", err)
	}
}
